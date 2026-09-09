package commit

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/verify"
)

var _ = Describe("session staging", func() {
	It("includes nested Claude sub-agent edits", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		repo := initCaptainSessionRepo()
		writeCaptainSessionFile(repo, "root.go")
		writeCaptainSessionFile(repo, "agent.go")
		writeCaptainSessionFile(repo, "unrelated.go")

		sessionID := "11111111-1111-1111-1111-111111111111"
		writeCaptainSessionFixture(home, sessionID, repo)

		Expect(stageFiles(repo, sessionID[:8], verify.CommitConfig{})).To(Succeed())
		staged, err := stagedFiles(repo)
		Expect(err).ToNot(HaveOccurred())
		Expect(staged).To(ConsistOf("root.go", "agent.go"))
	})

	It("uses Captain custom tool analysis for Codex writes", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		repo := initCaptainSessionRepo()
		writeCaptainSessionFile(repo, "shell.go")
		writeCaptainSessionFile(repo, "unrelated.go")

		sessionID := "22222222-2222-2222-2222-222222222222"
		writeCaptainCodexFixture(home, sessionID, repo, "shell.go")

		Expect(stageFiles(repo, sessionID[:8], verify.CommitConfig{})).To(Succeed())
		staged, err := stagedFiles(repo)
		Expect(err).ToNot(HaveOccurred())
		Expect(staged).To(Equal([]string{"shell.go"}))
	})
})

func initCaptainSessionRepo() string {
	dir := GinkgoT().TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
		{"config", "commit.gpgsign", "false"},
	} {
		runCaptainSessionGit(dir, args...)
	}
	writeCaptainSessionFile(dir, "README.md")
	runCaptainSessionGit(dir, "add", "README.md")
	runCaptainSessionGit(dir, "commit", "-m", "initial")
	return dir
}

func runCaptainSessionGit(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), "git %v: %s", args, output)
}

func writeCaptainSessionFile(dir, name string) {
	Expect(os.WriteFile(filepath.Join(dir, name), []byte("changed\n"), 0o644)).To(Succeed())
}

func writeCaptainSessionFixture(home, sessionID, repo string) {
	projectDir := filepath.Join(home, ".claude", "projects", "repo")
	subagentDir := filepath.Join(projectDir, sessionID, "subagents")
	Expect(os.MkdirAll(subagentDir, 0o755)).To(Succeed())

	writeCaptainSessionTranscript(filepath.Join(projectDir, sessionID+".jsonl"), map[string]any{
		"sessionId": sessionID,
		"uuid":      "root-edit",
		"timestamp": "2026-07-16T10:00:00Z",
		"message": map[string]any{"role": "assistant", "content": []map[string]any{{
			"type": "tool_use", "name": "Edit", "input": map[string]any{"file_path": filepath.Join(repo, "root.go")},
		}}},
	})
	writeCaptainSessionTranscript(filepath.Join(subagentDir, "agent-worker.jsonl"), map[string]any{
		"sessionId":   sessionID,
		"uuid":        "agent-edit",
		"timestamp":   "2026-07-16T10:01:00Z",
		"isSidechain": true,
		"agentId":     "worker",
		"message": map[string]any{"role": "assistant", "content": []map[string]any{{
			"type": "tool_use", "name": "Edit", "input": map[string]any{"file_path": filepath.Join(repo, "agent.go")},
		}}},
	})
}

func writeCaptainSessionTranscript(path string, entry map[string]any) {
	data, err := json.Marshal(entry)
	Expect(err).ToNot(HaveOccurred())
	Expect(os.WriteFile(path, append(data, '\n'), 0o644)).To(Succeed())
}

func writeCaptainCodexFixture(home, sessionID, repo, relPath string) {
	dir := filepath.Join(home, ".codex", "sessions", "2026", "07", "16")
	Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
	input := "const patch = \"*** Begin Patch\\n*** Update File: " + relPath + "\\n@@\\n-old\\n+new\\n*** End Patch\";\n" +
		"text(await tools.apply_patch(patch));\n"

	entries := []map[string]any{
		{"timestamp": "2026-07-16T10:00:00Z", "type": "session_meta", "payload": map[string]any{
			"id": sessionID, "cwd": repo, "model_provider": "openai",
		}},
		{"timestamp": "2026-07-16T10:01:00Z", "type": "response_item", "payload": map[string]any{
			"type": "custom_tool_call", "name": "exec", "input": input, "call_id": "call-patch",
		}},
	}
	path := filepath.Join(dir, "rollout-2026-07-16T10-00-00-"+sessionID+".jsonl")
	var content []byte
	for _, entry := range entries {
		line, err := json.Marshal(entry)
		Expect(err).ToNot(HaveOccurred())
		content = append(content, line...)
		content = append(content, '\n')
	}
	Expect(os.WriteFile(path, content, 0o644)).To(Succeed())
}
