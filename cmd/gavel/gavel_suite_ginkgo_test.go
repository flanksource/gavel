package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGavelCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gavel CLI Suite")
}
