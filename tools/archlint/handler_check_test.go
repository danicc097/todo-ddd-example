package archlint_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/danicc097/todo-ddd-example/tools/archlint"
)

func TestHandlerArch(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, archlint.HandlerAnalyzer, "github.com/danicc097/todo-ddd-example/internal/modules/...")
}
