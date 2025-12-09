#!/usr/bin/env bash
set -euo pipefail

# Demo script that exercises main API endpoints against a running server.
# Requirements: curl, jq, python3 (for generating a small wav file).

SERVER="${SERVER:-http://localhost:8080}"
EMAIL="${EMAIL:-demo@example.com}"
PASSWORD="${PASSWORD:-secret123}"
USERNAME="${USERNAME:-demo_user}"
DISPLAY_NAME="${DISPLAY_NAME:-Demo User}"

OTHER_EMAIL="${OTHER_EMAIL:-friend@example.com}"
OTHER_PASSWORD="${OTHER_PASSWORD:-secret123}"
OTHER_USERNAME="${OTHER_USERNAME:-friend_user}"

info() { printf "\n=== %s ===\n" "$*"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1" >&2; exit 1; }
}

require_cmd curl
require_cmd jq
require_cmd python3

TMP_AUDIO="$(mktemp /tmp/podvibe_demo.XXXXXX.wav)"
python3 - <<'PY'
import math, struct, wave, tempfile, os
tmp = os.environ["TMP_AUDIO"]
framerate = 8000
duration = 1
frequency = 440.0
with wave.open(tmp, "w") as w:
    w.setnchannels(1)
    w.setsampwidth(2)
    w.setframerate(framerate)
    for i in range(int(duration * framerate)):
        value = int(32767 * 0.3 * math.sin(2 * math.pi * frequency * (i / framerate)))
        w.writeframes(struct.pack("<h", value))
PY

register_user() {
  email="$1"; username="$2"; password="$3"; display="$4"
  curl -s -o /dev/null -w "%{http_code}" -X POST "$SERVER/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"username\":\"$username\",\"password\":\"$password\",\"display_name\":\"$display\"}"
}

login_json() {
  email="$1"; password="$2"
  curl -s -X POST "$SERVER/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$password\"}"
}

info "Register primary user (ignore 409 if already exists)"
register_user "$EMAIL" "$USERNAME" "$PASSWORD" "$DISPLAY_NAME" || true
info "Register secondary user"
register_user "$OTHER_EMAIL" "$OTHER_USERNAME" "$OTHER_PASSWORD" "Friend" || true

info "Login and get access token"
LOGIN_JSON=$(login_json "$EMAIL" "$PASSWORD")
ACCESS_TOKEN=$(echo "$LOGIN_JSON" | jq -r '.access_token')
USER_ID=$(echo "$LOGIN_JSON" | jq -r '.user.id')
AUTH_HEADER=("Authorization: Bearer $ACCESS_TOKEN")
echo "TOKEN=${ACCESS_TOKEN:0:16}..."

info "Get current user"
curl -s -H "${AUTH_HEADER[@]}" "$SERVER/users/me" | jq .

info "Update profile"
curl -s -X PATCH "$SERVER/users/me" -H "${AUTH_HEADER[@]}" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Demo Updated","bio":"Hello from demo script"}' | jq .

info "Create podcast"
PODCAST_JSON=$(curl -s -X POST "$SERVER/podcasts" -H "${AUTH_HEADER[@]}" \
  -F "title=Demo Podcast" -F "description=Demo description")
PODCAST_ID=$(echo "$PODCAST_JSON" | jq -r '.id')
echo "Podcast ID: $PODCAST_ID"

info "Create episode"
EPISODE_JSON=$(curl -s -X POST "$SERVER/podcasts/$PODCAST_ID/episodes" -H "${AUTH_HEADER[@]}" \
  -F "title=Demo Episode" -F "description=Short intro" -F "tags=#demo,#go" \
  -F "audio=@$TMP_AUDIO")
EPISODE_ID=$(echo "$EPISODE_JSON" | jq -r '.id')
echo "Episode ID: $EPISODE_ID"

info "Get episode"
curl -s "$SERVER/episodes/$EPISODE_ID" | jq .

info "Play episode (records play + history)"
curl -s -X POST "$SERVER/episodes/$EPISODE_ID/plays" -H "${AUTH_HEADER[@]}" -o /dev/null -w "%{http_code}\n"

info "Like episode"
curl -s -X POST "$SERVER/episodes/$EPISODE_ID/like" -H "${AUTH_HEADER[@]}" -o /dev/null -w "%{http_code}\n"

info "Comment episode"
curl -s -X POST "$SERVER/episodes/$EPISODE_ID/comments" -H "${AUTH_HEADER[@]}" \
  -H "Content-Type: application/json" -d '{"text":"Nice episode!"}' -o /dev/null -w "%{http_code}\n"

info "List comments"
curl -s "$SERVER/episodes/$EPISODE_ID/comments" | jq .

info "Popular episodes"
curl -s "$SERVER/episodes/popular?page=1&page_size=5" | jq .

info "Create playlist and add episode"
PLAYLIST_JSON=$(curl -s -X POST "$SERVER/playlists" -H "${AUTH_HEADER[@]}" \
  -H "Content-Type: application/json" -d '{"title":"Demo Playlist","description":"Mix"}')
PLAYLIST_ID=$(echo "$PLAYLIST_JSON" | jq -r '.id')
curl -s -X POST "$SERVER/playlists/$PLAYLIST_ID/items" -H "${AUTH_HEADER[@]}" \
  -H "Content-Type: application/json" -d "{\"episode_id\":$EPISODE_ID}" -o /dev/null -w "%{http_code}\n"
curl -s "$SERVER/playlists/$PLAYLIST_ID" | jq .

info "Follow another user"
FRIEND_LOGIN=$(login_json "$OTHER_EMAIL" "$OTHER_PASSWORD")
FRIEND_ID=$(echo "$FRIEND_LOGIN" | jq -r '.user.id')
curl -s -X POST "$SERVER/users/$FRIEND_ID/follow" -H "Authorization: Bearer $ACCESS_TOKEN" -o /dev/null -w "%{http_code}\n" || true

info "Feed (may be empty until following authors with episodes)"
curl -s -H "${AUTH_HEADER[@]}" "$SERVER/feed?page=1&page_size=10" | jq .

info "Cleanup temp audio"
rm -f "$TMP_AUDIO"

info "Done."
