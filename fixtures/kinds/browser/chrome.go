package browser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// chromeSearch is where resolveChrome looks, in order: an explicit binary,
// then the newest Chrome for Testing a Playwright install downloaded, then a
// system Chrome.
type chromeSearch struct {
	Binary          string
	PlaywrightRoots []string
	System          []string
}

// playwrightChromeLayouts are the executables inside a Playwright
// `chromium-<revision>` directory, across platforms and Playwright versions.
var playwrightChromeLayouts = []string{
	"chrome-mac*/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing",
	"chrome-mac*/Chromium.app/Contents/MacOS/Chromium",
	"chrome-linux*/chrome",
	"chrome-win*/chrome.exe",
}

func defaultChromeSearch(binary string) chromeSearch {
	search := chromeSearch{Binary: binary}
	if root := os.Getenv("PLAYWRIGHT_BROWSERS_PATH"); root != "" && root != "0" {
		search.PlaywrightRoots = append(search.PlaywrightRoots, root)
	}
	if home, err := os.UserHomeDir(); err == nil {
		if runtime.GOOS == "darwin" {
			search.PlaywrightRoots = append(search.PlaywrightRoots, filepath.Join(home, "Library", "Caches", "ms-playwright"))
		} else {
			search.PlaywrightRoots = append(search.PlaywrightRoots, filepath.Join(home, ".cache", "ms-playwright"))
		}
	}
	if runtime.GOOS == "darwin" {
		search.System = append(search.System,
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium")
	}
	for _, name := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
		if path, err := exec.LookPath(name); err == nil {
			search.System = append(search.System, path)
		} else {
			search.System = append(search.System, name+" (PATH)")
		}
	}
	return search
}

// resolveChrome returns the first executable Chrome; when there is none the
// error lists every location tried.
func resolveChrome(search chromeSearch) (string, error) {
	if search.Binary != "" {
		if !isExecutable(search.Binary) {
			return "", fmt.Errorf("browser binary %q does not exist or is not executable", search.Binary)
		}
		return search.Binary, nil
	}
	var tried []string
	for _, root := range search.PlaywrightRoots {
		if path := newestPlaywrightChrome(root); path != "" {
			return path, nil
		}
		tried = append(tried, filepath.Join(root, "chromium-*"))
	}
	for _, path := range search.System {
		if isExecutable(path) {
			return path, nil
		}
		tried = append(tried, path)
	}
	return "", fmt.Errorf("no Chrome found (set `binary:` or run `pnpm exec playwright install chromium`); tried: %s", strings.Join(tried, ", "))
}

// newestPlaywrightChrome picks the highest chromium-<revision> under root.
func newestPlaywrightChrome(root string) string {
	dirs, _ := filepath.Glob(filepath.Join(root, "chromium-*"))
	sort.Slice(dirs, func(i, j int) bool { return revision(dirs[i]) > revision(dirs[j]) })
	for _, dir := range dirs {
		for _, layout := range playwrightChromeLayouts {
			matches, _ := filepath.Glob(filepath.Join(dir, layout))
			for _, match := range matches {
				if isExecutable(match) {
					return match
				}
			}
		}
	}
	return ""
}

func revision(dir string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(dir), "chromium-"))
	if err != nil {
		return -1
	}
	return n
}
