#!/usr/bin/env bash
# Builds dist/YouTubeLite.zip: the launcher exe plus mpv, yt-dlp and the low-power config.
# Needs: go, curl, zip, python3. Downloads 7-Zip itself if 7zz isn't installed.
set -euo pipefail
cd "$(dirname "$0")/.."
trap 'echo "::error::build.sh failed at line $LINENO: $BASH_COMMAND"' ERR

OUT=dist/YouTubeLite
TMP=dist/tmp
rm -rf dist && mkdir -p "$OUT/bin/portable_config" "$TMP"

echo "== Building launcher"
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$OUT/YouTubeLite.exe" .

echo "== Getting 7-Zip"
SEVENZ=$(command -v 7zz || true)
if [ -z "$SEVENZ" ]; then
  curl -fsSL -o "$TMP/7z.tar.xz" https://github.com/ip7z/7zip/releases/download/24.09/7z2409-linux-x64.tar.xz
  mkdir -p "$TMP/7z" && tar -xf "$TMP/7z.tar.xz" -C "$TMP/7z"
  SEVENZ="$TMP/7z/7zz"
fi

echo "== Getting mpv"
# Latest shinchiro Windows build from GitHub; SourceForge mirror as fallback.
AUTH=()
[ -n "${GITHUB_TOKEN:-}" ] && AUTH=(-H "Authorization: Bearer $GITHUB_TOKEN")
MPV_URL=$(curl -fsSL "${AUTH[@]}" https://api.github.com/repos/shinchiro/mpv-winbuild-cmake/releases/latest \
  2>/dev/null | python3 -c 'import json,re,sys
for a in json.load(sys.stdin)["assets"]:
    if re.match(r"mpv-x86_64-\d", a["name"]): print(a["browser_download_url"]); break' 2>/dev/null || true)
if [ -z "$MPV_URL" ]; then
  curl -fsSL -o "$TMP/rss.xml" "https://sourceforge.net/projects/mpv-player-windows/rss?path=/64bit"
  MPV=$(grep -o -m1 'mpv-x86_64-[0-9][^<"/]*\.7z' "$TMP/rss.xml" | sed -n 1p)
  MPV_URL="https://sourceforge.net/projects/mpv-player-windows/files/64bit/$MPV/download"
fi
echo "mpv: $MPV_URL"
curl -fsSL -o "$TMP/mpv.7z" "$MPV_URL"
"$SEVENZ" e "$TMP/mpv.7z" -o"$OUT/bin" mpv.exe d3dcompiler_43.dll -y >/dev/null

echo "== Getting yt-dlp"
curl -fsSL -o "$OUT/bin/yt-dlp.exe" https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe

cp config/mpv.conf "$OUT/bin/portable_config/"
sed 's/\r*$/\r/' config/README.txt > "$OUT/README.txt"

(cd dist && zip -qr9 YouTubeLite.zip YouTubeLite)
rm -rf "$TMP"
ls -la dist/YouTubeLite.zip
