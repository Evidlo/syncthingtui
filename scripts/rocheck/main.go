// Read-only check against the real local syncthing via Discover().
package main

import (
	"fmt"

	"github.com/evidlo/syncthingtui/client"
)

func main() {
	addr, key, err := client.Discover()
	fmt.Printf("discover: addr=%q key=%q... err=%v\n", addr, key[:min(8, len(key))], err)
	if err != nil {
		return
	}
	c := client.New(addr, key)
	cfg, err := c.Config()
	fmt.Printf("config: folders=%d devices=%d err=%v\n", len(cfg.Folders), len(cfg.Devices), err)
	for _, f := range cfg.Folders {
		fmt.Printf("  folder id=%q label=%q paused=%v\n", f.ID, f.Label, f.Paused)
	}
	st, err := c.SystemStatus()
	fmt.Printf("status: myID=%.7s err=%v\n", st.MyID, err)
	for _, d := range cfg.Devices {
		fmt.Printf("  device id=%.7s name=%q\n", d.DeviceID, d.Name)
	}
}
