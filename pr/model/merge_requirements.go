package model

type AutoMergeInfo struct {
	Method    string `json:"mergeMethod"`
	EnabledAt string `json:"enabledAt"`
}

type BranchProtectionInfo struct {
	RequiredContexts               []string `json:"requiredStatusCheckContexts"`
	RequiredApprovals              int      `json:"requiredApprovingReviewCount"`
	RequiresApprovingReviews       bool     `json:"requiresApprovingReviews"`
	RequiresConversationResolution bool     `json:"requiresConversationResolution"`
	RequiresCommitSignatures       bool     `json:"requiresCommitSignatures"`
	RequiresStrictStatusChecks     bool     `json:"requiresStrictStatusChecks"`
}
