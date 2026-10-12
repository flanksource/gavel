package gitstate_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGitState(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "git gitstate")
}
