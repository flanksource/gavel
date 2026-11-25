//go:build integration

package cliintegration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSlotsCLI(t *testing.T) {
	// The gavel processes under test must contend for slots, not inherit the
	// lease of a gavel running this test.
	t.Setenv("GAVEL_SLOT_HELD", "")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gavel CLI slots integration")
}

var _ = Describe("Concurrent gavel test processes", func() {
	var root, home, binary, events string
	BeforeEach(func() {
		root, home = GinkgoT().TempDir(), GinkgoT().TempDir()
		binary = os.Getenv("GAVEL_SLOTS_BINARY")
		Expect(binary).NotTo(BeEmpty(), "set GAVEL_SLOTS_BINARY to the gavel binary built by make build")
		resolved, err := exec.LookPath(binary)
		Expect(err).NotTo(HaveOccurred(), "GAVEL_SLOTS_BINARY=%q", binary)
		binary, err = filepath.Abs(resolved)
		Expect(err).NotTo(HaveOccurred())
		Expect(binary).To(BeAnExistingFile())
		events = filepath.Join(root, "events.jsonl")
	})
	It("shares one project limit across linked worktrees", func() {
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("slots: {global: 6, projectDefault: 4}\n"), 0o600)).To(Succeed())
		repo := createRepository(root, "project", 2)
		worktree := filepath.Join(root, "worktree")
		git(repo, "worktree", "add", "--detach", worktree)
		runConcurrent(binary, home, events, []string{repo, worktree})
	})
	It("shares the global limit across distinct repositories", func() {
		Expect(os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("slots: {global: 2, projectDefault: 4}\n"), 0o600)).To(Succeed())
		runConcurrent(binary, home, events, []string{createRepository(root, "first", 4), createRepository(root, "second", 4)})
	})
})

func createRepository(root, name string, limit int) string {
	dir := filepath.Join(root, name)
	Expect(os.MkdirAll(dir, 0o700)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/slot-holder\n\ngo 1.22\n"), 0o600)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, "holder_test.go"), []byte(holderSource), 0o600)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(fmt.Sprintf("slots: {project: %d}\n", limit)), 0o600)).To(Succeed())
	git(dir, "init", "-q")
	git(dir, "add", "go.mod", "holder_test.go", ".gavel.yaml")
	git(dir, "-c", "user.name=Slot Test", "-c", "user.email=slots@example.com", "commit", "-qm", "initial")
	return dir
}

func git(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))
}

func runConcurrent(binary, home, events string, repos []string) {
	for _, repo := range repos {
		warm := exec.Command("go", "test", "-run=^$", ".")
		warm.Dir = repo
		warm.Env = append(os.Environ(), "GOWORK=off")
		output, err := warm.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
	}
	var commands []*exec.Cmd
	var outputs []*bytes.Buffer
	for i := range 6 {
		cmd := exec.Command(binary, "test", ".", "--recursive=false", "--framework", "go", "--cache=false", "--skip-hooks", "--timeout", "2m", "--", "-count=1")
		cmd.Dir = repos[i%len(repos)]
		cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME="+filepath.Join(home, "cache"), "GOWORK=off", "GAVEL_SLOT_EVENTS="+events)
		output := &bytes.Buffer{}
		cmd.Stdout, cmd.Stderr = output, output
		Expect(cmd.Start()).To(Succeed())
		DeferCleanup(func() { _ = cmd.Process.Kill() })
		commands, outputs = append(commands, cmd), append(outputs, output)
	}
	for i, cmd := range commands {
		Expect(cmd.Wait()).To(Succeed(), outputs[i].String())
	}
	data, err := os.ReadFile(events)
	Expect(err).NotTo(HaveOccurred())
	var records []event
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var record event
		Expect(json.Unmarshal([]byte(line), &record)).To(Succeed())
		records = append(records, record)
	}
	Expect(records).To(HaveLen(12))
	sort.Slice(records, func(i, j int) bool { return records[i].Time < records[j].Time })
	active, maximum := 0, 0
	for _, record := range records {
		if record.Enter {
			active++
		} else {
			active--
		}
		maximum = max(maximum, active)
	}
	Expect(active).To(BeZero())
	Expect(maximum).To(Equal(2))
}

type event struct {
	Enter bool
	Time  int64
}

const holderSource = `package holder
import (
 "encoding/json"
 "os"
 "testing"
 "time"
)
func TestHolding(t *testing.T) {
 f, err := os.OpenFile(os.Getenv("GAVEL_SLOT_EVENTS"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
 if err != nil { t.Fatal(err) }
 defer f.Close()
 write := func(enter bool) {
  data, err := json.Marshal(struct { Enter bool; Time int64 }{enter, time.Now().UnixNano()})
  if err != nil { t.Fatal(err) }
  if _, err := f.Write(append(data, '\n')); err != nil { t.Fatal(err) }
 }
 write(true)
 time.Sleep(800*time.Millisecond)
 write(false)
}
`
