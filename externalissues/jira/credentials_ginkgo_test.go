package jira

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira credential discovery", func() {
	DescribeTable("discovers complete environment credentials", func(env map[string]string, expected Credentials) {
		credentials, found, err := credentialsFromEnv(func(key string) string { return env[key] })
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(credentials).To(Equal(expected))
	},
		Entry("Jira API token", map[string]string{"JIRA_EMAIL": "api@example.test", "JIRA_API_TOKEN": "jira-secret"}, Credentials{Username: "api@example.test", APIToken: "jira-secret"}),
		Entry("Atlassian API token", map[string]string{"ATLASSIAN_EMAIL": "api@example.test", "ATLASSIAN_API_TOKEN": "atlassian-secret"}, Credentials{Username: "api@example.test", APIToken: "atlassian-secret"}),
		Entry("TWG API token", map[string]string{"TWG_USER": "api@example.test", "TWG_TOKEN": "twg-secret"}, Credentials{Username: "api@example.test", APIToken: "twg-secret"}),
		Entry("TWG OAuth", map[string]string{"TWG_OAUTH_ACCESS_TOKEN": "oauth-secret", "JIRA_CLOUD_ID": "cloud-id"}, Credentials{AccessToken: "oauth-secret", CloudID: "cloud-id"}),
		Entry("Jira takes precedence", map[string]string{"JIRA_USERNAME": "api@example.test", "JIRA_TOKEN": "jira-secret", "TWG_TOKEN": "unused"}, Credentials{Username: "api@example.test", APIToken: "jira-secret"}),
	)

	It("rejects partial credentials instead of switching identities", func() {
		_, _, err := credentialsFromEnv(func(key string) string {
			return map[string]string{"JIRA_API_TOKEN": "never-print-this-token", "TWG_USER": "different@example.test", "TWG_TOKEN": "other-token"}[key]
		})
		Expect(err).To(MatchError(ContainSubstring("JIRA_EMAIL")))
		Expect(err.Error()).NotTo(ContainSubstring("never-print-this-token"))
	})

	It("parses literal TWG exports without evaluating shell syntax", func() {
		values, err := parseTWGExport("# selected credentials\nunset TWG_TOKEN 2>/dev/null || true\nexport TWG_SITE='example.atlassian.net'\nexport TWG_OAUTH_ACCESS_TOKEN='literal'\\''quote$(command)'\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(Equal(map[string]string{"TWG_SITE": "example.atlassian.net", "TWG_OAUTH_ACCESS_TOKEN": "literal'quote$(command)"}))
		_, err = parseTWGExport("export TWG_TOKEN='secret'; run-something\n")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).NotTo(ContainSubstring("secret"))
	})
	It("decodes TWG ANSI-C quoting without retaining the dollar prefix or expanding commands", func() {
		values, err := parseTWGExport("export TWG_SITE=$'example.atlassian.net'\nexport TWG_OAUTH_ACCESS_TOKEN=$'literal\\'quote$(command)\\nline'\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(Equal(map[string]string{"TWG_SITE": "example.atlassian.net", "TWG_OAUTH_ACCESS_TOKEN": "literal'quote$(command)\nline"}))
	})
})
