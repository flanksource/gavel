package ui

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flanksource/commons-db/dbtest"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// trackedGitServer is a Server whose context carries a git state tracker over
// a freshly migrated database. The tracker is not started: specs drive its
// scans through reads and Touch, never through the background cadence.
type trackedGitServer struct {
	server  *Server
	tracker *gitstate.Tracker
	db      *gorm.DB
	// ctx is the process context: it outlives each request and ends with the
	// spec.
	ctx context.Context
}

func newTrackedGitServer() trackedGitServer {
	GinkgoHelper()
	handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_git_nav"})
	GinkgoT().Setenv(database.EnvDSN, handle.DSN())
	GinkgoT().Setenv(database.EnvDisable, "")
	GinkgoT().Setenv(database.LegacyEnvDisable, "")
	opened, err := database.Open(context.Background(), database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)
	tracker := gitstate.NewTracker(gitstate.Options{Store: gitstate.NewStore(opened.Gorm()), Context: ctx})
	return trackedGitServer{
		server:  &Server{ctx: gavelctx.New(ctx, gavelctx.WithGitTracker(tracker))},
		tracker: tracker, db: opened.Gorm(), ctx: ctx,
	}
}

// withoutGit points PATH at an empty directory for the rest of the spec, so
// any git a handler runs fails instead of answering.
func withoutGit() {
	GinkgoHelper()
	GinkgoT().Setenv("PATH", GinkgoT().TempDir())
}

// sseFrames reads the data frames of an event stream, one JSON payload per
// frame, until the body closes.
func sseFrames(resp *http.Response) <-chan string {
	frames := make(chan string, 16)
	go func() {
		defer GinkgoRecover()
		defer close(frames)
		scanner := bufio.NewScanner(resp.Body)
		var data []string
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "data: "):
				data = append(data, strings.TrimPrefix(line, "data: "))
			case line == "" && len(data) > 0:
				frames <- strings.Join(data, "\n")
				data = nil
			}
		}
	}()
	return frames
}

func decodeGenerations(frame string) map[string]int64 {
	GinkgoHelper()
	var generations map[string]int64
	Expect(json.Unmarshal([]byte(frame), &generations)).To(Succeed(), frame)
	return generations
}
