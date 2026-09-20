package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO target options", func() {
	It("requires targets at the command boundary and keeps selectors off run", func() {
		Expect(todosRunCmd.Args(todosRunCmd, nil)).To(HaveOccurred())
		Expect(todosCheckCmd.Args(todosCheckCmd, nil)).To(HaveOccurred())
		Expect(todosRunCmd.Flags().Lookup("status")).To(BeNil())
		Expect(todosRunCmd.Flags().Lookup("interactive")).To(BeNil())
	})

	It("keeps the check timeout on the typed command", func() {
		Expect(todosCheckCmd.Flags().Lookup("timeout")).NotTo(BeNil())
	})

	It("requires one concrete target for single TODO mutations", func() {
		_, err := (TodoTargetOptions{}).One()
		Expect(err).To(HaveOccurred())
		_, err = (TodoTargetOptions{IDs: []string{"first", "second"}}).One()
		Expect(err).To(HaveOccurred())
		ref, err := (TodoTargetOptions{IDs: []string{" first "}}).One()
		Expect(err).NotTo(HaveOccurred())
		Expect(ref).To(Equal("first"))
	})

	It("rejects empty and repeated selections before a bulk mutation", func() {
		for _, ids := range [][]string{nil, {" "}, {"first", "first"}} {
			_, err := (TodoTargetOptions{IDs: ids}).Many()
			Expect(err).To(HaveOccurred())
		}
		ids, err := (TodoTargetOptions{IDs: []string{" first ", "second"}}).Many()
		Expect(err).NotTo(HaveOccurred())
		Expect(ids).To(Equal([]string{"first", "second"}))
	})
})
