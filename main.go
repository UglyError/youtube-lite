// YouTube Lite: a tiny launcher that searches YouTube with yt-dlp and plays
// videos in mpv using a low-power, hardware-decoded 1080p/30fps profile.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	searchCount  = 10
	updateEvery  = 3 * 24 * time.Hour
	createNoWnd  = 0x08000000 // CREATE_NO_WINDOW
	updateMarker = "last-update.txt"
)

type result struct {
	id, title, channel, duration string
}

var (
	binDir string
	ytdlp  string
	mpv    string
	in     = bufio.NewReader(os.Stdin)
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		fail(err)
	}
	// Show video titles with emoji/non-Latin characters correctly.
	syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleOutputCP").Call(65001)
	binDir = filepath.Join(filepath.Dir(exe), "bin")
	ytdlp = filepath.Join(binDir, "yt-dlp.exe")
	mpv = filepath.Join(binDir, "mpv.exe")
	for _, p := range []string{ytdlp, mpv} {
		if _, err := os.Stat(p); err != nil {
			fail(fmt.Errorf("missing %s - keep the bin folder next to YouTubeLite.exe", p))
		}
	}

	fmt.Println("==============================================")
	fmt.Println("  YouTube Lite  -  1080p / 30fps / low power")
	fmt.Println("==============================================")
	maybeUpdate()

	// Links passed on the command line (drag & drop onto the exe) play straight away.
	if len(os.Args) > 1 {
		for _, a := range os.Args[1:] {
			play(a)
		}
	}

	for {
		fmt.Println()
		fmt.Println("Paste a YouTube link, or type something to search.")
		fmt.Println("(u = update yt-dlp, q = quit)")
		line := prompt("> ")
		switch strings.ToLower(line) {
		case "":
			continue
		case "q", "quit", "exit":
			return
		case "u", "update":
			update()
			continue
		}
		if isURL(line) {
			play(line)
		} else {
			searchAndPick(line)
		}
	}
}

func isURL(s string) bool {
	l := strings.ToLower(s)
	return strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "https://") ||
		strings.HasPrefix(l, "www.") || strings.HasPrefix(l, "youtu.be/") ||
		strings.HasPrefix(l, "youtube.com/") || strings.HasPrefix(l, "m.youtube.com/")
}

func searchAndPick(query string) {
	fmt.Printf("Searching for \"%s\"...\n", query)
	cmd := hidden(exec.Command(ytdlp, "--flat-playlist", "--no-warnings", "--ignore-config",
		"--print", "%(id)s\t%(title)s\t%(channel,uploader)s\t%(duration_string)s",
		fmt.Sprintf("ytsearch%d:%s", searchCount, query)))
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Search failed. Check your internet connection, or type u to update yt-dlp.")
		return
	}
	var res []result
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		f := strings.Split(strings.TrimRight(l, "\r"), "\t")
		if len(f) < 4 || f[0] == "" {
			continue
		}
		res = append(res, result{f[0], f[1], f[2], f[3]})
	}
	if len(res) == 0 {
		fmt.Println("No results.")
		return
	}
	for {
		fmt.Println()
		for i, r := range res {
			dur := r.duration
			if dur == "NA" {
				dur = "live"
			}
			fmt.Printf("%2d. %s\n    %s  [%s]\n", i+1, trim(r.title, 70), trim(r.channel, 40), dur)
		}
		choice := prompt("\nPick a number (Enter = new search): ")
		if choice == "" {
			return
		}
		n, err := strconv.Atoi(choice)
		if err != nil || n < 1 || n > len(res) {
			fmt.Println("Not a valid number.")
			continue
		}
		play("https://www.youtube.com/watch?v=" + res[n-1].id)
	}
}

func play(url string) {
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}
	fmt.Println("Playing... (close the player window or press q in it to come back here)")
	cmd := exec.Command(mpv, url)
	cmd.Dir = binDir
	if err := cmd.Run(); err != nil {
		fmt.Println("Playback ended with an error:", err)
		fmt.Println("If videos stop working, type u to update yt-dlp.")
	}
}

// YouTube changes often, so yt-dlp keeps itself current every few days.
func maybeUpdate() {
	marker := filepath.Join(binDir, updateMarker)
	if st, err := os.Stat(marker); err == nil && time.Since(st.ModTime()) < updateEvery {
		return
	}
	update()
}

func update() {
	fmt.Println("Checking for yt-dlp updates...")
	cmd := hidden(exec.Command(ytdlp, "-U"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Update check failed (will retry later):", lastLine(string(out)))
		return
	}
	fmt.Println(lastLine(string(out)))
	_ = os.WriteFile(filepath.Join(binDir, updateMarker), []byte(time.Now().Format(time.RFC3339)), 0o644)
}

func hidden(c *exec.Cmd) *exec.Cmd {
	c.Dir = binDir
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWnd}
	return c
}

func prompt(p string) string {
	fmt.Print(p)
	s, err := in.ReadString('\n')
	if err != nil && s == "" {
		os.Exit(0)
	}
	return strings.TrimSpace(s)
}

func trim(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-3]) + "..."
}

func lastLine(s string) string {
	l := strings.Split(strings.TrimSpace(s), "\n")
	return l[len(l)-1]
}

func fail(err error) {
	fmt.Println("Error:", err)
	fmt.Print("Press Enter to close.")
	in.ReadString('\n')
	os.Exit(1)
}
