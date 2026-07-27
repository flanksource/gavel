package ui

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/flanksource/gavel/procfile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("archived run cache", func() {
	It("serves Runs and Snapshot for one render from a single archive read", func() {
		loads := 0
		record := taskhistory.Record{
			Run: clickytask.RunMeta{
				ID: "cached-run", Name: "Cached", Status: string(clickytask.StatusSuccess),
				StartedAt: time.Now().Add(-time.Hour).Format(time.RFC3339Nano),
			},
			Snapshots:  []clickytask.TaskSnapshot{{ID: "cached-run", GroupID: "cached-run", Type: "group"}},
			ArchivedAt: time.Now().Add(-time.Hour),
		}
		source := &supervisorTaskSource{history: func(context.Context) ([]taskhistory.Record, error) {
			loads++
			return []taskhistory.Record{record}, nil
		}}

		originalProjectsPath := projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		DeferCleanup(func() { projectsPath = originalProjectsPath })

		runs, err := source.Runs(GinkgoT().Context(), clickytask.RunFilter{})
		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(HaveLen(1))
		snapshots, err := source.Snapshot(GinkgoT().Context(), "cached-run")
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshots).To(Equal(record.Snapshots))

		Expect(loads).To(Equal(1), "Runs and Snapshot within one TTL window must share one archive read")

		source.cachedAt = time.Now().Add(-2 * historyCacheTTL)
		_, err = source.Runs(GinkgoT().Context(), clickytask.RunFilter{})
		Expect(err).NotTo(HaveOccurred())
		Expect(loads).To(Equal(2), "an expired cache must re-read the archive")
	})
})

var _ = Describe("task manager API", func() {
	It("merges and controls task generations owned by detached process supervisors", func() {
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "Procfile"), []byte("worker: sh -c 'echo ready; sleep 30'\n"), 0o644)).To(Succeed())
		supervisor, err := procfile.NewSupervisor(procfile.Options{Root: root, Procfile: filepath.Join(root, "Procfile")})
		Expect(err).NotTo(HaveOccurred())
		Expect(supervisor.Start()).To(Succeed())
		DeferCleanup(supervisor.Shutdown)

		originalProjectsPath := projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		Expect(SaveProjects([]Project{
			{Name: "workspace", Dir: root},
			{Name: "without-procfile", Dir: GinkgoT().TempDir()},
		})).To(Succeed())
		DeferCleanup(func() { projectsPath = originalProjectsPath })
		archived := taskhistory.Record{
			Run: clickytask.RunMeta{
				ID: "archived-only", Name: "Archived task", Status: string(clickytask.StatusFailed),
				StartedAt: time.Now().Add(-time.Hour).Format(time.RFC3339Nano), Total: 1, Failed: 1,
			},
			Snapshots: []clickytask.TaskSnapshot{
				{
					ID: "archived-only", GroupID: "archived-only", Name: "Archived task", Type: "group",
					Status: string(clickytask.StatusFailed), Total: 1, Failed: 1,
				},
				{
					ID: "archived-step", GroupID: "archived-only", Name: "Archived step", Type: "task",
					Status: string(clickytask.StatusFailed), Stdout: "archived stdout\n", Stderr: "archived stderr\n",
				},
			},
			ArchivedAt: time.Now().Add(-time.Hour),
		}
		Expect(taskhistory.Write(root, archived)).To(Succeed())

		server := httptest.NewServer((&Server{taskSource: &supervisorTaskSource{
			history: func(context.Context) ([]taskhistory.Record, error) { return []taskhistory.Record{archived}, nil },
		}}).Handler())
		DeferCleanup(server.Close)
		var run clickytask.RunMeta
		Eventually(func(g Gomega) {
			response, err := http.Get(server.URL + "/api/v1/tasks?kind=supervised-process")
			g.Expect(err).NotTo(HaveOccurred())
			defer response.Body.Close()
			var runs []clickytask.RunMeta
			g.Expect(json.NewDecoder(response.Body).Decode(&runs)).To(Succeed())
			g.Expect(runs).NotTo(BeEmpty())
			run = runs[0]
		}).WithTimeout(3 * time.Second).Should(Succeed())
		Expect(run.Labels).To(HaveKeyWithValue("process", "worker"))

		response, err := http.Get(server.URL + "/api/v1/tasks/" + run.ID)
		Expect(err).NotTo(HaveOccurred())
		defer response.Body.Close()
		var snapshots []clickytask.TaskSnapshot
		Expect(json.NewDecoder(response.Body).Decode(&snapshots)).To(Succeed())
		Expect(snapshots).To(HaveLen(2))

		request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/tasks/"+run.ID+"/control", strings.NewReader(`{"action":"restart"}`))
		Expect(err).NotTo(HaveOccurred())
		request.Header.Set("Content-Type", "application/json")
		response, err = http.DefaultClient.Do(request)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Body.Close()).To(Succeed())
		Expect(response.StatusCode).To(Equal(http.StatusNoContent))
		Eventually(func() string {
			process, _ := supervisor.State().Process("worker")
			return process.TaskRunID
		}).WithTimeout(4 * time.Second).ShouldNot(Equal(run.ID))
		response, err = http.Get(server.URL + "/api/v1/tasks/archived-only")
		Expect(err).NotTo(HaveOccurred())
		defer response.Body.Close()
		snapshots = nil
		Expect(json.NewDecoder(response.Body).Decode(&snapshots)).To(Succeed())
		Expect(snapshots).To(Equal(archived.Snapshots))

		response, err = http.Get(server.URL + "/api/v1/tasks/stream?tasks=archived-only")
		Expect(err).NotTo(HaveOccurred())
		stream, err := io.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Body.Close()).To(Succeed())
		body := string(stream)
		metadataIndex := strings.Index(body, "event: task\ndata: {\"id\":\"archived-step\"")
		stdoutIndex := strings.Index(body, `"stream":"stdout","data":"archived stdout\n"`)
		stderrIndex := strings.Index(body, `"stream":"stderr","data":"archived stderr\n"`)
		doneIndex := strings.Index(body, "event: done")
		Expect(metadataIndex).To(BeNumerically(">=", 0), body)
		Expect(stdoutIndex).To(BeNumerically(">", metadataIndex), body)
		Expect(stderrIndex).To(BeNumerically(">", metadataIndex), body)
		Expect(doneIndex).To(BeNumerically(">", stdoutIndex), body)
		Expect(doneIndex).To(BeNumerically(">", stderrIndex), body)
	})
})
