package utils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDiffMap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DiffMap Suite")
}

var _ = Describe("SHA256 Compacting", func() {
	Describe("compactSHA256", func() {
		It("should compact SHA256 in Docker image digest", func() {
			input := "registry.io/myapp@sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
			result := CompactSHA256(input)
			Expect(result).To(Equal("registry.io/myapp@sha256:a1..b2"))
		})

		It("should compact plain SHA256 string", func() {
			input := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
			result := CompactSHA256(input)
			Expect(result).To(Equal("a1..b2"))
		})

		It("should compact multiple SHA256s in one string", func() {
			input := "hash1: abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789 hash2: 1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
			result := CompactSHA256(input)
			Expect(result).To(Equal("hash1: ab..89 hash2: 12..ef"))
		})

		It("should not modify strings without SHA256", func() {
			input := "registry.io/myapp:v1.2.3"
			result := CompactSHA256(input)
			Expect(result).To(Equal(input))
		})

		It("should not compact partial hex strings", func() {
			input := "short-hash-abc123"
			result := CompactSHA256(input)
			Expect(result).To(Equal(input))
		})

		It("should handle mixed case SHA256", func() {
			input := "A1B2C3D4E5F6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
			result := CompactSHA256(input)
			Expect(result).To(Equal("A1..b2"))
		})

		It("should handle empty string", func() {
			input := ""
			result := CompactSHA256(input)
			Expect(result).To(Equal(""))
		})
	})
})

var _ = Describe("SOPS Compacting", func() {
	Describe("CompactSOPS", func() {
		It("should compact SOPS encrypted value", func() {
			input := "ENC[AES256_GCM,data:0IZk2z9shg86RQ47DI4xPZYl2rh2x+XfcICWJ8RTxQMeA+njaV6QT1Lv8aZPfPP/4hpKb2akaO/27ZA1T/jt++07sQKGjI+kCjOOMEsDNrKDvPcPSlskkpnk5xuriuwaTYDQyeXcCYGx6uksq8HXkIXJN1ZynPSAlv/EytY19ys=,iv:KSgqvG6v+wja7PutWnTH2RdXds+5MgHrLvRo9wnyl4M=,tag:uz2SXIKyG6keHpGuFJ3xTw==,type:str]"
			result := CompactSOPS(input)
			Expect(result).To(Equal("ENC[AES256_GCM,data:0IZ..172 more chars,type:str]"))
		})

		It("should compact SOPS value with different data length", func() {
			input := "ENC[AES256_GCM,data:abc123,iv:xyz789,tag:def456,type:str]"
			result := CompactSOPS(input)
			Expect(result).To(Equal("ENC[AES256_GCM,data:abc..6 more chars,type:str]"))
		})

		It("should compact multiple SOPS values in one string", func() {
			input := "password: ENC[AES256_GCM,data:oldpass123,iv:iv1,tag:tag1,type:str] token: ENC[AES256_GCM,data:newtoken456,iv:iv2,tag:tag2,type:str]"
			result := CompactSOPS(input)
			Expect(result).To(Equal("password: ENC[AES256_GCM,data:old..10 more chars,type:str] token: ENC[AES256_GCM,data:new..11 more chars,type:str]"))
		})

		It("should not modify strings without SOPS values", func() {
			input := "regular password: mypassword123"
			result := CompactSOPS(input)
			Expect(result).To(Equal(input))
		})

		It("should handle empty string", func() {
			input := ""
			result := CompactSOPS(input)
			Expect(result).To(Equal(""))
		})

		It("should handle SOPS value with binary type", func() {
			input := "ENC[AES256_GCM,data:binarydata123456,iv:ivdata,tag:tagdata,type:binary]"
			result := CompactSOPS(input)
			Expect(result).To(Equal("ENC[AES256_GCM,data:bin..16 more chars,type:binary]"))
		})

		It("should preserve SOPS structure but compact long fields", func() {
			input := "secret: ENC[AES256_GCM,data:verylongdatastring1234567890abcdefghijklmnopqrstuvwxyz,iv:someinitializationvector,tag:someauthtag,type:str]"
			result := CompactSOPS(input)
			Expect(result).To(Equal("secret: ENC[AES256_GCM,data:ver..54 more chars,type:str]"))
		})

		It("should handle mixed SOPS and regular content", func() {
			input := "apiVersion: v1\ndata:\n  password: ENC[AES256_GCM,data:encryptedpass,iv:someiv,tag:sometag,type:str]\n  username: admin"
			result := CompactSOPS(input)
			Expect(result).To(ContainSubstring("ENC[AES256_GCM,data:enc..13 more chars,type:str]"))
			Expect(result).To(ContainSubstring("username: admin"))
		})
	})
})

var _ = Describe("Common Prefix/Suffix Detection", func() {
	Describe("findCommonPrefix", func() {
		It("should find common prefix in identical strings", func() {
			prefix := FindCommonPrefix("hello", "hello")
			Expect(prefix).To(Equal("hello"))
		})

		It("should find common prefix when strings share beginning", func() {
			prefix := FindCommonPrefix("registry.example.com/myapp:v1.2.3", "registry.example.com/myapp:v1.2.4")
			Expect(prefix).To(Equal("registry.example.com/myapp:v1.2."))
		})

		It("should return empty string when no common prefix", func() {
			prefix := FindCommonPrefix("abc", "xyz")
			Expect(prefix).To(Equal(""))
		})

		It("should handle empty strings", func() {
			prefix := FindCommonPrefix("", "hello")
			Expect(prefix).To(Equal(""))

			prefix = FindCommonPrefix("hello", "")
			Expect(prefix).To(Equal(""))
		})

		It("should handle single character difference", func() {
			prefix := FindCommonPrefix("a", "b")
			Expect(prefix).To(Equal(""))
		})
	})

	Describe("findCommonSuffix", func() {
		It("should find common suffix in identical strings", func() {
			suffix := FindCommonSuffix("hello", "hello")
			Expect(suffix).To(Equal("hello"))
		})

		It("should find common suffix when strings share ending", func() {
			suffix := FindCommonSuffix("old-config.yaml", "new-config.yaml")
			Expect(suffix).To(Equal("-config.yaml"))
		})

		It("should find only file extension when that's the only common suffix", func() {
			suffix := FindCommonSuffix("deployment.yaml", "service.yaml")
			Expect(suffix).To(Equal(".yaml"))
		})

		It("should return empty string when no common suffix", func() {
			suffix := FindCommonSuffix("abc", "xyz")
			Expect(suffix).To(Equal(""))
		})

		It("should handle empty strings", func() {
			suffix := FindCommonSuffix("", "hello")
			Expect(suffix).To(Equal(""))

			suffix = FindCommonSuffix("hello", "")
			Expect(suffix).To(Equal(""))
		})
	})

	Describe("formatValueDiff", func() {
		It("should highlight only differences for docker image tags", func() {
			result := HumanDiff("registry.example.com/myapp:v1.2.3", "registry.example.com/myapp:v1.2.4")
			text := result.String()

			Expect(text).To(ContainSubstring("registry.example.com/myapp:v1.2."))
			Expect(text).To(ContainSubstring("3"))
			Expect(text).To(ContainSubstring("4"))
		})

		It("should handle completely different strings", func() {
			result := HumanDiff("old", "new")
			text := result.String()

			Expect(text).To(ContainSubstring("old"))
			Expect(text).To(ContainSubstring("new"))
		})

		It("should handle identical strings", func() {
			result := HumanDiff("same", "same")
			text := result.String()

			Expect(text).NotTo(BeEmpty())
		})

		It("should handle strings with common suffix", func() {
			result := HumanDiff("old-config.yaml", "new-config.yaml")
			text := result.String()

			Expect(text).To(ContainSubstring("old-config"))
			Expect(text).To(ContainSubstring("new-config"))
			Expect(text).To(ContainSubstring(".yaml"))
		})

		It("should handle strings with both common prefix and suffix", func() {
			result := HumanDiff("prefix-old-suffix", "prefix-new-suffix")
			text := result.String()

			Expect(text).To(ContainSubstring("prefix-"))
			Expect(text).To(ContainSubstring("old"))
			Expect(text).To(ContainSubstring("new"))
			Expect(text).To(ContainSubstring("-suffix"))
		})

		It("should compact SHA256 values before comparing", func() {
			oldVal := "registry.io/myapp@sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
			newVal := "registry.io/myapp@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
			result := HumanDiff(oldVal, newVal)
			text := result.String()

			Expect(text).To(ContainSubstring("a1..b2"))
			Expect(text).To(ContainSubstring("12..ef"))
			Expect(text).NotTo(ContainSubstring("a1b2c3d4e5f6a1b2"))
		})
	})
})

var _ = Describe("DiffMap", func() {
	Describe("Diff", func() {
		It("should use smart formatting for modified string values", func() {
			before := DiffMap[any]{
				"image": "registry.example.com/myapp:v1.2.3",
			}
			after := DiffMap[any]{
				"image": "registry.example.com/myapp:v1.2.4",
			}

			diff := before.Diff(after)
			Expect(diff).To(HaveKey("image"))

			text := diff["image"].String()
			Expect(text).To(ContainSubstring("registry.example.com/myapp:v1.2."))
			Expect(text).To(ContainSubstring("3"))
			Expect(text).To(ContainSubstring("4"))
		})

		It("should handle added fields", func() {
			before := DiffMap[any]{}
			after := DiffMap[any]{
				"new_field": "value",
			}

			diff := before.Diff(after)
			Expect(diff).To(HaveKey("new_field"))
		})

		It("should handle removed fields", func() {
			before := DiffMap[any]{
				"old_field": "value",
			}
			after := DiffMap[any]{}

			diff := before.Diff(after)
			Expect(diff).To(HaveKey("old_field"))
		})

		It("should handle non-string values normally", func() {
			before := DiffMap[any]{
				"count": 5,
			}
			after := DiffMap[any]{
				"count": 10,
			}

			diff := before.Diff(after)
			Expect(diff).To(HaveKey("count"))
		})
	})

	Describe("Collapse with smart dot-notation", func() {
		It("should collapse single-child chains into dot notation", func() {
			dm := DiffMap[any]{
				"spec.ref.tag": "v1.0",
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("spec.ref.tag"))
			Expect(result["spec.ref.tag"]).To(Equal("v1.0"))
		})

		It("should preserve multi-child nodes as nested structures", func() {
			dm := DiffMap[any]{
				"spec.replicas": 3,
				"spec.image":    "myapp:v1",
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("spec"))
			spec, ok := result["spec"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(spec).To(HaveKey("replicas"))
			Expect(spec).To(HaveKey("image"))
			Expect(spec["replicas"]).To(Equal(3))
			Expect(spec["image"]).To(Equal("myapp:v1"))
		})

		It("should handle mixed single and multi-child nodes", func() {
			dm := DiffMap[any]{
				"spec.ref.tag":  "v1.0",
				"spec.replicas": 3,
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("spec"))
			spec, ok := result["spec"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(spec).To(HaveKey("ref.tag"))
			Expect(spec).To(HaveKey("replicas"))
			Expect(spec["ref.tag"]).To(Equal("v1.0"))
			Expect(spec["replicas"]).To(Equal(3))
		})

		It("should collapse deeply nested single-child chains", func() {
			dm := DiffMap[any]{
				"a.b.c.d.e.f": "deep",
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("a.b.c.d.e.f"))
			Expect(result["a.b.c.d.e.f"]).To(Equal("deep"))
		})

		It("should handle complex kubernetes-like structure", func() {
			dm := DiffMap[any]{
				"spec.template.spec.containers[0].image": "nginx:1.21",
				"spec.replicas":                          3,
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("spec"))
			spec, ok := result["spec"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(spec).To(HaveKey("template.spec.containers[0].image"))
			Expect(spec).To(HaveKey("replicas"))
		})

		It("should preserve structure when all nodes have multiple children", func() {
			dm := DiffMap[any]{
				"a.b": "v1",
				"a.c": "v2",
				"d.e": "v3",
				"d.f": "v4",
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("a"))
			Expect(result).To(HaveKey("d"))

			a, ok := result["a"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(a).To(HaveKey("b"))
			Expect(a).To(HaveKey("c"))

			d, ok := result["d"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(d).To(HaveKey("e"))
			Expect(d).To(HaveKey("f"))
		})

		It("should handle empty DiffMap", func() {
			dm := DiffMap[any]{}
			result := dm.Collapse()
			Expect(result).To(BeEmpty())
		})

		It("should handle single key without dots", func() {
			dm := DiffMap[any]{
				"simple": "value",
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("simple"))
			Expect(result["simple"]).To(Equal("value"))
		})

		It("should collapse partial chains correctly", func() {
			dm := DiffMap[any]{
				"a.b.c": "v1",
				"a.b.d": "v2",
				"a.e":   "v3",
			}

			result := dm.Collapse()
			Expect(result).To(HaveKey("a"))
			a, ok := result["a"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(a).To(HaveKey("b"))
			Expect(a).To(HaveKey("e"))

			b, ok := a["b"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(b).To(HaveKey("c"))
			Expect(b).To(HaveKey("d"))
		})
	})
})
