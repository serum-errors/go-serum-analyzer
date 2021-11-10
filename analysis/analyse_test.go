package analysis

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestVerifyAnalyzer(t *testing.T) {
	Analyzer.Flags.Set("strict", "true")
	dir := analysistest.TestData()
	analysistest.Run(t, dir, Analyzer, "001")
	analysistest.Run(t, dir, Analyzer, "docformat")
	analysistest.Run(t, dir, Analyzer, "dotimport/inner1", "dotimport")
	analysistest.Run(t, dir, Analyzer, "error_constructor")
	analysistest.Run(t, dir, Analyzer, "errortypes")
	analysistest.Run(t, dir, Analyzer, "examples")
	analysistest.Run(t, dir, Analyzer, "field_assignment")
	analysistest.Run(t, dir, Analyzer, "func_literal")
	analysistest.Run(t, dir, Analyzer, "interfaces/inner1", "interfaces")
	analysistest.Run(t, dir, Analyzer, "methods")
	analysistest.Run(t, dir, Analyzer, "multifile")
	analysistest.Run(t, dir, Analyzer, "multipackage/inner1", "multipackage")
	analysistest.Run(t, dir, Analyzer, "recursion")
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
