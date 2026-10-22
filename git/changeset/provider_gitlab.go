package changeset

import "fmt"

func gitlabComparison(env Environment) (Comparison, error) {
	head := env.Lookup("CI_COMMIT_SHA")
	if env.Lookup("CI_MERGE_REQUEST_IID") == "" {
		return previousComparison(env.Lookup("CI_COMMIT_BEFORE_SHA"), head), nil
	}
	base := env.Lookup("CI_MERGE_REQUEST_DIFF_BASE_SHA")
	if kind := env.Lookup("CI_MERGE_REQUEST_EVENT_TYPE"); kind == "merged_result" || kind == "merge_train" {
		base = env.Lookup("CI_MERGE_REQUEST_TARGET_BRANCH_SHA")
		head = env.Lookup("CI_MERGE_REQUEST_SOURCE_BRANCH_SHA")
	}
	if base == "" || head == "" {
		return Comparison{}, fmt.Errorf("MR requires its diff base and source SHA (target/source branch SHAs for merged results)")
	}
	return Comparison{Kind: "pull-request", Base: base, Head: head}, nil
}
