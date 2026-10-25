#!/usr/bin/env bash
# Builds dist/YouTubeLite.zip: the launcher exe plus mpv, yt-dlp and the low-power config.
# Needs: go, curl, zip, 7zz (or 7z).
set -euo pipefail
cd "$(dirname "$0")/.."
SEVENZ=$(command -v 7zz || command -v 7z)
OUT=dist/YouTubeLite
rm -rf dist && mkdir -p "$OUT/bin/portable_config" dist/tmp

GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$OUT/YouTubeLite.exe" .

# Latest mpv Windows build (shinchiro), mirrored on SourceForge.
MPV=$(curl -fsSL "https://sourceforge.net/projects/mpv-player-windows/rss?path=/64bit" \
  | grep -o 'mpv-x86_64-[0-9][^<"/]*\.7z' | head -1)
curl -fsSL -o dist/tmp/mpv.7z "https://sourceforge.net/projects/mpv-player-windows/files/64bit/$MPV/download"
"$SEVENZ" e dist/tmp/mpv.7z -o"$OUT/bin" mpv.exe d3dcompiler_43.dll -y >/dev/null

curl -fsSL -o "$OUT/bin/yt-dlp.exe" https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe

cp config/mpv.conf "$OUT/bin/portable_config/"
sed 's/\r*$/\r/' config/README.txt > "$OUT/README.txt"

(cd dist && zip -qr9 YouTubeLite.zip YouTubeLite)
rm -rf dist/tmp
echo "Built dist/YouTubeLite.zip ($MPV)"
