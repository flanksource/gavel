package fixtures

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/flanksource/clicky/process"
)

const profileInterval = 500 * time.Millisecond

type ProfileSample struct {
	SampledAt     time.Time         `json:"sampled_at"`
	CPUPercent    float64           `json:"cpu_percent"`
	MemoryPercent float64           `json:"memory_percent"`
	RSSBytes      uint64            `json:"rss_bytes"`
	Processes     []process.Process `json:"processes"`
}

type ProfileReport struct {
	IntervalMS        int64           `json:"interval_ms"`
	Samples           []ProfileSample `json:"samples"`
	PeakCPUPercent    float64         `json:"peak_cpu_percent"`
	PeakMemoryPercent float64         `json:"peak_memory_percent"`
	PeakRSSBytes      uint64          `json:"peak_rss_bytes"`
	DiskIO            *ProfileDiskIO  `json:"disk_io,omitempty"`
}

type profileSampler struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	report ProfileReport
	err    error
}

func startProfileSampler() *profileSampler {
	ctx, cancel := context.WithCancel(context.Background())
	sampler := &profileSampler{ctx: ctx, cancel: cancel, done: make(chan struct{}),
		report: ProfileReport{IntervalMS: profileInterval.Milliseconds(), Samples: []ProfileSample{}}}
	sampler.sample()
	go sampler.run()
	return sampler
}

func (s *profileSampler) run() {
	defer close(s.done)
	ticker := time.NewTicker(profileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.sample()
		}
	}
}

func (s *profileSampler) sample() {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	snapshot, err := process.Discover(ctx, process.SnapshotOptions{})
	if err != nil {
		if s.ctx.Err() == nil && !errors.Is(err, context.Canceled) {
			s.mu.Lock()
			s.err = fmt.Errorf("sample process tree: %w", err)
			s.mu.Unlock()
			s.cancel()
		}
		return
	}
	processes := snapshot.Subtree(os.Getpid())
	if len(processes) == 0 {
		s.mu.Lock()
		s.err = fmt.Errorf("gavel process %d missing from Clicky process snapshot", os.Getpid())
		s.mu.Unlock()
		s.cancel()
		return
	}
	pids := make([]int, 0, len(processes))
	for _, child := range processes {
		pids = append(pids, child.PID)
	}
	if err := snapshot.PopulateIO(ctx, pids); err != nil {
		s.mu.Lock()
		s.err = fmt.Errorf("sample process disk I/O: %w", err)
		s.mu.Unlock()
		s.cancel()
		return
	}
	processes = snapshot.Subtree(os.Getpid())
	sample := ProfileSample{SampledAt: time.Now().UTC(), Processes: processes}
	for _, child := range processes {
		sample.CPUPercent += child.CPUPercent
		sample.MemoryPercent += child.MemoryPercent
		sample.RSSBytes += child.RSSBytes
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.report.Samples = append(s.report.Samples, sample)
	s.report.PeakCPUPercent = max(s.report.PeakCPUPercent, sample.CPUPercent)
	s.report.PeakMemoryPercent = max(s.report.PeakMemoryPercent, sample.MemoryPercent)
	s.report.PeakRSSBytes = max(s.report.PeakRSSBytes, sample.RSSBytes)
}

func (s *profileSampler) Stop() (*ProfileReport, error) {
	s.cancel()
	<-s.done
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.report.Samples) > 0 {
		started := s.report.Samples[0].SampledAt
		finished := s.report.Samples[len(s.report.Samples)-1].SampledAt
		s.report.DiskIO = profileDiskIO(s.report.Samples, &ExecutionNode{StartedAt: &started, FinishedAt: &finished}, 0, runtime.GOOS)
	}
	return &s.report, s.err
}
