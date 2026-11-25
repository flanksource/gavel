package slots

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSlots(t *testing.T) {
	// Run under a gavel that holds a slot, the inherited lease would let every
	// Acquire here through and hide contention.
	t.Setenv(HeldEnv, "")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Machine-wide slots")
}

var _ = Describe("Slot values", func() {
	DescribeTable("resolves positive integers and core percentages", func(value Value, cores, expected int) {
		actual, err := Parse("slots.project", value, cores)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
	},
		Entry("fixed", Value("8"), 3, 8),
		Entry("half", Value("50%"), 8, 4),
		Entry("round down", Value("50%"), 7, 3),
		Entry("minimum one", Value("1%"), 2, 1),
		Entry("all cores", Value("100%"), 3, 3),
	)
	DescribeTable("rejects invalid values with their key", func(value Value) {
		_, err := Parse("slots.global", value, 8)
		Expect(err).To(MatchError(ContainSubstring("slots.global")))
	},
		Entry("zero", Value("0")), Entry("negative", Value("-1")),
		Entry("text", Value("abc")), Entry("over 100 percent", Value("150%")),
		Entry("zero percent", Value("0%")), Entry("empty", Value("")),
		Entry("decimal", Value("1.5")), Entry("fractional percent", Value("1.5%")),
		Entry("sign", Value("+1")), Entry("whitespace", Value(" 8")),
	)
	It("decodes JSON numbers and strings", func() {
		var config Config
		Expect(json.Unmarshal([]byte(`{"global":8,"projectDefault":"50%","project":"2"}`), &config)).To(Succeed())
		Expect(config).To(Equal(Config{Global: value("8"), ProjectDefault: value("50%"), Project: value("2")}))
	})
	It("resolves only home global and project defaults and lets project override its own limit", func() {
		limits, err := Resolve(Config{Global: value("4"), ProjectDefault: value("50%")}, Config{Global: value("20"), ProjectDefault: value("20"), Project: value("8")}, 6)
		Expect(err).NotTo(HaveOccurred())
		Expect(limits).To(Equal(Limits{Global: 4, Project: 8}))
		limits, err = Resolve(Config{ProjectDefault: value("50%")}, Config{}, 6)
		Expect(err).NotTo(HaveOccurred())
		Expect(limits).To(Equal(Limits{Project: 3}))
	})
	It("takes no locks or directories without configuration", func() {
		limits, err := Resolve(Config{}, Config{}, 8)
		Expect(err).NotTo(HaveOccurred())
		Expect(limits.Enabled()).To(BeFalse())
		lease, err := Acquire(GinkgoT().Context(), Request{WorkDir: "/does-not-exist", CacheDir: "/does-not-exist"})
		Expect(err).NotTo(HaveOccurred())
		Expect(lease.Close()).To(Succeed())
	})
})

func value(s string) *Value { v := Value(s); return &v }
