package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// sentinelName is written into the repo only if some layer ever evaluates the
// file argument as shell syntax; its absence is the injection assertion.
const sentinelName = "pwned-by-shell-injection"

// shellMetaName is a legal POSIX filename built entirely from shell
// metacharacters (command substitution, backticks, a statement separator). It is
// harmless as an argv element and catastrophic as a shell word.
var shellMetaName = "a;$(touch " + sentinelName + ")`touch " + sentinelName + "`.txt"

const plainName = "hello.txt"

// plainBody is the committed content of plainName; the patch must add it.
const plainBody = "hi\n"

var _ = Describe("CommitDiff file argument", func() {
	var dir, head string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		runGitIn := func(args ...string) {
			GinkgoHelper()
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred(), "git %v: %s", args, out)
		}
		runGitIn("init", "-q")
		runGitIn("config", "user.email", "test@example.com")
		runGitIn("config", "user.name", "Test User")
		runGitIn("config", "commit.gpgsign", "false")
		for name, body := range map[string]string{plainName: plainBody, shellMetaName: "meta\n"} {
			Expect(os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)).To(Succeed())
		}
		runGitIn("add", "-A")
		runGitIn("commit", "-m", "feat: add fixtures")

		cmd := exec.Command("git", "rev-parse", "HEAD")
		cmd.Dir = dir
		out, err := cmd.Output()
		Expect(err).ToNot(HaveOccurred())
		head = strings.TrimSpace(string(out))
	})

	It("passes a shell-metacharacter path as one inert argument", func() {
		diff, _, err := CommitDiff(dir, head, shellMetaName)
		Expect(err).ToNot(HaveOccurred())
		Expect(stripANSI(diff)).To(ContainSubstring(shellMetaName))
		_, statErr := os.Stat(filepath.Join(dir, sentinelName))
		Expect(os.IsNotExist(statErr)).To(BeTrue(),
			"a shell evaluated the file argument and created %s", sentinelName)
	})

	It("still returns the patch for an ordinary path", func() {
		diff, truncated, err := CommitDiff(dir, head, plainName)
		Expect(err).ToNot(HaveOccurred())
		Expect(truncated).To(BeFalse())
		plain := stripANSI(diff)
		Expect(plain).To(ContainSubstring(plainName))
		Expect(plain).To(ContainSubstring("+" + strings.TrimSuffix(plainBody, "\n")))
		Expect(plain).ToNot(ContainSubstring(shellMetaName), "the diff must stay scoped to one path")
	})

	It("still returns the whole commit when no path is given", func() {
		diff, _, err := CommitDiff(dir, head, "")
		Expect(err).ToNot(HaveOccurred())
		plain := stripANSI(diff)
		Expect(plain).To(ContainSubstring(plainName))
		Expect(plain).To(ContainSubstring(shellMetaName))
	})

	DescribeTable("rejects a path git would reinterpret instead of reading literally",
		func(file, wantReason string) {
			_, _, err := CommitDiff(dir, head, file)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(wantReason))
			// The message quotes the rejected value so control characters stay
			// visible; that quoted form is what must name what was received.
			Expect(err.Error()).To(ContainSubstring(strconv.Quote(file)),
				"the error must name the received value")
		},
		Entry("short option", "-p", "must not begin with"),
		Entry("long option", "--output=owned.txt", "must not begin with"),
		Entry("long option with separator", "--exec-path=/tmp", "must not begin with"),
		Entry("pathspec magic exclude", ":(exclude)"+plainName, "pathspec magic"),
		Entry("pathspec magic root", ":/"+plainName, "pathspec magic"),
		Entry("absolute path", "/etc/passwd", "not absolute"),
		Entry("parent traversal", "../../etc/passwd", "escape the repository"),
		Entry("nested parent traversal", "sub/../../"+plainName, "escape the repository"),
		Entry("embedded newline", plainName+"\n--all", "control character"),
		Entry("embedded NUL", plainName+"\x00--all", "control character"),
	)
})

var _ = DescribeTable("validateDiffPath accepts ordinary repository paths",
	func(file string) {
		Expect(validateDiffPath(file)).To(Succeed())
	},
	Entry("bare file", plainName),
	Entry("nested file", "pkg/sub/main.go"),
	Entry("dot-prefixed file", ".gavel.yaml"),
	Entry("single-dot segment", "./"+plainName),
	Entry("shell metacharacters", shellMetaName),
	Entry("dash inside the name", "pr-ui/todo-commits.go"),
)
