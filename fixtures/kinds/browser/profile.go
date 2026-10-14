package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Profile is the Chrome user-data-dir a browser step runs on.
type Profile struct {
	// Dir is the user-data-dir; default <tmp>/browser-profile-<step id>.
	Dir string `json:"dir,omitempty" jsonschema:"description=Chrome user-data-dir (default: a per-step temp dir)"`
	// Fresh wipes Dir before Chrome starts.
	Fresh bool `json:"fresh,omitempty" jsonschema:"description=Delete the profile directory before starting"`
	// From seeds the profile: a profile directory is copied into Dir; a
	// Playwright storage-state JSON file has its cookies and localStorage
	// applied once Chrome is up.
	From string `json:"from,omitempty" jsonschema:"description=Seed profile directory or Playwright storage-state JSON file"`
}

// preparedProfile is a profile directory ready for Chrome, plus the
// storage-state file (if any) to apply once Chrome is listening.
type preparedProfile struct {
	Dir          string
	StorageState string
}

func prepareProfile(profile Profile) (preparedProfile, error) {
	prepared := preparedProfile{Dir: profile.Dir}
	if profile.Fresh {
		if err := os.RemoveAll(profile.Dir); err != nil {
			return prepared, fmt.Errorf("wipe fresh profile %s: %w", profile.Dir, err)
		}
	}
	if err := os.MkdirAll(profile.Dir, 0o755); err != nil {
		return prepared, fmt.Errorf("create profile %s: %w", profile.Dir, err)
	}
	// Chrome rewrites it with the port it bound; a stale one would be read as
	// this run's port.
	if err := os.Remove(filepath.Join(profile.Dir, devToolsActivePort)); err != nil && !os.IsNotExist(err) {
		return prepared, fmt.Errorf("remove stale %s: %w", devToolsActivePort, err)
	}
	if profile.From == "" {
		return prepared, nil
	}
	info, err := os.Stat(profile.From)
	if err != nil {
		return prepared, fmt.Errorf("profile.from: %w", err)
	}
	if info.IsDir() {
		if err := os.CopyFS(profile.Dir, os.DirFS(profile.From)); err != nil {
			return prepared, fmt.Errorf("seed profile %s from %s: %w", profile.Dir, profile.From, err)
		}
		return prepared, nil
	}
	if filepath.Ext(profile.From) != ".json" {
		return prepared, fmt.Errorf("profile.from %s is neither a profile directory nor a Playwright storage-state .json file", profile.From)
	}
	prepared.StorageState = profile.From
	return prepared, nil
}

// storageState is Playwright's `storageState()` file format.
type storageState struct {
	Cookies []struct {
		Name     string  `json:"name"`
		Value    string  `json:"value"`
		Domain   string  `json:"domain"`
		Path     string  `json:"path"`
		Expires  float64 `json:"expires"`
		HTTPOnly bool    `json:"httpOnly"`
		Secure   bool    `json:"secure"`
		SameSite string  `json:"sameSite"`
	} `json:"cookies"`
	Origins []struct {
		Origin       string `json:"origin"`
		LocalStorage []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"localStorage"`
	} `json:"origins"`
}

// applyStorageState loads a storage-state file into the running Chrome's
// default context: cookies directly, localStorage by loading each origin in a
// throwaway tab whose document request is fulfilled locally, so the origin's
// server need not be running yet.
func applyStorageState(ctx context.Context, wsURL, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read storage state: %w", err)
	}
	var state storageState
	if err := json.Unmarshal(raw, &state); err != nil {
		return fmt.Errorf("parse storage state %s: %w", path, err)
	}
	allocCtx, cancelAlloc := chromedp.NewRemoteAllocator(ctx, wsURL, chromedp.NoModifyURL)
	defer cancelAlloc()
	// With a remote allocator this opens (and on cancel closes) a new tab;
	// it never closes the browser.
	tabCtx, cancelTab := chromedp.NewContext(allocCtx)
	defer cancelTab()

	if err := chromedp.Run(tabCtx, setCookies(state)); err != nil {
		return fmt.Errorf("apply storage-state cookies from %s: %w", path, err)
	}
	for _, origin := range state.Origins {
		items := map[string]string{}
		for _, item := range origin.LocalStorage {
			items[item.Name] = item.Value
		}
		if err := seedLocalStorage(tabCtx, origin.Origin, items); err != nil {
			return fmt.Errorf("apply storage-state localStorage for %s: %w", origin.Origin, err)
		}
	}
	return nil
}

func setCookies(state storageState) chromedp.Action {
	cookies := make([]*network.CookieParam, 0, len(state.Cookies))
	for _, c := range state.Cookies {
		param := &network.CookieParam{Name: c.Name, Value: c.Value, Domain: c.Domain, Path: c.Path, HTTPOnly: c.HTTPOnly, Secure: c.Secure}
		if c.SameSite != "" {
			param.SameSite = network.CookieSameSite(c.SameSite)
		}
		if c.Expires > 0 {
			expires := cdp.TimeSinceEpoch(secondsToTime(c.Expires))
			param.Expires = &expires
		}
		cookies = append(cookies, param)
	}
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if len(cookies) == 0 {
			return nil
		}
		return network.SetCookies(cookies).Do(ctx)
	})
}

func seedLocalStorage(tabCtx context.Context, origin string, items map[string]string) error {
	if len(items) == 0 {
		return nil
	}
	fulfilErr := make(chan error, 1)
	chromedp.ListenTarget(tabCtx, func(ev any) {
		paused, ok := ev.(*fetch.EventRequestPaused)
		if !ok {
			return
		}
		go func() {
			body := base64.StdEncoding.EncodeToString([]byte("<!doctype html><title>storage seed</title>"))
			fulfilErr <- chromedp.Run(tabCtx, fetch.FulfillRequest(paused.RequestID, 200).WithBody(body).
				WithResponseHeaders([]*fetch.HeaderEntry{{Name: "Content-Type", Value: "text/html"}}))
		}()
	})
	encoded, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return chromedp.Run(tabCtx,
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{{URLPattern: origin + "/*"}}),
		chromedp.Navigate(origin+"/"),
		chromedp.ActionFunc(func(context.Context) error { return <-fulfilErr }),
		chromedp.Evaluate(fmt.Sprintf(`Object.entries(%s).forEach(([k, v]) => localStorage.setItem(k, v))`, encoded), nil),
		fetch.Disable(),
	)
}

func secondsToTime(seconds float64) time.Time {
	return time.Unix(0, int64(seconds*float64(time.Second)))
}
