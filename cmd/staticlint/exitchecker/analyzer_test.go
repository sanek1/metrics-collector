package exitchecker_test

import (
	"os"
	"testing"

	"github.com/sanek1/metrics-collector/cmd/staticlint/exitchecker"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	_ = os.Setenv("GODEBUG", "analysisnoverify=1")
	testdata := analysistest.TestData()
	t.Logf("path to testdata: %s", testdata)

	t.Logf("checking package 'a'...")
	analysistest.Run(t, testdata, exitchecker.Analyzer, "a")

	t.Logf("checking package 'b'...")
	analysistest.Run(t, testdata, exitchecker.Analyzer, "b")
}
