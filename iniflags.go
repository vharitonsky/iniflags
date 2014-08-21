package iniflags

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

type Arg struct {
	Key     string
	Value   string
	LineNum int
}

var (
	config    = flag.String("config", "", "Path to ini config for using in go flags. May be relative to the current executable path")
	dumpflags = flag.Bool("dumpflags", false, "Dumps values for all flags defined in the app into stdout in ini-compatible syntax and terminates the app")

	importStack []string
)

func Parse() {
	flag.Parse()
	if !parseConfigFlags() {
		os.Exit(1)
	}
	if *dumpflags {
		dumpFlags()
		os.Exit(0)
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP)
	go sighupHandler(ch)
}

func sighupHandler(ch <-chan os.Signal) {
	for _ = range ch {
		log.Printf("Re-reading flags from config files\n")
		parseConfigFlags()
	}
}

func parseConfigFlags() bool {
	configPath := *config
	if !strings.HasPrefix(configPath, "./") {
		configPath = combinePath(os.Args[0], *config)
	}
	if configPath == "" {
		return true
	}
	parsedArgs, ok := getArgsFromConfig(configPath)
	if !ok {
		return false
	}
	allFlags, missingFlags := getFlags()
	for _, arg := range parsedArgs {
		if _, found := allFlags[arg.Key]; !found {
			log.Printf("Unknown flag name=[%s] found at line [%d] of file [%s]", arg.Key, arg.LineNum, configPath)
			return false
		}
		if _, found := missingFlags[arg.Key]; found {
			flag.Set(arg.Key, arg.Value)
		}
	}
	return true
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

	file, err := os.Open(configPath)
	if err != nil {
		log.Printf("Cannot open config file at [%s]: [%s]\n", configPath, err)
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
			importPath = combinePath(configPath, importPath)
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
		args = append(args, Arg{Key: key, Value: value, LineNum: lineNum})
	}

	importStack = importStack[:len(importStack)-1]
	return args, true
}

func combinePath(basePath, relPath string) string {
	if relPath == "" || relPath[0] == '/' {
		return relPath
	}
	return path.Join(path.Dir(basePath), relPath)
}

func getFlags() (allFlags, missingFlags map[string]bool) {
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	allFlags = make(map[string]bool)
	missingFlags = make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		allFlags[f.Name] = true
		if _, ok := setFlags[f.Name]; !ok {
			missingFlags[f.Name] = true
		}
	})
	return
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
