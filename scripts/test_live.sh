#!/bin/bash
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

# Integration smoke test: start a throwaway syncthing (temp home, random GUI
# port), run client/app checks against it, tear it down. Never touches the
# user's real syncthing.
set -u
DIR=$(mktemp -d /tmp/sttui-test.XXXXXX)
PORT=$((20000 + RANDOM % 20000))
cleanup() { kill "$ST_PID" 2>/dev/null; wait "$ST_PID" 2>/dev/null; rm -rf "$DIR"; }
trap cleanup EXIT

syncthing generate --home="$DIR" >/dev/null 2>&1 || syncthing -generate="$DIR" >/dev/null 2>&1
# force a known API key and GUI address
sed -i "s|<apikey>.*</apikey>|<apikey>testkey123</apikey>|" "$DIR/config.xml"
sed -i "s|<address>127.0.0.1:[0-9]*</address>|<address>127.0.0.1:$PORT</address>|" "$DIR/config.xml"

# new CLI (syncthing serve ...) vs old CLI (syncthing -home ...)
if syncthing serve --help >/dev/null 2>&1; then
  syncthing serve --home="$DIR" --no-browser --no-restart >/dev/null 2>&1 &
else
  syncthing -home="$DIR" -no-browser -no-restart >/dev/null 2>&1 &
fi
ST_PID=$!

for i in $(seq 1 30); do
  curl -s -H "X-API-Key: testkey123" "http://127.0.0.1:$PORT/rest/system/status" >/dev/null && break
  sleep 0.5
done

# second identity provides a valid device ID for add/remove tests
DIR2=$(mktemp -d /tmp/sttui-test2.XXXXXX)
trap 'cleanup; rm -rf "$DIR2"' EXIT
GEN2=$(syncthing generate --home="$DIR2" 2>&1 || syncthing -generate="$DIR2" 2>&1)
PEER_ID=$(echo "$GEN2" | grep -o '[A-Z2-7]\{7\}\(-[A-Z2-7]\{7\}\)\{7\}' | head -1)

echo "== syncthing up on :$PORT (home $DIR, peer $PEER_ID) =="
cd "$(dirname "$0")/.."
go run ./scripts/livecheck -address "127.0.0.1:$PORT" -api-key testkey123 -peer-id "$PEER_ID"
