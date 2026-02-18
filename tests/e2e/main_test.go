//go:build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestMain(m *testing.M) {
	os.Exit(testutils.VerifyTestMain(m))
}
