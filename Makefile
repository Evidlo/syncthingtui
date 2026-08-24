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

# --- Debian packaging -------------------------------------------------------
# One-shot release. Commit your upstream work on master, then from master:
#
#     make debian V=1.0.2
#
# Idempotent: every step checks whether it already happened, so re-running
# after a failure resumes instead of duplicating work.
#
# This drives the whole release from master; it checks out the packaging
# branch internally and always returns you to master, even on failure. The
# packaging branch receives upstream sources only through uscan + gbp here -
# never merge master into it by hand, that breaks the gbp layout.

DEBEMAIL ?= evan_debian@widloski.com
DEBFULLNAME ?= Evan Widloski
DEB_BUILD_DIR ?= ../build-area
DEB_BRANCH ?= debian
# origin has two push URLs (GitHub + Salsa), so one push reaches both.
DEB_REMOTES ?= origin

.PHONY: debian debian-check debian-tag debian-package
.NOTPARALLEL:

debian: debian-check debian-tag debian-package
	@echo "released $(V); packages in $(DEB_BUILD_DIR)"

debian-check:
	@test -n "$(V)" || { echo "usage: make debian V=1.0.2"; exit 1; }
	@test "$$(git branch --show-current)" = master \
	  || { echo "must be on master"; exit 1; }
	@git diff --quiet && git diff --cached --quiet \
	  || { echo "working tree dirty - commit first"; exit 1; }

# Tag master and publish, so uscan can fetch the tag over the network below.
debian-tag:
	@git rev-parse -q --verify refs/tags/v$(V) >/dev/null \
	  || git tag -a v$(V) -m "syncthingtui $(V)"
	@for r in $(DEB_REMOTES); do git push $$r master v$(V); done

# Everything that must happen on the packaging branch, in one shell so the
# trap can guarantee we end up back on master.
debian-package:
	@set -e; \
	git checkout -q $(DEB_BRANCH); \
	trap 'git checkout -q master' EXIT; \
	if git rev-parse -q --verify refs/tags/upstream/$(V) >/dev/null; then \
	  echo "upstream/$(V) already imported"; \
	else \
	  gbp import-orig --uscan --no-interactive; \
	fi; \
	case "$$(dpkg-parsechangelog -SVersion)" in \
	  $(V)-*) echo "changelog already at $(V)" ;; \
	  *) DEBEMAIL="$(DEBEMAIL)" DEBFULLNAME="$(DEBFULLNAME)" \
	       dch -v $(V)-1 -D unstable "New upstream release."; \
	     git commit -q -m "debian: new upstream release $(V)" debian/changelog ;; \
	esac; \
	gbp buildpackage --git-export-dir=$(DEB_BUILD_DIR) -us -uc; \
	lintian -I --pedantic $(DEB_BUILD_DIR)/syncthingtui_$(V)-*.changes; \
	for r in $(DEB_REMOTES); do \
	  git push $$r master $(DEB_BRANCH) upstream/latest v$(V) upstream/$(V); \
	done
