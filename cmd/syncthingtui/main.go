// Command syncthingtui is a TUI for syncthing.
//
//	go run ./cmd/syncthingtui                     # live TUI (auto-discovers local syncthing)
//	go run ./cmd/syncthingtui -fake               # fake-data mode (no syncthing needed)
//	go run ./cmd/syncthingtui -address 127.0.0.1:8384 -api-key XYZ
//	go run ./cmd/syncthingtui -screen <name>      # static render (80x60, fake data)
//	go run ./cmd/syncthingtui -list               # list static screens
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evidlo/syncthingtui/app"
	"github.com/evidlo/syncthingtui/client"
)

func main() {
	screen := flag.String("screen", "", "render a named screen statically and exit")
	size := flag.String("size", "80x60", "WxH for -screen")
	list := flag.Bool("list", false, "list static screen names")
	fake := flag.Bool("fake", false, "use fake data instead of a live syncthing")
	address := flag.String("address", "", "syncthing GUI address (default: from config.xml)")
	apiKey := flag.String("api-key", os.Getenv("SYNCTHING_API_KEY"), "API key (default: from config.xml)")
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
			dAddr, dKey, err := client.Discover()
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot find local syncthing (%v); use -address/-api-key or -fake\n", err)
				os.Exit(1)
			}
			if addr == "" {
				addr = dAddr
			}
			if key == "" {
				key = dKey
			}
		}
		a = app.NewLive(client.New(addr, key))
	}
	if _, err := tea.NewProgram(a, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
