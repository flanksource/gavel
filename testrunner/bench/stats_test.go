package bench

import (
	"math"
	"testing"
)

// Reference values computed with scipy 1.13:
//
//	from scipy import stats
//	a = [100, 102, 98, 101, 99]
//	b = [110, 112, 108, 111, 109]
//	stats.ttest_ind(a, b, equal_var=False)
//	# TtestResult(statistic=-7.745..., pvalue=5.26e-5, df=8.0)
//	stats.gmean([0.9, 1.1, 1.0])  # 0.99666...
func approx(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.10g, want %.10g (±%.g)", name, got, want, tol)
	}
}

func TestMeanStddev(t *testing.T) {
	xs := []float64{100, 102, 98, 101, 99}
	approx(t, "mean", Mean(xs), 100, 1e-12)
	approx(t, "stddev", Stddev(xs), math.Sqrt(2.5), 1e-12)

	if Stddev([]float64{5}) != 0 {
		t.Error("single-element stddev should be 0")
	}
}

func TestWelchT(t *testing.T) {
	a := []float64{100, 102, 98, 101, 99}
	b := []float64{110, 112, 108, 111, 109}
	stat, df := WelchT(a, b)
	approx(t, "t-statistic", stat, -10, 1e-9)
	approx(t, "df", df, 8, 1e-9)

	p := StudentTwoTailP(stat, df)
	// scipy: stats.t.sf(10, 8)*2 ≈ 8.4882e-6
	approx(t, "pvalue", p, 8.488181528e-6, 1e-9)
}

func TestStudentTwoTailP_Identical(t *testing.T) {
	// When both samples are identical, t=0, p=1.
	a := []float64{100, 100, 100, 100}
	b := []float64{100, 100, 100, 100}
	stat, df := WelchT(a, b)
	if stat != 0 || df != 0 {
		t.Errorf("identical zero-variance samples should return (0,0), got (%v,%v)", stat, df)
	}
	// StudentTwoTailP(0, positive df) should be 1 (both tails sum).
	approx(t, "p at t=0", StudentTwoTailP(0, 10), 1, 1e-12)
}

func TestStudentTwoTailP_KnownValues(t *testing.T) {
	// scipy: stats.t.sf(2.0, 10)*2 ≈ 0.07338803
	approx(t, "t=2,df=10", StudentTwoTailP(2.0, 10), 0.07338803, 1e-7)
	// stats.t.sf(1.96, 1000)*2 ≈ 0.05029
	approx(t, "t=1.96,df=1000", StudentTwoTailP(1.96, 1000), 0.05029, 1e-4)
}

func TestGeomean(t *testing.T) {
	// (0.9*1.1*1.0)^(1/3) ≈ 0.9966554934
	approx(t, "geomean", Geomean([]float64{0.9, 1.1, 1.0}), 0.9966554934, 1e-9)
	// scipy: stats.gmean([2, 8]) = 4
	approx(t, "geomean 2,8", Geomean([]float64{2, 8}), 4, 1e-12)

	if Geomean([]float64{}) != 0 {
		t.Error("empty geomean should be 0")
	}
	if Geomean([]float64{1, -1}) != 0 {
		t.Error("negative input geomean should be 0")
	}
}
