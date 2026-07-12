package app

import (
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evidlo/syncthingtui/client"
)

// closeMsg asks the root App to pop the current subview.
type closeMsg struct{ flash string }

// App routes between the main tabbed view and one subview at a time.
type App struct {
	main            MainModel
	sub             tea.Model
	width, height   int
	client          *client.Client
	prevIn, prevOut int64
	prevAt          time.Time
}

// New builds a fake-data app (used by static screen presets).
func New() App { return App{main: NewMain(nil)} }

// NewLive builds an app backed by a syncthing instance.
func NewLive(c *client.Client) App {
	return App{main: NewMain(c), client: c}
}

func (a App) Init() tea.Cmd {
	if a.client != nil {
		return tea.Batch(a.main.Init(), fetchCmd(a.client))
	}
	return a.main.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		m, _ := a.main.Update(msg)
		a.main = m.(MainModel)
		if a.sub != nil {
			a.sub, _ = a.sub.Update(msg)
		}
		return a, nil
	case openMsg:
		a.sub = msg.view
		sub, cmd := a.sub.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
		a.sub = sub
		return a, tea.Batch(a.sub.Init(), cmd)
	case closeMsg:
		a.sub = nil
		a.main.flash = msg.flash
		if a.client != nil { // reflect any saved changes immediately
			return a, fetchCmd(a.client)
		}
		return a, nil
	case dataMsg:
		if msg.err != nil {
			a.main.flash = "syncthing: " + msg.err.Error()
			return a, pollCmd()
		}
		now := time.Now()
		if !a.prevAt.IsZero() {
			dt := now.Sub(a.prevAt).Seconds()
			down := humanRate(float64(msg.inTotal-a.prevIn) / dt)
			up := humanRate(float64(msg.outTotal-a.prevOut) / dt)
			a.main.rates = "↓" + down + " ↑" + up
			msg.stats[0][1] = fmt.Sprintf("%s (%s total)", down, humanBytes(msg.inTotal))
			msg.stats[1][1] = fmt.Sprintf("%s (%s total)", up, humanBytes(msg.outTotal))
		}
		a.prevIn, a.prevOut, a.prevAt = msg.inTotal, msg.outTotal, now
		a.main.setData(msg)
		return a, pollCmd()
	case pollTickMsg:
		return a, fetchCmd(a.client)
	case tea.KeyMsg:
		if a.sub != nil {
			switch msg.String() {
			case "esc", "q", "backspace":
				// while editing a field, these keys belong to the editor
				// (esc cancels the edit, q/backspace type/delete)
				if ed, ok := a.sub.(interface{ Editing() bool }); ok && ed.Editing() {
					break
				}
				a.sub = nil
				return a, nil
			case "ctrl+c":
				return a, tea.Quit
			}
			var cmd tea.Cmd
			a.sub, cmd = a.sub.Update(msg)
			return a, cmd
		}
	}
	if a.sub != nil {
		var cmd tea.Cmd
		a.sub, cmd = a.sub.Update(msg)
		return a, cmd
	}
	m, cmd := a.main.Update(msg)
	if mm, ok := m.(MainModel); ok {
		a.main = mm
	}
	return a, cmd
}

func (a App) View() string {
	if a.sub != nil {
		return a.sub.View()
	}
	return a.main.View()
}

// ── Static rendering (layout review without a live terminal) ────────────────

var presets = map[string]func() App{
	"main-folders":    func() App { return New() },
	"main-thisdevice": func() App { a := New(); a.main.active = tabThisDevice; return a },
	"main-devices":    func() App { a := New(); a.main.active = tabDevices; return a },
	"main-alerts":     func() App { a := New(); a.main.active = tabAlerts; return a },
	"main-actions":    func() App { a := New(); a.main.active = tabActions; return a },
	"edit-folder":     func() App { a := New(); a.sub = NewEditFolder("Photos", false); return a },
	"add-folder":      func() App { a := New(); a.sub = NewEditFolder("", true); return a },
	"edit-device":     func() App { a := New(); a.sub = NewEditDevice("nas", false); return a },
	"settings":        func() App { a := New(); a.sub = NewSettings(); return a },
	"show-id":         func() App { a := New(); a.sub = NewShowID(DeviceID); return a },
	"about":           func() App { a := New(); a.sub = NewAbout("v2.0.13, Linux (64-bit)", AboutPaths); return a },
}

func ScreenNames() []string {
	names := make([]string, 0, len(presets))
	for n := range presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// RenderScreen statically renders a named app state at the given size.
func RenderScreen(name string, width, height int) (string, error) {
	preset, ok := presets[name]
	if !ok {
		return "", fmt.Errorf("unknown screen %q", name)
	}
	a := preset()
	a.Init()
	if a.sub != nil {
		a.sub.Init()
	}
	m, _ := a.Update(tea.WindowSizeMsg{Width: width, Height: height})
	a = m.(App)
	return a.View(), nil
}
