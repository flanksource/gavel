package bench

import (
	"math"
	"sort"
)

const (
	DefaultThreshold = 5.0  // percent
	DefaultPValue    = 0.05 // alpha for significance
)

// Compare computes per-benchmark deltas between base and head runs.
// Benchmarks are matched by Name. Benchmarks present on only one side are still
// reported with OnlyIn set but do not contribute to the geomean.
//
// threshold is the regression threshold in percent; pass 0 for the default (5%).
func Compare(base, head []BenchRun, threshold float64) BenchComparison {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	baseIdx := indexByName(base)
	headIdx := indexByName(head)

	var deltas []BenchDelta
	var ratios []float64
	var hasRegression bool

	seen := make(map[string]bool, len(baseIdx)+len(headIdx))
	for name, b := range baseIdx {
		seen[name] = true
		h, ok := headIdx[name]
		if !ok {
			deltas = append(deltas, BenchDelta{Name: name, Package: b.Package, OnlyIn: "base"})
			continue
		}
		d := computeDelta(b, h)
		if d.IsRegression(threshold) {
			hasRegression = true
		}
		deltas = append(deltas, d)
		if d.BaseMean > 0 {
			ratios = append(ratios, d.HeadMean/d.BaseMean)
		}
	}
	for name, h := range headIdx {
		if seen[name] {
			continue
		}
		deltas = append(deltas, BenchDelta{Name: name, Package: h.Package, OnlyIn: "head"})
	}

	sort.SliceStable(deltas, func(i, j int) bool { return deltas[i].Name < deltas[j].Name })

	var geomeanPct float64
	if len(ratios) > 0 {
		geomeanPct = (Geomean(ratios) - 1) * 100
	}

	return BenchComparison{
		Threshold:     threshold,
		Deltas:        deltas,
		GeomeanDelta:  geomeanPct,
		HasRegression: hasRegression,
	}
}

func computeDelta(base, head BenchRun) BenchDelta {
	d := BenchDelta{
		Name:    base.Name,
		Package: base.Package,
		Samples: min(len(base.Samples), len(head.Samples)),
	}
	if len(base.Samples) == 0 || len(head.Samples) == 0 {
		return d
	}
	d.BaseMean = Mean(base.Samples)
	d.HeadMean = Mean(head.Samples)
	if d.BaseMean > 0 {
		d.BaseStddev = Stddev(base.Samples) / d.BaseMean * 100
	}
	if d.HeadMean > 0 {
		d.HeadStddev = Stddev(head.Samples) / d.HeadMean * 100
	}
	if d.BaseMean > 0 {
		d.DeltaPct = (d.HeadMean - d.BaseMean) / d.BaseMean * 100
	}
	if len(base.Samples) >= 2 && len(head.Samples) >= 2 {
		t, df := WelchT(base.Samples, head.Samples)
		if df > 0 && !math.IsNaN(t) && !math.IsInf(t, 0) {
			d.PValue = StudentTwoTailP(t, df)
			d.Significant = d.PValue < DefaultPValue
		}
	}
	return d
}

func indexByName(runs []BenchRun) map[string]BenchRun {
	m := make(map[string]BenchRun, len(runs))
	for _, r := range runs {
		m[r.Name] = r
	}
	return m
}
