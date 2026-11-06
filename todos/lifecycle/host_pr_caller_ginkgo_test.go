// Tests a local step on a linked PR whose run executes whole on substrate: the run's program has no
// credentials, so this host answers its PR calls with the step's own PR client.
package lifecycle

import (
	"context"
	"encoding/json"
	"io"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

// statusPR is a step's PR client that answers the PR's status.
type statusPR struct {
	provider.Client
	asked []int
}

func (p *statusPR) Status(_ context.Context, number int, _ model.StatusOptions) (*model.StatusSnapshot, error) {
	p.asked = append(p.asked, number)
	return &model.StatusSnapshot{PR: &model.PRInfo{Number: number, State: "OPEN", HeadRefOID: "abc"}}, nil
}

var _ = g.Describe("a local step on a linked PR, run whole elsewhere", func() {
	g.It("answers its run's PR calls with the step's own PR client", func() {
		client := &statusPR{}
		prepared := &preparedStep{
			class: types.ModeRun, agent: "claude",
			request: api.Spec{
				Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}, Prompt: api.Prompt{User: "Address the review"},
				Setup: &shell.Setup{Cwd: g.GinkgoT().TempDir()},
			},
			pr: &prExecution{client: client, number: 7, target: &provider.RunTarget{RepositoryURL: "https://github.com/acme/app.git", Branch: "topic", HeadSHA: "abc"}, workDir: g.GinkgoT().TempDir()},
		}
		exec := todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		input := mustRunInput((&Host{}).runInput(exec, &types.TODO{ID: "example"}, prepared, RunOptions{}))

		answer, ok := input.input.HostCalls[runclient.HostCallName]
		o.Expect(ok).To(o.BeTrue(), "the run can call this host for its PR")
		resp, err := runclient.OverHostCall(func(_ string, data json.RawMessage) (json.RawMessage, error) {
			return answer(context.Background(), data)
		}).Do(context.Background(), "GET", "pr/status?number=7&threads=true", nil)
		o.Expect(err).NotTo(o.HaveOccurred())
		body, err := io.ReadAll(resp.Body)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(string(body)).To(o.ContainSubstring(`"headRefOid":"abc"`))
		o.Expect(client.asked).To(o.Equal([]int{7}))
	})
})
