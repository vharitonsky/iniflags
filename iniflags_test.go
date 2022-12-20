package iniflags

import (
	"flag"
	"fmt"
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

func TestGetTrailingComments(t *testing.T) {
	hashCommented := "v = v # test_comment"
	comment := getTrailingComment(hashCommented)
	if comment != " test_comment" {
		t.Fatalf("Supposed to get ' test_comment', got '%s'", comment)
	}
	colonCommented := "v = v ; test_comment"
	comment = getTrailingComment(colonCommented)
	if comment != " test_comment" {
		t.Fatalf("Supposed to get ' test_comment', got '%s'", comment)
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
	fixedVal, comment, ok := unquoteValue(val, 0, "")
	if !ok || fixedVal != "val#;\"\n" {
		t.Fatalf("Value should be unquoted and stripped, got '%s'", fixedVal)
	}
	expected := " test\n"
	if comment != expected {
		t.Fatalf("Supposed to get '%q', got '%q'", expected, comment)
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
	var expected string
	var checkedVar0, checkedVar1, checkedVar2, checkedVar3, checkedVar4 bool
	for _, arg := range args {
		t.Log(arg.Key, fmt.Sprintf("%q", arg.Value))
		switch arg.Key {
		case "var0":
			expected = "val0"
			if arg.Value != expected {
				t.Fatalf("Val of 'var0' should be '%q', got %q", expected, arg.Value)
			}
			expected = " comment"
			if arg.Comment != expected {
				t.Fatalf("Comment of 'var0' should be '%q', got %q", expected, arg.Comment)
			}
			checkedVar0 = true
		case "var1":
			expected = "val#1\n\\\"\nx"
			if arg.Value != expected {
				t.Fatalf("Invalid val for var1 should be '%q', got '%q'", expected, arg.Value)
			}
			expected = " this is a test comment"
			if arg.Comment != expected {
				t.Fatalf("Comment of 'var1' should be '%q', got %q", expected, arg.Comment)
			}
			checkedVar1 = true
		case "var2":
			expected = "1234"
			if arg.Value != expected {
				t.Fatalf("Val of 'var2' should be '%q', got %q", expected, arg.Value)
			}
			checkedVar2 = true
		case "var3":
			expected = ""
			if arg.Value != expected {
				t.Fatalf("Val of 'var3' should be '%q', got %q", expected, arg.Value)
			}
			expected = " empty value"
			if arg.Comment != expected {
				t.Fatalf("Comment of 'var3' should be '%q', got %q", expected, arg.Comment)
			}
			checkedVar3 = true
		case "var4":
			expected = "multi,var|12345"
			if arg.Value != expected {
				t.Fatalf("Val of 'var4' should be '%q', got %q", expected, arg.Value)
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

func TestMissingConfig(t *testing.T) {
	parsed = false
	*configUpdateInterval = 0
	*allowMissingConfig = false
	SetConfigFile("non-existent.ini")
	err := SafeParse()
	if err == nil {
		t.Fatal("SafeParse claimed to parse non-existent file.")
	}
}
