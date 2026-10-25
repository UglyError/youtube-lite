//go:build windows

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	mpvPipe      = `\\.\pipe\youtube-lite-mpv`
	updateEvery  = 3 * 24 * time.Hour
	createNoWnd  = 0x08000000 // CREATE_NO_WINDOW
	updateMarker = "last-update.txt"
	mpvLoadError = 2 // mpv exit code when nothing could be played
)

// play sends url to the running player window, or opens a new one.
// With queue set, the video is added after the current one instead of replacing it.
func play(url string, queue bool, cookieFile string, signedIn bool) {
	mode := "replace"
	if queue {
		mode = "append-play"
	}
	if sendToPlayer("loadfile", url, mode) {
		return
	}

	args := []string{"--input-ipc-server=" + mpvPipe, "--log-file=" + playerLog}
	// yt-dlp needs a JavaScript runtime to unlock YouTube's video streams.
	if deno := filepath.Join(binDir, "deno.exe"); fileExists(deno) {
		args = append(args, "--ytdl-raw-options-append=js-runtimes=deno:"+deno)
	}
	if cookieFile != "" {
		args = append(args, "--ytdl-raw-options-append=cookies="+cookieFile)
		if signedIn {
			// Videos played here show up in your YouTube watch history.
			args = append(args, "--ytdl-raw-options-append=mark-watched=")
		}
	}
	args = append(args, url)
	cmd := exec.Command(mpvPath, args...)
	cmd.Dir = binDir
	if err := cmd.Start(); err != nil {
		messageBox(appName, "Couldn't start the player:\n"+err.Error(), mbIconError)
		return
	}
	go func() {
		started := time.Now()
		err := cmd.Wait()
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == mpvLoadError && time.Since(started) < 2*time.Minute {
			reason := playerErrors()
			logf("playback failed for %s: %s", url, reason)
			if reason != "" {
				reason = "\n\nDetails:\n" + reason
			}
			messageBox(appName,
				"That video couldn't be played. The video downloader is checking for an update now, so try again in a minute."+reason,
				mbIconWarning)
			update()
		}
	}()
}

// playerErrors returns the last few error lines from the player's log.
func playerErrors() string {
	b, err := os.ReadFile(playerLog)
	if err != nil {
		return ""
	}
	var errs []string
	for _, l := range strings.Split(string(b), "\n") {
		// mpv log lines look like "[   1.234][e][ytdl_hook] ERROR: ..."
		if i := strings.Index(l, "][e]["); i >= 0 {
			msg := strings.TrimSpace(l[i+5:])
			if j := strings.Index(msg, "] "); j >= 0 {
				msg = msg[j+2:]
			}
			errs = append(errs, msg)
		}
	}
	if len(errs) > 4 {
		errs = errs[len(errs)-4:]
	}
	return strings.Join(errs, "\n")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// sendToPlayer runs one command in the open mpv window. It reports false if no player is open.
func sendToPlayer(cmd ...string) bool {
	f, err := os.OpenFile(mpvPipe, os.O_RDWR, 0)
	if err != nil {
		return false
	}
	defer f.Close()
	b, _ := json.Marshal(map[string]any{"command": cmd})
	_, err = f.Write(append(b, '\n'))
	return err == nil
}

// YouTube changes often, so yt-dlp keeps itself current every few days.
func maybeUpdate() {
	if st, err := os.Stat(filepath.Join(binDir, updateMarker)); err == nil && time.Since(st.ModTime()) < updateEvery {
		return
	}
	update()
}

func update() {
	cmd := exec.Command(ytdlpPath, "-U")
	cmd.Dir = binDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWnd}
	out, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(out), "up to date") {
		return
	}
	_ = os.WriteFile(filepath.Join(binDir, updateMarker), []byte(time.Now().Format(time.RFC3339)), 0o644)
}
