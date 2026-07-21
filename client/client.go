// Package client is a minimal syncthing REST API client covering the
// endpoints the TUI needs.
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	base string
	key  string
	http *http.Client
}

// New accepts "host:port" (http) or a full "http(s)://host:port" base URL.
// Certificate verification is skipped: syncthing's GUI cert is self-signed
// (and often expired); the API key is the effective authentication.
func New(address, apiKey string) *Client {
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}
	return &Client{
		base: address,
		key:  apiKey,
		http: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (c *Client) do(method, path string, body, into any) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, c.base+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s %s: %s", method, path, resp.Status)
	}
	if into != nil {
		return json.NewDecoder(resp.Body).Decode(into)
	}
	return nil
}

func (c *Client) get(path string, into any) error { return c.do("GET", path, nil, into) }

// ── Types (subset of the syncthing REST schema) ──────────────────────────────

type FolderDevice struct {
	DeviceID string `json:"deviceID"`
}

type VersioningCfg struct {
	Type             string            `json:"type"`
	Params           map[string]string `json:"params"`
	CleanupIntervalS int               `json:"cleanupIntervalS"`
}

type FolderCfg struct {
	ID               string         `json:"id"`
	Label            string         `json:"label"`
	Path             string         `json:"path"`
	Type             string         `json:"type"`
	Paused           bool           `json:"paused"`
	Devices          []FolderDevice `json:"devices"`
	Order            string         `json:"order"`
	RescanIntervalS  int            `json:"rescanIntervalS"`
	FSWatcherEnabled bool           `json:"fsWatcherEnabled"`
	IgnorePerms      bool           `json:"ignorePerms"`
	Versioning       VersioningCfg  `json:"versioning"`
}

type DeviceCfg struct {
	DeviceID          string   `json:"deviceID"`
	Name              string   `json:"name"`
	Addresses         []string `json:"addresses"`
	Paused            bool     `json:"paused"`
	Compression       string   `json:"compression"`
	Introducer        bool     `json:"introducer"`
	AutoAcceptFolders bool     `json:"autoAcceptFolders"`
	NumConnections    int      `json:"numConnections"`
	MaxRecvKbps       int      `json:"maxRecvKbps"`
	MaxSendKbps       int      `json:"maxSendKbps"`
	Untrusted         bool     `json:"untrusted"`
}

type Config struct {
	Folders []FolderCfg `json:"folders"`
	Devices []DeviceCfg `json:"devices"`
}

type ServiceStatus struct {
	Error *string `json:"error"`
}

type SystemStatus struct {
	MyID                    string                   `json:"myID"`
	Uptime                  int64                    `json:"uptime"`
	DiscoveryStatus         map[string]ServiceStatus `json:"discoveryStatus"`
	ConnectionServiceStatus map[string]ServiceStatus `json:"connectionServiceStatus"`
}

type Connections struct {
	Total struct {
		InBytesTotal  int64 `json:"inBytesTotal"`
		OutBytesTotal int64 `json:"outBytesTotal"`
	} `json:"total"`
	Connections map[string]struct {
		Connected bool   `json:"connected"`
		Address   string `json:"address"`
	} `json:"connections"`
}

type DBStatus struct {
	State                 string `json:"state"`
	GlobalBytes           int64  `json:"globalBytes"`
	LocalBytes            int64  `json:"localBytes"`
	NeedBytes             int64  `json:"needBytes"`
	NeedTotalItems        int64  `json:"needTotalItems"`
	ReceiveOnlyTotalItems int64  `json:"receiveOnlyTotalItems"`
	PullErrors            int64  `json:"pullErrors"`
	LocalFiles            int64  `json:"localFiles"`
}

type SizeCfg struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type OptionsCfg struct {
	ListenAddresses       []string `json:"listenAddresses"`
	GlobalAnnounceServers []string `json:"globalAnnounceServers"`
	GlobalAnnounceEnabled bool     `json:"globalAnnounceEnabled"`
	LocalAnnounceEnabled  bool     `json:"localAnnounceEnabled"`
	RelaysEnabled         bool     `json:"relaysEnabled"`
	NATEnabled            bool     `json:"natEnabled"`
	MaxRecvKbps           int      `json:"maxRecvKbps"`
	MaxSendKbps           int      `json:"maxSendKbps"`
	LimitBandwidthInLan   bool     `json:"limitBandwidthInLan"`
	StartBrowser          bool     `json:"startBrowser"`
	URAccepted            int      `json:"urAccepted"`
	AutoUpgradeIntervalH  int      `json:"autoUpgradeIntervalH"`
	UpgradeToPreReleases  bool     `json:"upgradeToPreReleases"`
	MinHomeDiskFree       SizeCfg  `json:"minHomeDiskFree"`
}

type GUICfg struct {
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
	UseTLS   bool   `json:"useTLS"`
	Theme    string `json:"theme"`
	APIKey   string `json:"apiKey"`
}

type Version struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

type PendingDevice struct {
	Time    time.Time `json:"time"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
}

type PendingFolderOffer struct {
	Time  time.Time `json:"time"`
	Label string    `json:"label"`
}

type PendingFolder struct {
	OfferedBy map[string]PendingFolderOffer `json:"offeredBy"`
}

type DeviceStats struct {
	LastSeen time.Time `json:"lastSeen"`
}

type SystemError struct {
	When    time.Time `json:"when"`
	Message string    `json:"message"`
}

type Completion struct {
	Completion float64 `json:"completion"`
}

// ── Endpoints ────────────────────────────────────────────────────────────────

func (c *Client) Config() (Config, error) {
	var v Config
	return v, c.get("/rest/config", &v)
}

func (c *Client) SystemStatus() (SystemStatus, error) {
	var v SystemStatus
	return v, c.get("/rest/system/status", &v)
}

func (c *Client) Connections() (Connections, error) {
	var v Connections
	return v, c.get("/rest/system/connections", &v)
}

func (c *Client) DBStatus(folder string) (DBStatus, error) {
	var v DBStatus
	return v, c.get("/rest/db/status?folder="+folder, &v)
}

func (c *Client) Version() (Version, error) {
	var v Version
	return v, c.get("/rest/system/version", &v)
}

func (c *Client) PendingDevices() (map[string]PendingDevice, error) {
	var v map[string]PendingDevice
	return v, c.get("/rest/cluster/pending/devices", &v)
}

func (c *Client) PendingFolders() (map[string]PendingFolder, error) {
	var v map[string]PendingFolder
	return v, c.get("/rest/cluster/pending/folders", &v)
}

func (c *Client) DeviceStats() (map[string]DeviceStats, error) {
	var v map[string]DeviceStats
	return v, c.get("/rest/stats/device", &v)
}

func (c *Client) DeviceCompletion(deviceID string) (Completion, error) {
	var v Completion
	return v, c.get("/rest/db/completion?device="+deviceID, &v)
}

func (c *Client) Rescan(folder string) error {
	path := "/rest/db/scan"
	if folder != "" {
		path += "?folder=" + folder
	}
	return c.do("POST", path, nil, nil)
}

func (c *Client) FolderByID(id string) (FolderCfg, error) {
	var v FolderCfg
	return v, c.get("/rest/config/folders/"+id, &v)
}

func (c *Client) DeviceByID(id string) (DeviceCfg, error) {
	var v DeviceCfg
	return v, c.get("/rest/config/devices/"+id, &v)
}

// PostFolder adds a folder (or replaces one with the same ID); accepts a
// FolderCfg or a map following the folder config schema.
func (c *Client) PostFolder(f any) error {
	return c.do("POST", "/rest/config/folders", f, nil)
}

// PatchFolder applies a partial update; keys follow the folder config schema.
func (c *Client) PatchFolder(id string, patch map[string]any) error {
	return c.do("PATCH", "/rest/config/folders/"+id, patch, nil)
}

func (c *Client) DeleteFolder(id string) error {
	return c.do("DELETE", "/rest/config/folders/"+id, nil, nil)
}

// PostDevice adds a device; accepts a DeviceCfg or a schema-shaped map.
func (c *Client) PostDevice(d any) error {
	return c.do("POST", "/rest/config/devices", d, nil)
}

func (c *Client) PatchDevice(id string, patch map[string]any) error {
	return c.do("PATCH", "/rest/config/devices/"+id, patch, nil)
}

func (c *Client) DeleteDevice(id string) error {
	return c.do("DELETE", "/rest/config/devices/"+id, nil, nil)
}

func (c *Client) Ignores(folder string) ([]string, error) {
	var v struct {
		Ignore []string `json:"ignore"`
	}
	return v.Ignore, c.get("/rest/db/ignores?folder="+folder, &v)
}

func (c *Client) SetIgnores(folder string, lines []string) error {
	return c.do("POST", "/rest/db/ignores?folder="+folder,
		map[string][]string{"ignore": lines}, nil)
}

// ShareFolderWithDevice adds deviceID to an existing folder's device list
// (accepting a pending "share folder" offer).
func (c *Client) ShareFolderWithDevice(folderID, deviceID string) error {
	f, err := c.FolderByID(folderID)
	if err != nil {
		return err
	}
	for _, d := range f.Devices {
		if d.DeviceID == deviceID {
			return nil
		}
	}
	return c.PatchFolder(folderID, map[string]any{
		"devices": append(f.Devices, FolderDevice{DeviceID: deviceID}),
	})
}

// RawConfig fetches the whole config as a generic map (preserves keys the
// typed Config struct doesn't model, for safe read-modify-PUT updates).
func (c *Client) RawConfig() (map[string]any, error) {
	var v map[string]any
	return v, c.get("/rest/config", &v)
}

func (c *Client) PutRawConfig(cfg map[string]any) error {
	return c.do("PUT", "/rest/config", cfg, nil)
}

// IgnorePendingDevice permanently ignores a pending device
// (config.remoteIgnoredDevices), suppressing further notifications.
func (c *Client) IgnorePendingDevice(deviceID, name, address string) error {
	cfg, err := c.RawConfig()
	if err != nil {
		return err
	}
	arr, _ := cfg["remoteIgnoredDevices"].([]any)
	cfg["remoteIgnoredDevices"] = append(arr, map[string]any{
		"deviceID": deviceID, "name": name, "address": address,
		"time": time.Now().Format(time.RFC3339),
	})
	return c.PutRawConfig(cfg)
}

// IgnorePendingFolder permanently ignores a folder offered by deviceID
// (that device's ignoredFolders list).
func (c *Client) IgnorePendingFolder(deviceID, folderID, label string) error {
	cfg, err := c.RawConfig()
	if err != nil {
		return err
	}
	devs, _ := cfg["devices"].([]any)
	for _, d := range devs {
		dm, ok := d.(map[string]any)
		if !ok || dm["deviceID"] != deviceID {
			continue
		}
		arr, _ := dm["ignoredFolders"].([]any)
		dm["ignoredFolders"] = append(arr, map[string]any{
			"id": folderID, "label": label,
			"time": time.Now().Format(time.RFC3339),
		})
	}
	return c.PutRawConfig(cfg)
}

func (c *Client) SetFolderPaused(id string, paused bool) error {
	return c.do("PATCH", "/rest/config/folders/"+id, map[string]bool{"paused": paused}, nil)
}

func (c *Client) SetDevicePaused(id string, paused bool) error {
	return c.do("PATCH", "/rest/config/devices/"+id, map[string]bool{"paused": paused}, nil)
}

func (c *Client) Options() (OptionsCfg, error) {
	var v OptionsCfg
	return v, c.get("/rest/config/options", &v)
}

func (c *Client) GUIConfig() (GUICfg, error) {
	var v GUICfg
	return v, c.get("/rest/config/gui", &v)
}

func (c *Client) PatchOptions(patch map[string]any) error {
	return c.do("PATCH", "/rest/config/options", patch, nil)
}

func (c *Client) PatchGUI(patch map[string]any) error {
	return c.do("PATCH", "/rest/config/gui", patch, nil)
}

func (c *Client) Restart() error {
	return c.do("POST", "/rest/system/restart", nil, nil)
}

func (c *Client) Shutdown() error {
	return c.do("POST", "/rest/system/shutdown", nil, nil)
}

func (c *Client) Errors() ([]SystemError, error) {
	var v struct {
		Errors []SystemError `json:"errors"`
	}
	return v.Errors, c.get("/rest/system/error", &v)
}

func (c *Client) ClearErrors() error {
	return c.do("POST", "/rest/system/error/clear", nil, nil)
}

// PostError registers a system error (plain-text body; used by tests).
func (c *Client) PostError(msg string) error {
	req, err := http.NewRequest("POST", c.base+"/rest/system/error", strings.NewReader(msg))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.key)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("POST /rest/system/error: %s", resp.Status)
	}
	return nil
}

// Paths returns the filesystem paths syncthing is using (config, db, ...).
func (c *Client) Paths() (map[string]string, error) {
	var v map[string]string
	return v, c.get("/rest/system/paths", &v)
}

func (c *Client) DismissPendingDevice(id string) error {
	return c.do("DELETE", "/rest/cluster/pending/devices?device="+id, nil, nil)
}

func (c *Client) DismissPendingFolder(id string) error {
	return c.do("DELETE", "/rest/cluster/pending/folders?folder="+id, nil, nil)
}

// ── Local config discovery ───────────────────────────────────────────────────

type guiConfig struct {
	GUI struct {
		TLS     bool   `xml:"tls,attr"`
		Address string `xml:"address"`
		APIKey  string `xml:"apikey"`
	} `xml:"gui"`
}

// Discover reads address and API key from the local syncthing config.xml
// (new ~/.local/state and legacy ~/.config locations).
func Discover() (address, apiKey string, err error) {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".local/state/syncthing/config.xml"),
		filepath.Join(home, ".config/syncthing/config.xml"),
	}
	for _, p := range paths {
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			continue
		}
		var cfg guiConfig
		if xerr := xml.Unmarshal(data, &cfg); xerr != nil {
			continue
		}
		if cfg.GUI.APIKey != "" {
			scheme := "http://"
			if cfg.GUI.TLS {
				scheme = "https://"
			}
			return scheme + cfg.GUI.Address, cfg.GUI.APIKey, nil
		}
	}
	return "", "", fmt.Errorf("no syncthing config.xml with apikey found")
}
