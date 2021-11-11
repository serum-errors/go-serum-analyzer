package analysis

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestVerifyAnalyzer(t *testing.T) {
	Analyzer.Flags.Set("strict", "true")
	dir := analysistest.TestData()
	analysistest.Run(t, dir, Analyzer,
		"001",
		"annotation",
		"docformat",
		"dotimport/inner1", "dotimport",
		"error_constructor",
		"errortypes",
		"examples",
		"field_assignment",
		"func_literal",
		"interfaces/inner1", "interfaces",
		"methods",
		"multifile",
		"multipackage/inner1", "multipackage",
		"recursion",
	)
}

func TestIsErrorCodeValid(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		{"error", true},
		{"valid-error", true},
		{"ValidError", true},
		{"-invalid", false},
		{"invalid-", false},
		{"3invalid", false},
		{"a", true},
		{"-", false},
		{"invalid$error", false},
		{"invalid error", false},
		{"some-2-error", true},
	}

	for _, test := range tests {
		if isErrorCodeValid(test.code) != test.valid {
			t.Errorf("isErrorCodeValid(%q) should return %v but did not", test.code, test.valid)
		}
	}
}
