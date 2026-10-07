package gitstate_test

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	fresh    = 10 * time.Second
	maxStale = 2 * time.Minute
)

// fakeCompute counts computations and, while gate is non-nil, blocks each one
// until the gate is closed. The base it reports is the computation's ordinal,
// so a spec can tell which computation a state came from.
type fakeCompute struct {
	mu        sync.Mutex
	calls     atomic.Int32
	gate      chan struct{}
	err       error
	worktrees []gavelgit.Worktree
}

func (f *fakeCompute) compute(_ context.Context, dir string) (gitstate.State, error) {
	n := f.calls.Add(1)
	f.mu.Lock()
	gate, err := f.gate, f.err
	worktrees := append([]gavelgit.Worktree{{Path: dir, Branch: "main", Primary: true}}, f.worktrees...)
	f.mu.Unlock()
	if gate != nil {
		<-gate
	}
	if err != nil {
		return gitstate.State{}, err
	}
	state := gitstate.State{Base: string(rune('0' + n)), CurrentBranch: "main"}
	for _, wt := range worktrees {
		state.Worktrees = append(state.Worktrees, gitstate.Worktree{Worktree: wt})
	}
	return state, nil
}

func (f *fakeCompute) block() chan struct{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gate = make(chan struct{})
	return f.gate
}

func (f *fakeCompute) unblock() {
	f.mu.Lock()
	defer f.mu.Unlock()
	close(f.gate)
	f.gate = nil
}

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

var _ = Describe("Service", func() {
	var (
		fake    *fakeCompute
		clock   *fakeClock
		service *gitstate.Service
		dir     string
		ctx     context.Context
	)

	BeforeEach(func() {
		fake = &fakeCompute{}
		clock = &fakeClock{now: time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)}
		service = gitstate.New(gitstate.Options{Compute: fake.compute, Fresh: fresh, MaxStale: maxStale, Now: clock.Now})
		var err error
		dir, err = filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		ctx = context.Background()
	})

	get := func() gitstate.State {
		GinkgoHelper()
		state, err := service.Get(ctx, dir)
		Expect(err).NotTo(HaveOccurred())
		return state
	}

	It("runs one computation for concurrent Gets of a directory", func() {
		gate := fake.block()
		var wg sync.WaitGroup
		bases := make([]string, 8)
		for i := range bases {
			wg.Add(1)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()
				bases[i] = get().Base
			}()
		}
		Eventually(fake.calls.Load).Should(BeEquivalentTo(1))
		time.Sleep(50 * time.Millisecond)
		close(gate)
		wg.Wait()

		Expect(fake.calls.Load()).To(BeEquivalentTo(1))
		Expect(bases).To(HaveEach("1"))
	})

	It("serves a fresh entry without computing and stamps it with the service clock", func() {
		first := get()
		clock.Advance(fresh - time.Second)

		Expect(get()).To(Equal(first))
		Expect(fake.calls.Load()).To(BeEquivalentTo(1))
		Expect(first.ComputedAt).To(Equal(time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)))
	})

	It("serves a stale entry immediately and refreshes it once in the background", func() {
		Expect(get().Base).To(Equal("1"))
		clock.Advance(fresh + time.Second)
		gate := fake.block()

		Expect(get().Base).To(Equal("1"))
		Expect(get().Base).To(Equal("1"))
		Eventually(fake.calls.Load).Should(BeEquivalentTo(2))
		close(gate)

		Eventually(func() string { return get().Base }).Should(Equal("2"))
		Expect(fake.calls.Load()).To(BeEquivalentTo(2))
	})

	It("blocks on a fresh computation once an entry is past MaxStale", func() {
		get()
		clock.Advance(maxStale)

		Expect(get().Base).To(Equal("2"))
		Expect(fake.calls.Load()).To(BeEquivalentTo(2))
	})

	It("never caches an error", func() {
		fake.err = errors.New("git exploded")
		_, err := service.Get(ctx, dir)
		Expect(err).To(MatchError("git exploded"))

		fake.err = nil
		Expect(get().Base).To(Equal("2"))
	})

	It("does not let a canceled caller fail the shared computation", func() {
		gate := fake.block()
		canceled, cancel := context.WithCancel(ctx)
		errs := make(chan error, 1)
		go func() {
			_, err := service.Get(canceled, dir)
			errs <- err
		}()
		Eventually(fake.calls.Load).Should(BeEquivalentTo(1))
		cancel()
		Eventually(errs).Should(Receive(MatchError(context.Canceled)))
		close(gate)

		Eventually(func() string { return get().Base }).Should(Equal("1"))
		Expect(fake.calls.Load()).To(BeEquivalentTo(1))
	})

	It("drops the project that owns an invalidated linked worktree path", func() {
		linked := filepath.Join(filepath.Dir(dir), "linked")
		fake.worktrees = []gavelgit.Worktree{{Path: linked, Branch: "feature"}}
		get()

		service.Invalidate(linked)

		Expect(get().Base).To(Equal("2"))
	})

	It("keeps entries an invalidation does not name", func() {
		get()

		service.Invalidate(filepath.Join(dir, "elsewhere"))

		Expect(get().Base).To(Equal("1"))
	})

	It("does not store a computation that was in flight when its project was invalidated", func() {
		gate := fake.block()
		results := make(chan string, 1)
		go func() {
			defer GinkgoRecover()
			results <- get().Base
		}()
		Eventually(fake.calls.Load).Should(BeEquivalentTo(1))

		service.Invalidate(dir)
		close(gate)
		Eventually(results).Should(Receive(Equal("1")))

		Expect(get().Base).To(Equal("2"))
	})

	Describe("WorktreeOf", func() {
		It("re-reads once to find a worktree created after the state was cached", func() {
			get()
			created, err := filepath.EvalSymlinks(GinkgoT().TempDir())
			Expect(err).NotTo(HaveOccurred())
			fake.worktrees = []gavelgit.Worktree{{Path: created, Branch: "feature"}}

			wt, found, err := service.WorktreeOf(ctx, dir, created)

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(wt.Branch).To(Equal("feature"))
			Expect(fake.calls.Load()).To(BeEquivalentTo(2))
		})

		It("reports a path that is no worktree as not found after one re-read", func() {
			_, found, err := service.WorktreeOf(ctx, dir, GinkgoT().TempDir())

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
			Expect(fake.calls.Load()).To(BeEquivalentTo(2))
		})

		It("answers from the cache when the worktree is already known", func() {
			get()

			wt, found, err := service.WorktreeOf(ctx, dir, dir)

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(wt.Primary).To(BeTrue())
			Expect(fake.calls.Load()).To(BeEquivalentTo(1))
		})
	})
})
