package fixtures

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PTY input script", func() {
	run := func(script string, input ...PTYInput) (*Capture, error) {
		return CaptureANSI(CaptureOptions{Width: 80, Height: 24, Command: []string{"/bin/sh", "-c", script}, Input: input})
	}

	It("types each answer once its prompt appears", func() {
		capture, err := run(`printf 'name? '; read n; printf 'colour? '; read c; echo "hi $n in $c"`,
			PTYInput{Expect: "name?", Send: "ada", Keys: []string{"enter"}},
			PTYInput{Expect: "colour?", Send: "teal", Keys: []string{"enter"}},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(capture.ExitCode).To(Equal(0))
		Expect(stripFixtureANSI(capture.Raw())).To(ContainSubstring("hi ada in teal"))
	})

	It("matches each expectation only after the previous one", func() {
		// "ready" is printed once: the second item must not match the first
		// item's text again, so the command exits with it unmet.
		_, err := run(`printf 'ready '; read a; echo "got $a"`,
			PTYInput{Expect: "ready", Send: "x", Keys: []string{"enter"}},
			PTYInput{Expect: "ready", Send: "y", Keys: []string{"enter"}},
		)

		Expect(err).To(MatchError(And(
			ContainSubstring(`pty input 2: expected "ready" before exit`),
			ContainSubstring("got x"),
		)))
	})

	It("refuses an unknown key before starting the command", func() {
		_, err := run(`echo never`, PTYInput{Expect: "never", Keys: []string{"return"}})

		Expect(err).To(MatchError(ContainSubstring(`input 1: unknown key "return"`)))
	})

	It("answers a cursor-position query split across reads", func() {
		var term bytes.Buffer
		script := newPTYScript(nil, &term)

		Expect(script.feed([]byte("prompt\x1b["))).To(Succeed())
		Expect(term.String()).To(BeEmpty())
		Expect(script.feed([]byte("6n"))).To(Succeed())

		Expect(term.String()).To(Equal("\x1b[1;1R"))
	})

	// termenv asks for the background (OSC 11) and then the cursor position,
	// and takes a first reply that is not OSC to mean the terminal has none:
	// it stops reading, and the OSC reply is left for the program's input.
	It("answers queries in the order they were asked", func() {
		for range 20 {
			var term bytes.Buffer
			script := newPTYScript(nil, &term)

			Expect(script.feed([]byte("\x1b]11;?\x1b\\\x1b[6n"))).To(Succeed())

			Expect(term.String()).To(Equal("\x1b]11;rgb:0000/0000/0000\x1b\\\x1b[1;1R"))
		}
	})
})
