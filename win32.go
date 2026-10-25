//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	shell32 = windows.NewLazySystemDLL("shell32.dll")

	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procGetMessageW                   = user32.NewProc("GetMessageW")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procPostQuitMessage               = user32.NewProc("PostQuitMessage")
	procPostMessageW                  = user32.NewProc("PostMessageW")
	procShowWindow                    = user32.NewProc("ShowWindow")
	procUpdateWindow                  = user32.NewProc("UpdateWindow")
	procSetWindowPos                  = user32.NewProc("SetWindowPos")
	procMessageBoxW                   = user32.NewProc("MessageBoxW")
	procLoadCursorW                   = user32.NewProc("LoadCursorW")
	procGetDpiForWindow               = user32.NewProc("GetDpiForWindow")
	procGetDpiForSystem               = user32.NewProc("GetDpiForSystem")
	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procExtractIconW                  = shell32.NewProc("ExtractIconW")
	procShellExecuteW                 = shell32.NewProc("ShellExecuteW")
)

const (
	wmDestroy    = 0x0002
	wmMove       = 0x0003
	wmSize       = 0x0005
	wmDpiChanged = 0x02E0
	wmApp        = 0x8000

	wsOverlappedWindow = 0x00CF0000
	cwUseDefault       = 0x80000000
	swShowNormal       = 1
	swpNoZOrder        = 0x0004
	swpNoActivate      = 0x0010

	mbOK          = 0x00000000
	mbYesNo       = 0x00000004
	mbIconError   = 0x00000010
	mbIconWarning = 0x00000030
	mbIconInfo    = 0x00000040
	idYes         = 6
)

type wndClassExW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type point struct{ x, y int32 }

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type rect struct{ left, top, right, bottom int32 }

func utf16(s string) *uint16 {
	p, _ := windows.UTF16PtrFromString(s)
	return p
}

func messageBox(title, text string, flags uintptr) int {
	r, _, _ := procMessageBoxW.Call(0, uintptr(unsafe.Pointer(utf16(text))), uintptr(unsafe.Pointer(utf16(title))), flags)
	return int(r)
}

func openInBrowser(url string) {
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16("open"))), uintptr(unsafe.Pointer(utf16(url))), 0, 0, swShowNormal)
}

// enableDPIAwareness keeps the window sharp on scaled (125%, 150%...) displays.
func enableDPIAwareness() {
	if procSetProcessDpiAwarenessContext.Find() == nil {
		procSetProcessDpiAwarenessContext.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 (-4)
	}
}

func dpiFor(hwnd uintptr) float64 {
	if procGetDpiForWindow.Find() == nil && hwnd != 0 {
		if d, _, _ := procGetDpiForWindow.Call(hwnd); d != 0 {
			return float64(d)
		}
	}
	if procGetDpiForSystem.Find() == nil {
		if d, _, _ := procGetDpiForSystem.Call(); d != 0 {
			return float64(d)
		}
	}
	return 96
}

// createMainWindow registers a window class and creates a normal top-level window.
func createMainWindow(title, iconSource string, wndProc func(hwnd, msg, wp, lp uintptr) uintptr) uintptr {
	var hinst windows.Handle
	_ = windows.GetModuleHandleEx(0, nil, &hinst)

	icon, _, _ := procExtractIconW.Call(uintptr(hinst), uintptr(unsafe.Pointer(utf16(iconSource))), 0)
	if icon <= 1 {
		icon = 0
	}
	cursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_ARROW

	class := utf16("YouTubeLiteWindow")
	wc := wndClassExW{
		lpfnWndProc:   windows.NewCallback(wndProc),
		hInstance:     hinst,
		hIcon:         icon,
		hIconSm:       icon,
		hCursor:       cursor,
		lpszClassName: class,
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	scale := dpiFor(0) / 96
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(class)),
		uintptr(unsafe.Pointer(utf16(title))),
		wsOverlappedWindow,
		cwUseDefault, cwUseDefault,
		uintptr(int(1280*scale)), uintptr(int(820*scale)),
		0, 0, uintptr(hinst), 0,
	)
	return hwnd
}

func showWindow(hwnd uintptr) {
	procShowWindow.Call(hwnd, swShowNormal)
	procUpdateWindow.Call(hwnd)
}

func runMessageLoop() {
	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func defWindowProc(hwnd, msg, wp, lp uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(hwnd, msg, wp, lp)
	return r
}
