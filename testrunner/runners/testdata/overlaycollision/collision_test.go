package overlaycollision

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Package-level names that a dot-imported Ginkgo, or an unaliased import in
// the injected reporter, would redeclare.
type Labels []string

var (
	types  = Labels{"types"}
	fmt    = "fmt"
	os     = "os"
	json   = "json"
	base64 = "base64"
)

func TestCollision(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Overlay collision fixture")
}

var _ = ginkgo.Describe("collision", func() {
	ginkgo.It("keeps package names that shadow reporter imports", func() {
		gomega.Expect([]string{fmt, os, json, base64}).To(gomega.HaveLen(len(types) + 3))
	})
})
