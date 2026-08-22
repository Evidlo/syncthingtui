# syncthingtui - a terminal user interface for Syncthing
# Copyright (C) 2026 Evan Widloski
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.
#
# SPDX-License-Identifier: GPL-3.0-or-later

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X github.com/evidlo/syncthingtui/app.TUIVersion=$(VERSION)

.PHONY: syncthingtui clean release

syncthingtui:
	go build -ldflags '$(LDFLAGS)' ./cmd/syncthingtui

clean:
	rm -f syncthingtui

release: syncthingtui
	gh release create $(VERSION) syncthingtui --generate-notes
