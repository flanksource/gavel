package labels_test

import (
	"sort"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/todos/labels"
)

func TestLabels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO label resolution")
}

var _ = Describe("Resolver", func() {
	workspace := labels.Definitions{
		{Name: "bug", Color: labels.ColorPink, Icon: "test", Description: "workspace override"},
		{Name: "area", Color: labels.ColorTeal, Icon: "docs"},
	}
	global := labels.Definitions{
		{Name: "bug", Color: labels.ColorIndigo, Description: "global override"},
		{Name: "flaky", Color: labels.ColorAmber},
		{Name: "source", Color: labels.ColorSlate},
	}

	Describe("precedence", func() {
		resolver := labels.NewResolver(workspace, global)

		It("prefers a workspace row over a global row of the same name", func() {
			def := resolver.Resolve("bug")
			Expect(def.Color).To(Equal(labels.ColorPink))
			Expect(def.Scope).To(Equal(labels.ScopeWorkspace))
			Expect(def.Description).To(Equal("workspace override"))
		})

		It("prefers a global row over the built-in of the same name", func() {
			// "docs" is a builtin (sky); only global defines "flaky".
			Expect(resolver.Resolve("flaky").Color).To(Equal(labels.ColorAmber))
			Expect(resolver.Resolve("flaky").Scope).To(Equal(labels.ScopeGlobal))
		})

		It("falls back to the built-in when nothing is stored", func() {
			def := resolver.Resolve("docs")
			Expect(def.Color).To(Equal(labels.ColorSky))
			Expect(def.Scope).To(Equal(labels.ScopeBuiltin))
			Expect(def.Icon).To(Equal("docs"))
		})
	})

	Describe("namespace keys", func() {
		resolver := labels.NewResolver(workspace, global)

		It("colours a key:value label from its key and records the match", func() {
			def := resolver.Resolve("source:todo")
			Expect(def.Color).To(Equal(labels.ColorSlate))
			Expect(def.Scope).To(Equal(labels.ScopeGlobal))
			Expect(def.MatchedKey).To(Equal("source"))
			Expect(def.Name).To(Equal("source:todo"), "the full label stays the rendered name")
		})

		It("colours a key/value label from its key", func() {
			def := resolver.Resolve("area/ui")
			Expect(def.Color).To(Equal(labels.ColorTeal))
			Expect(def.Scope).To(Equal(labels.ScopeWorkspace))
			Expect(def.MatchedKey).To(Equal("area"))
		})

		It("gives every member of one namespace the same hue", func() {
			Expect(resolver.Resolve("area/ui").Color).To(Equal(resolver.Resolve("area/api").Color))
		})

		It("prefers an exact match over the key match", func() {
			scoped := labels.NewResolver(labels.Definitions{
				{Name: "area", Color: labels.ColorTeal},
				{Name: "area/ui", Color: labels.ColorRose},
			}, nil)
			Expect(scoped.Resolve("area/ui").Color).To(Equal(labels.ColorRose))
			Expect(scoped.Resolve("area/ui").MatchedKey).To(BeEmpty())
		})
	})

	Describe("derived fallback", func() {
		resolver := labels.NewResolver(nil, nil)

		It("never returns an empty colour for an undefined label", func() {
			def := resolver.Resolve("something-nobody-defined")
			Expect(def.Color).NotTo(BeEmpty())
			Expect(labels.IsColor(string(def.Color))).To(BeTrue())
			Expect(def.Scope).To(Equal(labels.ScopeDerived))
			Expect(def.Icon).To(BeEmpty())
		})

		It("is stable across calls", func() {
			Expect(resolver.Resolve("wibble").Color).To(Equal(resolver.Resolve("wibble").Color))
		})

		It("normalizes case and surrounding space", func() {
			Expect(resolver.Resolve("  Flaky  ").Name).To(Equal("flaky"))
			Expect(resolver.Resolve("  BUG  ").Color).To(Equal(resolver.Resolve("bug").Color))
		})
	})

	Describe("ResolveAll", func() {
		resolver := labels.NewResolver(workspace, global)

		It("preserves label order", func() {
			defs := resolver.ResolveAll([]string{"flaky", "bug", "docs"})
			Expect(defs.Names()).To(Equal([]string{"flaky", "bug", "docs"}))
		})

		It("drops blank labels rather than rendering empty chips", func() {
			Expect(resolver.ResolveAll([]string{"bug", "   ", ""}).Names()).To(Equal([]string{"bug"}))
		})

		It("returns nil for no labels", func() {
			Expect(resolver.ResolveAll(nil)).To(BeNil())
		})
	})

	Describe("All", func() {
		It("shadows built-ins with stored rows and sorts by name", func() {
			all := labels.NewResolver(workspace, global).All()

			Expect(sort.SliceIsSorted(all, func(i, j int) bool { return all[i].Name < all[j].Name })).To(BeTrue())

			byName := map[string]labels.Definition{}
			for _, def := range all {
				byName[def.Name] = def
			}
			Expect(byName).To(HaveKey("bug"))
			Expect(byName["bug"].Color).To(Equal(labels.ColorPink), "workspace wins")
			Expect(byName["bug"].Scope).To(Equal(labels.ScopeWorkspace))
			Expect(byName["flaky"].Scope).To(Equal(labels.ScopeGlobal))
			Expect(byName["security"].Scope).To(Equal(labels.ScopeBuiltin))
		})

		It("lists each name exactly once", func() {
			all := labels.NewResolver(workspace, global).All()
			seen := map[string]int{}
			for _, def := range all {
				seen[def.Name]++
			}
			for name, count := range seen {
				Expect(count).To(Equal(1), "duplicate definition for %s", name)
			}
		})

		It("never lists a derived presentation", func() {
			resolver := labels.NewResolver(nil, nil)
			resolver.Resolve("never-stored")
			for _, def := range resolver.All() {
				Expect(def.Scope).NotTo(Equal(labels.ScopeDerived))
			}
		})
	})
})
