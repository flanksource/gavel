package repomap

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

func (conf *ArchConf) Pretty() api.Text {
	if conf == nil {
		return clicky.Text("(empty configuration)")
	}

	t := clicky.Text(conf.repoPath).NewLine()
	indent := "  "

	// Git configuration
	hasGit := conf.Git.Commits.Enabled || len(conf.Git.Commits.AllowedTypes) > 0
	if hasGit {
		t = t.Append("📝 Git Configuration", "font-bold text-blue-600").NewLine()
		if conf.Git.Commits.Enabled {
			t = t.Append(indent + "Commits Validation: Enabled").NewLine()
			if len(conf.Git.Commits.AllowedTypes) > 0 {
				t = t.Append(indent+"  Allowed Types: ", "text-muted").Append(strings.Join(conf.Git.Commits.AllowedTypes, ", ")).NewLine()
			}
			if len(conf.Git.Commits.Blocklist) > 0 {
				t = t.Append(indent+"  Blocklist: ", "text-muted").Append(strings.Join(conf.Git.Commits.Blocklist, ", ")).NewLine()
			}
			if conf.Git.Commits.RequiredScope {
				t = t.Append(indent + "  Required Scope: Yes").NewLine()
			}
			if conf.Git.Commits.RequiredReference {
				t = t.Append(indent + "  Required Reference: Yes").NewLine()
			}
		}
		t = t.NewLine()
	}

	// Build configuration
	hasBuild := conf.Build.Enabled || conf.Build.Tool != "" || len(conf.Build.Commands) > 0
	if hasBuild {
		t = t.Append("🔨 Build Configuration", "font-bold text-green-600").NewLine()
		if conf.Build.Enabled {
			t = t.Append(indent + "Enabled: Yes").NewLine()
		}
		if conf.Build.Tool != "" {
			t = t.Append(indent+"Tool: ", "text-muted").Append(conf.Build.Tool).NewLine()
		}
		if len(conf.Build.Commands) > 0 {
			t = t.Append(indent + "Commands:").NewLine()
			for name, cmd := range conf.Build.Commands {
				t = t.Append(indent+"  "+name+": ", "text-muted").Append(cmd).NewLine()
			}
		}
		t = t.NewLine()
	}

	// Golang configuration
	hasGolang := conf.Golang.Enabled || len(conf.Golang.Blocklist) > 0 || conf.Golang.Testing.Tool != ""
	if hasGolang {
		t = t.Append("🐹 Golang Configuration", "font-bold text-cyan-600").NewLine()
		if conf.Golang.Enabled {
			t = t.Append(indent + "Enabled: Yes").NewLine()
		}
		if len(conf.Golang.Blocklist) > 0 {
			t = t.Append(indent+"Blocklist: ", "text-muted").Append(strings.Join(conf.Golang.Blocklist, ", ")).NewLine()
		}
		if conf.Golang.Testing.Tool != "" {
			t = t.Append(indent+"Testing Tool: ", "text-muted").Append(conf.Golang.Testing.Tool).NewLine()
			if len(conf.Golang.Testing.Files) > 0 {
				t = t.Append(indent+"  Files: ", "text-muted").Append(strings.Join(conf.Golang.Testing.Files, ", ")).NewLine()
			}
		}
		t = t.NewLine()
	}

	// Scopes configuration
	if len(conf.Scopes.Rules) > 0 || len(conf.Scopes.AllowedScopes) > 0 {
		t = t.Append("🎯 Scopes Configuration", "font-bold text-purple-600").NewLine()

		if len(conf.Scopes.AllowedScopes) > 0 {
			t = t.Append(indent+"Allowed Scopes: ", "text-muted").Append(strings.Join(conf.Scopes.AllowedScopes, ", ")).NewLine()
		}

		if len(conf.Scopes.Rules) > 0 {
			t = t.Append(indent + "Rules:").NewLine()
			for scope, rules := range conf.Scopes.Rules {
				t = t.Append(indent+"  "+string(scope)+": ", "font-medium")
				t = t.Append(fmt.Sprintf("%d pattern(s)", len(rules)), "text-muted").NewLine()
				for i, rule := range rules {
					if i < 5 { // Limit to first 5 rules per scope to avoid cluttering
						t = t.Append(indent + "    - " + rule.Path).NewLine()
					} else if i == 5 {
						t = t.Append(indent+"    ... and ", "text-muted").Append(fmt.Sprintf("%d more", len(rules)-5), "text-muted").NewLine()
						break
					}
				}
			}
		}
		t = t.NewLine()
	}

	// Technology configuration
	if len(conf.Tech.Rules) > 0 {
		t = t.Append("⚙️  Technology Configuration", "font-bold text-orange-600").NewLine()
		t = t.Append(indent + "Rules:").NewLine()
		for tech, rules := range conf.Tech.Rules {
			t = t.Append(indent+"  "+string(tech)+": ", "font-medium")
			t = t.Append(fmt.Sprintf("%d pattern(s)", len(rules)), "text-muted").NewLine()
			for i, rule := range rules {
				if i < 5 { // Limit to first 5 rules per tech to avoid cluttering
					t = t.Append(indent + "    - " + rule.Path).NewLine()
				} else if i == 5 {
					t = t.Append(indent+"    ... and ", "text-muted").Append(fmt.Sprintf("%d more", len(rules)-5), "text-muted").NewLine()
					break
				}
			}
		}
	}

	return t
}
