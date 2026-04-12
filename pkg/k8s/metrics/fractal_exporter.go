// Package metrics provides a custom Kubernetes metrics exporter for fractal
// complexity (Box-Counting Dimension, D₀) used by the HPA controller.
package metrics

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

const (
	// MetricName is the Prometheus / custom-metrics API metric name consumed by HPA.
	MetricName = "logostalisman_fractal_complexity_metric"

	// D0VarianceThreshold is the D₀ variance value above which the HPA triggers a scale-up.
	D0VarianceThreshold = 0.5

	// DefaultPort is the default HTTP port for the metrics server.
	DefaultPort = 9101

	// boxCountingLevels is the number of box-size levels used in the BCD algorithm.
	boxCountingLevels = 8
)

// LatentBatch represents a batch of normalised latent codes in [0,1]^d.
type LatentBatch struct {
	Points    [][]float64 // shape: [n, d]
	Timestamp time.Time
}

// D0Observation holds a single Box-Counting Dimension measurement.
type D0Observation struct {
	D0        float64
	Timestamp time.Time
}

// FractalExporter calculates D₀ for incoming latent batches and exposes the
// resulting variance as a Prometheus-compatible gauge.
type FractalExporter struct {
	mu           sync.RWMutex
	observations []D0Observation
	windowSize   int // number of observations kept for variance calculation
	currentD0    float64
	d0Variance   float64
	server       *http.Server
}

// NewFractalExporter creates a FractalExporter with the supplied observation window.
func NewFractalExporter(windowSize int) *FractalExporter {
	if windowSize <= 0 {
		windowSize = 30
	}
	return &FractalExporter{
		windowSize: windowSize,
	}
}

// RecordBatch calculates D₀ for the batch and appends it to the observation
// window, then recomputes the rolling variance.
func (fe *FractalExporter) RecordBatch(batch LatentBatch) {
	d0 := ComputeBoxCountingDimension(batch.Points)
	fe.recordD0Locked(d0, batch.Timestamp)
}

// RecordD0 directly records a pre-computed D₀ value into the observation window.
// This is useful when D₀ is calculated externally or in tests.
func (fe *FractalExporter) RecordD0(d0 float64, ts time.Time) {
	fe.recordD0Locked(d0, ts)
}

// recordD0Locked appends a D₀ observation and recomputes variance.
func (fe *FractalExporter) recordD0Locked(d0 float64, ts time.Time) {
	obs := D0Observation{D0: d0, Timestamp: ts}

	fe.mu.Lock()
	defer fe.mu.Unlock()

	fe.observations = append(fe.observations, obs)
	if len(fe.observations) > fe.windowSize {
		fe.observations = fe.observations[len(fe.observations)-fe.windowSize:]
	}
	fe.currentD0 = d0
	fe.d0Variance = computeVariance(fe.observations)
}

// D0Variance returns the current rolling variance of D₀ observations.
func (fe *FractalExporter) D0Variance() float64 {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	return fe.d0Variance
}

// CurrentD0 returns the most recently computed Box-Counting Dimension.
func (fe *FractalExporter) CurrentD0() float64 {
	fe.mu.RLock()
	defer fe.mu.RUnlock()
	return fe.currentD0
}

// ScaleUpTriggered returns true when D₀ variance exceeds the threshold.
func (fe *FractalExporter) ScaleUpTriggered() bool {
	return fe.D0Variance() > D0VarianceThreshold
}

// ServeMetrics starts an HTTP server that exposes Prometheus-format metrics.
// It blocks until ctx is cancelled.
func (fe *FractalExporter) ServeMetrics(ctx context.Context, port int) error {
	if port <= 0 {
		port = DefaultPort
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", fe.handleMetrics)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	fe.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("fractal_exporter: listening on :%d", port)
		if err := fe.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return fe.server.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// handleMetrics writes Prometheus-compatible text output.
func (fe *FractalExporter) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	fe.mu.RLock()
	d0 := fe.currentD0
	variance := fe.d0Variance
	observations := len(fe.observations)
	fe.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP %s Box-Counting Dimension (D0) variance used for fractal-aware HPA\n", MetricName)
	fmt.Fprintf(w, "# TYPE %s gauge\n", MetricName)
	fmt.Fprintf(w, "%s{dimension=\"d0_variance\"} %g\n", MetricName, variance)
	fmt.Fprintf(w, "%s{dimension=\"d0_current\"} %g\n", MetricName, d0)
	fmt.Fprintf(w, "%s{dimension=\"observation_count\"} %d\n", MetricName, observations)
}

// ComputeBoxCountingDimension estimates the Box-Counting Dimension of a set of
// normalised points in [0,1]^d using a log-log linear regression over
// boxCountingLevels grid resolutions.
//
// Algorithm:
//  1. For each level k (1 … boxCountingLevels), set box size r_k = 2^(-k).
//  2. Count N(r_k) = number of distinct boxes containing ≥ 1 point.
//  3. Estimate D₀ = -d(log N(r)) / d(log r) via linear regression of
//     log(N(r_k)) against log(1/r_k).
func ComputeBoxCountingDimension(points [][]float64) float64 {
	if len(points) == 0 {
		return 0
	}

	logSizes := make([]float64, boxCountingLevels)
	logCounts := make([]float64, boxCountingLevels)

	for k := 1; k <= boxCountingLevels; k++ {
		r := math.Pow(2, -float64(k))
		n := countBoxes(points, r)
		if n == 0 {
			n = 1 // avoid log(0)
		}
		logSizes[k-1] = math.Log(1.0 / r)   // log(1/r) on x-axis
		logCounts[k-1] = math.Log(float64(n)) // log(N) on y-axis
	}

	return linearRegressionSlope(logSizes, logCounts)
}

// countBoxes counts the number of d-dimensional boxes of side length r that
// contain at least one point from the supplied slice.
//
// All points in a batch are expected to have the same dimensionality. Only the
// first boxCountingLevels coordinate dimensions are considered; coordinates
// beyond that index are ignored. Points with fewer dimensions have their
// remaining coordinates implicitly treated as zero, which is consistent within
// a homogeneous batch.
func countBoxes(points [][]float64, r float64) int {
	type key struct{ coords [boxCountingLevels]int64 }
	occupied := make(map[key]struct{}, len(points))

	for _, p := range points {
		var k key
		d := len(p)
		if d > boxCountingLevels {
			d = boxCountingLevels
		}
		for i := 0; i < d; i++ {
			k.coords[i] = int64(math.Floor(p[i] / r))
		}
		occupied[k] = struct{}{}
	}
	return len(occupied)
}

// linearRegressionSlope returns the slope of the OLS regression line for
// y = slope*x + intercept.
func linearRegressionSlope(x, y []float64) float64 {
	n := float64(len(x))
	if n == 0 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

// computeVariance calculates the sample variance of D₀ observations.
func computeVariance(obs []D0Observation) float64 {
	n := len(obs)
	if n < 2 {
		return 0
	}

	var sum float64
	for _, o := range obs {
		sum += o.D0
	}
	mean := sum / float64(n)

	var sq float64
	for _, o := range obs {
		diff := o.D0 - mean
		sq += diff * diff
	}
	return sq / float64(n-1) // sample variance
}
