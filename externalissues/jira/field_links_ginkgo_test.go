package jira

import (
	"encoding/json"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("Jira counter links", func(id, name, raw, count, address string) {
	issue := Issue{Issue: externalissues.Issue{Fields: map[string]externalissues.Field{
		id: {ID: id, Name: name, Value: json.RawMessage(raw)},
	}}}
	for _, full := range []bool{false, true} {
		fields, err := issue.renderFields(RenderOptions{Full: full})
		Expect(err).NotTo(HaveOccurred())
		Expect(fields.(api.TextTable).Rows).To(HaveLen(1))
		Expect(fields.Markdown()).To(ContainSubstring("| " + name + " | [" + count + "](" + address + ") |"))
		Expect(fields.HTML()).To(ContainSubstring(`<a href="` + address + `">` + count + `</a>`))
		Expect(fields.String()).To(SatisfyAll(ContainSubstring(name), ContainSubstring(count)))
		issue.Full = full
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		var decoded Issue
		Expect(json.Unmarshal([]byte(output), &decoded)).To(Succeed())
		Expect(decoded.Fields).To(HaveLen(1))
		for _, field := range decoded.Fields {
			Expect(string(field.Value)).To(MatchJSON(raw))
		}
	}
},
	Entry("zero votes", "votes", "Votes", `{"hasVoted":false,"self":"https://example.test/issue/EXAMPLE-42/votes","votes":0}`, "0", "https://example.test/issue/EXAMPLE-42/votes"),
	Entry("watchers", "watches", "Watchers", `{"isWatching":false,"self":"https://example.test/issue/EXAMPLE-42/watchers","watchCount":1}`, "1", "https://example.test/issue/EXAMPLE-42/watchers"),
)
