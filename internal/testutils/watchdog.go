package testutils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

// nolint: gochecknoinits
func init() {
	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot, _ := filepath.Abs(filepath.Join(filepath.Dir(thisFile), "../../"))

	scriptPath := filepath.Join(projectRoot, "scripts/test_watchdog.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		panic("Watchdog script not found at " + scriptPath)
	}

	lastRunPath := filepath.Join(projectRoot, ".test_last_run")
	now := strconv.FormatInt(time.Now().Unix(), 10)
	_ = os.WriteFile(lastRunPath, []byte(now), 0o644)

	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = projectRoot

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // detach from the go test process
	}

	if err := cmd.Start(); err != nil {
		panic("Failed to start watchdog script: " + err.Error())
	}
}
