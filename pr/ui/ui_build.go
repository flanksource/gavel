package ui

import (
	"crypto/sha256"
	"encoding/hex"
)

// devUIBuildID is the build id while `--dev` proxies the UI to Vite: there the
// served bundle is Vite's, not the embed, and the dev index.html carries the
// same constant, so a dev tab's page and hello always agree and never reload.
const devUIBuildID = "dev"

// uiBuildID identifies the embedded dashboard bundle. It is stamped into every
// served page (<meta name="gavel-ui-build">) and the /api/events hello, so a
// page loaded from an older bundle sees the mismatch and reloads. prui.js is a
// small shim importing the content-hashed entry chunk, so hashing it together
// with prui.css changes whenever the JS or the CSS does.
var uiBuildID = embeddedUIBuildID()

func embeddedUIBuildID() string {
	hash := sha256.New()
	for _, name := range []string{"dist/prui.js", "dist/prui.css"} {
		data, err := distFS.ReadFile(name)
		if err != nil {
			panic("ui: read embedded " + name + " for the UI build id: " + err.Error()) // static embed; failure is a build error
		}
		hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil))[:12]
}

// uiBuild is the build id this server's /api/events hello reports.
func (s *Server) uiBuild() string {
	if s.devProxy != nil {
		return devUIBuildID
	}
	return uiBuildID
}
