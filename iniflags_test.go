package iniflags

import (
	"testing"
)

func TestRemoveTrailingComments(t *testing.T) {
	hash_commented := "v = v # test_comment"
	clean := removeTrailingComments(hash_commented)
	if clean != "v = v" {
		t.Errorf("Supposed to get 'v = v ', got '%s'", clean)
		t.Fail()

	}
	colon_commented := "v = v ; test_comment"
	clean = removeTrailingComments(colon_commented)
	if clean != "v = v" {
		t.Errorf("Supposed to get 'v = v ', got '%s'", clean)
		t.Fail()
	}

}

func TestUnquoteValue(t *testing.T) {
	val := "\"val\"\n"
	fixed_val := unquoteValue(val, 0, "")
	if fixed_val != "val" {
		t.Error("Value should be unquoted and stripped, got", fixed_val)
		t.Fail()
	}
}

func TestGetFlags(t *testing.T) {
	Parse()
	allFlags, missingFlags := getFlags()
	if _, found := missingFlags["config"]; !found {
		t.Error("'config' flag should be missing in tests")
		t.Fail()
	}
	if _, found := allFlags["config"]; !found {
		t.Error("'config' flag should be present in tests")
		t.Fail()
	}

}

func TestGetArgsFromConfig(t *testing.T) {
	args := getArgsFromConfig("test_config.ini")
	checked_var0, checked_var1, checked_var2 := false, false, false
	for _, arg := range args {
		t.Log(arg.Key, arg.Value)
		if arg.Key == "var0" {
			if arg.Value != "val0" {
				t.Errorf("Val of 'var0' should be 'val0', got '%s'", arg.Value)
				t.Fail()
			} else {
				checked_var0 = true
			}
		}
		if arg.Key == "var1" {
			if arg.Value != "val#1\n\\\"\nx" {
				t.Errorf("Invalid val for var1='%s'", arg.Value)
				t.Fail()
			} else {
				checked_var1 = true
			}
		}
		if arg.Key == "var2" {
			if arg.Value != "1234" {
				t.Errorf("Val of 'var2' should be '1234', got '%s'", arg.Value)
				t.Fail()
			} else {
				checked_var2 = true
			}
		}
	}
	if !checked_var0 || !checked_var1 || !checked_var2 {
		t.Errorf("Not all vals checked: args=[%v], %v, %v, %v", args, checked_var0, checked_var1, checked_var2)
		t.Fail()
	}
}
