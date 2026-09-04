package types

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFixtureTypes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fixture Types Suite")
}
