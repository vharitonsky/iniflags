package iniflags

import (
	"testing"
)

func TestUnquoteValue(t *testing.T){
	val := "\"val\"\n"
	fixed_val := unquoteValue(val)
	if fixed_val != "val"{
		t.Error("Value should be unquoted and stripped, got", fixed_val)
		t.Fail()
	}
}

func TestGetMissingFlags(t *testing.T){
	Parse()
	missingFlags := getMissingFlags()
	if _, found := missingFlags["config"]; !found{
		t.Error("'config' flag should be missing in tests")
		t.Fail()
	}
}

func TestGetArgsFromConfig(t *testing.T){
	args := getArgsFromConfig("test_config.ini")
	checked_var0, checked_var1 := false, false
	for _, arg := range args{
		if arg.key == "var0"{
			if arg.value != "val0"{
				t.Error("Val of 'var0' should be 'val0', got", arg.value)
				t.Fail()
			}else{
				checked_var0 = true
			}
		}
		if arg.key == "var1"{
			if arg.value != "val1"{
				t.Error("Val of 'var1' should be 'val1', got", arg.value)
				t.Fail()
			}else{
				checked_var1 = true
			}
		}
	}
	if !checked_var0 || !checked_var1{
		t.Error("Not all vals checked:", checked_var0, checked_var1)
		t.Fail()
	}
}