// syncthingtui - a terminal user interface for Syncthing
// Copyright (C) 2026 Evan Widloski
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: GPL-3.0-or-later

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
