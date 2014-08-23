package iniflags

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

type Arg struct {
	Key      string
	Value    string
	FilePath string
	LineNum  int
}

var (
	config    = flag.String("config", "", "Path to ini config for using in go flags. May be relative to the current executable path")
	dumpflags = flag.Bool("dumpflags", false, "Dumps values for all flags defined in the app into stdout in ini-compatible syntax and terminates the app")

	configReadCallbacks []func()
	importStack         []string
)

// Use instead of flag.Parse().
func Parse() {
	flag.Parse()
	if !parseConfigFlags() {
		os.Exit(1)
	}
	if *dumpflags {
		dumpFlags()
		os.Exit(0)
	}
	issueConfigReadCallbacks()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP)
	go sighupHandler(ch)
}

// Registers the callback, which is called after each config read.
// An app can register arbitrary number of callbacks.
// Usually these callbacks should be registered in init() functions.
// The callbacks should be used for re-applying new config values across
// the application.
func AddConfigReadCallback(f func()) {
	configReadCallbacks = append(configReadCallbacks, f)
}

func issueConfigReadCallbacks() {
	for _, f := range configReadCallbacks {
		f()
	}
}

func sighupHandler(ch <-chan os.Signal) {
	for _ = range ch {
		log.Printf("Re-reading flags from config files\n")
		if parseConfigFlags() {
			issueConfigReadCallbacks()
		}
	}
}

func parseConfigFlags() bool {
	var ok bool
	configPath := *config
	if !strings.HasPrefix(configPath, "./") {
		if configPath, ok = combinePath(os.Args[0], *config); !ok {
			return false
		}
	}
	if configPath == "" {
		return true
	}
	parsedArgs, ok := getArgsFromConfig(configPath)
	if !ok {
		return false
	}
	missingFlags := getMissingFlags()

	ok = true
	oldFlagValues := make(map[string]string)
	for _, arg := range parsedArgs {
		f := flag.Lookup(arg.Key)
		if f == nil {
			log.Printf("Unknown flag name=[%s] found at line [%d] of file [%s]", arg.Key, arg.LineNum, arg.FilePath)
			ok = false
			continue
		}

		oldFlagValues[arg.Key] = f.Value.String()
		if _, found := missingFlags[f.Name]; found {
			if err := f.Value.Set(arg.Value); err != nil {
				log.Printf("Error when parsing flag [%s] value [%s] at line [%d] of file [%s]: [%s]", arg.Key, arg.Value, arg.LineNum, arg.FilePath, err)
				ok = false
			}
		}
	}

	if !ok {
		// restore old flag values
		for k, v := range oldFlagValues {
			flag.Set(k, v)
		}
	}

	return ok
}

func checkImportRecursion(configPath string) bool {
	for _, path := range importStack {
		if path == configPath {
			log.Printf("Import recursion found for [%s]: %v", configPath, importStack)
			return false
		}
	}
	return true
}

func getArgsFromConfig(configPath string) (args []Arg, ok bool) {
	if !checkImportRecursion(configPath) {
		return nil, false
	}
	importStack = append(importStack, configPath)
	defer func() {
		importStack = importStack[:len(importStack)-1]
	}()

	file := openConfigFile(configPath)
	if file == nil {
		return nil, false
	}
	defer file.Close()
	r := bufio.NewReader(file)

	var lineNum int
	for {
		lineNum++
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			if err == io.EOF {
				break
			}
			log.Printf("Error when reading file [%s] at line %d: [%s]\n", configPath, lineNum, err)
			return nil, false
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#import ") {
			importPath, ok := unquoteValue(line[7:], lineNum, configPath)
			if !ok {
				return nil, false
			}
			if importPath, ok = combinePath(configPath, importPath); !ok {
				return nil, false
			}
			importArgs, ok := getArgsFromConfig(importPath)
			if !ok {
				return nil, false
			}
			args = append(args, importArgs...)
			continue
		}
		if line == "" || line[0] == ';' || line[0] == '#' || line[0] == '[' {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("Cannot split [%s] at line %d into key and value in config file [%s]", line, lineNum, configPath)
			return nil, false
		}
		key := strings.TrimSpace(parts[0])
		value, ok := unquoteValue(parts[1], lineNum, configPath)
		if !ok {
			return nil, false
		}
		args = append(args, Arg{Key: key, Value: value, FilePath: configPath, LineNum: lineNum})
	}

	return args, true
}

func openConfigFile(path string) io.ReadCloser {
	if isHttp(path) {
		resp, err := http.Get(path)
		if err != nil {
			log.Printf("Cannot load config file at [%s]: [%s]\n", path, err)
			return nil
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("Unexpected http status code when obtaining config file [%s]: %d. Expected %d\n", path, resp.StatusCode, http.StatusOK)
			return nil
		}
		return resp.Body
	}

	file, err := os.Open(path)
	if err != nil {
		log.Printf("Cannot open config file at [%s]: [%s]\n", path, err)
		return nil
	}
	return file
}

func combinePath(basePath, relPath string) (string, bool) {
	if isHttp(basePath) {
		base, err := url.Parse(basePath)
		if err != nil {
			log.Printf("Error when parsing http base path [%s]: %s\n", basePath, err)
			return "", false
		}
		rel, err := url.Parse(relPath)
		if err != nil {
			log.Printf("Error when parsing http rel path [%s] for base [%s]: %s\n", relPath, basePath, err)
			return "", false
		}
		return base.ResolveReference(rel).String(), true
	}

	if relPath == "" || relPath[0] == '/' || isHttp(relPath) {
		return relPath, true
	}
	return path.Join(path.Dir(basePath), relPath), true
}

func isHttp(path string) bool {
	return strings.HasPrefix(strings.ToLower(path), "http://") || strings.HasPrefix(strings.ToLower(path), "https://")
}

func getMissingFlags() map[string]bool {
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	missingFlags := make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		if _, ok := setFlags[f.Name]; !ok {
			missingFlags[f.Name] = true
		}
	})
	return missingFlags
}

func dumpFlags() {
	flag.VisitAll(func(f *flag.Flag) {
		if f.Name != "config" && f.Name != "dumpflags" {
			fmt.Printf("%s = %s  # %s\n", f.Name, quoteValue(f.Value.String()), escapeUsage(f.Usage))
		}
	})
}

func escapeUsage(s string) string {
	return strings.Replace(s, "\n", "\n    # ", -1)
}

func quoteValue(v string) string {
	if !strings.ContainsAny(v, "\n#;") && strings.TrimSpace(v) == v {
		return v
	}
	v = strings.Replace(v, "\\", "\\\\", -1)
	v = strings.Replace(v, "\n", "\\n", -1)
	v = strings.Replace(v, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", v)
}

func unquoteValue(v string, lineNum int, configPath string) (string, bool) {
	v = strings.TrimSpace(v)
	if v[0] != '"' {
		return removeTrailingComments(v), true
	}
	n := strings.LastIndex(v, "\"")
	if n == -1 {
		log.Printf("Unclosed string found [%s] at line %d in config file [%s]", v, lineNum, configPath)
		return "", false
	}
	v = v[1:n]
	v = strings.Replace(v, "\\\"", "\"", -1)
	v = strings.Replace(v, "\\n", "\n", -1)
	return strings.Replace(v, "\\\\", "\\", -1), true
}

func removeTrailingComments(v string) string {
	v = strings.Split(v, "#")[0]
	v = strings.Split(v, ";")[0]
	return strings.TrimSpace(v)
}
