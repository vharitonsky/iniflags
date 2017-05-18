package iniflags

import (
	"flag"
	"testing"
	"time"
)

func TestRemoveTrailingComments(t *testing.T) {
	hashCommented := "v = v # test_comment"
	clean := removeTrailingComments(hashCommented)
	if clean != "v = v" {
		t.Fatalf("Supposed to get 'v = v ', got '%s'", clean)
	}
	colonCommented := "v = v ; test_comment"
	clean = removeTrailingComments(colonCommented)
	if clean != "v = v" {
		t.Fatalf("Supposed to get 'v = v ', got '%s'", clean)
	}
}

func TestBOM(t *testing.T) {
	args, ok := getArgsFromConfig("test_bom.ini")
	if !ok {
		t.Fail()
	}
	if len(args) != 1 {
		t.Fatalf("Unexpected number of args parsed: %d. Expected 1", len(args))
	}
	if args[0].Key != "bom" {
		t.Fatalf("Unexpected key name parsed: %q. Expected \"bom\"", args[0].Key)
	}
	if args[0].Value != "привет" {
		t.Fatalf("Unexpected value parsed: %q. Expected \"привет\"", args[0].Value)
	}
}

func TestUnquoteValue(t *testing.T) {
	val := "\"val#;\\\"\\n\"    # test\n"
	fixedVal, ok := unquoteValue(val, 0, "")
	if !ok || fixedVal != "val#;\"\n" {
		t.Fatalf("Value should be unquoted and stripped, got '%s'", fixedVal)
	}
}

func TestGetFlags(t *testing.T) {
	parsed = false
	Parse()
	missingFlags := getMissingFlags()
	if _, found := missingFlags["config"]; !found {
		t.Fatalf("'config' flag should be missing in tests")
	}
}

func TestGetArgsFromConfig(t *testing.T) {
	args, ok := getArgsFromConfig("test_config.ini")
	if !ok {
		t.Fail()
	}
	var checkedVar0, checkedVar1, checkedVar2, checkedVar3, checkedVar4 bool
	for _, arg := range args {
		t.Log(arg.Key, arg.Value)
		switch arg.Key {
		case "var0":
			if arg.Value != "val0" {
				t.Fatalf("Val of 'var0' should be 'val0', got %q", arg.Value)
			}
			checkedVar0 = true
		case "var1":
			if arg.Value != "val#1\n\\\"\nx" {
				t.Fatalf("Invalid val for var1=%q", arg.Value)
			}
			checkedVar1 = true
		case "var2":
			if arg.Value != "1234" {
				t.Fatalf("Val of 'var2' should be '1234', got %q", arg.Value)
			}
			checkedVar2 = true
		case "var3":
			if arg.Value != "" {
				t.Fatalf("Val of 'var3' should be '', got %q", arg.Value)
			}
			checkedVar3 = true
		case "var4":
			if arg.Value != "multi,var|12345" {
				t.Fatalf("Val of 'var4' should be 'multi,var,12345', got %q", arg.Value)
			}
			checkedVar4 = true
		}
	}
	if !checkedVar0 || !checkedVar1 || !checkedVar2 || !checkedVar3 || !checkedVar4 {
		t.Fatalf("Not all vals checked: args=[%v], %v, %v, %v, %v, %v",
			args, checkedVar0, checkedVar1, checkedVar2, checkedVar3, checkedVar4)
	}
}

func TestIsHTTP(t *testing.T) {
	if !isHTTP("http://example.com") {
		t.Fatalf("http://example.com should must be recognized as http path")
	}
	if !isHTTP("hTtpS://example.com") {
		t.Fatalf("hTtpS://example.com should must be recognized as http path")
	}
}

var x = flag.String("x", "baz", "for TestSetConfigFile")

func TestSetConfigFile(t *testing.T) {
	parsed = false
	SetConfigFile("./test_setconfigfile.ini")
	Parse()
	if *x != "foobar" {
		t.Fatalf("Unexpected x=[%s]. Expected [foobar]", *x)
	}
}

func TestSetAllowMissingConfigFile(t *testing.T) {
	parsed = false
	*allowMissingConfig = false
	SetAllowMissingConfigFile(true)
	if *allowMissingConfig != true {
		t.Fatal("SetAllowUnknownFlags failed to update global.")
	}
}

func TestSetAllowUnknownFlags(t *testing.T) {
	parsed = false
	*allowUnknownFlags = false
	SetAllowUnknownFlags(true)
	if *allowUnknownFlags != true {
		t.Fatal("SetAllowUnknownFlags failed to update global.")
	}
}

func TestSetConfigUpdateInterval(t *testing.T) {
	parsed = false
	*configUpdateInterval = time.Second
	SetConfigUpdateInterval(time.Minute)
	if *configUpdateInterval != time.Minute {
		t.Fatal("SetConfigUpdateInterval failed to update global.")
	}
}
