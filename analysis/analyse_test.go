package analysis

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestToyAnalyzer(t *testing.T) {
	t.Skip("this was for experiments only, not really a test")
	analysistest.Run(t, analysistest.TestData()+"/toy", ToyAnalyzer)
}

func TestVerifyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData()+"/001", VerifyAnalyzer)
}
