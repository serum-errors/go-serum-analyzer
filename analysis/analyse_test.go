package analysis

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestToyAnalyzer(t *testing.T) {
	t.Skip("this was for experiments only, not really a test")
	analysistest.Run(t, analysistest.TestData(), ToyAnalyzer, "toy")
}

func TestVerifyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "001")
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "docformat")
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "dotimport")
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "errortypes")
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "field_assignment")
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "methods")
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "multipackage/inner1", "multipackage")
	analysistest.Run(t, analysistest.TestData(), VerifyAnalyzer, "recursion")
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
