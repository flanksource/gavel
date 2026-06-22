package cmux

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

type runnerCall struct {
	cwd    string
	binary string
	args   []string
}

type recordingRunner struct {
	calls []runnerCall
	out   map[string]string
}

func (r *recordingRunner) run(_ context.Context, cwd, binary string, _ time.Duration, args ...string) (string, error) {
	call := runnerCall{cwd: cwd, binary: binary, args: append([]string(nil), args...)}
	r.calls = append(r.calls, call)
	if r.out != nil {
		if stdout, ok := r.out[joinArgs(args)]; ok {
			return stdout, nil
		}
	}
	return "ok", nil
}

func TestParseWorkspaceRef(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want WorkspaceRef
	}{
		{
			name: "json",
			out:  `{"workspaceId":"ws-json","surfaceId":"sf-json"}`,
			want: WorkspaceRef{WorkspaceID: "ws-json", SurfaceID: "sf-json", Raw: `{"workspaceId":"ws-json","surfaceId":"sf-json"}`},
		},
		{
			name: "labeled text",
			out:  "created workspace=ws-text surface=sf-text",
			want: WorkspaceRef{WorkspaceID: "ws-text", SurfaceID: "sf-text", Raw: "created workspace=ws-text surface=sf-text"},
		},
		{
			name: "plain ref",
			out:  "ws-plain\n",
			want: WorkspaceRef{Raw: "ws-plain"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseWorkspaceRef(tc.out); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ParseWorkspaceRef() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestClientNewWorkspaceAndSendUseCmuxArgs(t *testing.T) {
	runner := &recordingRunner{out: map[string]string{
		joinArgs([]string{"new-workspace", "--cwd", "/repo", "--name", "repo", "--focus", "true", "--command", "cmux claude-teams --model opus", "--id-format", "both"}): "workspace=ws1 surface=sf1",
	}}
	client := &Client{Runner: runner.run}

	ref, err := client.NewWorkspace(context.Background(), NewWorkspaceOpts{
		Cwd:     "/repo",
		Name:    "repo",
		Command: "cmux claude-teams --model opus",
		Focus:   true,
	})
	if err != nil {
		t.Fatalf("NewWorkspace() error = %v", err)
	}
	if ref.String() != "ws1" {
		t.Fatalf("workspace ref = %q, want ws1", ref.String())
	}

	if err := client.Send(context.Background(), ref.String(), "hello\n"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if len(runner.calls) != 2 {
		t.Fatalf("calls = %d, want 2", len(runner.calls))
	}
	wantSend := []string{"send", "--workspace", "ws1", "--", "hello\n"}
	if !reflect.DeepEqual(runner.calls[1].args, wantSend) {
		t.Fatalf("send args = %#v, want %#v", runner.calls[1].args, wantSend)
	}
}

func joinArgs(args []string) string {
	return strings.Join(args, "\x00")
}
