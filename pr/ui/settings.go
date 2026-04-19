package ui

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/commons/logger"
)

type UISettings struct {
	Repos  []string `json:"repos,omitempty"`
	Author string   `json:"author,omitempty"`
	Any    bool     `json:"any,omitempty"`
	Bots   bool     `json:"bots,omitempty"`
	// IgnoredOrgs are GitHub org logins the user has chosen to hide from
	// the header chooser and exclude from ResolveDefaultOrg. Persists
	// across daemon restarts so a user's "don't show this org" decision
	// sticks — common example is a legacy personal org the user still
	// belongs to but doesn't care about.
	IgnoredOrgs []string `json:"ignoredOrgs,omitempty"`
}

var settingsPath = filepath.Join(os.Getenv("HOME"), ".config", "gavel", "ui.settings.json")

func LoadSettings() UISettings {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return UISettings{}
	}
	var s UISettings
	if err := json.Unmarshal(data, &s); err != nil {
		logger.Warnf("failed to parse %s: %v", settingsPath, err)
		return UISettings{}
	}
	return s
}

func SaveSettings(s UISettings) {
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		logger.Warnf("failed to create config dir: %v", err)
		return
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		logger.Warnf("failed to marshal settings: %v", err)
		return
	}
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		logger.Warnf("failed to write %s: %v", settingsPath, err)
	}
}
