package slots

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type holderRequest struct {
	Request
	Events  string
	Ready   string
	Hold    time.Duration
	Timeout time.Duration
}

type holderEvent struct {
	Enter bool
	Time  int64
}

func TestMain(m *testing.M) {
	if raw := os.Getenv("GAVEL_SLOT_TEST_HOLDER"); raw != "" {
		if err := runHolder(raw); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func runHolder(raw string) (err error) {
	var request holderRequest
	if err := json.Unmarshal([]byte(raw), &request); err != nil {
		return err
	}
	lease, err := Acquire(context.Background(), request.Request)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lease.Close()) }()
	if err := os.WriteFile(request.Ready, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		return err
	}
	if err := recordHolder(request.Events, true); err != nil {
		return err
	}
	ctx := context.Background()
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}
	err = exec.CommandContext(ctx, "sleep", fmt.Sprint(request.Hold.Seconds())).Run()
	if request.Timeout > 0 && ctx.Err() == context.DeadlineExceeded {
		err = nil
	}
	return errors.Join(err, recordHolder(request.Events, false))
}

func recordHolder(path string, enter bool) error {
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	data, err := json.Marshal(holderEvent{Enter: enter, Time: time.Now().UnixNano()})
	if err == nil {
		_, err = f.Write(append(data, '\n'))
	}
	return errors.Join(err, f.Close())
}

var _ = Describe("Slot subprocess integration", func() {
	var root, cache string
	BeforeEach(func() { root = GinkgoT().TempDir(); cache = filepath.Join(root, "cache") })

	It("shares the project limit between two worktrees", func() {
		repo := slotRepo(root, "repo")
		worktree := filepath.Join(root, "worktree")
		slotGit(repo, "worktree", "add", "--detach", worktree)
		firstDir, first, err := ProjectDir(context.Background(), repo)
		Expect(err).NotTo(HaveOccurred())
		secondDir, second, err := ProjectDir(context.Background(), worktree)
		Expect(err).NotTo(HaveOccurred())
		Expect(second).To(Equal(first))
		Expect(secondDir).To(Equal(firstDir))
		assertHolderConcurrency(root, cache, []string{repo, worktree}, Limits{Global: 5, Project: 2}, 2)
	})
	It("enforces the global limit across repositories", func() {
		assertHolderConcurrency(root, cache, []string{slotRepo(root, "one"), slotRepo(root, "two")}, Limits{Global: 2, Project: 4}, 2)
	})
	It("reclaims a SIGKILLed holder and its stale PID record", func() {
		request := holderRequest{Request: Request{Limits: Limits{Global: 1}, WorkDir: root, CacheDir: cache, Command: "killed holder"}, Ready: filepath.Join(root, "ready"), Hold: time.Minute}
		cmd, output := startHolder(request)
		Eventually(func() bool { _, err := os.Stat(request.Ready); return err == nil }).WithTimeout(10*time.Second).Should(BeTrue(), output.String)
		Expect(cmd.Process.Kill()).To(Succeed())
		Expect(cmd.Wait()).To(HaveOccurred())
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		lease, err := Acquire(ctx, request.Request)
		Expect(err).NotTo(HaveOccurred())
		Expect(lease.Close()).To(Succeed())
	})
	It("releases a slot when its subprocess times out", func() {
		request := holderRequest{Request: Request{Limits: Limits{Global: 1}, WorkDir: root, CacheDir: cache, Command: "timed holder"}, Ready: filepath.Join(root, "ready"), Hold: time.Minute, Timeout: 100 * time.Millisecond}
		cmd, output := startHolder(request)
		Expect(cmd.Wait()).To(Succeed(), output.String())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lease, err := Acquire(ctx, request.Request)
		Expect(err).NotTo(HaveOccurred())
		Expect(lease.Close()).To(Succeed())
	})
	It("does not reserve global capacity for a blocked project", func() {
		one, two := slotRepo(root, "one"), slotRepo(root, "two")
		request := Request{Limits: Limits{Global: 2, Project: 1}, WorkDir: one, CacheDir: cache, Command: "first"}
		lease, err := Acquire(context.Background(), request)
		Expect(err).NotTo(HaveOccurred())
		defer lease.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()
		done := make(chan error, 1)
		go func(request Request) { _, err := Acquire(ctx, request); done <- err }(request)
		request.WorkDir = two
		other, err := Acquire(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(other.Close()).To(Succeed())
		Expect(<-done).To(MatchError(context.DeadlineExceeded))
	})
})

func slotRepo(root, name string) string {
	dir := filepath.Join(root, name)
	Expect(os.MkdirAll(dir, 0o700)).To(Succeed())
	slotGit(dir, "init", "-q")
	slotGit(dir, "-c", "user.name=Slot Test", "-c", "user.email=slots@example.com", "commit", "--allow-empty", "-qm", "initial")
	return dir
}

func slotGit(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))
}

func startHolder(request holderRequest) (*exec.Cmd, *bytes.Buffer) {
	encoded, err := json.Marshal(request)
	Expect(err).NotTo(HaveOccurred())
	executable, err := os.Executable()
	Expect(err).NotTo(HaveOccurred())
	cmd := exec.Command(executable)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Dir = request.WorkDir
	cmd.Env = append(os.Environ(), "GAVEL_SLOT_TEST_HOLDER="+string(encoded))
	output := &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = output, output
	Expect(cmd.Start()).To(Succeed())
	DeferCleanup(func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) })
	return cmd, output
}

func assertHolderConcurrency(root, cache string, repos []string, limits Limits, expected int) {
	var commands []*exec.Cmd
	var outputs []*bytes.Buffer
	events := filepath.Join(root, "events.jsonl")
	for i := range 8 {
		cmd, output := startHolder(holderRequest{Request: Request{Limits: limits, WorkDir: repos[i%len(repos)], CacheDir: cache, Command: "concurrent holder"}, Events: events, Ready: filepath.Join(root, fmt.Sprint(i)), Hold: 200 * time.Millisecond})
		commands, outputs = append(commands, cmd), append(outputs, output)
	}
	for i, cmd := range commands {
		Expect(cmd.Wait()).To(Succeed(), outputs[i].String())
	}
	file, err := os.Open(events)
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()
	var records []holderEvent
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event holderEvent
		Expect(json.Unmarshal(scanner.Bytes(), &event)).To(Succeed())
		records = append(records, event)
	}
	Expect(scanner.Err()).NotTo(HaveOccurred())
	Expect(records).To(HaveLen(16))
	sort.Slice(records, func(i, j int) bool { return records[i].Time < records[j].Time })
	active, maximum := 0, 0
	for _, event := range records {
		if event.Enter {
			active++
		} else {
			active--
		}
		maximum = max(maximum, active)
	}
	Expect(active).To(BeZero())
	Expect(maximum).To(Equal(expected))
}
