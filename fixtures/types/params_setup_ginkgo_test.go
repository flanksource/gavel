package types

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

var _ = Describe("a setup connection picked through params", func() {
	const server = "https://cluster.example.invalid:6443"

	// writeKubernetesConnection saves a kubernetes connection in the fixture
	// directory's .connections, the catalog a `format: connection` picker lists.
	writeKubernetesConnection := func(dir, name string) {
		kubeconfig := strings.Join([]string{
			"apiVersion: v1", "kind: Config", "current-context: example",
			"clusters: [{name: example, cluster: {server: " + server + "}}]",
			"contexts: [{name: example, context: {cluster: example, user: example}}]",
			"users: [{name: example, user: {token: example-token}}]",
		}, "\n")
		body := "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\nmetadata:\n  name: " + name +
			"\nspec:\n  type: kubernetes\n  certificate:\n    value: |\n" + indent(kubeconfig, "      ") + "\n"
		Expect(os.MkdirAll(filepath.Join(dir, ".connections"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, ".connections", name+".yaml"), []byte(body), 0o600)).To(Succeed())
	}

	It("resolves the picked kubernetes connection and hands later steps its kubeconfig", func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		writeKubernetesConnection(dir, "acme-cluster")
		path := filepath.Join(dir, "scale.fixture.yaml")
		Expect(os.WriteFile(path, []byte(`
kind: Fixture
params:
  type: object
  required: [k8s]
  properties:
    k8s: {type: string, format: connection, x-connection-types: [kubernetes]}
setup:
  connections:
    kubernetes: {$cel: "params.k8s"}
steps:
  - id: kubeconfig
    kind: exec
    script: 'cat "$KUBECONFIG"'
`), 0o600)).To(Succeed())

		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: dir, MaxWorkers: 1,
			Params: map[string]any{"k8s": "connection://acme-cluster"}})
		Expect(err).ToNot(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).ToNot(HaveOccurred())
		var stdout, failure string
		tree.Walk(func(node *fixtures.FixtureNode) {
			if node.Results != nil && node.Name == "kubeconfig" {
				stdout, failure = node.Results.Stdout, node.Results.Error
			}
		})
		Expect(tree.GetStats().Passed).To(Equal(2), failure)
		Expect(stdout).To(ContainSubstring(server))
	})
})

func indent(text, prefix string) string {
	return prefix + strings.ReplaceAll(text, "\n", "\n"+prefix)
}
