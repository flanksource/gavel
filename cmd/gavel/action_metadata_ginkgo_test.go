// Checks the GitHub Action's metadata in action.yml: its inputs' defaults and what its
// "Run gavel" step runs and is given.
package main

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

type actionStep struct {
	Name string            `yaml:"name"`
	Run  string            `yaml:"run"`
	Env  map[string]string `yaml:"env"`
}

type actionMetadata struct {
	Inputs map[string]struct {
		Default string `yaml:"default"`
	} `yaml:"inputs"`
	Runs struct {
		Steps []actionStep `yaml:"steps"`
	} `yaml:"runs"`
}

// runGavelStep reads action.yml and returns it with its "Run gavel" step.
func runGavelStep() (actionMetadata, actionStep) {
	contents, err := os.ReadFile(filepath.Join("..", "..", "action.yml"))
	Expect(err).NotTo(HaveOccurred())
	var action actionMetadata
	Expect(yaml.Unmarshal(contents, &action)).To(Succeed())
	for _, step := range action.Runs.Steps {
		if step.Name == "Run gavel" {
			return action, step
		}
	}
	Fail(`action.yml has no "Run gavel" step`)
	return action, actionStep{}
}

var _ = Describe("GitHub Action metadata", func() {
	It("defaults to a twenty minute watchdog around the gavel invocation", func() {
		action, step := runGavelStep()
		Expect(action.Inputs).To(HaveKey("timeout-minutes"))
		Expect(action.Inputs["timeout-minutes"].Default).To(Equal("20"))

		Expect(step.Run).NotTo(BeEmpty())
		Expect(strings.Join(strings.Fields(step.Run), " ")).To(ContainSubstring("gavel action-watchdog --timeout \"${GAVEL_TIMEOUT_MINUTES}m\" --"))
	})

	// Unauthenticated, GitHub's API allows 60 requests an hour per runner IP, so the tools gavel
	// installs, such as golangci-lint, would often fail to resolve their versions.
	It("gives gavel the action's GitHub token for the tools it installs", func() {
		_, step := runGavelStep()
		Expect(step.Env).To(HaveKeyWithValue("GH_TOKEN", "${{ inputs.github-token }}"))
	})
})
