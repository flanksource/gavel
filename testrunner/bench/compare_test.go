package bench

import (
	"math"
	"testing"
)

func TestCompare_Identical(t *testing.T) {
	samples := []float64{100, 102, 98, 101, 99}
	base := []BenchRun{{Name: "BenchmarkFoo", Samples: samples}}
	head := []BenchRun{{Name: "BenchmarkFoo", Samples: samples}}

	cmp := Compare(base, head, 5)
	if len(cmp.Deltas) != 1 {
		t.Fatalf("want 1 delta, got %d", len(cmp.Deltas))
	}
	d := cmp.Deltas[0]
	if d.DeltaPct != 0 {
		t.Errorf("DeltaPct = %v, want 0", d.DeltaPct)
	}
	if d.Significant {
		t.Error("identical samples should not be significant")
	}
	if cmp.HasRegression {
		t.Error("identical samples should not flag regression")
	}
	if math.Abs(cmp.GeomeanDelta) > 1e-9 {
		t.Errorf("geomean = %v, want 0", cmp.GeomeanDelta)
	}
}

func TestCompare_Regression(t *testing.T) {
	base := []BenchRun{{Name: "BenchmarkFoo", Samples: []float64{100, 101, 99, 100, 100}}}
	head := []BenchRun{{Name: "BenchmarkFoo", Samples: []float64{120, 121, 119, 120, 120}}}

	cmp := Compare(base, head, 5)
	d := cmp.Deltas[0]
	if d.DeltaPct < 15 || d.DeltaPct > 25 {
		t.Errorf("DeltaPct = %v, want ≈20", d.DeltaPct)
	}
	if !d.Significant {
		t.Errorf("expected significant, p=%v", d.PValue)
	}
	if !cmp.HasRegression {
		t.Error("expected regression flag")
	}
	if cmp.GeomeanDelta < 15 || cmp.GeomeanDelta > 25 {
		t.Errorf("geomean = %v, want ≈20", cmp.GeomeanDelta)
	}
}

func TestCompare_Improvement(t *testing.T) {
	base := []BenchRun{{Name: "BenchmarkFoo", Samples: []float64{100, 101, 99, 100, 100}}}
	head := []BenchRun{{Name: "BenchmarkFoo", Samples: []float64{80, 81, 79, 80, 80}}}

	cmp := Compare(base, head, 5)
	d := cmp.Deltas[0]
	if d.DeltaPct > -15 || d.DeltaPct < -25 {
		t.Errorf("DeltaPct = %v, want ≈-20", d.DeltaPct)
	}
	if !d.Significant {
		t.Error("expected significant improvement")
	}
	if cmp.HasRegression {
		t.Error("improvement should not flag regression")
	}
	if !d.IsImprovement(5) {
		t.Error("expected IsImprovement true")
	}
}

func TestCompare_BelowThreshold(t *testing.T) {
	// 2% difference, significant statistically but below the 5% threshold.
	base := []BenchRun{{Name: "B", Samples: []float64{100, 100.1, 99.9, 100, 100}}}
	head := []BenchRun{{Name: "B", Samples: []float64{102, 102.1, 101.9, 102, 102}}}
	cmp := Compare(base, head, 5)
	if cmp.HasRegression {
		t.Error("2% delta below 5% threshold should not flag regression")
	}
	d := cmp.Deltas[0]
	if !d.Significant {
		t.Error("expected statistical significance")
	}
	if d.IsRegression(5) {
		t.Error("expected not a regression at 5% threshold")
	}
}

func TestCompare_OnlyInOneSide(t *testing.T) {
	base := []BenchRun{
		{Name: "Shared", Samples: []float64{100, 100, 100}},
		{Name: "RemovedInHead", Samples: []float64{50, 50, 50}},
	}
	head := []BenchRun{
		{Name: "Shared", Samples: []float64{100, 100, 100}},
		{Name: "NewInHead", Samples: []float64{75, 75, 75}},
	}
	cmp := Compare(base, head, 5)
	if len(cmp.Deltas) != 3 {
		t.Fatalf("want 3 deltas, got %d", len(cmp.Deltas))
	}
	byName := map[string]BenchDelta{}
	for _, d := range cmp.Deltas {
		byName[d.Name] = d
	}
	if byName["RemovedInHead"].OnlyIn != "base" {
		t.Errorf("RemovedInHead.OnlyIn = %q, want base", byName["RemovedInHead"].OnlyIn)
	}
	if byName["NewInHead"].OnlyIn != "head" {
		t.Errorf("NewInHead.OnlyIn = %q, want head", byName["NewInHead"].OnlyIn)
	}
	if byName["Shared"].OnlyIn != "" {
		t.Errorf("Shared should not have OnlyIn set")
	}
}

func TestCompare_SingleSample(t *testing.T) {
	// -count=1: no stddev, no p-value, but we should still get a delta.
	base := []BenchRun{{Name: "B", Samples: []float64{100}}}
	head := []BenchRun{{Name: "B", Samples: []float64{120}}}
	cmp := Compare(base, head, 5)
	d := cmp.Deltas[0]
	if d.DeltaPct < 15 || d.DeltaPct > 25 {
		t.Errorf("DeltaPct = %v", d.DeltaPct)
	}
	if d.Significant {
		t.Error("single sample should never be flagged significant")
	}
	if cmp.HasRegression {
		t.Error("single-sample 20% delta should not trigger HasRegression without p-value")
	}
}

func TestCompare_ThresholdDefault(t *testing.T) {
	base := []BenchRun{{Name: "B", Samples: []float64{100, 100, 100}}}
	head := []BenchRun{{Name: "B", Samples: []float64{100, 100, 100}}}
	cmp := Compare(base, head, 0)
	if cmp.Threshold != DefaultThreshold {
		t.Errorf("Threshold = %v, want %v", cmp.Threshold, DefaultThreshold)
	}
}
