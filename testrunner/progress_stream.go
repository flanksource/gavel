package testrunner

import (
	"io"

	"github.com/flanksource/gavel/testrunner/parsers"
)

type progressStream interface {
	io.Writer
	Flush()
}

// newProgressStream returns a stdout tap that turns a running package's output
// into live progress for the streamer, or nil when the framework has none.
func (o *TestOrchestrator) newProgressStream(framework Framework, pkgPath string) progressStream {
	if o.streamer == nil {
		return nil
	}
	switch framework {
	case parsers.Ginkgo:
		return parsers.NewGinkgoStreamWriter(pkgPath, o.streamer)
	case parsers.Vitest:
		return parsers.NewVitestStreamWriter(o.WorkDir, pkgPath, o.streamer)
	}
	return nil
}
