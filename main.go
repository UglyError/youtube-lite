//go:build windows

// YouTube Lite: browse YouTube (signed in, with your own feed and subscriptions) in a
// lightweight window, and watch videos in mpv with a low-power, hardware-decoded
// 1080p/30fps profile instead of the website's player.
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/go-webview2/pkg/edge"
	"github.com/wailsapp/go-webview2/webviewloader"
	"golang.org/x/sys/windows"
)

//go:embed inject.js
var injectJS string

const (
	appName      = "YouTube Lite"
	startURL     = "https://www.youtube.com/"
	webview2Page = "https://go.microsoft.com/fwlink/p/?LinkId=2124703"
)

var (
	binDir, mpvPath, ytdlpPath string
	dataDir, cookiePath        string
	playerLog                  string

	mainWnd  uintptr
	chromium *edge.Chromium
	cookies  *edge.ICoreWebView2CookieManager
)

func init() {
	// WebView2 and the window must live on one OS thread.
	runtime.LockOSThread()
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		fatal(err.Error())
	}
	binDir = filepath.Join(filepath.Dir(exe), "bin")
	mpvPath = filepath.Join(binDir, "mpv.exe")
	ytdlpPath = filepath.Join(binDir, "yt-dlp.exe")
	for _, p := range []string{mpvPath, ytdlpPath} {
		if _, err := os.Stat(p); err != nil {
			fatal("Missing " + p + "\n\nKeep the \"bin\" folder next to YouTubeLite.exe.")
		}
	}

	// Sign-in and the cookie copy for the player stay in your own Windows profile.
	dataDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "YouTubeLite")
	_ = os.MkdirAll(dataDir, 0o700)
	cookiePath = filepath.Join(dataDir, "cookies.txt")
	playerLog = filepath.Join(dataDir, "player-log.txt")

	if v, err := webviewloader.GetAvailableCoreWebView2BrowserVersionString(""); err != nil || v == "" {
		if messageBox(appName, "YouTube Lite needs the Microsoft Edge WebView2 Runtime (built into Windows 11 and most Windows 10 PCs).\n\nOpen the download page now?", mbYesNo|mbIconWarning) == idYes {
			openInBrowser(webview2Page)
		}
		os.Exit(1)
	}

	go maybeUpdate()

	enableDPIAwareness()
	_ = windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED)
	mainWnd = createMainWindow(appName, mpvPath, wndProc)
	if mainWnd == 0 {
		fatal("Couldn't create the window.")
	}

	// The window must be visible before WebView2 attaches to it, or the page can stay blank.
	showWindow(mainWnd)

	chromium = edge.NewChromium()
	chromium.DataPath = filepath.Join(dataDir, "WebView")
	chromium.SetErrorCallback(func(err error) {
		logf("webview error: %v", err)
		fatal("The browser view stopped working:\n" + err.Error())
	})
	chromium.MessageCallback = func(message string, _ *edge.ICoreWebView2, _ *edge.ICoreWebView2WebMessageReceivedEventArgs) {
		onPageMessage(message)
	}
	chromium.NavigationCompletedCallback = onNavigationCompleted
	chromium.ProcessFailedCallback = func(_ *edge.ICoreWebView2, args *edge.ICoreWebView2ProcessFailedEventArgs) {
		kind, _ := args.GetProcessFailedKind()
		logf("webview process failed, kind %d", kind)
	}
	chromium.Embed(mainWnd)
	chromium.SetBackgroundColour(15, 15, 15, 255)
	applyDPI()
	loadZoom()

	if s, err := chromium.GetSettings(); err == nil {
		_ = s.PutAreDevToolsEnabled(false)
		_ = s.PutIsStatusBarEnabled(false)
		_ = s.PutAreHostObjectsAllowed(false)
	}
	if cookies, err = chromium.GetCookieManager(); err != nil {
		cookies = nil
	}

	chromium.Init(injectJS)
	chromium.Navigate(startURL)
	chromium.Resize()
	chromium.Focus()
	logf("started")

	// A link dragged onto the exe (or "Open with") plays straight away.
	for _, a := range os.Args[1:] {
		if strings.Contains(a, "youtu") {
			startPlayback(a, false)
		}
	}

	runMessageLoop()
}

func wndProc(hwnd, msg, wp, lp uintptr) uintptr {
	switch msg {
	case wmSize:
		if chromium != nil {
			chromium.Resize()
		}
	case wmMove:
		if chromium != nil {
			_ = chromium.NotifyParentWindowPositionChanged()
		}
	case wmDpiChanged:
		r := (*rect)(unsafe.Pointer(lp))
		procSetWindowPos.Call(hwnd, 0, uintptr(r.left), uintptr(r.top),
			uintptr(r.right-r.left), uintptr(r.bottom-r.top), swpNoZOrder|swpNoActivate)
		applyDPI()
		return 0
	case wmClose:
		if chromium != nil {
			saveZoom()
		}
	case wmDestroy:
		if chromium != nil {
			chromium.ShuttingDown()
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	return defWindowProc(hwnd, msg, wp, lp)
}

var firstPageShown bool

// navigationArgs mirrors ICoreWebView2NavigationCompletedEventArgs, whose getters go-webview2 doesn't expose.
type navigationArgs struct {
	vtbl *struct {
		QueryInterface, AddRef, Release                  uintptr
		GetIsSuccess, GetWebErrorStatus, GetNavigationID uintptr
	}
}

func onNavigationCompleted(_ *edge.ICoreWebView2, args *edge.ICoreWebView2NavigationCompletedEventArgs) {
	nav := (*navigationArgs)(unsafe.Pointer(args))
	var ok, status int32
	syscall.SyscallN(nav.vtbl.GetIsSuccess, uintptr(unsafe.Pointer(nav)), uintptr(unsafe.Pointer(&ok)))
	if ok == 0 {
		syscall.SyscallN(nav.vtbl.GetWebErrorStatus, uintptr(unsafe.Pointer(nav)), uintptr(unsafe.Pointer(&status)))
		logf("page failed to load, web error status %d", status)
	}
	if firstPageShown {
		return
	}
	firstPageShown = true
	// Toggling visibility once makes WebView2 start painting
	// (https://github.com/MicrosoftEdge/WebView2Feedback/issues/1077).
	_ = chromium.Hide()
	_ = chromium.Show()
	chromium.Resize()
	chromium.Focus()
}

// logf appends a line to %LOCALAPPDATA%\YouTubeLite\log.txt, kept small.
func logf(format string, a ...any) {
	path := filepath.Join(dataDir, "log.txt")
	if st, err := os.Stat(path); err == nil && st.Size() > 256<<10 {
		_ = os.Remove(path)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s  %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, a...))
}

// Page zoom on top of Windows scaling. Ctrl + mouse wheel, Ctrl +/- and Ctrl+0 change it,
// and the last setting is remembered.
const defaultZoom = 1.25

func zoomFile() string { return filepath.Join(dataDir, "zoom.txt") }

func loadZoom() {
	zoom := defaultZoom
	if b, err := os.ReadFile(zoomFile()); err == nil {
		if z, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64); err == nil && z >= 0.25 && z <= 5 {
			zoom = z
		}
	}
	_ = chromium.GetController().PutZoomFactor(zoom)
}

func saveZoom() {
	if z, err := chromium.GetController().GetZoomFactor(); err == nil && z > 0 {
		_ = os.WriteFile(zoomFile(), []byte(strconv.FormatFloat(z, 'f', 2, 64)), 0o600)
	}
}

// applyDPI scales the page to match Windows display scaling.
func applyDPI() {
	if chromium == nil || chromium.GetController() == nil {
		return
	}
	if c3 := chromium.GetController().GetICoreWebView2Controller3(); c3 != nil {
		// Let WebView2 follow the monitor's scale (125%, 150%...) itself, starting from the current one.
		_ = c3.PutRasterizationScale(dpiFor(mainWnd) / 96)
		_ = c3.PutShouldDetectMonitorScaleChanges(true)
	}
	chromium.Resize()
}

type pageMessage struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Queue bool   `json:"queue"`
}

func onPageMessage(raw string) {
	var m pageMessage
	if json.Unmarshal([]byte(raw), &m) != nil || m.Type != "play" || m.URL == "" {
		return
	}
	startPlayback(m.URL, m.Queue)
}

// startPlayback refreshes the player's copy of your sign-in, then plays the video.
func startPlayback(url string, queue bool) {
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}
	if cookies == nil {
		go play(url, queue, "", false)
		return
	}
	fetchCookies(cookies, func(list []cookie, err error) {
		file, signedIn := "", false
		if err == nil {
			if signedIn, err = writeCookieFile(cookiePath, list); err == nil {
				file = cookiePath
			}
		}
		go play(url, queue, file, signedIn)
	})
}

func fatal(text string) {
	messageBox(appName, text, mbIconError)
	os.Exit(1)
}
