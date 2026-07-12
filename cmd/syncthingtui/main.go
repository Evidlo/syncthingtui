// Command syncthingtui runs the interactive stage-3 mockup.
//
//	go run ./cmd/syncthingtui                     # interactive TUI
//	go run ./cmd/syncthingtui -screen <name>      # static render (80x60)
//	go run ./cmd/syncthingtui -screen <name> -size 100x40
//	go run ./cmd/syncthingtui -list               # list static screens
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanw/syncthingtui/app"
)

func main() {
	screen := flag.String("screen", "", "render a named screen statically and exit")
	size := flag.String("size", "80x60", "WxH for -screen")
	list := flag.Bool("list", false, "list static screen names")
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
	if _, err := tea.NewProgram(app.New(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
