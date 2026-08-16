# Syncthingtui
# Evan Widloski - 2026-08-16

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X github.com/evidlo/syncthingtui/app.TUIVersion=$(VERSION)

.PHONY: syncthingtui clean release

syncthingtui:
	go build -ldflags '$(LDFLAGS)' ./cmd/syncthingtui

clean:
	rm -f syncthingtui

release: syncthingtui
	gh release create $(VERSION) syncthingtui --generate-notes
