package commit

import (
	"errors"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/flanksource/gavel/status"
	zone "github.com/lrstanley/bubblezone/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("interactive patch preview", func() {
	var m treeModel

	BeforeEach(func(ctx SpecContext) {
		m = newTreeModel([]status.FileStatus{{Path: "dir/nested.go"}, {Path: "alpha.go"}, {Path: "bravo.go"}})
		m.preview = newTreePreview(ctx)
		m.zones = zone.New()
		DeferCleanup(m.zones.Close)
	})

	It("shows only the focused file's patch and ignores late results after navigation", func(ctx SpecContext) {
		m, _ = updateKey(m, "down") // dir/nested.go
		first := m.preview.request
		Expect(ansi.Strip(m.View().Content)).To(ContainSubstring("Loading patch…"))

		m, _ = updateKey(m, "down") // alpha.go
		second := m.preview.request
		request := treePatchRequest{id: second, file: *m.preview.node.File}
		msg := request.format(ctx, "@@ -1 +1 @@\n-before\n+after\n", nil)
		model, cmd := m.Update(msg)
		Expect(cmd).To(BeNil())
		m = model.(treeModel)
		Expect(m.View().Content).To(MatchRegexp(`\x1b\[[0-9;]+m\+`))

		model, _ = m.Update(treePatchMsg{request: first, content: "stale patch"})
		m = model.(treeModel)
		Expect(ansi.Strip(m.View().Content)).To(And(ContainSubstring("+after"), Not(ContainSubstring("stale patch"))))

		m, _ = updateKey(m, "/")
		m, _ = updateKey(m, "missing-file")
		Expect(ansi.Strip(m.View().Content)).To(And(ContainSubstring("No files match"), Not(ContainSubstring("+after"))))
	})

	It("uses mouse hover to focus and a checkbox click to toggle selection", func() {
		layout := m.layout()
		y := len(layout.header) + 1 // dir/nested.go
		hover := tea.MouseMotionMsg{X: 10, Y: y}
		view := m.View()
		Expect(view.AltScreen).To(BeTrue())
		Expect(view.MouseMode).To(Equal(tea.MouseModeAllMotion))
		Expect(view.Content).NotTo(MatchRegexp(`\x1b\[\d+z`))
		Eventually(func() bool { return m.zones.Get(rowZoneID(m.visible[1])).InBounds(hover) }).Should(BeTrue())

		model, _ := m.Update(hover)
		m = model.(treeModel)
		Expect(m.preview.node.Path).To(Equal("dir/nested.go"))
		Expect(m.selectedPaths()).To(BeEmpty())

		model, _ = m.Update(tea.MouseClickMsg{X: 4, Y: y, Button: tea.MouseLeft})
		m = model.(treeModel)
		Expect(m.selectedPaths()).To(Equal([]string{"dir/nested.go"}))
	})

	It("reflows the panes and scrolls the same patch with keys and the wheel", func() {
		m, _ = updateKey(m, "down")
		model, _ := m.Update(tea.WindowSizeMsg{Width: 112, Height: 20})
		m = model.(treeModel)
		model, _ = m.Update(treePatchMsg{request: m.preview.request, content: strings.Repeat(strings.Repeat("x", 170)+"\n", 60)})
		m = model.(treeModel)
		Expect(m.layout().sideBySide).To(BeTrue())

		wheel := tea.MouseWheelMsg{X: 108, Y: 10, Button: tea.MouseWheelDown}
		m.View()
		Eventually(func() bool { return m.zones.Get(patchZoneID).InBounds(wheel) }).Should(BeTrue())
		model, _ = m.Update(wheel)
		m = model.(treeModel)
		Expect(m.preview.pane.YOffset()).To(BeNumerically(">", 0))
		Expect(m.currentNode().Path).To(Equal("dir/nested.go"))

		model, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 18})
		m = model.(treeModel)
		Expect(m.layout().sideBySide).To(BeFalse())
		Expect(m.layout().patchWidth).To(Equal(60))
		model, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
		m = model.(treeModel)
		Expect(m.preview.pane.YOffset()).To(BeNumerically(">", 3))
		model, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift})
		m = model.(treeModel)
		Expect(m.preview.pane.XOffset()).To(BeNumerically(">", 0))
		lines := strings.Split(m.View().Content, "\n")
		Expect(lines).To(HaveLen(18))
		for _, line := range lines {
			Expect(ansi.StringWidth(line)).To(BeNumerically("<=", 60))
		}
	})

	It("renders diff status and keeps raw terminal controls out of the formatted patch", func(ctx SpecContext) {
		request := treePatchRequest{file: status.FileStatus{Path: "main.go"}}
		Expect(ansi.Strip(request.format(ctx, "", errors.New("git failed")).content)).To(ContainSubstring("git failed"))
		Expect(ansi.Strip(request.format(ctx, "", nil).content)).To(ContainSubstring("No net changes"))
		Expect(request.format(ctx, "diff --git a/image.bin b/image.bin\nBinary files a/image.bin and b/image.bin differ\n", nil).label).To(Equal("Binary"))
		Expect(request.format(ctx, strings.Repeat("+data\n", 60_000), nil).label).To(Equal("Truncated"))
		raw := request.format(ctx, "@@ -0,0 +1 @@\n+\tvalue\x1b[2J\n", nil).content
		Expect(raw).NotTo(ContainSubstring("\x1b[2J"))
		Expect(ansi.Strip(raw)).To(ContainSubstring("+    value"))
	})

	It("combines Go syntax colors with change backgrounds and preserves the patch text", func(ctx SpecContext) {
		const raw = `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,5 +1,5 @@
 package main
-func main() { println("before") }
+func main() { println("after") }
 var example = ` + "`\n" + `-return "inside old string"
+return "inside new string"
 ` + "`\n" + `\ No newline at end of file` + "\n"
		request := treePatchRequest{file: status.FileStatus{Path: "main.go"}}
		rendered := request.format(ctx, raw, nil).content
		Expect(ansi.Strip(rendered)).To(Equal(strings.TrimSuffix(raw, "\n")))

		lines := strings.Split(rendered, "\n")
		oldKeyword := patchForeground(lines[5], "func")
		newKeyword := patchForeground(lines[6], "func")
		oldString := patchForeground(lines[5], `"before"`)
		newString := patchForeground(lines[6], `"after"`)
		Expect(oldKeyword).NotTo(BeEmpty())
		Expect(newKeyword).To(Equal(oldKeyword))
		Expect(oldString).To(And(Not(BeEmpty()), Not(Equal(oldKeyword))))
		Expect(newString).To(Equal(oldString))
		Expect(patchForeground(lines[8], `return "inside old string"`)).To(Equal(oldString))
		Expect(patchForeground(lines[9], `return "inside new string"`)).To(Equal(newString))
		background := regexp.MustCompile(`48;2;\d+;\d+;\d+`)
		Expect(background.FindString(lines[5])).NotTo(BeEmpty())
		Expect(background.FindString(lines[6])).To(And(Not(BeEmpty()), Not(Equal(background.FindString(lines[5])))))
	})

	It("uses the old path for rename deletions and keeps diff styling for an unknown new extension", func(ctx SpecContext) {
		request := treePatchRequest{file: status.FileStatus{Path: "main.gavel-preview-unknown", PreviousPath: "main.go"}}
		const raw = "@@ -1 +1 @@\n-func before() {}\n+func after() {}"
		rendered := request.format(ctx, raw, nil).content
		Expect(ansi.Strip(rendered)).To(Equal(raw))

		lines := strings.Split(rendered, "\n")
		Expect(patchForeground(lines[1], "func")).NotTo(BeEmpty())
		Expect(lines[2]).To(And(MatchRegexp(`\x1b\[[0-9;]+m\+func after\(\) \{\}`), ContainSubstring("48;2;")))
	})
})

func patchForeground(line, token string) string {
	span := regexp.MustCompile(`\x1b\[([0-9;]+)m` + regexp.QuoteMeta(token)).FindStringSubmatch(line)
	if span == nil {
		return ""
	}
	return regexp.MustCompile(`38;2;\d+;\d+;\d+`).FindString(span[1])
}
