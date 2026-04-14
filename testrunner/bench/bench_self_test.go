package bench

import "testing"

// BenchmarkGeomean is a trivial self-benchmark used by the gavel bench e2e test.
// It runs fast (sub-millisecond) and is stable enough for dogfooding the pipeline.
func BenchmarkGeomean(b *testing.B) {
	xs := []float64{1.0, 1.1, 0.9, 1.05, 0.95, 1.02, 0.98}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Geomean(xs)
	}
}

func BenchmarkMean(b *testing.B) {
	xs := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Mean(xs)
	}
}
