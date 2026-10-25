YouTube Lite - low-power YouTube player (Windows 10/11, 64-bit)
================================================================

HOW TO USE
  1. Unzip this whole folder somewhere (e.g. Documents). Keep "bin" next to the exe.
  2. Double-click YouTubeLite.exe
  3. Paste a YouTube link, or type a search and pick a number.
  You can also drag a link onto YouTubeLite.exe.

PLAYER CONTROLS (in the video window)
  Space = pause      Left/Right = seek 5s    Up/Down = seek 1 min
  f = fullscreen     9 / 0 = volume          m = mute
  q = close video and go back to search
  Videos remember where you stopped.

WHY IT USES LESS POWER THAN A BROWSER
  - No browser, no ads, no page scripts, no recommendations loading.
  - Always picks 1080p at 30fps using H.264, which every GPU decodes in
    dedicated hardware, so the CPU barely does anything.
  - Rendering uses the cheapest scaling and turns off extra filters.

NOTE ON 60FPS VIDEOS
  YouTube doesn't keep a 30fps copy of videos uploaded at 60fps. For those,
  you get 1080p60 (still hardware decoded).

IF VIDEOS STOP PLAYING
  YouTube changes things often. The app updates its downloader (yt-dlp)
  every few days on its own; you can also type  u  to update right now.

FIRST LAUNCH
  Windows may show "Windows protected your PC" because the exe isn't
  code-signed. Click "More info" -> "Run anyway".

Built with mpv (https://mpv.io) and yt-dlp (https://github.com/yt-dlp/yt-dlp).
