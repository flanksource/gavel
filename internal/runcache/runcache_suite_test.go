package runcache_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRuncache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runcache Suite")
}
