package testutils

import (
	"net/http"
	"os"
	"testing"

	"go.uber.org/goleak"
)

// VerifyTestMain standardizes goroutine leak checks.
// We avoid closing the global pool here because other parallel test processes
// might still be using the shared container.
func VerifyTestMain(m *testing.M) int {
	code := m.Run()

	if tr, ok := http.DefaultTransport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}

	if err := goleak.Find(
		goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"),
		goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go.(*Reaper).connect.func1"),
		goleak.IgnoreTopFunction("github.com/testcontainers/testcontainers-go/internal/testrand.init"),
	); err != nil {
		os.Stderr.WriteString("goleak: " + err.Error() + "\n")
		return 1
	}

	return code
}
