package procfile

// Test-only seams. This file is compiled only into the package's test binary,
// so these identifiers are not part of the public API.

// HomeForPath exposes homeForPath for external tests.
var HomeForPath = homeForPath

// ParseEnvOutput exposes parseEnvOutput for external tests.
var ParseEnvOutput = parseEnvOutput

// SetShellForHome substitutes the login-shell resolver used by LoadUserEnv and
// returns a function that restores the previous resolver.
func SetShellForHome(f func(home string) string) (restore func()) {
	prev := shellForHome
	shellForHome = f
	return func() { shellForHome = prev }
}

// RestartProc triggers a restart of the named process, exercising the same path
// as a control-socket restart request, for external tests.
func (s *Supervisor) RestartProc(name string) {
	if m, ok := s.byName[name]; ok {
		s.restartProc(m)
	}
}
