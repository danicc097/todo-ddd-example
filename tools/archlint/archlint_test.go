package archlint_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/danicc097/todo-ddd-example/tools/archlint"
)

func TestArchLinters(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	t.Run("HandlerAnalyzer", func(t *testing.T) {
		analysistest.Run(t, testdata, archlint.HandlerAnalyzer, "github.com/danicc097/todo-ddd-example/internal/modules/todo/...")
	})

	t.Run("GinLeakAnalyzer", func(t *testing.T) {
		analysistest.Run(t, testdata, archlint.GinLeakAnalyzer, "github.com/danicc097/todo-ddd-example/internal/modules/ginleak/...")
	})

	t.Run("CQRSPurityAnalyzer", func(t *testing.T) {
		analysistest.Run(t, testdata, archlint.CQRSPurityAnalyzer, "github.com/danicc097/todo-ddd-example/internal/modules/cqrs/...")
	})

	t.Run("DTOBleedAnalyzer", func(t *testing.T) {
		analysistest.Run(t, testdata, archlint.DTOBleedAnalyzer, "github.com/danicc097/todo-ddd-example/internal/modules/dtobleed/...")
	})
}
