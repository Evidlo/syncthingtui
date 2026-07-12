package client_test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestLiveIntegration runs the full REST integration suite against a
// throwaway syncthing instance (scripts/test_live.sh + scripts/livecheck).
// Skipped when syncthing isn't installed or with -short.
func TestLiveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("-short")
	}
	if _, err := exec.LookPath("syncthing"); err != nil {
		t.Skip("syncthing not installed")
	}
	out, err := exec.Command("bash", "../scripts/test_live.sh").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "ALL OK") {
		t.Fatalf("integration failed (%v):\n%s", err, out)
	}
}
