package changeset

import (
	"fmt"
	"strings"
)

func azureComparison(env Environment) (Comparison, error) {
	if env.Lookup("BUILD_REASON") != "PullRequest" {
		return Comparison{Kind: "commit", Head: env.Lookup("BUILD_SOURCEVERSION")}, nil
	}
	base, head := env.Lookup("SYSTEM_PULLREQUEST_TARGETBRANCH"), env.Lookup("SYSTEM_PULLREQUEST_SOURCECOMMITID")
	if base == "" || head == "" {
		return Comparison{}, fmt.Errorf("PR requires SYSTEM_PULLREQUEST_TARGETBRANCH and SYSTEM_PULLREQUEST_SOURCECOMMITID")
	}
	base = "refs/remotes/origin/" + strings.TrimPrefix(base, "refs/heads/")
	return Comparison{Kind: "pull-request", Base: base, Head: head}, nil
}
