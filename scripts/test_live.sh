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

# second identity provides a valid device ID for add/remove tests
DIR2=$(mktemp -d /tmp/sttui-test2.XXXXXX)
trap 'cleanup; rm -rf "$DIR2"' EXIT
GEN2=$(syncthing generate --home="$DIR2" 2>&1 || syncthing -generate="$DIR2" 2>&1)
PEER_ID=$(echo "$GEN2" | grep -o '[A-Z2-7]\{7\}\(-[A-Z2-7]\{7\}\)\{7\}' | head -1)

echo "== syncthing up on :$PORT (home $DIR, peer $PEER_ID) =="
cd "$(dirname "$0")/.."
go run ./scripts/livecheck -address "127.0.0.1:$PORT" -api-key testkey123 -peer-id "$PEER_ID"
