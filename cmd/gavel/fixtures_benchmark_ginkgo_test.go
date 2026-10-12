package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixtures measurement flags", func() {
	It("exposes benchmark and fixture limits on the fixture command", func() {
		for _, name := range []string{"benchmark", "baseline", "profile", "max-deviation-pct", "max-time", "max-memory", "max-io-read", "max-io-write", "max-sql-time", "max-sql-query", "max-sql-queries", "max-slow-sql"} {
			Expect(fixturesCmd.Flags().Lookup(name)).NotTo(BeNil(), name)
		}
	})
})
