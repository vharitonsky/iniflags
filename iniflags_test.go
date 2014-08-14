package iniflags

import (
	"testing"
)

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
	checked_var0, checked_var1 := false, false
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
			if arg.Value != "val1" {
				t.Error("Val of 'var1' should be 'val1', got '%s'", arg.Value)
				t.Fail()
			} else {
				checked_var1 = true
			}
		}
	}
	if !checked_var0 || !checked_var1 {
		t.Error("Not all vals checked:", args, checked_var0, checked_var1)
		t.Fail()
	}
}
