YouTube Lite - low-power YouTube player (Windows 10/11, 64-bit)
================================================================

HOW TO USE
  1. Unzip this whole folder somewhere (e.g. Documents). Keep "bin" next to the exe.
  2. Double-click YouTubeLite.exe. It opens YouTube in its own window.
  3. Click "Sign in" (top right) to log into your YouTube / Google account.
     You stay signed in next time.
  4. Browse your Home feed, Subscriptions, playlists or search as normal.
     Click any video and it opens in the player window.
     Shift+click a video to add it to the queue instead.

PLAYER CONTROLS
  Move the mouse over the player to show the control bar: seek bar,
  play/pause, volume, subtitles, speed, playlist, next/previous, fullscreen.
  Right-click (or the menu button) opens the full menu.

  Keyboard:
  Space = pause      Left/Right = seek 5s    Up/Down = seek 1 min
  f = fullscreen     9 / 0 = volume          m = mute
  q = close the player
  Videos remember where you stopped. Videos you watch are added to your
  YouTube watch history while you're signed in.

WHY IT USES LESS POWER THAN A BROWSER
  - The YouTube page never plays video itself (hover previews are off too).
  - The player always picks 1080p at 30fps in H.264, which every GPU decodes
    in dedicated hardware, so the CPU barely does anything.
  - Rendering uses the cheapest scaling and turns off extra filters.

NOTE ON 60FPS VIDEOS
  YouTube doesn't keep a 30fps copy of videos uploaded at 60fps. For those,
  you get 1080p60 (still hardware decoded).

PRIVACY
  Your sign-in is stored only on this PC, in
  %LOCALAPPDATA%\YouTubeLite. Delete that folder to sign out completely.

IF VIDEOS STOP PLAYING
  YouTube changes things often. The app updates its downloader (yt-dlp)
  every few days, and right away if a video fails to play.

FIRST LAUNCH
  Windows may show "Windows protected your PC" because the exe isn't
  code-signed. Click "More info" -> "Run anyway".
  The app needs the Microsoft Edge WebView2 Runtime. It's built into
  Windows 11 and most Windows 10 PCs; if missing, the app offers the link.

Built with mpv (https://mpv.io), uosc (https://github.com/tomasklaen/uosc)
yt-dlp (https://github.com/yt-dlp/yt-dlp) and Deno (https://deno.com).
