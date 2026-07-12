#!/bin/bash
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

echo "== syncthing up on :$PORT (home $DIR) =="
cd "$(dirname "$0")/.."
go run ./scripts/livecheck -address "127.0.0.1:$PORT" -api-key testkey123
