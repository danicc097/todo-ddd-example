package ws_test

import (
	"os"
	"testing"

	"github.com/danicc097/todo-ddd-example/internal/testutils"
)

func TestMain(m *testing.M) {
	os.Exit(testutils.VerifyTestMain(m))
}
