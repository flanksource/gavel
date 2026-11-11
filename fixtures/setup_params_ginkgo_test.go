package fixtures

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons-db/connection"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("a setup that reads params", func() {
	const answered = "from-params"

	// answerParams stands in for the form kind (kinds/form imports this package):
	// the params step answers with greeting = answered.
	answerParams := func() {
		useKind(fakeKind{name: FormKind, execute: passing(map[string]any{"greeting": answered}, nil)})
	}

	It("is prepared after the params form, and its environment reaches the later steps", func() {
		answerParams()
		var seen string
		useKind(fakeKind{name: "probe", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			cfg, err := KindConfig[map[string]any](step)
			seen, _ = cfg["url"].(string)
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, err
		}})
		path := writeDocument("params-setup.fixture.yaml", `
kind: Fixture
params:
  type: object
  properties:
    greeting: {type: string}
setup:
  envVars:
    - name: GREETING
      value: "{{.params.greeting}}"
steps:
  - { id: use, kind: probe, url: "{{.env.GREETING}}" }
`)
		runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: filepath.Dir(path), MaxWorkers: 1})
		Expect(err).ToNot(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).ToNot(HaveOccurred())
		Expect(tree.GetStats().Passed).To(Equal(2))
		Expect(seen).To(Equal(answered))
	})

	It("is prepared after reused params when the run resumes past the params form", func() {
		asked := 0
		useKind(fakeKind{name: FormKind, execute: func(context.Context, *StepContext) (StepOutcome, error) {
			asked++
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
		}})
		var seen string
		useKind(fakeKind{name: "probe", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			cfg, err := KindConfig[map[string]any](step)
			seen, _ = cfg["url"].(string)
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, err
		}})
		path := writeDocument("resumed-params-setup.fixture.yaml", `
kind: Fixture
params:
  type: object
  properties:
    greeting: {type: string}
setup:
  envVars:
    - name: GREETING
      value: "{{.params.greeting}}"
steps:
  - { id: use, kind: probe, url: "{{.env.GREETING}}" }
`)
		resume := &Resume{FromStep: "use", Prior: map[string]PriorStep{ParamsStepID: {Outputs: map[string]any{"greeting": answered}}}}
		runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: filepath.Dir(path), MaxWorkers: 1, Resume: resume})
		Expect(err).ToNot(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).ToNot(HaveOccurred())
		Expect(tree.GetStats().Passed).To(Equal(2))
		Expect(asked).To(BeZero())
		Expect(seen).To(Equal(answered))
	})

	It("leaves the connection unconfigured when its $cel value is nil", func() {
		answerParams()
		useKind(fakeKind{name: "probe", execute: passing(nil, nil)})
		path := writeDocument("nil-connection.fixture.yaml", `
kind: Fixture
params:
  type: object
  properties:
    greeting: {type: string}
    cluster: {type: string}
setup:
  connections:
    kubernetes: {$cel: "params.cluster"}
steps:
  - { id: use, kind: probe }
`)
		runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: filepath.Dir(path), MaxWorkers: 1})
		Expect(err).ToNot(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).ToNot(HaveOccurred())
		Expect(tree.GetStats().Passed).To(Equal(2))
	})

	const reference = "connection://acme"

	DescribeTable("decodes a setup's connections",
		func(authored string, kubernetes *connection.KubernetesConnection, fromConfigItem *string, keptAsAuthored bool) {
			var spec SetupSpec
			Expect(json.Unmarshal([]byte(authored), &spec)).To(Succeed())
			Expect(spec.Connections.Kubernetes).To(Equal(kubernetes))
			Expect(spec.Connections.FromConfigItem).To(Equal(fromConfigItem))
			Expect(setupReadsParams(&spec)).To(Equal(keptAsAuthored))
		},
		Entry("a reference string as the kind's connection",
			`{"connections":{"kubernetes":"`+reference+`"}}`,
			&connection.KubernetesConnection{KubeconfigConnection: connection.KubeconfigConnection{ConnectionName: reference}}, nil, false),
		Entry("a null as no connection", `{"connections":{"kubernetes":null}}`, nil, nil, false),
		Entry("an object unchanged",
			`{"connections":{"kubernetes":{"connection":"`+reference+`"}}}`,
			&connection.KubernetesConnection{KubeconfigConnection: connection.KubeconfigConnection{ConnectionName: reference}}, nil, false),
		Entry("a string field that is no connection kind as itself", `{"connections":{"fromConfigItem":"item"}}`, nil, new("item"), false),
		Entry("a $cel value kept for the params form", `{"connections":{"kubernetes":{"$cel":"params.cluster"}}}`, nil, nil, true),
	)

	It("refuses a setup key it does not know when parsed", func() {
		var spec SetupSpec
		Expect(json.Unmarshal([]byte(`{"connections":{"kubernetes":{"$cel":"params.cluster"}},"conections":{}}`), &spec)).
			To(MatchError(ContainSubstring(`unknown field "conections"`)))
	})

	DescribeTable("is refused before any step runs",
		func(body, message string) {
			answerParams()
			path := writeDocument("refused.fixture.yaml", body)
			runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: filepath.Dir(path), MaxWorkers: 1})
			Expect(err).ToNot(HaveOccurred())
			_, err = runner.Run()
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("without params to read", `
kind: Fixture
setup:
  envVars: [{name: GREETING, value: "{{.params.greeting}}"}]
steps:
  - { id: use, kind: form }
`, "declares no params:"),
		Entry("when it would relocate the run", `
kind: Fixture
params: {type: object, properties: {dir: {type: string}}}
setup:
  cwd: "{{.params.dir}}"
steps:
  - { id: use, kind: form }
`, "cannot declare checkout: or cwd:"),
		Entry("when a $cel value would relocate the run", `
kind: Fixture
params: {type: object, properties: {dir: {type: string}}}
setup:
  cwd: {$cel: "params.dir"}
steps:
  - { id: use, kind: form }
`, "cannot declare checkout: or cwd:"),
	)
})
