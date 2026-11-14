package models

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseLocation", func() {
	DescribeTable("parses a shortcut",
		func(shortcut string, want ViolationLocation) {
			Expect(ParseLocation(shortcut)).To(Equal(want))
		},
		Entry("file:line", "path/to/file.sql:42", ViolationLocation{File: "path/to/file.sql", Line: 42}),
		Entry("file:line:col", "file.sql:42:7", ViolationLocation{File: "file.sql", Line: 42, Column: 7}),
		Entry("a bare file", "README.md", ViolationLocation{File: "README.md"}),
		Entry("branch:path:line:col", "main:rules/pc/data.xml:12:3", ViolationLocation{Ref: "main", File: "rules/pc/data.xml", Line: 12, Column: 3}),
		Entry("slashed branch:path:line", "feature/x:rules/a.xml:4", ViolationLocation{Ref: "feature/x", File: "rules/a.xml", Line: 4}),
		Entry("tag:path:line", "v1.2.3:a.xml:9", ViolationLocation{Ref: "v1.2.3", File: "a.xml", Line: 9}),
		Entry("short sha:path:line", "abc1234:file.sql:5", ViolationLocation{Ref: "abc1234", File: "file.sql", Line: 5}),
		Entry("full sha:path:line:col", "0123456789abcdef0123456789abcdef01234567:x/y.go:1:2",
			ViolationLocation{Ref: "0123456789abcdef0123456789abcdef01234567", File: "x/y.go", Line: 1, Column: 2}),
		Entry("a:b:line is a path containing a colon, not a ref", "a:b:3", ViolationLocation{File: "a:b", Line: 3}),
		Entry("a Windows drive path is not a ref", `C:\work\q.sql:12`, ViolationLocation{File: `C:\work\q.sql`, Line: 12}),
		Entry("a Windows drive path with forward slashes", "C:/work/q.sql:12:4", ViolationLocation{File: "C:/work/q.sql", Line: 12, Column: 4}),
		Entry("entity", "entity://Policy/6F9A-11", ViolationLocation{Entity: &EntityTarget{Kind: "Policy", ID: "6F9A-11"}}),
		Entry("url with css", "https://app.example/entity/policy/1#css=.row[data-id=x]",
			ViolationLocation{URL: &URLTarget{URL: "https://app.example/entity/policy/1", CSS: ".row[data-id=x]"}}),
		Entry("url with xpath", "http://app.example/p#xpath=//tr[2]", ViolationLocation{URL: &URLTarget{URL: "http://app.example/p", XPath: "//tr[2]"}}),
		Entry("url with testid", "https://app.example/p#testid=save", ViolationLocation{URL: &URLTarget{URL: "https://app.example/p", TestID: "save"}}),
		Entry("url keeps a plain anchor", "https://app.example/docs#section-2", ViolationLocation{URL: &URLTarget{URL: "https://app.example/docs#section-2"}}),
	)

	DescribeTable("rejects a malformed shortcut",
		func(shortcut, problem string) {
			_, err := ParseLocation(shortcut)
			Expect(err).To(MatchError(ContainSubstring(problem)))
		},
		Entry("empty", "", "empty"),
		Entry("entity without id", "entity://Policy", "entity://<kind>/<id>"),
		Entry("zero line", "file.sql:0", "line must be positive"),
		Entry("url without host", "https://#css=.x", "host"),
		Entry("unknown selector", "https://app.example/p#bogus=1&css=.x", "selector"),
	)
})

var _ = Describe("ViolationLocation", func() {
	It("refuses a URL target with two selector strategies", func() {
		location := ViolationLocation{URL: &URLTarget{URL: "https://app.example", CSS: ".a", XPath: "//a"}}
		Expect(location.Validate()).To(MatchError(ContainSubstring("css and xpath")))
	})

	DescribeTable("refuses what the lint UI cannot render",
		func(location ViolationLocation, problem string) {
			Expect(location.Validate()).To(MatchError(ContainSubstring(problem)))
		},
		Entry("name without role", ViolationLocation{URL: &URLTarget{URL: "https://x.example", Name: "Save"}}, "name without role"),
		Entry("a relative url", ViolationLocation{URL: &URLTarget{URL: "/entity/policy/1"}}, "not an absolute URL"),
		Entry("an empty image region", ViolationLocation{Image: &ImageTarget{Src: "s.png", X: 1, Y: 1}}, "width, height > 0"),
	)

	It("refuses a location naming two targets", func() {
		location := ViolationLocation{File: "a.sql", Entity: &EntityTarget{Kind: "Policy", ID: "1"}}
		Expect(location.Validate()).To(MatchError(ContainSubstring("file and entity")))
	})

	It("renders each target", func() {
		Expect([]string{
			ViolationLocation{File: "a.sql", Line: 3}.String(),
			ViolationLocation{Ref: "main", File: "a.sql", Line: 3}.String(),
			ViolationLocation{Entity: &EntityTarget{Kind: "Policy", ID: "9"}}.String(),
			ViolationLocation{URL: &URLTarget{URL: "https://x.example/p", Role: "button", Name: "Save"}}.String(),
			ViolationLocation{Image: &ImageTarget{Src: "shot.png", X: 1, Y: 2, Width: 30, Height: 40}}.String(),
		}).To(Equal([]string{
			"a.sql:3",
			"main:a.sql:3",
			"entity://Policy/9",
			"https://x.example/p#role=button&name=Save",
			"shot.png#1,2,30x40",
		}))
	})

	It("decodes shortcut strings and objects, and round-trips every structured target", func() {
		var decoded []ViolationLocation
		Expect(json.Unmarshal([]byte(`["main:a/b.xml:2", {"entity": {"kind": "Activity", "id": "A1", "name": "x"}}]`), &decoded)).To(Succeed())
		Expect(decoded).To(Equal([]ViolationLocation{
			{Ref: "main", File: "a/b.xml", Line: 2},
			{Entity: &EntityTarget{Kind: "Activity", ID: "A1", Name: "x"}},
		}))

		targets := []ViolationLocation{
			{File: "q.sql", Line: 4, Column: 2, EndLine: 9, Ref: "v1", Message: "here"},
			{Entity: &EntityTarget{Kind: "Policy", ID: "1", Name: "P"}},
			{URL: &URLTarget{URL: "https://x.example", Role: "row", Name: "first"}},
			{Image: &ImageTarget{Src: "s.png", X: 1, Y: 2, Width: 3, Height: 4}},
		}
		raw, err := json.Marshal(targets)
		Expect(err).NotTo(HaveOccurred())
		var back []ViolationLocation
		Expect(json.Unmarshal(raw, &back)).To(Succeed())
		Expect(back).To(Equal(targets))
	})

	It("keeps a non-file location's JSON free of an empty file key", func() {
		raw, err := json.Marshal(ViolationLocation{Entity: &EntityTarget{Kind: "Policy", ID: "1"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).To(Equal(`{"entity":{"kind":"Policy","id":"1"}}`))
	})
})
