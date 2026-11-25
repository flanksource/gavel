package fixtures

import (
	"context"
	"errors"

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

	Describe("config rendering", func() {
		BeforeEach(func() {
			Expect(step.Scope.SetOutputs("databases", map[string]any{"rows": []any{
				map[string]any{"name": "master"}, map[string]any{"name": "uat"},
			}})).To(Succeed())
		})

		render := func(config string) (string, error) {
			step.Spec.Config = []byte(config)
			out, err := renderConfig(step, nil)
			return string(out), err
		}

		It("replaces a {$cel} object with its typed result and templates the strings around it", func() {
			out, err := render(`{"title":"on {{.env.HOST}}","enum":{"$cel":"databases.rows.map(r, r.name)"},"n":{"$cel":"size(databases.rows) + 1"}}`)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(MatchJSON(`{"title":"on 127.0.0.1","enum":["master","uat"],"n":3}`))
		})

		DescribeTable("fails naming the config path",
			func(config, message string) {
				_, err := render(config)
				Expect(err).To(MatchError(ContainSubstring(message)))
			},
			Entry("a $cel beside another key", `{"a":{"b":{"$cel":"1","x":2}}}`, "a: b: $cel must be the only key of its object, got 2 keys"),
			Entry("a $cel that is not an expression", `{"a":{"$cel":3}}`, "a: $cel must be a CEL expression, got 3"),
			Entry("an expression that does not evaluate", `{"a":[{"$cel":"nope.rows"}]}`, "a: [0]: $cel nope.rows:"),
		)
	})
})

var _ = Describe("StepScope teardown", func() {
	It("runs cleanup registered after teardown and rejects a late temp dir", func() {
		scope := NewStepScope(nil)
		_, err := scope.Teardown(context.Background())
		Expect(err).ToNot(HaveOccurred())
		called := false
		Expect(scope.AddCleanup("late", func(context.Context) error { called = true; return errors.New("late cleanup failed") })).To(MatchError("late cleanup failed"))
		Expect(called).To(BeTrue())
		_, err = scope.TempDir()
		Expect(err).To(HaveOccurred())
	})
})
