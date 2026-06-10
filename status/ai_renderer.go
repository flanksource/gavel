package status

import (
	"sync"

	"github.com/flanksource/clicky/api"
	clickytask "github.com/flanksource/clicky/task"
)

// statusRenderer renders the status table through clicky's task manager. It
// implements clickytask.LiveRenderer so clicky owns the terminal, the
// ClearLines accounting, and the logger serializer — which is what keeps AI
// agent log lines (emitted to the logger during ExecutePrompt) from corrupting
// the in-place redraw.
//
// State is driven by AISummaryUpdate values applied via Apply; the clicky
// render loop calls RenderLive concurrently on its tick, so both paths lock mu.
// The task snapshot clicky passes is ignored — file state lives in result.
type statusRenderer struct {
	mu     sync.Mutex
	result *Result
}

// NewStatusRenderer builds a LiveRenderer that paints the status table from
// result and folds AI summary updates into it via Apply.
func NewStatusRenderer(result *Result) *statusRenderer { //nolint:revive // returns unexported type by design; callers use it only as clickytask.LiveRenderer
	return &statusRenderer{result: result}
}

// Apply folds one AI summary update into the result. Safe to call concurrently
// with RenderLive/RenderFinal.
func (s *statusRenderer) Apply(update AISummaryUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result.ApplyAISummaryUpdate(update)
}

func (s *statusRenderer) RenderLive(_ []*clickytask.Task) api.Text {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result.Pretty()
}

func (s *statusRenderer) RenderFinal(_ []*clickytask.Task) api.Text {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result.Pretty()
}
