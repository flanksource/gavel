package main

import (
	"time"

	"github.com/flanksource/gavel/sdk/status"
	"github.com/flanksource/gavel/slots"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gavel slots rows", func() {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	project := "/src/acme/.git"
	pools := []slots.PoolStatus{
		{Name: "global", Limit: 2, Held: []slots.Slot{
			{Index: 0, Owner: slots.Owner{PID: 101, Command: "go test ./a", Started: now.Add(-90 * time.Second)}},
		}},
		{Name: "project/abc", Project: project, Limit: 1,
			Held:    []slots.Slot{{Index: 0, Owner: slots.Owner{PID: 101, Command: "go test ./a", Project: project, Started: now.Add(-90 * time.Second)}}},
			Waiting: []slots.Waiter{{PID: 202, Lease: status.Lease{Pool: "project/abc", Command: "golangci-lint run", Since: now.Add(-5 * time.Second)}}},
		},
		{Name: "project/def", Project: "/srv/bare.git", Limit: 4},
	}

	It("lists each holder and waiter under its pool's use, and idle pools once", func() {
		Expect(slotRows(pools, now)).To(Equal([]slotRow{
			{Pool: "global", Use: "1/2", Slot: "0", State: "held", PID: 101, Age: "1m30s", Command: "go test ./a"},
			{Pool: "project/abc", Use: "1/1 +1 waiting", Project: "/src/acme", Slot: "0", State: "held", PID: 101, Age: "1m30s", Command: "go test ./a"},
			{Pool: "project/abc", Use: "1/1 +1 waiting", Project: "/src/acme", Slot: "-", State: "waiting", PID: 202, Age: "5s", Command: "golangci-lint run"},
			{Pool: "project/def", Use: "0/4", Project: "/srv/bare.git", Slot: "-", State: "idle"},
		}))
	})

	It("keeps the global pool and the current repository's with --project", func() {
		Expect(projectPools(pools, "project/def")).To(HaveExactElements(HaveField("Name", "global"), HaveField("Name", "project/def")))
	})
})
