// Command livecheck exercises the client and live-data layer against a
// running syncthing instance (see scripts/test_live.sh) and prints a summary.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/evanw/syncthingtui/client"
)

func check(name string, err error) {
	if err != nil {
		fmt.Printf("FAIL %-28s %v\n", name, err)
		os.Exit(1)
	}
	fmt.Printf("ok   %s\n", name)
}

func main() {
	address := flag.String("address", "127.0.0.1:8384", "")
	apiKey := flag.String("api-key", "", "")
	flag.Parse()
	c := client.New(*address, *apiKey)

	cfg, err := c.Config()
	check("GET /rest/config", err)
	fmt.Printf("     folders=%d devices=%d\n", len(cfg.Folders), len(cfg.Devices))

	status, err := c.SystemStatus()
	check("GET /rest/system/status", err)
	fmt.Printf("     myID=%s uptime=%ds\n", status.MyID[:7], status.Uptime)

	_, err = c.Connections()
	check("GET /rest/system/connections", err)

	v, err := c.Version()
	check("GET /rest/system/version", err)
	fmt.Printf("     version=%s\n", v.Version)

	_, err = c.PendingDevices()
	check("GET /rest/cluster/pending/devices", err)
	_, err = c.PendingFolders()
	check("GET /rest/cluster/pending/folders", err)
	_, err = c.DeviceStats()
	check("GET /rest/stats/device", err)

	// /rest/system/paths only exists on syncthing >= v1.19; optional.
	if _, err = c.Paths(); err != nil {
		fmt.Printf("skip GET /rest/system/paths (%v)\n", err)
	} else {
		fmt.Println("ok   GET /rest/system/paths")
	}

	// Notice round-trip: post an error, see it, clear it, see none.
	check("POST /rest/system/error", c.PostError("livecheck test error"))
	errs, err := c.Errors()
	check("GET /rest/system/error", err)
	if len(errs) == 0 || errs[len(errs)-1].Message != "livecheck test error" {
		fmt.Println("FAIL posted error not returned")
		os.Exit(1)
	}
	fmt.Println("ok   posted error visible")
	check("POST /rest/system/error/clear", c.ClearErrors())
	errs, _ = c.Errors()
	if len(errs) != 0 {
		fmt.Println("FAIL errors not cleared")
		os.Exit(1)
	}
	fmt.Println("ok   errors cleared")

	if len(cfg.Folders) > 0 {
		id := cfg.Folders[0].ID
		_, err = c.DBStatus(id)
		check("GET /rest/db/status", err)
		check("POST /rest/db/scan", c.Rescan(id))
		check("PATCH folder paused=true", c.SetFolderPaused(id, true))
		cfg2, _ := c.Config()
		if !cfg2.Folders[0].Paused {
			fmt.Println("FAIL pause not reflected in config")
			os.Exit(1)
		}
		fmt.Println("ok   pause reflected in config")
		check("PATCH folder paused=false", c.SetFolderPaused(id, false))
	}
	fmt.Println("ALL OK")
}
