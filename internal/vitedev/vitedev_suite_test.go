package vitedev

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestViteDev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vitedev")
}
