package jsonlint

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJSON(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "JSON Linter")
}

var _ = Describe("JSON linter", func() {
	var adapter *JSON
	var write func(string, string) string
	var run func(string, string) ([]models.Violation, error)
	BeforeEach(func() {
		adapter = New(GinkgoT().TempDir())
		write = func(name, contents string) string {
			path := filepath.Join(adapter.WorkDir, name)
			Expect(os.MkdirAll(filepath.Dir(path), 0o700)).To(Succeed())
			Expect(os.WriteFile(path, []byte(contents), 0o600)).To(Succeed())
			return path
		}
		run = func(name, contents string) ([]models.Violation, error) {
			adapter.Files = []string{write(name, contents)}
			return adapter.Run(commonsContext.NewContext(context.Background()), nil)
		}
	})

	DescribeTable("checks syntax without a schema", func(name, contents string, count int) {
		violations, err := run(name, contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(count))
		Expect(adapter.GetFileCount()).To(Equal(1))
		for _, violation := range violations {
			Expect(violation.Source).To(Equal("json"))
			Expect(violation.Rule.Method).To(Equal("syntax"))
			Expect(violation.Severity).To(Equal(models.SeverityError))
			Expect(violation.Line).To(BeNumerically(">", 0))
		}
	},
		Entry("JSON", "data.json", `{"enabled":true}`, 0),
		Entry("JSON scalar", "data.json", `null`, 0),
		Entry("bad JSON", "data.json", `{"enabled":}`, 1),
		Entry("trailing JSON", "data.json", `{} {}`, 1),
		Entry("YAML", "data.yaml", "enabled: true\n", 0),
		Entry("YML", "data.yml", "enabled: true\n", 0),
		Entry("YAML numeric keys without a schema", "data.yaml", "ports:\n  80: http\n", 0),
		Entry("YAML special numbers without a schema", "data.yaml", "value: .nan\n", 0),
		Entry("bad YAML", "data.yaml", "enabled: [\n", 1),
		Entry("duplicate YAML keys", "data.yaml", "enabled: true\nenabled: false\n", 1),
		Entry("YAML documents", "data.yaml", "enabled: true\n---\nenabled: false\n", 0),
		Entry("NDJSON with blank lines", "data.ndjson", "{}\n\n[]\nnull\n", 0),
		Entry("bad NDJSON records", "data.ndjson", "{}\n\n{bad}\n{} {}\n", 2),
		Entry("large NDJSON record", "data.ndjson", `{"text":"`+strings.Repeat("x", 70000)+`"}`, 0),
	)

	It("reports original NDJSON line numbers and continues after malformed records", func() {
		violations, err := run("data.ndjson", "{}\n\n{bad}\n{bad}\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveExactElements(HaveField("Line", 3), HaveField("Line", 4)))
	})

	DescribeTable("validates using an explicit schema", func(name, contents string, line int) {
		adapter.Schema = write("schema.json", `{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`)
		violations, err := run(name, contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveExactElements(And(HaveField("Line", line), HaveField("Source", "json"), HaveField("Rule.Method", "schema"))))
		Expect(*violations[0].Message).To(ContainSubstring("/name"))
	},
		Entry("JSON source location", "data.json", "{\n  \"name\": 123\n}", 2),
		Entry("YAML source location", "data.yaml", "# example\nname: 123\n", 2),
		Entry("each YAML document", "data.yaml", "name: valid\n---\nname: 123\n", 3),
		Entry("each NDJSON record", "data.ndjson", "{\"name\":\"valid\"}\n\n{\"name\":123}\n", 3),
	)

	It("uses file-relative YAML headers and schema-relative references", func() {
		write("nested/schema.json", `{"$ref":"types.json#/$defs/name"}`)
		write("nested/types.json", `{"$defs":{"name":{"type":"object","required":["name"]}}}`)
		violations, err := run("nested/data.yaml", "# yaml-language-server: $schema=./schema.json\nwrong: value\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(1))
		Expect(*violations[0].Message).To(ContainSubstring("name"))
	})

	It("lets the explicit schema override a broken header", func() {
		adapter.Schema = write("schema.json", `true`)
		violations, err := run("data.yaml", "# yaml-language-server: $schema=missing.json\nname: valid\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(BeEmpty())
	})

	DescribeTable("uses object schema declarations", func(name, contents string) {
		write("nested/schema.json", `{"required":["name"]}`)
		violations, err := run(name, contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(1))
	}, Entry("JSON", "nested/data.json", `{"$schema":"schema.json"}`),
		Entry("NDJSON", "nested/data.ndjson", "{\"$schema\":\"schema.json\",\"name\":\"valid\"}\n{\"$schema\":\"schema.json\"}\n"))

	DescribeTable("preserves integer precision", func(name, contents string) {
		adapter.Schema = write("schema.json", `{"properties":{"number":{"const":9007199254740993}}}`)
		violations, err := run(name, contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(BeEmpty())
	}, Entry("JSON", "data.json", `{"number":9007199254740993}`), Entry("YAML", "data.yaml", "number: 9007199254740993\n"))

	DescribeTable("supports declared schema dialects", func(dialect string) {
		adapter.Schema = write("schema.json", `{"$schema":"`+dialect+`","type":"string"}`)
		violations, err := run("data.json", `123`)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(1))
	}, Entry("draft-07", "http://json-schema.org/draft-07/schema#"), Entry("2020-12", "https://json-schema.org/draft/2020-12/schema"))

	DescribeTable("fails on unusable schemas", func(schema string) {
		adapter.Schema = write("schema.json", schema)
		_, err := run("data.json", `{}`)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("schema.json"))
	}, Entry("malformed", `{`), Entry("invalid", `{"type":"unknown"}`), Entry("missing reference", `{"$ref":"missing.json"}`))

	It("fails on a missing declared schema", func() {
		_, err := run("data.yaml", "# yaml-language-server: $schema=missing.json\nname: value\n")
		Expect(err).To(MatchError(ContainSubstring("missing.json")))
	})

	It("loads the declared schema even for a YAML file containing only a header", func() {
		_, err := run("data.yaml", "# yaml-language-server: $schema=missing.json\n")
		Expect(err).To(MatchError(ContainSubstring("missing.json")))
	})

	DescribeTable("rejects empty YAML schema declarations", func(reference string) {
		_, err := run("data.yaml", "# yaml-language-server: $schema="+reference+"\nname: example\n")
		Expect(err).To(HaveOccurred())
	}, Entry("bare", ""), Entry("quoted", `""`), Entry("blank", `" "`))

	It("loads remote schemas and references once per run", func() {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			if r.URL.Path == "/schema.json" {
				_, err := w.Write([]byte(`{"$ref":"types.json"}`))
				Expect(err).NotTo(HaveOccurred())
				return
			}
			_, err := w.Write([]byte(`{"type":"object","required":["name"]}`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		adapter.Schema = server.URL + "/schema.json"
		violations, err := run("data.ndjson", "{}\n{}\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(2))
		Expect(requests.Load()).To(Equal(int32(2)))
	})

	It("surfaces HTTP failures", func() {
		server := httptest.NewServer(http.NotFoundHandler())
		DeferCleanup(server.Close)
		adapter.Schema = server.URL + "/schema.json"
		_, err := run("data.json", `{}`)
		Expect(err).To(MatchError(ContainSubstring("404")))
	})

	It("cancels an in-flight remote schema request", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
		DeferCleanup(server.Close)
		adapter.Schema = server.URL + "/schema.json"
		adapter.Files = []string{write("data.json", `{}`)}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		DeferCleanup(cancel)
		_, err := adapter.Run(commonsContext.NewContext(ctx), nil)
		Expect(err).To(HaveOccurred())
		Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))
	})

	It("expands directory and glob targets without duplicate files", func() {
		write("nested/valid.json", `{}`)
		invalid := write("nested/bad.yaml", "broken: [\n")
		adapter.Files = []string{"nested", "nested/*.yaml"}
		violations, err := adapter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveExactElements(HaveField("File", invalid)))
		Expect(adapter.GetFileCount()).To(Equal(2))
	})

	It("respects gitignore for automatic discovery and validates explicitly named ignored files", func() {
		Expect(os.Mkdir(filepath.Join(adapter.WorkDir, ".git"), 0o700)).To(Succeed())
		write(".gitignore", "ignored/\n")
		write("valid.json", `{}`)
		ignored := write("ignored/bad.json", `{bad}`)
		violations, err := adapter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(BeEmpty())
		Expect(adapter.GetFileCount()).To(Equal(1))
		adapter.Files = []string{ignored}
		violations, err = adapter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(1))
	})

	It("compiles an explicit schema even for an empty NDJSON file", func() {
		adapter.Schema = write("schema.json", `{"type":"unknown"}`)
		_, err := run("data.ndjson", "")
		Expect(err).To(HaveOccurred())
	})

	DescribeTable("handles YAML aliases and exact scalar values", func(contents, schema string, count int) {
		adapter.Schema = write("schema.json", schema)
		violations, err := run("data.yaml", contents)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(count))
	},
		Entry("merge keys", "defaults: &defaults\n  name: valid\nvalue:\n  <<: *defaults\n", `{"properties":{"value":{"required":["name"]}}}`, 0),
		Entry("cyclic aliases", "value: &value [*value]\n", `true`, 1),
		Entry("integer above uint64", "value: !!int 18446744073709551616\n", `{"properties":{"value":{"const":18446744073709551616}}}`, 0),
		Entry("non-JSON numeric value", "value: .nan\n", `true`, 1),
		Entry("YAML exponent with trailing decimal", "value: 1.e+3\n", `{"properties":{"value":{"const":1000}}}`, 0),
		Entry("YAML float with leading zero", "value: 01.50\n", `{"properties":{"value":{"const":1.5}}}`, 0),
		Entry("exact decimal", "value: .123456789123456789\n", `{"properties":{"value":{"const":0.123456789123456789}}}`, 0),
	)
})
