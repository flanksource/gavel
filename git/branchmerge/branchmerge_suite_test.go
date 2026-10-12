package branchmerge_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBranchMerge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "git branchmerge")
}
