// effect_size.go
// Statistical functions for computing effect sizes, t-tests, and p-values
// used for Phase II validation analysis (MNIST, CIFAR-10 benchmarks).

package engine

import "math"

// StatResult holds results from a Welch two-sample t-test.
type StatResult struct {
	TStatistic float64
	DegreesOfFreedom float64
	PValueTwoSided   float64
	PValueOneSided   float64
	CohensD          float64
	MeanDiff         float64
	PooledStdDev     float64
}

// SliceMean returns the arithmetic mean of a slice.
// Returns 0 for an empty slice.
func SliceMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// SliceVariance returns the sample variance (Bessel's correction, ddof=1).
// Returns 0 for slices with fewer than two elements.
func SliceVariance(data []float64) float64 {
	n := len(data)
	if n < 2 {
		return 0
	}
	m := SliceMean(data)
	sumSq := 0.0
	for _, v := range data {
		d := v - m
		sumSq += d * d
	}
	return sumSq / float64(n-1)
}

// SliceStdDev returns the sample standard deviation (ddof=1).
func SliceStdDev(data []float64) float64 {
	return math.Sqrt(SliceVariance(data))
}

// SlicePooledStdDev returns the pooled standard deviation of two groups:
//
//	σ_pooled = sqrt((σ₁² + σ₂²) / 2)
func SlicePooledStdDev(group1, group2 []float64) float64 {
	v1 := SliceVariance(group1)
	v2 := SliceVariance(group2)
	return math.Sqrt((v1 + v2) / 2)
}

// CohensD computes Cohen's d effect size between two independent groups:
//
//	d = (mean₁ − mean₂) / σ_pooled
//
// Returns 0 if the pooled standard deviation is zero.
func CohensD(group1, group2 []float64) float64 {
	pooled := SlicePooledStdDev(group1, group2)
	if pooled == 0 {
		return 0
	}
	return (SliceMean(group1) - SliceMean(group2)) / pooled
}

// WelchTTest performs Welch's two-sample t-test and returns a StatResult
// containing the t-statistic, Welch-Satterthwaite degrees of freedom,
// two-sided p-value, one-sided p-value (group1 > group2), and Cohen's d.
//
// Returns zero values if either group has fewer than 2 observations.
func WelchTTest(group1, group2 []float64) StatResult {
	n1, n2 := float64(len(group1)), float64(len(group2))
	if n1 < 2 || n2 < 2 {
		return StatResult{}
	}

	m1, m2 := SliceMean(group1), SliceMean(group2)
	v1, v2 := SliceVariance(group1), SliceVariance(group2)

	se1 := v1 / n1 // variance of the mean for group 1
	se2 := v2 / n2 // variance of the mean for group 2
	seSq := se1 + se2

	if seSq == 0 {
		return StatResult{}
	}

	tStat := (m1 - m2) / math.Sqrt(seSq)

	// Welch-Satterthwaite degrees of freedom
	df := (seSq * seSq) / (se1*se1/(n1-1) + se2*se2/(n2-1))

	pTwo := tDistTwoSidedPValue(tStat, df)
	pOne := tDistOneSidedPValue(tStat, df)

	return StatResult{
		TStatistic:       tStat,
		DegreesOfFreedom: df,
		PValueTwoSided:   pTwo,
		PValueOneSided:   pOne,
		CohensD:          CohensD(group1, group2),
		MeanDiff:         m1 - m2,
		PooledStdDev:     SlicePooledStdDev(group1, group2),
	}
}

// tDistTwoSidedPValue computes P(|T_df| ≥ |t|) using the regularized
// incomplete beta function: I_{df/(df+t²)}(df/2, 1/2).
func tDistTwoSidedPValue(t, df float64) float64 {
	x := df / (df + t*t)
	return regularizedIncompleteBeta(x, df/2, 0.5)
}

// tDistOneSidedPValue computes P(T_df ≥ t), i.e. the right-tail probability.
func tDistOneSidedPValue(t, df float64) float64 {
	return tDistTwoSidedPValue(t, df) / 2
}

// regularizedIncompleteBeta computes I_x(a, b) = B(x; a, b) / B(a, b)
// using the continued-fraction method from Numerical Recipes (Press et al.).
func regularizedIncompleteBeta(x, a, b float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}

	lgA, _ := math.Lgamma(a)
	lgB, _ := math.Lgamma(b)
	lgAB, _ := math.Lgamma(a + b)
	lbeta := lgA + lgB - lgAB
	bt := math.Exp(math.Log(x)*a + math.Log(1-x)*b - lbeta)

	// Use the symmetry relation I_x(a,b) = 1 - I_{1-x}(b,a) for stability
	// when x is large relative to (a+1)/(a+b+2).
	if x < (a+1)/(a+b+2) {
		return bt * betacf(a, b, x) / a
	}
	return 1 - bt*betacf(b, a, 1-x)/b
}

// betacf evaluates the continued fraction for the incomplete beta function
// using Lentz's modified algorithm (Numerical Recipes §6.4).
func betacf(a, b, x float64) float64 {
	const maxIter = 200
	const eps = 3.0e-7
	const fpmin = 1.0e-30

	qab := a + b
	qap := a + 1.0
	qam := a - 1.0

	c := 1.0
	d := 1.0 - qab*x/qap
	if math.Abs(d) < fpmin {
		d = fpmin
	}
	d = 1.0 / d
	h := d

	for mi := 1; mi <= maxIter; mi++ {
		mf := float64(mi)
		m2 := 2 * mf

		// Even step
		aa := mf * (b - mf) * x / ((qam + m2) * (a + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpmin {
			d = fpmin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpmin {
			c = fpmin
		}
		d = 1.0 / d
		h *= d * c

		// Odd step
		aa = -(a + mf) * (qab + mf) * x / ((a + m2) * (qap + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpmin {
			d = fpmin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpmin {
			c = fpmin
		}
		d = 1.0 / d
		del := d * c
		h *= del
		if math.Abs(del-1.0) <= eps {
			break
		}
	}

	return h
}
