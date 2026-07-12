# WORK.md

## Status

Stage 2 (mockups). Go module + bubbletea/lipgloss/bubbles.
Render any screen statically: `go run ./cmd/mockup <name> [WxH]` (default 80x60, `-list` to enumerate).
Screens: `mockups/screens.go`, fake data: `mockups/data.go`. Keep all screens ≤80 cols.

Feedback protocol: present mockup sets with exact commands, pause for user feedback between sets.
Stage 4 in progress. Done: client/ package (REST subset + config.xml Discover);
app live mode (App polls every 3s → dataMsg; rates from byte-total deltas;
MainModel.action() runs against client or fake fallback). Wired: pause/resume
folder+device, rescan one/all, alert Dismiss, live folders/devices/stats/alerts.
CLI: syncthingtui [-fake | -address X -api-key Y] (default: Discover from
~/.local/state|.config/syncthing/config.xml; honors gui tls attr → https,
cert verification skipped — syncthing GUI certs are self-signed/expired,
API key is the auth). scripts/rocheck = read-only GET check vs real instance. Integration smoke test:
`./scripts/test_live.sh` — throwaway syncthing home in /tmp, random port, runs
scripts/livecheck (all endpoints + pause round-trip). Local syncthing is
v1.18 (old single-dash CLI; script handles both).

Notices: all system errors (/rest/system/error) combine into ONE "Notice"
alert (REST can only clear all at once; per-error clear impossible), OK =
/rest/system/error/clear. livecheck covers post-error→visible→clear→empty.
About view: Actions > About — syncthingtui version (app.TUIVersion) +
syncthing build info on top, paths below (/rest/system/paths; 404 on <v1.19,
shown as unavailable). Tests so far: scripts/test_live.sh + livecheck
(integration smoke vs throwaway instance), static width checks; no go tests
yet (stage-4 item 5). GUI.yaml synced with current app 2026-07-12 (TUI
notes on main-view buttons; Restore Versions modal documented, TUI TODO).

Form Save/Remove wired for folders + devices: editors load live config on open
(openFolderEditor/openDeviceEditor cmds), fields carry functional config keys
(FormField.K/.Vals/.Num — replaces comment-only annotations), Save PATCHes
(or POSTs when new; folder ignores via /rest/db/ignores), Remove DELETEs,
close triggers refetch. Alert "Add Device" now truly accepts the pending
device (opens Add Device prefilled with its ID; Save POSTs it). livecheck
covers folder+device lifecycle (create/patch/verify/delete; needs valid
peer ID — test_live.sh generates a 2nd identity).

Remaining stage-4 work:
1. Settings Save (options/gui PATCH; fiddly: minHomeDiskFree object, urAccepted ints, upgrade split key) + Settings load live values
2. Alert Share Folder / Ignore actions (share = PATCH folder devices[]; ignore lists remoteIgnoredDevices / folder ignoredFolders)
3. /rest/events long-poll instead of 3s polling
4. Actions tab: Show ID uses real ID+QR, restart/shutdown/logs endpoints
5. Proper Go integration tests (port scripts/livecheck into go test with harness)
6. Restore Versions view (per-folder Versions button; see GUI.yaml) — design + wire

Stage-3 walkthrough approved by user 2026-07-12. Stage-3 app details:
- Interactive app: `go run ./cmd/syncthingtui` (q quits, altscreen)
- Static review: `go run ./cmd/syncthingtui -screen <name> [-size WxH]`, `-list` for names
- `app/` package: App router (app.go) → MainModel tabs (main_view.go) → subviews
  (subviews.go: huh forms w/ FormView chrome; small_views.go: ShowID, About)
- Forms use huh (ThemeBase + accent tweaks); readonly = dim huh Note
- Known gaps/caveats to review with user: sub-tab number hotkeys disabled in form
  subviews (digits type into fields; tab/⇧tab switches sections); Save/Remove
  buttons not wired (esc closes, fake data static); alert "Add Device" opens Add
  Device form; Actions>Advanced temporarily opens About/Paths
Subview keys: tab/⇧tab = next/prev section (both hints kept alongside ¹²³⁴ hotkeys).
TODO: revisit form-field styling later — user dislikes current ▕value▏ pipe-surround look (placeholder until a field-style set).
Form elements must visually distinguish editable vs readonly fields (web GUI greys out non-editable); mockups grey out readonly fields entirely (formField readonly flag).
Keep rejected mockup variants in the registry (user request) — don't delete after a decision.

## Open questions for user

(none currently)

## Decisions

- Upstream parity: no tooling (deemed brittle). Form fields in app/subviews.go
  carry trailing comments naming their /rest/config keys; AGENTS.md holds the
  last-synced syncthing version + the parity-check procedure. Original project
  brief lives in PROMPT.md (user renamed it from AGENTS.md).

- Alerts TUI analog: fifth tab, badge when non-empty — gray count, space: "Alerts (2)"
- Stack: Go + bubbletea (bubblon for nesting)
- Forms: custom browse-then-edit layer in app/form.go (huh dropped — always-live
  fields conflict with jk browsing; wizard StateCompleted blanks persistent forms;
  multi-select can't embed editable values). ↑↓/jk browse, enter edits/toggles,
  enter commits (ctrl+s in textarea), esc cancels edit then backs out. Field kinds:
  Text/Pass/RO/Sel/Bool/Area; readonly renders dim "(readonly)"
- Min terminal size: 80×60
- Tab bar: boxed active tab + superscript number hotkeys (from ../syncthingtui_old)
- "Add Folder" / "Rescan All" are list items (+ / ⟳ icons, blank line above), not key shortcuts
- Tab switching everywhere (main + subviews): number keys and tab/⇧tab only (no ←→); hint "tab/⇧tab:switch tabs"
- esc/q/backspace synonymous everywhere for quit/back (except while editing a field, where they belong to the editor and only esc cancels); hints say "esc:quit" (main) / "esc:cancel/back" (subviews)
- Folder/device detail panes end with button rows (←→ select, enter confirm): Edit always first — folders [Edit, Pause|Resume, Rescan], devices [Edit, Pause|Resume]
- Edit Folder has a Sharing tab: checkbox per known device (GUI.yaml updated — only per-device encryption passwords remain do-not-implement)
- Status bar: key hints win over net-rates display when width is tight (rates dropped)
- Folders layout: **folders-v3-split** (left list / right detail) is the baseline
- Alerts layout: **alerts-v3-split**; left-panel entries as "Device (name)" / "Folder (label)"
- Remote Devices: **devices-split** (mirrors folders baseline, no variants)
- This Device: **thisdevice-v1-kv**; Actions: **actions-v1-list**
- Show ID includes a QR code (user request; GUI.yaml updated) — half-block render, real impl e.g. skip2/go-qrcode
- Forms: current field rendering is throwaway placeholder; real forms via huh (charmbracelet), restyled at stage 3
- Subviews: **editfolder-v1-infobar** — info bar (title left, boxed "Backᵉˢᶜ" right) replaces tab bar; boxed-style sub-tabs with ¹²³⁴ hotkeys; keep tab:next-section hint alongside number keys
- WORK.md must stay concise

## Done

- Added `alerts:` section to GUI.yaml from `../syncthing/gui/default/index.html`
  (REST `/rest/cluster/pending/{devices,folders}`, events `Pending*Changed`)
- Mockup harness + folder/alert variants built
- Reviewed ../syncthingtui_old: reuse its 3-layer AppModel→View→Widget pattern
  and WindowSizeMsg broadcast for the real implementation (see its CLAUDE.md)
