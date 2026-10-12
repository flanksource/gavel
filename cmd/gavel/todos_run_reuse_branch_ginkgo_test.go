package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todos run --reuse-branch", func() {
	It("declares the flag and forwards it to the run options", func() {
		Expect(todosRunCmd.Flags().Lookup("reuse-branch")).NotTo(BeNil())
		Expect(todosRunOptions(TodosRunOptions{ReuseBranch: true}).ReuseBranch).To(BeTrue())
		Expect(todosRunOptions(TodosRunOptions{}).ReuseBranch).To(BeFalse())
	})
})
