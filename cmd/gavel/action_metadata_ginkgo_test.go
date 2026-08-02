package main

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("GitHub Action metadata", func() {
	It("defaults to a twenty minute watchdog around the gavel invocation", func() {
		contents, err := os.ReadFile(filepath.Join("..", "..", "action.yml"))
		Expect(err).NotTo(HaveOccurred())

		var action struct {
			Inputs map[string]struct {
				Default string `yaml:"default"`
			} `yaml:"inputs"`
			Runs struct {
				Steps []struct {
					Name string `yaml:"name"`
					Run  string `yaml:"run"`
				} `yaml:"steps"`
			} `yaml:"runs"`
		}
		Expect(yaml.Unmarshal(contents, &action)).To(Succeed())
		Expect(action.Inputs).To(HaveKey("timeout-minutes"))
		Expect(action.Inputs["timeout-minutes"].Default).To(Equal("20"))

		var runStep string
		for _, step := range action.Runs.Steps {
			if step.Name == "Run gavel" {
				runStep = step.Run
				break
			}
		}
		Expect(runStep).NotTo(BeEmpty())
		Expect(strings.Join(strings.Fields(runStep), " ")).To(ContainSubstring("gavel action-watchdog --timeout \"${GAVEL_TIMEOUT_MINUTES}m\" --"))
	})
})
