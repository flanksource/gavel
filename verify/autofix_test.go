package verify

import (
	"strings"
	"testing"
)

func makeResult(checks map[string]bool, ratings map[string]int, completenessPass bool) *VerifyResult {
	r := &VerifyResult{
		Checks:  make(map[string]CheckResult),
		Ratings: make(map[string]RatingResult),
		Completeness: CompletenessResult{
			Pass:    completenessPass,
			Summary: "completeness summary",
		},
	}
	for id, pass := range checks {
		cr := CheckResult{Pass: pass}
		if !pass {
			cr.Evidence = []Evidence{{File: "main.go", Line: 10, Message: id + " failed"}}
		}
		r.Checks[id] = cr
	}
	for dim, score := range ratings {
		rr := RatingResult{Score: score}
		if score < 80 {
			rr.Findings = []Evidence{{File: "main.go", Message: dim + " low"}}
		}
		r.Ratings[dim] = rr
	}
	r.Score = ComputeOverallScore(*r)
	return r
}

func TestCheckConverged(t *testing.T) {
	opts := AutoFixOptions{ScoreThreshold: 80}

	tests := []struct {
		name   string
		result *VerifyResult
		want   bool
	}{
		{
			name:   "score above threshold",
			result: makeResult(map[string]bool{"a": true}, map[string]int{"sec": 90}, true),
			want:   true,
		},
		{
			name:   "all checks pass even if score below",
			result: makeResult(map[string]bool{"a": true}, map[string]int{"sec": 30}, true),
			want:   true,
		},
		{
			name:   "failing check and low score",
			result: makeResult(map[string]bool{"a": false}, map[string]int{"sec": 30}, false),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkConverged(tt.result, opts); got != tt.want {
				t.Errorf("checkConverged() = %v, want %v (score=%d)", got, tt.want, tt.result.Score)
			}
		})
	}
}

func TestExtractFindings(t *testing.T) {
	r := makeResult(
		map[string]bool{"a": false, "b": true},
		map[string]int{"sec": 50, "dup": 90},
		false,
	)

	findings := extractFindings(r, 1)

	ids := map[string]bool{}
	for _, f := range findings {
		ids[f.ID] = true
		if f.DiscoveredTurn != 1 {
			t.Errorf("finding %q: DiscoveredTurn = %d, want 1", f.ID, f.DiscoveredTurn)
		}
		if f.FixedTurn != 0 {
			t.Errorf("finding %q: FixedTurn = %d, want 0", f.ID, f.FixedTurn)
		}
	}

	for _, expected := range []string{"a", "rating:sec", "completeness"} {
		if !ids[expected] {
			t.Errorf("missing expected finding %q, got %v", expected, ids)
		}
	}
	if ids["b"] {
		t.Error("passing check 'b' should not appear in findings")
	}
	if ids["rating:dup"] {
		t.Error("rating:dup (score 90) should not appear in findings")
	}
}

func TestUpdateFindings(t *testing.T) {
	// Turn 1: checks a,b fail; rating:sec low; completeness fails
	r1 := makeResult(
		map[string]bool{"a": false, "b": false},
		map[string]int{"sec": 50},
		false,
	)
	loop := &FixLoopResult{
		Turns:    []TurnResult{{Turn: 1, Score: r1.Score, Result: r1}},
		Findings: extractFindings(r1, 1),
	}

	// Turn 2: a fixed, b still fails; sec improved; completeness fixed; new check c fails
	r2 := makeResult(
		map[string]bool{"a": true, "b": false, "c": false},
		map[string]int{"sec": 85},
		true,
	)
	updateFindings(loop, r2, 2)

	byID := map[string]TrackedFinding{}
	for _, f := range loop.Findings {
		byID[f.ID] = f
	}

	// a was fixed in turn 2
	if f := byID["a"]; f.FixedTurn != 2 {
		t.Errorf("finding 'a': FixedTurn = %d, want 2", f.FixedTurn)
	}

	// b still unfixed
	if f := byID["b"]; f.FixedTurn != 0 {
		t.Errorf("finding 'b': FixedTurn = %d, want 0", f.FixedTurn)
	}

	// rating:sec fixed (score went above 80)
	if f := byID["rating:sec"]; f.FixedTurn != 2 {
		t.Errorf("finding 'rating:sec': FixedTurn = %d, want 2", f.FixedTurn)
	}

	// completeness fixed
	if f := byID["completeness"]; f.FixedTurn != 2 {
		t.Errorf("finding 'completeness': FixedTurn = %d, want 2", f.FixedTurn)
	}

	// c is new in turn 2
	if f, ok := byID["c"]; !ok {
		t.Error("new finding 'c' should be tracked")
	} else if f.DiscoveredTurn != 2 || f.FixedTurn != 0 {
		t.Errorf("finding 'c': discovered=%d fixed=%d, want discovered=2 fixed=0", f.DiscoveredTurn, f.FixedTurn)
	}
}

func TestBuildFixPrompt(t *testing.T) {
	r := makeResult(
		map[string]bool{"tests-added": false, "null-safety": true},
		map[string]int{"security": 50},
		false,
	)

	t.Run("first turn includes findings and evidence", func(t *testing.T) {
		loop := &FixLoopResult{Turns: []TurnResult{{Turn: 1, Score: r.Score, Result: r}}}
		prompt := buildFixPrompt(r, RunOptions{Config: VerifyConfig{Prompt: "custom context"}}, loop, 1)

		if !strings.Contains(prompt, "tests-added") {
			t.Error("prompt should contain failed check ID 'tests-added'")
		}
		if !strings.Contains(prompt, "security") {
			t.Error("prompt should contain low-rated dimension 'security'")
		}
		if !strings.Contains(prompt, "Completeness") {
			t.Error("prompt should contain completeness section")
		}
		if !strings.Contains(prompt, "custom context") {
			t.Error("prompt should contain additional context from config")
		}
		if strings.Contains(prompt, "Previous attempts") {
			t.Error("first turn should not include previous attempts section")
		}
	})

	t.Run("later turns include previous attempts", func(t *testing.T) {
		loop := &FixLoopResult{
			Turns: []TurnResult{
				{Turn: 1, Score: 40, Result: r},
				{Turn: 2, Score: 55, Result: r},
			},
			Findings: []TrackedFinding{
				{ID: "x", FixedTurn: 2, DiscoveredTurn: 1},
			},
		}
		prompt := buildFixPrompt(r, RunOptions{}, loop, 3)

		if !strings.Contains(prompt, "Previous attempts") {
			t.Error("later turns should include previous attempts section")
		}
		if !strings.Contains(prompt, "Turn 1: score 40") {
			t.Error("prompt should show score history")
		}
		if !strings.Contains(prompt, "Already fixed") {
			t.Error("prompt should mention already-fixed findings")
		}
	})
}

func TestBuildFixArgs(t *testing.T) {
	tests := []struct {
		name      string
		adapter   Adapter
		model     string
		patchOnly bool
		wantParts []string
	}{
		{
			name:      "claude interactive",
			adapter:   Claude{},
			model:     "claude-sonnet-4",
			patchOnly: false,
			wantParts: []string{"-p", "--allowedTools", "--model", "claude-sonnet-4"},
		},
		{
			name:      "claude patch-only",
			adapter:   Claude{},
			model:     "",
			patchOnly: true,
			wantParts: []string{"-p", "--output-format", "json"},
		},
		{
			name:      "codex full-auto",
			adapter:   Codex{},
			model:     "codex-mini",
			patchOnly: false,
			wantParts: []string{"exec", "--full-auto", "-m", "codex-mini"},
		},
		{
			name:      "codex patch-only",
			adapter:   Codex{},
			model:     "",
			patchOnly: true,
			wantParts: []string{"exec", "--"},
		},
		{
			name:      "gemini",
			adapter:   Gemini{},
			model:     "gemini-2.5-flash",
			patchOnly: false,
			wantParts: []string{"-p", "-m", "gemini-2.5-flash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.adapter.BuildFixArgs(tt.model, "fix this", tt.patchOnly)
			joined := strings.Join(args, " ")
			for _, part := range tt.wantParts {
				if !strings.Contains(joined, part) {
					t.Errorf("args %q should contain %q", joined, part)
				}
			}
		})
	}
}

func TestFormatEvidence(t *testing.T) {
	tests := []struct {
		name     string
		evidence []Evidence
		want     string
	}{
		{
			name:     "with file and line",
			evidence: []Evidence{{File: "a.go", Line: 5, Message: "bad"}},
			want:     "a.go:5 — bad",
		},
		{
			name:     "with file only",
			evidence: []Evidence{{File: "b.go", Message: "issue"}},
			want:     "b.go — issue",
		},
		{
			name:     "message only",
			evidence: []Evidence{{Message: "general"}},
			want:     "general",
		},
		{
			name: "multiple joined",
			evidence: []Evidence{
				{File: "a.go", Line: 1, Message: "x"},
				{Message: "y"},
			},
			want: "a.go:1 — x; y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEvidence(tt.evidence)
			if got != tt.want {
				t.Errorf("formatEvidence() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFixLoopResultPretty(t *testing.T) {
	r1 := makeResult(map[string]bool{"a": false}, map[string]int{"sec": 50}, false)
	r2 := makeResult(map[string]bool{"a": true}, map[string]int{"sec": 90}, true)

	loop := FixLoopResult{
		Turns: []TurnResult{
			{Turn: 1, Score: 30, Result: r1},
			{Turn: 2, Score: 90, Result: r2},
		},
		Findings: []TrackedFinding{
			{ID: "a", DiscoveredTurn: 1, FixedTurn: 2, Evidence: "check failed"},
			{ID: "rating:sec", DiscoveredTurn: 1, FixedTurn: 2},
			{ID: "unfixed-thing", DiscoveredTurn: 1, Evidence: "still broken"},
		},
		Passed: true,
		Reason: "score 90 >= 80 after turn 1",
	}

	// Just verify Pretty() doesn't panic
	_ = loop.Pretty()
}
