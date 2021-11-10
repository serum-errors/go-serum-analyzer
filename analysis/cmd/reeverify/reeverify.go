// The analyse command runs the error code analyzer.
package main

import (
	"github.com/warpfork/go-ree/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(analysis.Analyzer) }
