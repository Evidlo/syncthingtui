// Command livecheck exercises the client and live-data layer against a
// running syncthing instance (see scripts/test_live.sh) and prints a summary.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/evidlo/syncthingtui/client"
)

func rawDeviceHasIgnoredFolder(raw map[string]any, deviceID, folderID string) bool {
	devs, _ := raw["devices"].([]any)
	for _, d := range devs {
		dm, _ := d.(map[string]any)
		if dm["deviceID"] != deviceID {
			continue
		}
		igns, _ := dm["ignoredFolders"].([]any)
		for _, e := range igns {
			em, _ := e.(map[string]any)
			if em["id"] == folderID {
				return true
			}
		}
	}
	return false
}

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
	peerID := flag.String("peer-id", "", "valid device ID for add/remove tests")
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
	// Folder lifecycle: create, patch, ignores, delete.
	tmp, _ := os.MkdirTemp("", "livecheck-folder")
	defer os.RemoveAll(tmp)
	check("POST /rest/config/folders", c.PostFolder(map[string]any{
		"id": "lcheck-test1", "label": "livecheck", "path": tmp,
	}))
	check("PATCH /rest/config/folders/{id}", c.PatchFolder("lcheck-test1", map[string]any{"label": "livecheck2"}))
	f, err := c.FolderByID("lcheck-test1")
	check("GET /rest/config/folders/{id}", err)
	if f.Label != "livecheck2" {
		fmt.Println("FAIL folder patch not reflected")
		os.Exit(1)
	}
	fmt.Println("ok   folder patch reflected")
	check("POST /rest/db/ignores", c.SetIgnores("lcheck-test1", []string{"*.tmp"}))
	ign, err := c.Ignores("lcheck-test1")
	check("GET /rest/db/ignores", err)
	if len(ign) != 1 || ign[0] != "*.tmp" {
		fmt.Println("FAIL ignores round-trip")
		os.Exit(1)
	}
	fmt.Println("ok   ignores round-trip")
	check("DELETE /rest/config/folders/{id}", c.DeleteFolder("lcheck-test1"))

	// Device lifecycle: add, patch, delete (needs a valid, check-digit-passing ID).
	if *peerID == "" {
		fmt.Println("skip device lifecycle (no -peer-id)")
	} else {
		check("POST /rest/config/devices", c.PostDevice(map[string]any{
			"deviceID": *peerID, "name": "livecheck-dev",
		}))
		check("PATCH /rest/config/devices/{id}", c.PatchDevice(*peerID, map[string]any{"name": "livecheck-dev2"}))
		d, err := c.DeviceByID(*peerID)
		check("GET /rest/config/devices/{id}", err)
		if d.Name != "livecheck-dev2" {
			fmt.Println("FAIL device patch not reflected")
			os.Exit(1)
		}
		fmt.Println("ok   device patch reflected")

		// Share-folder-with-device (the alert "Share" action): create a
		// folder, share it with the peer, verify the device list.
		tmp2, _ := os.MkdirTemp("", "livecheck-share")
		defer os.RemoveAll(tmp2)
		check("POST folder for share test", c.PostFolder(map[string]any{
			"id": "lcheck-share1", "label": "sharetest", "path": tmp2,
		}))
		check("ShareFolderWithDevice", c.ShareFolderWithDevice("lcheck-share1", *peerID))
		f2, _ := c.FolderByID("lcheck-share1")
		found := false
		for _, d := range f2.Devices {
			found = found || d.DeviceID == *peerID
		}
		if !found {
			fmt.Println("FAIL shared device not in folder device list")
			os.Exit(1)
		}
		fmt.Println("ok   share reflected in folder devices")

		// Ignore-pending-folder (the alert "Ignore" action on a folder offer).
		check("IgnorePendingFolder", c.IgnorePendingFolder(*peerID, "some-offer", "Offered"))
		raw, _ := c.RawConfig()
		if !rawDeviceHasIgnoredFolder(raw, *peerID, "some-offer") {
			fmt.Println("FAIL ignoredFolders entry missing")
			os.Exit(1)
		}
		fmt.Println("ok   ignore folder reflected in device config")

		check("DELETE folder for share test", c.DeleteFolder("lcheck-share1"))
		check("DELETE /rest/config/devices/{id}", c.DeleteDevice(*peerID))

		// Ignore-pending-device (the alert "Ignore" action on a device).
		check("IgnorePendingDevice", c.IgnorePendingDevice(*peerID, "ignored-dev", ""))
		raw, _ = c.RawConfig()
		igns, _ := raw["remoteIgnoredDevices"].([]any)
		found = false
		for _, e := range igns {
			em, _ := e.(map[string]any)
			found = found || em["deviceID"] == *peerID
		}
		if !found {
			fmt.Println("FAIL remoteIgnoredDevices entry missing")
			os.Exit(1)
		}
		fmt.Println("ok   ignore device reflected in config")
	}

	fmt.Println("ALL OK")
}
