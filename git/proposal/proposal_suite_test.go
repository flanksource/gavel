package proposal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProposal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Git Proposal Suite")
}
