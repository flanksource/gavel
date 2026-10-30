package testui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// AssetsMount is where a Server serves the UI bundle: the stable testui.js
// entry and the hashed chunks/*.js it imports relative to itself. A host that
// mounts a Server under a prefix serves the bundle at <prefix>/_testui/.
const AssetsMount = "/_testui"

// BundlePath is the entry module's path under a Server's mount.
const BundlePath = AssetsMount + "/testui.js"

// dist holds the embed build (vite.config.ts): testui.js, chunks/*.js, and the
// unserved library build under lib/.
//
//go:embed all:dist
var distFS embed.FS

var bundleFS = func() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("testui: embedded dist FS: " + err.Error()) // static embed; failure is a build error
	}
	return sub
}()

// handleAssets serves the entry and its chunks. Hashed chunks are immutable;
// the stable entry gets a short cache so a rebuilt binary is picked up.
func handleAssets(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, AssetsMount+"/")
	// Only the embed build is served: not the library build under lib/, and no
	// directory listings. ServeFileFS rejects ".." in the request path.
	if name != "testui.js" && (!strings.HasPrefix(name, "chunks/") || !strings.HasSuffix(name, ".js")) {
		http.NotFound(w, r)
		return
	}
	if strings.HasPrefix(name, "chunks/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	http.ServeFileFS(w, r, bundleFS, name)
}
