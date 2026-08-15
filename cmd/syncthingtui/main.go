// Command syncthingtui is a TUI for syncthing.
//
//	go run ./cmd/syncthingtui                     # live TUI (auto-discovers local syncthing)
//	go run ./cmd/syncthingtui -fake               # fake-data mode (no syncthing needed)
//	go run ./cmd/syncthingtui -address 127.0.0.1:8384 -api-key XYZ
//	go run ./cmd/syncthingtui -config /path/to/config.xml
//	go run ./cmd/syncthingtui -screen <name>      # static render (80x60, fake data)
//	go run ./cmd/syncthingtui -list               # list static screens
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evidlo/syncthingtui/app"
	"github.com/evidlo/syncthingtui/client"
)

// minSyncthing is the oldest syncthing the TUI supports: the
// /rest/cluster/pending/{devices,folders} endpoints behind the Alerts tab
// landed in v1.18.0; older instances 404 there and Alerts is silently blank.
var minSyncthing = [3]int{1, 18, 0}

// parseVersion extracts major.minor.patch from a syncthing version string such
// as "v1.12.1-ds1". Returns ok=false when it can't be parsed.
func parseVersion(s string) ([3]int, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexAny(s, "-+ "); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return [3]int{}, false
	}
	var v [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return [3]int{}, false
		}
		v[i] = n
	}
	return v, true
}

// checkVersion fails when the connected syncthing is older than minSyncthing.
// An unreadable or unparseable version is not treated as fatal.
func checkVersion(c *client.Client) error {
	ver, err := c.Version()
	if err != nil {
		return fmt.Errorf("cannot read syncthing version: %w", err)
	}
	v, ok := parseVersion(ver.Version)
	if !ok {
		return nil
	}
	if v[0] < minSyncthing[0] ||
		(v[0] == minSyncthing[0] && (v[1] < minSyncthing[1] ||
			(v[1] == minSyncthing[1] && v[2] < minSyncthing[2]))) {
		return fmt.Errorf("syncthing %s is too old; syncthingtui needs v%d.%d.%d or newer",
			ver.Version, minSyncthing[0], minSyncthing[1], minSyncthing[2])
	}
	return nil
}

func main() {
	screen := flag.String("screen", "", "render a named screen statically and exit")
	size := flag.String("size", "80x60", "WxH for -screen")
	list := flag.Bool("list", false, "list static screen names")
	fake := flag.Bool("fake", false, "use fake data instead of a live syncthing")
	address := flag.String("address", "", "syncthing GUI address (default: from config.xml)")
	apiKey := flag.String("api-key", os.Getenv("SYNCTHING_API_KEY"), "API key (default: from config.xml)")
	config := flag.String("config", "", "path to syncthing config.xml (default: auto-discover)")
	flag.Parse()

	if *list {
		for _, n := range app.ScreenNames() {
			fmt.Println(n)
		}
		return
	}
	if *screen != "" {
		width, height := 80, 60
		fmt.Sscanf(*size, "%dx%d", &width, &height)
		out, err := app.RenderScreen(*screen, width, height)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}
	a := app.New()
	if !*fake {
		addr, key := *address, *apiKey
		if addr == "" || key == "" {
			var dAddr, dKey string
			var err error
			if *config != "" {
				dAddr, dKey, err = client.FromConfig(*config)
			} else {
				dAddr, dKey, err = client.Discover()
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot find local syncthing (%v); use -address/-api-key, -config, or -fake\n", err)
				os.Exit(1)
			}
			if addr == "" {
				addr = dAddr
			}
			if key == "" {
				key = dKey
			}
		}
		c := client.New(addr, key)
		if err := checkVersion(c); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		a = app.NewLive(c)
	}
	if _, err := tea.NewProgram(a, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
