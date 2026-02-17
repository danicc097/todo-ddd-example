//go:build e2e

package e2e

import (
	"net/http"
	"os"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	code := m.Run()

	if tr, ok := http.DefaultTransport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}

	if err := goleak.Find(); err != nil {
		os.Stderr.WriteString("goleak: " + err.Error() + "\n")
		os.Exit(1)
	}

	os.Exit(code)
}
