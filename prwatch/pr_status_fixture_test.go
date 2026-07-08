package prwatch

import (
	"strings"
	"testing"
)

func TestBuildPRStatusVerificationBlock(t *testing.T) {
	block := BuildPRStatusVerificationBlock(PRStatusVerification{
		PRNumber:   123,
		Repo:       "flanksource/gavel",
		CommentIDs: []int64{3, 1, 3},
		Actions:    []string{"*"},
	})

	want := "gavel pr status 123 --repo flanksource/gavel --comments 1,3 --actions '*'"
	if !strings.Contains(block, want) {
		t.Fatalf("block missing command %q:\n%s", want, block)
	}
	if !strings.Contains(block, "```exec") {
		t.Fatalf("block should be an exec fixture:\n%s", block)
	}
}

func TestBuildPRStatusVerificationBlockQuotesActions(t *testing.T) {
	block := BuildPRStatusVerificationBlock(PRStatusVerification{
		PRNumber: 9,
		Actions:  []string{"Go Test / pg 15"},
	})

	if !strings.Contains(block, "--actions 'Go Test / pg 15'") {
		t.Fatalf("action selector was not shell-quoted:\n%s", block)
	}
}

func TestBuildPRStatusVerificationBlockEmptyWithoutSelectors(t *testing.T) {
	if got := BuildPRStatusVerificationBlock(PRStatusVerification{PRNumber: 1}); got != "" {
		t.Fatalf("expected empty block without selectors, got %q", got)
	}
}
