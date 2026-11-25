package blame

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestBlame(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PR reviewer recommendations")
}
