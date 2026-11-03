package commit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("git lock holders", func() {
	DescribeTable("classifies lsof file access without treating readers as writers",
		func(output string, expected []string) {
			holders, err := parseLockHolders(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(holders).To(Equal(expected))
		},
		Entry("no open descriptors", "", nil),
		Entry("read-only watchers", "p4256\nf3522\nar\np34736\nf1355\nar\n", nil),
		Entry("write-only descriptor", "p101\nf3\naw\n", []string{"101"}),
		Entry("read/write descriptor", "p102\nf4\nau\n", []string{"102"}),
		Entry("unknown access blocks removal", "p103\nf5\na \n", []string{"103"}),
		Entry("missing access blocks removal", "p104\nf6\n", []string{"104"}),
		Entry("unrecognized access blocks removal", "p105\nf7\na?\n", []string{"105"}),
		Entry("one PID with several writable descriptors is reported once", "p106\nf8\nar\nf9\naw\nf10\nau\np107\nf11\nar\np108\nf12\naw\n", []string{"106", "108"}),
	)

	DescribeTable("rejects malformed ownership data",
		func(output string) {
			_, err := parseLockHolders(output)
			Expect(err).To(HaveOccurred())
		},
		Entry("descriptor without process", "f3\naw\n"),
		Entry("invalid process ID", "pnot-a-pid\nf3\naw\n"),
		Entry("invalid PID suffix", "p101extra\nf3\naw\n"),
		Entry("zero process ID", "p0\nf3\naw\n"),
		Entry("access without descriptor", "p101\nar\n"),
		Entry("process without descriptor", "p101\n"),
		Entry("empty descriptor", "p101\nf\nar\n"),
		Entry("duplicate access field", "p101\nf3\naw\nar\n"),
		Entry("unexpected output", "warning: partial result\n"),
	)
})
