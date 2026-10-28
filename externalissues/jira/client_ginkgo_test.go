package jira

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cexec "github.com/flanksource/clicky/exec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJira(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jira Suite")
}

var _ = Describe("twg client", func() {
	It("passes options as arguments and returns the full response data", func() {
		var arguments []any
		client := Client{Site: "example", Profile: "work", run: func(args ...any) (*cexec.ExecResult, error) {
			arguments = args
			return &cexec.ExecResult{Stdout: `{"apiVersion":"v2","data":[{"key":"EXAMPLE-1"}]}`}, nil
		}}
		data, err := client.Call(context.Background(), "jira", "workitem", "get", "EXAMPLE-1", "--full")
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Valid(data)).To(BeTrue())
		Expect(string(data)).To(Equal(`[{"key":"EXAMPLE-1"}]`))
		Expect(arguments).To(ContainElements("--site", "example", "--profile", "work", "--output-summary", "none", "--output", "json", "EXAMPLE-1"))
	})

	It("returns errors for failed processes and malformed envelopes", func() {
		client := Client{run: func(...any) (*cexec.ExecResult, error) {
			return &cexec.ExecResult{Stderr: "denied"}, errors.New("exit code 1")
		}}
		_, err := client.Call(context.Background(), "jira", "workitem", "get", "EXAMPLE-1")
		Expect(err).To(MatchError(ContainSubstring("denied")))
		client.run = func(...any) (*cexec.ExecResult, error) { return &cexec.ExecResult{Stdout: `{}`}, nil }
		_, err = client.Call(context.Background(), "jira", "workitem", "get", "EXAMPLE-1")
		Expect(err).To(HaveOccurred())
	})

	It("requests the full issue and exposes all returned fields", func() {
		var arguments []any
		client := Client{run: func(args ...any) (*cexec.ExecResult, error) {
			arguments = args
			return &cexec.ExecResult{Stdout: `{"data":[{"key":"EXAMPLE-1","summary":"Example","customfield_123":"value"}]}`}, nil
		}}
		issue, err := client.Get(context.Background(), "EXAMPLE-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(issue)).To(MatchJSON(`{"key":"EXAMPLE-1","summary":"Example","customfield_123":"value"}`))
		Expect(arguments).To(ContainElements("get", "EXAMPLE-1", "--full"))
	})

	It("rejects a get response without exactly one issue", func() {
		client := Client{run: func(...any) (*cexec.ExecResult, error) {
			return &cexec.ExecResult{Stdout: `{"data":[]}`}, nil
		}}
		_, err := client.Get(context.Background(), "EXAMPLE-1")
		Expect(err).To(HaveOccurred())
	})
})
