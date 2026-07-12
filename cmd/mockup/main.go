// Command mockup renders a named mockup screen at a fixed size for static
// layout review (see AGENTS.md: static TUI rendering requirement).
//
//	go run ./cmd/mockup <screen> [WxH]   # default 80x60
//	go run ./cmd/mockup -list
package main

import (
	"fmt"
	"os"

	"github.com/evidlo/syncthingtui/mockups"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "-list" {
		for _, n := range mockups.Names() {
			fmt.Println(n)
		}
		return
	}
	width, height := 80, 60
	if len(args) > 1 {
		fmt.Sscanf(args[1], "%dx%d", &width, &height)
	}
	render, ok := mockups.Screens[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown screen %q; run with -list\n", args[0])
		os.Exit(1)
	}
	fmt.Println(render(width, height))
}
