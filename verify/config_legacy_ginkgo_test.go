package verify

import (
	"fmt"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// resolveConfigPath walks a dotted .gavel.yaml path through the live config
// structs using the same reflection the unknown-field inspector uses, so the
// answer is exactly "would the decoder accept this key".
func resolveConfigPath(path string) error {
	inspector := fieldInspector{format: "json", unknown: map[string]bool{}}
	current := reflect.TypeOf(GavelConfig{})
	for _, segment := range strings.Split(path, ".") {
		wire := inspector.wireType(current)
		if wire == nil || wire.Kind() != reflect.Struct {
			return fmt.Errorf("%s: %s is not an object", path, segment)
		}
		fields, _ := inspector.structFields(wire, map[reflect.Type]bool{})
		next, ok := fields[segment]
		if !ok {
			return fmt.Errorf("%s: no field %q", path, segment)
		}
		current = next
	}
	return nil
}

var _ = Describe("legacy config fields", func() {
	It("resolves the path the inspector itself accepts", func() {
		Expect(resolveConfigPath("commit.summary")).To(Succeed())
		Expect(resolveConfigPath("commit.nonsense")).NotTo(Succeed())
	})

	It("points every rename at a field that still exists", func() {
		for path, field := range LegacyConfigFields {
			if field.ReplacedBy == "" {
				continue
			}
			Expect(resolveConfigPath(field.ReplacedBy)).To(Succeed(),
				"%s claims it was renamed to %s, which no longer resolves", path, field.ReplacedBy)
		}
	})

	It("does not claim a field is gone while it still exists", func() {
		for path, field := range LegacyConfigFields {
			Expect(resolveConfigPath(path)).NotTo(Succeed(),
				"%s is listed as retired but the config still accepts it", path)
			if field.ReplacedBy == "" {
				Expect(field.Note).NotTo(BeEmpty(), "%s is removed without saying why", path)
			}
		}
	})

	DescribeTable("renders a hint that names the replacement",
		func(path, expected string) {
			Expect(legacyFieldHints[path]).To(Equal(expected))
		},
		Entry("rename", "commit.summaryPrompt", "renamed to commit.summary"),
		Entry("move to another section", "commit.prContentPrompt", "renamed to pr.content"),
		Entry("removal", "commit.compatibility", "removed; compatibility analysis no longer runs"),
		Entry("removed section", "verify", "removed; use the top-level ai: spec and todos.verify"),
	)

	It("annotates a retired field and leaves an unrecognised one bare", func() {
		var warnings []string
		var cfg GavelConfig
		err := DecodeJSON([]byte(`{"commit":{"summaryPrompt":{},"futureSetting":true}}`), &cfg,
			DecodeOptions{
				Source:  "project.yaml",
				Renames: legacyFieldHints,
				Warn:    func(message string) { warnings = append(warnings, message) },
			})

		Expect(err).NotTo(HaveOccurred())
		Expect(warnings).To(ConsistOf(
			"project.yaml: unknown field commit.futureSetting",
			"project.yaml: unknown field commit.summaryPrompt (renamed to commit.summary)",
		))
	})
})
