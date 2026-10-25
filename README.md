# YouTube Lite

A tiny Windows YouTube player built to use as little PC power as possible without hurting the viewing experience.

- **1080p at 30fps.** It picks the H.264 version of each video, which every GPU since about 2010 decodes in dedicated hardware, so the CPU barely works.
- **No browser.** No ads, page scripts or background recommendations.
- **Cheap rendering.** Bilinear scaling, no debanding, dithering or interpolation.
- Search from the app, or paste or drag in a link. Videos remember where you stopped.

## Download

Get `YouTubeLite.zip` from the [Releases](../../releases) page. Unzip it and run `YouTubeLite.exe` (Windows 10/11, 64-bit).

Windows may say "Windows protected your PC" because the exe isn't signed. Click **More info → Run anyway**.

## Controls

| Key | Action |
| --- | --- |
| Space | Pause |
| ← / → | Seek 5 s |
| ↑ / ↓ | Seek 1 min |
| f | Fullscreen |
| 9 / 0 | Volume |
| q | Close video, back to search |

In the launcher, type `u` to update yt-dlp right away. It also updates itself every 3 days.

## Notes

- Videos uploaded at 60fps have no 30fps copy on YouTube, so those play at 1080p60 (still hardware decoded).
- Videos never uploaded in 1080p play at the best quality available.

## Building

```sh
./scripts/build.sh   # needs go, curl, zip and python3; output: dist/YouTubeLite.zip
```

Every push to `main` (or a `v*` tag) makes GitHub Actions build the zip and publish it as a release.

## Files

- `main.go`: the launcher (search, pick, play, auto-update yt-dlp)
- `config/mpv.conf`: the low-power playback profile

Built on [mpv](https://mpv.io) and [yt-dlp](https://github.com/yt-dlp/yt-dlp).
