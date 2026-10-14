package fixtures

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StepContext rendering", func() {
	var step *StepContext

	BeforeEach(func() {
		scope := NewStepScope([]string{"HOST=127.0.0.1", "PORT=8080"})
		Expect(scope.SetOutputs("api", map[string]any{"path": "/v1"})).To(Succeed())
		step = &StepContext{Scope: scope, TmpDir: "/tmp/x", Test: FixtureTest{SourceDir: "/src"}}
	})

	It("Template renders Go templates only and leaves every $ to the shell", func() {
		out, err := step.Template(`for f in *; do echo $f $? $1 $$; done; echo $UNSET {{.env.PORT}}{{.steps.api.path}}`)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal(`for f in *; do echo $f $? $1 $$; done; echo $UNSET 8080/v1`))
	})

	It("Expand renders templates, then expands $NAME and ${NAME} from the fixture Env", func() {
		out, err := step.Expand(`http://$HOST:${PORT}{{.steps.api.path}} cost=$5 pid=$$ status=$? arg=$1 lone=$ {{.tmp}}`)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal(`http://127.0.0.1:8080/v1 cost=$5 pid=$$ status=$? arg=$1 lone=$ /tmp/x`))
	})

	It("Expand fails on an unset variable instead of expanding it to nothing", func() {
		_, err := step.Expand(`http://$HOST:${MISSING_PORT}/$ALSO_MISSING`)
		Expect(err).To(MatchError(ContainSubstring("MISSING_PORT, ALSO_MISSING")))
	})

	It("fails a template that references a value that is not set", func() {
		_, err := step.Template(`{{.env.NOPE}}`)
		Expect(err).To(MatchError(ContainSubstring("not set")))
	})
})
