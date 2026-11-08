package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dashboard browser startup", func() {
	DescribeTable("opens the browser only when explicitly requested", func(opts PRListOptions, wantOpen bool) {
		if runtime.GOOS == "windows" {
			Skip("browser launcher fixture uses a POSIX executable")
		}
		dir := GinkgoT().TempDir()
		called := filepath.Join(dir, "browser-url")
		GinkgoT().Setenv("GAVEL_BROWSER_CALLED", called)
		GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		for _, name := range []string{"open", "xdg-open"} {
			Expect(os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$GAVEL_BROWSER_CALLED\"\n"), 0o700)).To(Succeed())
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		const dashboardURL = "http://localhost:19093"
		Expect(runDashboardFrontend(ctx, nil, dashboardURL, opts)).To(Succeed())
		if wantOpen {
			Eventually(func() string {
				data, _ := os.ReadFile(called)
				return string(data)
			}).WithTimeout(time.Second).Should(Equal(dashboardURL + "\n"))
		} else {
			Consistently(func() bool {
				_, err := os.Stat(called)
				return os.IsNotExist(err)
			}).WithTimeout(150 * time.Millisecond).Should(BeTrue())
		}
	},
		Entry("UI alone", PRListOptions{UI: true}, false),
		Entry("serve defaults", applyServeDefaults(PRListOptions{}), false),
		Entry("explicit open", PRListOptions{Open: true}, true),
		Entry("UI and explicit open", PRListOptions{UI: true, Open: true}, true),
	)

	DescribeTable("exposes an opt-in open flag", func(args []string) {
		cmd, _, err := rootCmd.Find(args)
		Expect(err).NotTo(HaveOccurred())
		flag := cmd.Flags().Lookup("open")
		Expect(flag).NotTo(BeNil())
		Expect(flag.DefValue).To(Equal("false"))
	},
		Entry("pr list", []string{"pr", "list"}),
		Entry("serve", []string{"serve"}),
	)
})
