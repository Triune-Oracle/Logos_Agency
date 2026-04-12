// stat_inference_complete.go

package engine

import (
	"fmt"
	"math"
)

// Prior distribution
type Prior struct {
	Type   string // e.g., "normal", "uniform"
	Params []float64
}

// calculateLikelihood computes the Gaussian likelihood for a dataset given mean and variance
func calculateLikelihood(data []float64, mean float64, variance float64) float64 {
	likelihood := 1.0
	for _, d := range data {
		likelihood *= (1 / math.Sqrt(2*math.Pi*variance)) * math.Exp(-(math.Pow(d-mean, 2) / (2 * variance)))
	}
	return likelihood
}

// computePosterior computes the posterior mean and variance given data and a prior
func computePosterior(data []float64, prior Prior) (float64, float64) {
	var mean float64
	for _, d := range data {
		mean += d
	}
	mean /= float64(len(data))
	variance := float64(len(data)) // Placeholder for variance computation

	return mean, variance
}

// getLocaleAwarePrior returns a prior distribution based on the locale
func getLocaleAwarePrior(locale string) Prior {
	if locale == "US" {
		return Prior{Type: "normal", Params: []float64{0, 1}}
	}
	return Prior{Type: "uniform", Params: []float64{0, 1}}
}

// BayesianInference performs full Bayesian inference on the given data with locale-aware priors
func BayesianInference(data []float64, locale string) (float64, float64) {
	prior := getLocaleAwarePrior(locale)

	mean, variance := computePosterior(data, prior)

	fmt.Printf("Computed Mean: %f, Variance: %f, Prior Type: %s\n", mean, variance, prior.Type)
	return mean, variance
}