package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("launch native option schemas", func() {
	It("excludes process flags and workspace overrides from native configurations", func() {
		for _, kind := range []string{"test", "lint"} {
			schema, err := (launchCommandSchemas{}).Schema(kind)
			Expect(err).NotTo(HaveOccurred())
			properties := schema["properties"].(map[string]any)
			Expect(properties).NotTo(HaveKey("db-url"))
			Expect(properties).NotTo(HaveKey("cwd"))
			Expect(properties).NotTo(HaveKey("work-dir"))
			Expect(properties).NotTo(HaveKey("format"))
			if kind == "test" {
				Expect(properties).To(HaveKey("framework"))
				Expect(properties).NotTo(HaveKey("lint"))
			} else {
				Expect(properties).To(HaveKey("fix"))
				Expect(properties).NotTo(HaveKey("linters"))
			}
		}
	})
})
