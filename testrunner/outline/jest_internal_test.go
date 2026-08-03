package outline

import (
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("Jest source outline", func() {
	It("enumerates nested TypeScript tests and modifiers without executing source", func() {
		source := []byte(`
import { describe as suite, test as spec } from '@jest/globals'

suite.only("calculator", () => {
  const fake = "test('not a test', () => {})"
  const pattern = () => { return /test("not a test")/ }
  spec("adds numbers", () => expect(add(1, 2)).toBe(3))
  spec.skip('rejects text', () => {})
  spec.each([[1, 2, 3]])("adds %i and %i", (a: number, b: number, want: number) => {})
})

test(` + "`dynamic ${name}`" + `, () => {})
`)

		entries, err := parseJestSource("src/calculator.test.ts", source)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(2))
		Expect(entries[0].Name).To(Equal("calculator"))
		Expect(entries[0].Container).To(BeTrue())
		Expect(entries[0].Focused).To(BeTrue())
		Expect(entries[0].Children).To(HaveLen(3))
		Expect(entries[0].Children[0].Framework).To(Equal(parsers.Jest))
		Expect(entries[0].Children[0].Name).To(Equal("adds numbers"))
		Expect(entries[0].Children[0].Suite).To(Equal([]string{"calculator"}))
		Expect(entries[0].Children[1].Pending).To(BeTrue())
		Expect(entries[0].Children[2].Dynamic).To(BeTrue())
		Expect(entries[1].Dynamic).To(BeTrue())
		Expect(entries[1].Name).To(Equal("<dynamic>"))
	})

	It("fails loudly on unterminated source", func() {
		_, err := parseJestSource("broken.test.ts", []byte(`test("broken, () => {})`))
		Expect(err).To(MatchError(ContainSubstring("broken.test.ts")))
	})
})
