# YouTube Lite

A Windows YouTube app built to use as little PC power as possible without hurting the viewing experience.

- **Your real YouTube.** Sign in with your Google account and get your Home feed, Subscriptions, playlists and search, in a lightweight window.
- **Low-power playback.** Clicking a video opens it in [mpv](https://mpv.io) at **1080p / 30fps in H.264**, which every GPU since about 2010 decodes in dedicated hardware, so the CPU barely works. The website's own player and hover previews never run.
- **Modern playback controls** from [uosc](https://github.com/tomasklaen/uosc): seek bar, volume, subtitles, speed, playlist, fullscreen.
- Page zoom (Ctrl + mouse wheel, Ctrl +/-) is remembered; starts at 125%.
- Shift+click a video to queue it. Watched videos go to your YouTube history while you're signed in.

## Download

Get `YouTubeLite.zip` from the [latest release](../../releases/latest). Unzip it and run `YouTubeLite.exe` (Windows 10/11, 64-bit).

Windows may say "Windows protected your PC" because the exe isn't signed. Click **More info → Run anyway**.

## Controls

Move the mouse over the player to show the control bar. Right-click opens the full menu.

| Key | Action |
| --- | --- |
| Space | Pause |
| ← / → | Seek 5 s |
| ↑ / ↓ | Seek 1 min |
| f | Fullscreen |
| 9 / 0 | Volume |
| q | Close the player |

## How it works

- The window is Microsoft Edge WebView2 showing youtube.com. `inject.js` catches clicks on videos and sends them to the app instead of letting the page play them.
- The app copies your YouTube sign-in cookies from WebView2 into `%LOCALAPPDATA%\YouTubeLite\cookies.txt` for [yt-dlp](https://github.com/yt-dlp/yt-dlp), so members-only, age-restricted and private videos work. Nothing leaves your PC except normal YouTube traffic. Delete that folder to sign out.
- mpv runs with `config/mpv.conf`: zero-copy D3D11 hardware decoding and cheap bilinear scaling, with no debanding, dithering or interpolation.
- yt-dlp uses the bundled [Deno](https://deno.com) runtime to unlock YouTube's streams, and updates itself every 3 days and right away when a video fails to play. Player errors are logged to `%LOCALAPPDATA%\YouTubeLite\player-log.txt`.

## Notes

- Videos uploaded at 60fps have no 30fps copy on YouTube, so those play at 1080p60 (still hardware decoded).
- Videos never uploaded in 1080p play at the best quality available.

## Building

```sh
./scripts/build.sh   # needs go, curl, zip and python3; output: dist/YouTubeLite.zip
```

Every push to `main` (or a `v*` tag) makes GitHub Actions build the zip and publish it as a release.

## Files

- `main.go`, `win32.go`: the app window (WebView2)
- `inject.js`: hands video clicks to the player and stops in-page playback
- `cookies.go`: copies your sign-in for yt-dlp
- `player.go`: starts and controls mpv, keeps yt-dlp updated
- `config/mpv.conf`: the low-power playback profile
- `config/script-opts/uosc.conf`: player control bar layout
