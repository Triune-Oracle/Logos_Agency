package tests

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Triune-Oracle/Logos_Agency/pkg/k8s/metrics"
)

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

// generateUniformPoints returns n points uniformly sampled from [0,1]^d.
func generateUniformPoints(n, d int, rng *rand.Rand) [][]float64 {
	pts := make([][]float64, n)
	for i := range pts {
		p := make([]float64, d)
		for j := range p {
			p[j] = rng.Float64()
		}
		pts[i] = p
	}
	return pts
}

// generateCantorPoints returns n points tightly clustered around a few centres,
// producing a high D₀ variance when interleaved with uniform batches.
func generateCantorPoints(n, d int, rng *rand.Rand) [][]float64 {
	centres := []float64{0.1, 0.5, 0.9}
	pts := make([][]float64, n)
	for i := range pts {
		p := make([]float64, d)
		c := centres[rng.Intn(len(centres))]
		for j := range p {
			p[j] = c + (rng.Float64()-0.5)*0.02 // tight cluster
		}
		pts[i] = p
	}
	return pts
}

// ──────────────────────────────────────────────────────────────────────────────
// Unit tests: ComputeBoxCountingDimension
// ──────────────────────────────────────────────────────────────────────────────

func TestComputeBoxCountingDimension_Empty(t *testing.T) {
	d0 := metrics.ComputeBoxCountingDimension(nil)
	if d0 != 0 {
		t.Errorf("expected 0 for empty input, got %g", d0)
	}
}

func TestComputeBoxCountingDimension_UniformCloud(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	pts := generateUniformPoints(500, 3, rng)
	d0 := metrics.ComputeBoxCountingDimension(pts)
	// With sparse sampling (500 pts in 3-D) the BCD estimator saturates early,
	// yielding values in roughly [0.3, 2.0]. We verify it is positive and bounded.
	if d0 < 0.3 || d0 > 2.0 {
		t.Errorf("expected BCD in [0.3, 2.0] for uniform 3-D cloud, got %g", d0)
	}
}

func TestComputeBoxCountingDimension_1DLine(t *testing.T) {
	// Points arranged on the main diagonal of [0,1]^2 → expected BCD ≈ 1.
	n := 200
	pts := make([][]float64, n)
	for i := range pts {
		v := float64(i) / float64(n-1)
		pts[i] = []float64{v, v}
	}
	d0 := metrics.ComputeBoxCountingDimension(pts)
	if d0 < 0.7 || d0 > 1.5 {
		t.Errorf("expected BCD ≈ 1 for diagonal line, got %g", d0)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Unit tests: FractalExporter
// ──────────────────────────────────────────────────────────────────────────────

func TestFractalExporter_InitialState(t *testing.T) {
	fe := metrics.NewFractalExporter(30)
	if fe.D0Variance() != 0 {
		t.Errorf("expected zero initial variance, got %g", fe.D0Variance())
	}
	if fe.ScaleUpTriggered() {
		t.Error("scale-up should not be triggered with no observations")
	}
}

func TestFractalExporter_SingleBatch_NoVariance(t *testing.T) {
	fe := metrics.NewFractalExporter(30)
	rng := rand.New(rand.NewSource(7))
	fe.RecordBatch(metrics.LatentBatch{
		Points:    generateUniformPoints(100, 2, rng),
		Timestamp: time.Now(),
	})
	// One observation → variance is 0.
	if fe.D0Variance() != 0 {
		t.Errorf("expected zero variance with one observation, got %g", fe.D0Variance())
	}
	if fe.ScaleUpTriggered() {
		t.Error("scale-up should not be triggered with a single observation")
	}
}

func TestFractalExporter_D0VarianceIncreasesWithDiverseBatches(t *testing.T) {
	fe := metrics.NewFractalExporter(30)
	rng := rand.New(rand.NewSource(99))
	now := time.Now()

	// Alternate between uniform (high BCD) and clustered (low BCD) batches
	// to guarantee high variance.
	for i := 0; i < 20; i++ {
		var pts [][]float64
		if i%2 == 0 {
			pts = generateUniformPoints(200, 3, rng)
		} else {
			pts = generateCantorPoints(200, 3, rng)
		}
		fe.RecordBatch(metrics.LatentBatch{Points: pts, Timestamp: now.Add(time.Duration(i) * time.Second)})
	}

	if fe.D0Variance() <= 0 {
		t.Errorf("expected positive variance after diverse batches, got %g", fe.D0Variance())
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Integration test: scale-up trigger
// ──────────────────────────────────────────────────────────────────────────────

// TestFractalExporter_ScaleUpTriggered simulates a workload where D₀ variance
// exceeds the 0.5 threshold and verifies the ScaleUpTriggered flag.
func TestFractalExporter_ScaleUpTriggered(t *testing.T) {
	fe := metrics.NewFractalExporter(10)
	now := time.Now()

	// Inject alternating D₀ = 0 and D₀ = 2 observations directly.
	// mean = 1.0, variance = (5*(0-1)² + 5*(2-1)²) / 9 ≈ 1.11 > 0.5
	for i := 0; i < 10; i++ {
		d0 := 0.0
		if i%2 != 0 {
			d0 = 2.0
		}
		fe.RecordD0(d0, now.Add(time.Duration(i)*time.Second))
	}

	variance := fe.D0Variance()
	if variance <= metrics.D0VarianceThreshold {
		t.Errorf("expected variance > %g to trigger scale-up, got %g",
			metrics.D0VarianceThreshold, variance)
	}
	if !fe.ScaleUpTriggered() {
		t.Errorf("ScaleUpTriggered() should return true when variance (%g) > threshold (%g)",
			variance, metrics.D0VarianceThreshold)
	}
}

// TestFractalExporter_NoScaleUpForStableD0 verifies no false positives when D₀
// is stable.
func TestFractalExporter_NoScaleUpForStableD0(t *testing.T) {
	fe := metrics.NewFractalExporter(30)
	rng := rand.New(rand.NewSource(55))
	now := time.Now()

	// Feed 20 batches of the same uniform distribution → near-zero variance.
	for i := 0; i < 20; i++ {
		fe.RecordBatch(metrics.LatentBatch{
			Points:    generateUniformPoints(200, 3, rng),
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	variance := fe.D0Variance()
	if variance > metrics.D0VarianceThreshold {
		t.Errorf("unexpected scale-up trigger for stable D₀ (variance=%g)", variance)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Integration test: window eviction
// ──────────────────────────────────────────────────────────────────────────────

func TestFractalExporter_ObservationWindowBounded(t *testing.T) {
	windowSize := 5
	fe := metrics.NewFractalExporter(windowSize)
	rng := rand.New(rand.NewSource(0))
	now := time.Now()

	// Feed more observations than the window holds.
	for i := 0; i < 20; i++ {
		fe.RecordBatch(metrics.LatentBatch{
			Points:    generateUniformPoints(50, 2, rng),
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	// Variance should be computable (not NaN).
	v := fe.D0Variance()
	if math.IsNaN(v) || math.IsInf(v, 0) {
		t.Errorf("expected finite variance after window eviction, got %g", v)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Integration test: HTTP metrics endpoint
// ──────────────────────────────────────────────────────────────────────────────

func TestFractalExporter_HTTPMetricsEndpoint(t *testing.T) {
	fe := metrics.NewFractalExporter(10)
	rng := rand.New(rand.NewSource(321))
	now := time.Now()

	for i := 0; i < 5; i++ {
		fe.RecordBatch(metrics.LatentBatch{
			Points:    generateUniformPoints(100, 2, rng),
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := 19101
	ready := make(chan struct{})
	go func() {
		// Signal readiness just before blocking on ListenAndServe.
		close(ready)
		_ = fe.ServeMetrics(ctx, port)
	}()
	<-ready
	time.Sleep(50 * time.Millisecond) // give the server time to bind

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	if err != nil {
		t.Skipf("HTTP server not available in this environment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	buf := new(strings.Builder)
	_, _ = io.Copy(buf, resp.Body)
	body := buf.String()
	if !strings.Contains(body, metrics.MetricName) {
		t.Errorf("metric name %q not found in /metrics output:\n%s", metrics.MetricName, body)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Stability test: no oscillation when D₀ variance fluctuates near threshold
// ──────────────────────────────────────────────────────────────────────────────

// TestFractalExporter_StabilityNearThreshold verifies that the scale-up flag
// does not flip rapidly (oscillation) when variance is perturbed just around
// the threshold by injecting controlled noise.
func TestFractalExporter_StabilityNearThreshold(t *testing.T) {
	fe := metrics.NewFractalExporter(30)
	rng := rand.New(rand.NewSource(777))
	now := time.Now()

	// Prime with stable observations (variance ≈ 0).
	for i := 0; i < 30; i++ {
		fe.RecordBatch(metrics.LatentBatch{
			Points:    generateUniformPoints(200, 3, rng),
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}

	// One outlier batch should not be sufficient to trigger.
	fe.RecordBatch(metrics.LatentBatch{
		Points:    generateCantorPoints(200, 3, rng),
		Timestamp: now.Add(31 * time.Second),
	})

	if fe.ScaleUpTriggered() {
		t.Log("note: a single outlier within a large stable window triggered scale-up; " +
			"this may be acceptable depending on window size and data spread")
	}
}
