package jira

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func jiraProgressSnapshot(name string) task.TaskSnapshot {
	snapshots := task.SnapshotAll()
	for index := len(snapshots) - 1; index >= 0; index-- {
		snapshot := snapshots[index]
		if snapshot.Type == "task" && snapshot.Name == name {
			return snapshot
		}
	}
	return task.TaskSnapshot{}
}

var _ = Describe("Jira task progress", func() {
	BeforeEach(func() {
		task.ClearTasks()
		DeferCleanup(task.ClearTasks)
	})

	It("shows batch, issue, attachment and comment progress in one group", func() {
		var batchDuringSecond, commentsDuringSecond, downloadDuringCopy task.TaskSnapshot
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasPrefix(r.URL.Path, "/rest/api/3/issue/") && !strings.HasSuffix(r.URL.Path, "/comment") && !strings.HasSuffix(r.URL.Path, "/remotelink"):
				key := filepath.Base(r.URL.Path)
				if key == "EXAMPLE-2" {
					batchDuringSecond = jiraProgressSnapshot("Sync Jira issues")
				}
				fmt.Fprintf(w, `{"id":"1","key":%q,"fields":{"summary":"Example","project":{"id":"10001","key":"EXAMPLE"},"attachment":[{"id":"20001","filename":"notes.txt"}]}}`, key)
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field", strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				if r.URL.Query().Get("startAt") == "0" {
					fmt.Fprint(w, `{"total":2,"comments":[{"id":"c1","body":"First"}]}`)
				} else {
					commentsDuringSecond = jiraProgressSnapshot("Load Jira comments EXAMPLE-2")
					fmt.Fprint(w, `{"total":2,"comments":[{"id":"c2","body":"Second"}]}`)
				}
			case r.URL.Path == "/rest/api/3/attachment/content/20001":
				downloadDuringCopy = jiraProgressSnapshot("Download Jira attachment 20001")
				fmt.Fprint(w, "attachment content")
			default:
				http.NotFound(w, r)
			}
		})
		dir := GinkgoT().TempDir()
		results, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1", "EXAMPLE-2"}, Dir: dir, Attachments: true, Store: CatalogStore{Dir: dir}})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))
		Expect(batchDuringSecond.Status).To(Equal(string(task.StatusRunning)))
		Expect([]int{batchDuringSecond.Progress, batchDuringSecond.MaxValue}).To(Equal([]int{1, 2}))
		Expect(batchDuringSecond.Description).To(ContainSubstring("EXAMPLE-2"))
		Expect([]int{commentsDuringSecond.Progress, commentsDuringSecond.MaxValue}).To(Equal([]int{1, 2}))
		Expect(downloadDuringCopy.Status).To(Equal(string(task.StatusRunning)))
		Expect(downloadDuringCopy.Description).To(ContainSubstring("notes.txt"))
		groupID := jiraProgressSnapshot("Sync Jira issues").GroupID
		Expect(groupID).NotTo(BeEmpty())
		for _, snapshot := range task.SnapshotAll(groupID) {
			Expect(snapshot.Status).To(Equal(string(task.StatusSuccess)), snapshot.Name)
			Expect(snapshot.GroupID).To(Equal(groupID))
		}
		Expect(jiraProgressSnapshot("Sync EXAMPLE-2").Description).To(ContainSubstring(results[1].Path))
	})

	It("keeps completed progress on failure and never starts the remaining issues", func() {
		var requested []string
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			requested = append(requested, r.URL.Path)
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				fmt.Fprint(w, `{"key":"EXAMPLE-1","fields":{"summary":"Example","project":{"id":"10001"}}}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field", "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			default:
				http.Error(w, "permission denied", http.StatusForbidden)
			}
		})
		dir := GinkgoT().TempDir()
		results, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1", "EXAMPLE-2", "EXAMPLE-3"}, Dir: dir, Store: CatalogStore{Dir: dir}})
		Expect(err).To(MatchError(ContainSubstring("EXAMPLE-2 (2/3)")))
		Expect(results).To(HaveLen(1))
		Expect(results[0].Path).To(BeAnExistingFile())
		Expect(requested).NotTo(ContainElement("/rest/api/3/issue/EXAMPLE-3"))
		batch := jiraProgressSnapshot("Sync Jira issues")
		Expect(batch.Status).To(Equal(string(task.StatusFailed)))
		Expect([]int{batch.Progress, batch.MaxValue}).To(Equal([]int{1, 3}))
		Expect(batch.Description).To(ContainSubstring("1/3 issues saved"))
		Expect(jiraProgressSnapshot("Sync EXAMPLE-1").Status).To(Equal(string(task.StatusSuccess)))
		Expect(jiraProgressSnapshot("Sync EXAMPLE-2").Status).To(Equal(string(task.StatusFailed)))
		Expect(jiraProgressSnapshot("Sync EXAMPLE-3").ID).To(BeEmpty())
	})

	DescribeTable("reports downloaded bytes while the response is still streaming", func(knownLength bool) {
		const firstChunk = "first attachment chunk"
		const lastChunk = "last attachment chunk"
		total := len(firstChunk) + len(lastChunk)
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			if knownLength {
				w.Header().Set("Content-Length", strconv.Itoa(total))
			}
			fmt.Fprint(w, firstChunk)
			w.(http.Flusher).Flush()
			Eventually(func() task.TaskSnapshot { return jiraProgressSnapshot("Download Jira attachment 20001") }).Should(HaveField("Description", ContainSubstring(fmt.Sprintf("%d bytes", len(firstChunk)))))
			if knownLength {
				snapshot := jiraProgressSnapshot("Download Jira attachment 20001")
				Expect([]int{snapshot.Progress, snapshot.MaxValue}).To(Equal([]int{len(firstChunk), total}))
			}
			fmt.Fprint(w, lastChunk)
		})
		path := filepath.Join(GinkgoT().TempDir(), "notes.txt")
		_, err := client.DownloadAttachment(context.Background(), DownloadOptions{ID: "20001", Output: path})
		Expect(err).NotTo(HaveOccurred())
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal(firstChunk + lastChunk))
		Expect(jiraProgressSnapshot("Download Jira attachment 20001").Status).To(Equal(string(task.StatusSuccess)))
	}, Entry("with a content length", true), Entry("without a content length", false))

	It("propagates request cancellation and leaves no running steps", func() {
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			cancel()
			<-r.Context().Done()
		})
		dir := GinkgoT().TempDir()
		_, err := client.Sync(ctx, SyncOptions{Keys: []string{"EXAMPLE-1", "EXAMPLE-2"}, Dir: dir, Store: CatalogStore{Dir: dir}})
		Expect(err).To(MatchError(ContainSubstring(context.Canceled.Error())))
		Expect(jiraProgressSnapshot("Sync Jira issues").ID).NotTo(BeEmpty())
		for _, snapshot := range task.SnapshotAll() {
			Expect(snapshot.Status).To(Equal(string(task.StatusCancelled)), snapshot.Name)
		}
		Expect(jiraProgressSnapshot("Sync EXAMPLE-2").ID).To(BeEmpty())
	})
})
