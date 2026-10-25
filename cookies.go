//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/wailsapp/go-webview2/pkg/edge"
	"golang.org/x/sys/windows"
)

// The cookie list is fetched with ICoreWebView2CookieManager::GetCookies, which is
// asynchronous: it reports back through an ICoreWebView2GetCookiesCompletedHandler.
// go-webview2's own GetCookies wrapper treats it as synchronous, so the call and the
// completion handler are implemented here.

type cookieManager struct {
	vtbl *struct {
		QueryInterface, AddRef, Release                             uintptr
		CreateCookie, CopyCookie, GetCookies, AddOrUpdateCookie     uintptr
		DeleteCookie, DeleteCookies, DeleteCookiesWithDomainAndPath uintptr
		DeleteAllCookies                                            uintptr
	}
}

type comHandler struct {
	vtbl *comHandlerVtbl
}

type comHandlerVtbl struct {
	QueryInterface, AddRef, Release, Invoke uintptr
}

type cookie struct {
	domain, path, name, value string
	secure                    bool
	expires                   float64
}

var (
	cookiesHandler = &comHandler{vtbl: &comHandlerVtbl{
		QueryInterface: windows.NewCallback(func(this, _, out uintptr) uintptr {
			*(*uintptr)(unsafe.Pointer(out)) = this
			return 0
		}),
		AddRef:  windows.NewCallback(func(uintptr) uintptr { return 1 }),
		Release: windows.NewCallback(func(uintptr) uintptr { return 1 }),
		Invoke:  windows.NewCallback(onCookiesReady),
	}}
	emptyURI       = utf16("")
	cookieWaiters  []func([]cookie, error)
	cookieInFlight bool
)

// fetchCookies asks WebView2 for every cookie in the profile. done runs on the UI thread.
func fetchCookies(cm *edge.ICoreWebView2CookieManager, done func([]cookie, error)) {
	cookieWaiters = append(cookieWaiters, done)
	if cookieInFlight {
		return
	}
	cookieInFlight = true
	m := (*cookieManager)(unsafe.Pointer(cm))
	hr, _, _ := syscall.SyscallN(m.vtbl.GetCookies,
		uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(emptyURI)), uintptr(unsafe.Pointer(cookiesHandler)))
	if int32(hr) < 0 {
		cookieInFlight = false
		finishCookies(nil, fmt.Errorf("GetCookies failed: %#x", hr))
	}
}

func onCookiesReady(_, hr, listPtr uintptr) uintptr {
	cookieInFlight = false
	if int32(hr) < 0 || listPtr == 0 {
		finishCookies(nil, fmt.Errorf("reading cookies failed: %#x", hr))
		return 0
	}
	list := (*edge.ICoreWebView2CookieList)(unsafe.Pointer(listPtr))
	n, err := list.GetCount()
	if err != nil {
		finishCookies(nil, err)
		return 0
	}
	var out []cookie
	for i := uint32(0); i < n; i++ {
		c, err := list.GetItem(i)
		if err != nil || c == nil {
			continue
		}
		var ck cookie
		ck.domain, _ = c.GetDomain()
		ck.path, _ = c.GetPath()
		ck.name, _ = c.GetName()
		ck.value, _ = c.GetValue()
		ck.secure, _ = c.GetIsSecure()
		ck.expires, _ = c.GetExpires()
		c.Release()
		out = append(out, ck)
	}
	finishCookies(out, nil)
	return 0
}

func finishCookies(c []cookie, err error) {
	waiters := cookieWaiters
	cookieWaiters = nil
	for _, w := range waiters {
		w(c, err)
	}
}

func isYouTubeOrGoogle(domain string) bool {
	d := strings.TrimPrefix(strings.ToLower(domain), ".")
	return d == "youtube.com" || strings.HasSuffix(d, ".youtube.com") ||
		d == "google.com" || strings.HasSuffix(d, ".google.com")
}

// writeCookieFile saves the YouTube/Google cookies in the Netscape format yt-dlp reads.
// It reports whether the user is signed in to YouTube.
func writeCookieFile(path string, cookies []cookie) (signedIn bool, err error) {
	var b strings.Builder
	b.WriteString("# Netscape HTTP Cookie File\n# Written by YouTube Lite from its own sign-in. Do not share this file.\n\n")
	for _, c := range cookies {
		if !isYouTubeOrGoogle(c.domain) || strings.ContainsAny(c.name+c.value, "\t\n") {
			continue
		}
		if strings.HasSuffix(strings.ToLower(c.domain), "youtube.com") && (c.name == "SID" || c.name == "__Secure-3PSID") {
			signedIn = true
		}
		sub := "FALSE"
		if strings.HasPrefix(c.domain, ".") {
			sub = "TRUE"
		}
		sec := "FALSE"
		if c.secure {
			sec = "TRUE"
		}
		exp := int64(0)
		if c.expires > 0 {
			exp = int64(c.expires)
		}
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n", c.domain, sub, c.path, sec, exp, c.name, c.value)
	}
	if b.Len() == 0 {
		return false, errors.New("no cookies")
	}
	return signedIn, os.WriteFile(path, []byte(b.String()), 0o600)
}
