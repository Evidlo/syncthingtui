This project will be a syncthing TUI tool that will have (close to) complete parity with the official syncthing web GUI.

@GUI.yaml contains a hierarchal description of syncthing web GUI menus/views/forms/labels to be used to inform the TUI structure.

Latest syncthing source can be found in ../syncthing if needed.

Agents are bad at evaluating the friendliness of TUI design, so a human will be involved in the final selection process of GUI→TUI mappings from a set of agent-generated mockups of:
- low level design items: forms/buttons/fields/radios/etc
- high level design items: tabs/views/modals

Maintain a concise WORK.md as you work for consumption by future agents in new sessions.  Log questions to user and answers there concisely.

GUI.yaml is lacking a description of the alerts box used to accept new devices/folders.  This needs to be added to GUI.yaml and a suitable TUI analog decided upon.

# Work Stages

1. research and context gathering for the agent, conceptual discussions with user
2. mockup design phase - agent builds multiple versions of TUI mockups and collects feedback from user
3. complete TUI mockup with fake data, final feedback from user
4. wiring TUI to syncthing REST endpoints

# Design Requirements and Miscellaneous

- static TUI rendering - agent should be able to review any screen in the TUI for correct layout in a virtual terminal (e.g. 80x60 window) to avoid user being involved in fixing layout bugs
- syncthing integration tests - actions should be tested against syncthing REST interface for correct behavior
- straightforward updating of code to follow syncthing changes - code will need to follow upstream as syncthing adds/removes features, we should make it easy to validate/update the TUI to maintain syncthing parity (we can discuss options more)
- consider using bubblon for nested charmbracelet views - https://donderom.com/posts/managing-nested-models-with-bubble-tea/ (unless there is a more canonical way to do nesting)
