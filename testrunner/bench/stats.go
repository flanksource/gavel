package bench

import "math"

// Mean returns the arithmetic mean. Panics if xs is empty — callers must check len.
func Mean(xs []float64) float64 {
	var sum float64
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

// Stddev returns the sample standard deviation (Bessel-corrected, N-1). Returns 0 if len(xs) < 2.
func Stddev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := Mean(xs)
	var ss float64
	for _, x := range xs {
		d := x - m
		ss += d * d
	}
	return math.Sqrt(ss / float64(len(xs)-1))
}

// WelchT returns (t, df) for Welch's unequal-variance two-sample t-test.
// Returns (0, 0) if either sample has fewer than 2 elements or both variances are zero.
func WelchT(a, b []float64) (t, df float64) {
	if len(a) < 2 || len(b) < 2 {
		return 0, 0
	}
	ma, mb := Mean(a), Mean(b)
	va, vb := variance(a, ma), variance(b, mb)
	na, nb := float64(len(a)), float64(len(b))
	if va == 0 && vb == 0 {
		return 0, 0
	}
	se := math.Sqrt(va/na + vb/nb)
	t = (ma - mb) / se
	num := va/na + vb/nb
	den := (va*va)/(na*na*(na-1)) + (vb*vb)/(nb*nb*(nb-1))
	df = (num * num) / den
	return t, df
}

func variance(xs []float64, mean float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return ss / float64(len(xs)-1)
}

// StudentTwoTailP returns the two-tailed p-value for a t statistic with df degrees of freedom.
// Uses the regularised incomplete beta function: p = I_{df/(df+t²)}(df/2, 1/2).
// Returns 1.0 if df <= 0 (insufficient data).
func StudentTwoTailP(t, df float64) float64 {
	if df <= 0 {
		return 1
	}
	x := df / (df + t*t)
	return betaI(x, df/2, 0.5)
}

// betaI is the regularised incomplete beta function I_x(a, b). Adapted from the
// continued-fraction formulation in "Numerical Recipes in C", §6.4. Accurate to
// ~1e-10 for the domains we care about (df in [2, 200], |t| in [0, 20]).
func betaI(x, a, b float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	bt := math.Exp(
		lgamma(a+b) - lgamma(a) - lgamma(b) +
			a*math.Log(x) + b*math.Log(1-x),
	)
	if x < (a+1)/(a+b+2) {
		return bt * betaCF(x, a, b) / a
	}
	return 1 - bt*betaCF(1-x, b, a)/b
}

func lgamma(x float64) float64 {
	v, _ := math.Lgamma(x)
	return v
}

func betaCF(x, a, b float64) float64 {
	const maxIter = 200
	const eps = 3e-12
	qab := a + b
	qap := a + 1
	qam := a - 1
	c := 1.0
	d := 1 - qab*x/qap
	if math.Abs(d) < 1e-30 {
		d = 1e-30
	}
	d = 1 / d
	h := d
	for m := 1; m <= maxIter; m++ {
		mf := float64(m)
		m2 := 2 * mf
		aa := mf * (b - mf) * x / ((qam + m2) * (a + m2))
		d = 1 + aa*d
		if math.Abs(d) < 1e-30 {
			d = 1e-30
		}
		c = 1 + aa/c
		if math.Abs(c) < 1e-30 {
			c = 1e-30
		}
		d = 1 / d
		h *= d * c
		aa = -(a + mf) * (qab + mf) * x / ((a + m2) * (qap + m2))
		d = 1 + aa*d
		if math.Abs(d) < 1e-30 {
			d = 1e-30
		}
		c = 1 + aa/c
		if math.Abs(c) < 1e-30 {
			c = 1e-30
		}
		d = 1 / d
		del := d * c
		h *= del
		if math.Abs(del-1) < eps {
			break
		}
	}
	return h
}

// Geomean returns the geometric mean of xs. Returns 0 if any element is <= 0 or xs is empty.
func Geomean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var logSum float64
	for _, x := range xs {
		if x <= 0 {
			return 0
		}
		logSum += math.Log(x)
	}
	return math.Exp(logSum / float64(len(xs)))
}
