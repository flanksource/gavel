//go:build !darwin || !cgo || !gavel_menubar

package main

import (
	"context"
	"time"

	"github.com/flanksource/commons/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("dashboard without compiled menu bar support", func() {
	It("warns and keeps serving until shutdown", func() {
		output := gbytes.NewBuffer()
		previous := logger.GetOutput()
		logger.SetOutput(output)
		DeferCleanup(func() { logger.SetOutput(previous) })
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		done := make(chan error, 1)
		go func() {
			done <- runDashboardFrontend(ctx, nil, "http://localhost:19093", PRListOptions{MenuBar: true})
		}()

		Eventually(output).Should(gbytes.Say("menu bar requires a macOS build"))
		Expect(string(output.Contents())).To(ContainSubstring("WRN"))
		Consistently(done).WithTimeout(150 * time.Millisecond).ShouldNot(Receive())
		cancel()
		Eventually(done).Should(Receive(BeNil()))
	})
})
