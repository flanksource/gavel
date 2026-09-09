package labels

// builtins are the well-known label presentations, applied when nothing is
// stored for a name. They are deliberately NOT seeded into todo_labels: they
// are the third link in the resolution chain, so the table only ever holds what
// someone actually edited, editing a builtin inserts a row that shadows it, and
// deleting that row restores the default. No seed migration, no drift.
//
// Icons are keys in clicky's icon registry (api/icons.All), which supplies both
// the terminal glyph and the Iconify name — hence "debug" for bug and "lock"
// for security, since the registry has no "bug" or "shield".
var builtins = map[string]Definition{
	"bug":      {Name: "bug", Color: ColorRed, Icon: "debug", Description: "Something is broken.", Scope: ScopeBuiltin},
	"security": {Name: "security", Color: ColorRose, Icon: "lock", Description: "Security-sensitive work.", Scope: ScopeBuiltin},
	"docs":     {Name: "docs", Color: ColorSky, Icon: "docs", Description: "Documentation.", Scope: ScopeBuiltin},
	"perf":     {Name: "perf", Color: ColorAmber, Icon: "performance", Description: "Performance work.", Scope: ScopeBuiltin},
	"test":     {Name: "test", Color: ColorEmerald, Icon: "test", Description: "Test coverage or test infrastructure.", Scope: ScopeBuiltin},
	"refactor": {Name: "refactor", Color: ColorViolet, Icon: "refactor", Description: "Internal restructuring, no behaviour change.", Scope: ScopeBuiltin},
	"ui":       {Name: "ui", Color: ColorFuchsia, Icon: "style", Description: "User interface work.", Scope: ScopeBuiltin},
	"api":      {Name: "api", Color: ColorBlue, Icon: "http", Description: "API surface.", Scope: ScopeBuiltin},
	"ci":       {Name: "ci", Color: ColorTeal, Icon: "ci", Description: "Build or continuous integration.", Scope: ScopeBuiltin},
	"breaking": {Name: "breaking", Color: ColorOrange, Icon: "warning", Description: "Breaking change.", Scope: ScopeBuiltin},
}

// Builtins returns a copy of the built-in definitions, keyed by name.
func Builtins() map[string]Definition {
	out := make(map[string]Definition, len(builtins))
	for name, def := range builtins {
		out[name] = def
	}
	return out
}

// BuiltinNames returns the built-in label names, sorted.
func BuiltinNames() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sortNames(names)
	return names
}
