# syncthingtui

Go/bubbletea TUI aiming for near-parity with the syncthing web GUI.
Parity spec: GUI.yaml contains hierarchal description of TUI elements (keep up to date after changes).

# Structure

- client/       minimal syncthing REST client + config.xml discovery (TLS-aware)
- app/          the application
  - app.go        root router (main view ⇄ one subview), static screen presets
  - main_view.go  5 tabs: Folders, This Device, Remote Devices, Alerts, Actions
  - subviews.go   Edit Folder/Device, Settings (FormView = infobar + sub-tabs)
  - form.go       custom browse-then-edit form (jk browse, enter edit; NOT huh)
  - live.go       3s polling → dataMsg; fake data in data.go used when client nil
  - small_views.go Show ID (QR), About
- cmd/syncthingtui  entry point; cmd/mockup + mockups/ = frozen stage-2 mockups
- scripts/      test_live.sh, livecheck, rocheck (see Tests)

# Running

- go run ./cmd/syncthingtui              # live; discovers local syncthing config.xml
- go run ./cmd/syncthingtui -fake        # fake data, no syncthing needed
- go run ./cmd/syncthingtui -screen <name> [-size 80x60]   # static render (-list) for layout validation

# Tests

- ./scripts/test_live.sh   integration smoke: throwaway syncthing in /tmp
  (never the user's real one) + scripts/livecheck: all endpoints, pause and
  error-clear round-trips. Local syncthing is v1.18 (old CLI; script handles both).
- go run ./scripts/rocheck   read-only GETs against the user's real instance.
- Width check: render every -screen at 80x60, strip ANSI, assert ≤80 cols.
- No go tests yet; when adding features, add an assert to livecheck per action.

# Conventions

- Min terminal 80x60; every screen must render ≤80 cols (check before done).
- User-approved UX: tab/⇧tab + number keys switch tabs; esc/q/backspace = back
  (only esc while editing); Edit is first in button rows; readonly fields dim.
- Keep rejected design variants; record decisions in WORK.md, not code comments.
- Design changes need user feedback: mock up variants, give exact run commands.

# Syncing with Upstream

**Last synced against: syncthing v2.0.13** — update this line after each parity
check. Procedure: (1) diff /rest/config keys against the `.K("key")` mappings
on form fields in app/subviews.go (Settings still uses `// key` comments); (2) diff ../syncthing/gui/default/
since the last-synced version, using GUI.yaml as the checklist; (3) update GUI.yaml,
the TUI, and this line.
