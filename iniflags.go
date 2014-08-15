package iniflags

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type Arg struct {
	Key     string
	Value   string
	LineNum int
}

var (
	config = flag.String("config", "", "Path to ini config for using in go flags. May be relative to the current executable path")
)

func Parse() {
	flag.Parse()
	configPath := *config
	if configPath == "" {
		return
	}
	if configPath[0] != '/' {
		configPath = path.Join(path.Dir(os.Args[0]), configPath)
	}
	parsedArgs := getArgsFromConfig(configPath)
	allFlags, missingFlags := getFlags()
	for _, arg := range parsedArgs {
		if _, found := allFlags[arg.Key]; !found {
			log.Fatalf("Unknown flag name=[%s] found at line [%d] of file [%s]", arg.Key, arg.LineNum, configPath)
		}
		if _, found := missingFlags[arg.Key]; found {
			flag.Set(arg.Key, arg.Value)
		}
	}
}

func getArgsFromConfig(configPath string) []Arg {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Cannot open config file at [%s]: [%s]\n", configPath, err)
	}
	defer file.Close()
	r := bufio.NewReader(file)

	var args []Arg
	var lineNum int
	for {
		lineNum++
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error when reading file [%s] at line %d: [%s]\n", configPath, lineNum, err)
		}
		line = strings.TrimSpace(line)
		if line == "" || line[0] == ';' || line[0] == '#' || line[0] == '[' {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("Cannot split [%s] at line %d into key and value in config file [%s]", line, lineNum, configPath)
		}
		key := strings.TrimSpace(parts[0])
		value := unquoteValue(strings.TrimSpace(parts[1]), lineNum, configPath)
		args = append(args, Arg{Key: key, Value: value, LineNum: lineNum})
	}
	return args
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

func unquoteValue(v string, lineNum int, configPath string) string {
	if v[0] != '"' {
		return removeTrailingComments(v)
	}
	n := strings.LastIndex(v, "\"")
	if n == -1 {
		log.Fatalf("Unclosed string found [%s] at line %d in config file [%s]", v, lineNum, configPath)
	}
	v = v[1:n]
	v = strings.Replace(v, "\\\"", "\"", -1)
	v = strings.Replace(v, "\\n", "\n", -1)
	return strings.Replace(v, "\\\\", "\\", -1)
}

func removeTrailingComments(v string) string {
	v = strings.Split(v, "#")[0]
	v = strings.Split(v, ";")[0]
	return strings.TrimSpace(v)
}
