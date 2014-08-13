package iniflags

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type Arg struct {
	key, value string
}

var (
	config = flag.String("config", "dev.ini", "Path to config.")
)

var (
	LINES_REGEXP = regexp.MustCompile("[\\r\\n]")
	KV_REGEXP    = regexp.MustCompile("\\s*=\\s*")
)

func Parse() {
	flag.Parse()
	parsedArgs := getArgsFromConfig(*config)
	not_set_flags := getNotSetFlags()
	for _, arg := range parsedArgs {
		if _, found := not_set_flags[arg.key]; found {
			flag.Set(arg.key, arg.value)
		}
	}
}

func getArgsFromConfig(configPath string) []Arg {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("cannot open config file at [%s]: [%s]\n", configPath, err)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Error when reading config file [%s]: [%s]\n", configPath, err)
	}

	var args []Arg
	for _, line := range LINES_REGEXP.Split(string(data), -1) {
		if line == "" || line[0] == ';' || line[0] == '#' || line[0] == '[' {
			continue
		}
		parts := KV_REGEXP.Split(line, 2)
		if len(parts) != 2 {
			log.Fatalf("Cannot split line=[%s] into key and value in config file [%s]", line, configPath)
		}
		key := parts[0]
		value := unquoteValue(parts[1])
		args = append(args, Arg{key: key, value: value})
	}
	return args
}

func getNotSetFlags() map[string]bool {
	not_set_flags := make(map[string]bool, 0)
	flag.VisitAll(func(f *flag.Flag) {
		not_set_flags[f.Name] = true
	})
	flag.Visit(func(f *flag.Flag) {
		delete(not_set_flags, f.Name)
	})
	return not_set_flags
}

func unquoteValue(v string) string {
	if v[0] != '"' {
		return v
	}
	n := strings.LastIndex(v, "\"")
	if n == -1 {
		return v
	}
	v = v[1:n]
	v = strings.Replace(v, "\\\"", "\"", -1)
	return strings.Replace(v, "\\n", "\n", -1)
}
