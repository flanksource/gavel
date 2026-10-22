package overlaylabels

import "testing"

func BenchmarkLabels(b *testing.B) {
	for b.Loop() {
		_ = Labels{"app": "example"}
	}
}
