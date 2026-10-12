//go:build darwin && cgo && gavel_menubar

package main

import (
	"os"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("menu-bar signal shutdown", func() {
	It("reaches the timeout when Wails Quit blocks", func() {
		signals := make(chan os.Signal, 2)
		quitStarted := make(chan struct{})
		releaseQuit := make(chan struct{})
		DeferCleanup(func() { close(releaseQuit) })
		done := make(chan struct{})

		go func() {
			waitForMenuBarShutdown(signals, func() {
				close(quitStarted)
				<-releaseQuit
			}, 20*time.Millisecond)
			close(done)
		}()

		signals <- syscall.SIGTERM
		Eventually(quitStarted).WithTimeout(time.Second).Should(BeClosed())
		Eventually(done).WithTimeout(200 * time.Millisecond).Should(BeClosed())
	})

	It("reaches the second signal when Wails Quit blocks", func() {
		signals := make(chan os.Signal, 2)
		quitStarted := make(chan struct{})
		releaseQuit := make(chan struct{})
		DeferCleanup(func() { close(releaseQuit) })
		done := make(chan struct{})

		go func() {
			waitForMenuBarShutdown(signals, func() {
				close(quitStarted)
				<-releaseQuit
			}, time.Second)
			close(done)
		}()

		signals <- syscall.SIGINT
		Eventually(quitStarted).WithTimeout(time.Second).Should(BeClosed())
		signals <- syscall.SIGTERM
		Eventually(done).WithTimeout(200 * time.Millisecond).Should(BeClosed())
	})
})
