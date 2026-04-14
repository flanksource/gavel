package bench

import (
	"fmt"
	"sort"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

// BenchRun is the result of running one benchmark one or more times.
// One BenchRun per benchmark name, with per-iteration samples from go test -bench=. -count=N.
type BenchRun struct {
	Name        string    `json:"name"`                   // e.g. "BenchmarkRLS/10k_rows"
	Package     string    `json:"package"`                // e.g. "github.com/flanksource/duty/bench"
	Samples     []float64 `json:"samples"`                // ns/op per -count iteration
	Iterations  int       `json:"iterations,omitempty"`   // b.N from the final run
	BytesPerOp  int64     `json:"bytes_per_op,omitempty"`
	AllocsPerOp int64     `json:"allocs_per_op,omitempty"`
	MBPerSec    float64   `json:"mb_per_sec,omitempty"`
}

// BenchDelta is the comparison between one benchmark's base and head runs.
type BenchDelta struct {
	Name        string  `json:"name" pretty:"label=Name"`
	Package     string  `json:"package,omitempty" pretty:"label=Package"`
	BaseMean    float64 `json:"base_mean" pretty:"label=Base,format=ns/op"`
	BaseStddev  float64 `json:"base_stddev,omitempty" pretty:"label=±base%"`
	HeadMean    float64 `json:"head_mean" pretty:"label=Head,format=ns/op"`
	HeadStddev  float64 `json:"head_stddev,omitempty" pretty:"label=±head%"`
	DeltaPct    float64 `json:"delta_pct" pretty:"label=Δ%"`
	PValue      float64 `json:"p_value,omitempty" pretty:"label=p"`
	Samples     int     `json:"samples,omitempty" pretty:"label=n"`      // min(len(base.Samples), len(head.Samples))
	Significant bool    `json:"significant,omitempty" pretty:"label=sig"`
	OnlyIn      string  `json:"only_in,omitempty" pretty:"label=OnlyIn"` // "base" or "head" if missing on one side
}

// BenchComparison is the full result: per-benchmark deltas + geomean summary.
type BenchComparison struct {
	BaseLabel     string       `json:"base_label,omitempty"`
	HeadLabel     string       `json:"head_label,omitempty"`
	Threshold     float64      `json:"threshold"`
	Deltas        []BenchDelta `json:"deltas"`
	GeomeanDelta  float64      `json:"geomean_delta"`
	HasRegression bool         `json:"has_regression"`
}

func (d BenchDelta) IsRegression(threshold float64) bool {
	return d.Significant && d.DeltaPct > threshold
}

func (d BenchDelta) IsImprovement(threshold float64) bool {
	return d.Significant && d.DeltaPct < -threshold
}

// Pretty renders a single delta row.
func (d BenchDelta) Pretty() api.Text {
	s := clicky.Text("")
	if d.OnlyIn != "" {
		s = s.Append(icons.Skip, "text-orange-500").Space().
			Append(d.Name, "bold").Space().
			Append(fmt.Sprintf("(only in %s)", d.OnlyIn), "text-muted")
		return s
	}
	switch {
	case d.IsRegression(0):
		s = s.Append(icons.Fail, "text-red-500")
	case d.IsImprovement(0):
		s = s.Append(icons.Pass, "text-green-600")
	default:
		s = s.Append("·", "text-muted")
	}
	s = s.Space().Append(d.Name, "bold")
	s = s.Space().Append(fmt.Sprintf("%s → %s", formatNs(d.BaseMean), formatNs(d.HeadMean)), "text-muted")

	deltaStyle := "text-muted"
	if d.Significant {
		if d.DeltaPct > 0 {
			deltaStyle = "text-red-500 bold"
		} else if d.DeltaPct < 0 {
			deltaStyle = "text-green-600 bold"
		}
	}
	s = s.Space().Append(fmt.Sprintf("%+.2f%%", d.DeltaPct), deltaStyle)

	if d.PValue > 0 {
		s = s.Space().Append(fmt.Sprintf("(p=%.3g, n=%d)", d.PValue, d.Samples), "text-muted")
	} else if d.Samples < 2 {
		s = s.Space().Append("(n=1, no p-value)", "text-muted")
	}
	return s
}

// Pretty renders the full comparison: header, regressions, improvements, geomean footer.
func (c BenchComparison) Pretty() api.Text {
	text := clicky.Text("Benchmark comparison", "bold")
	if c.BaseLabel != "" || c.HeadLabel != "" {
		text = text.Space().Append(fmt.Sprintf("%s → %s", c.BaseLabel, c.HeadLabel), "text-muted")
	}
	text = text.Space().Append(fmt.Sprintf("(threshold=±%.1f%%)", c.Threshold), "text-muted")
	text = text.NewLine()

	var regressions, improvements, neutral []BenchDelta
	for _, d := range c.Deltas {
		switch {
		case d.IsRegression(c.Threshold):
			regressions = append(regressions, d)
		case d.IsImprovement(c.Threshold):
			improvements = append(improvements, d)
		default:
			neutral = append(neutral, d)
		}
	}
	sort.SliceStable(regressions, func(i, j int) bool { return regressions[i].DeltaPct > regressions[j].DeltaPct })
	sort.SliceStable(improvements, func(i, j int) bool { return improvements[i].DeltaPct < improvements[j].DeltaPct })

	if len(regressions) > 0 {
		text = text.NewLine().Append(fmt.Sprintf("Regressions (%d)", len(regressions)), "bold text-red-500").NewLine()
		for _, d := range regressions {
			text = text.Add(d.Pretty()).NewLine()
		}
	}
	if len(improvements) > 0 {
		inner := clicky.Text("")
		for _, d := range improvements {
			inner = inner.Add(d.Pretty()).NewLine()
		}
		text = text.NewLine().Add(api.Collapsed{
			Label:   fmt.Sprintf("Improvements (%d)", len(improvements)),
			Content: inner,
		}).NewLine()
	}
	if len(neutral) > 0 {
		inner := clicky.Text("")
		for _, d := range neutral {
			inner = inner.Add(d.Pretty()).NewLine()
		}
		text = text.NewLine().Add(api.Collapsed{
			Label:   fmt.Sprintf("Unchanged (%d)", len(neutral)),
			Content: inner,
		}).NewLine()
	}

	geomeanStyle := "text-muted"
	if c.GeomeanDelta > c.Threshold {
		geomeanStyle = "text-red-500 bold"
	} else if c.GeomeanDelta < -c.Threshold {
		geomeanStyle = "text-green-600 bold"
	}
	text = text.NewLine().Append(fmt.Sprintf("geomean: %+.2f%%", c.GeomeanDelta), geomeanStyle)
	if c.HasRegression {
		text = text.Space().Append("[REGRESSION]", "text-red-500 bold")
	}
	return text
}

func formatNs(ns float64) string {
	switch {
	case ns >= 1e9:
		return fmt.Sprintf("%.2fs", ns/1e9)
	case ns >= 1e6:
		return fmt.Sprintf("%.2fms", ns/1e6)
	case ns >= 1e3:
		return fmt.Sprintf("%.2fµs", ns/1e3)
	default:
		return fmt.Sprintf("%.2fns", ns)
	}
}
