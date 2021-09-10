package analysis_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/warpfork/go-ree/analysis"
)

func TestToyAnalyzer(t *testing.T) {
	t.Skip("this was for experiments only, not really a test")
	analysistest.Run(t, analysistest.TestData(), analysis.ToyAnalyzer, "toy")
}

func TestVerifyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analysis.VerifyAnalyzer, "001", "docformat")
	analysistest.Run(t, analysistest.TestData(), analysis.VerifyAnalyzer, "multipackage/inner1", "multipackage")

	// TODO: All of the examples in the following test currently lead to endless loops in our analyzer.
	// analysistest.Run(t, analysistest.TestData(), analysis.VerifyAnalyzer, "recursion")
}
