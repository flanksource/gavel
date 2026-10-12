package land_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "todos land")
}
