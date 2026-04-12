# Fractal-Aware Horizontal Pod Autoscaling

## Overview

LogosTalisman's distributed training pods are scaled dynamically based on the
**Box-Counting Dimension (D₀) variance** of each worker's latent-code batch,
rather than standard CPU or memory utilisation. When the geometry of the
latent space changes rapidly (high D₀ variance), the cluster adds training
workers to exploit the broader exploration surface; when the space stabilises
(low variance) workers are consolidated.

---

## Components

| File | Purpose |
|------|---------|
| `pkg/k8s/metrics/fractal_exporter.go` | Go library: BCD calculation + Prometheus metrics HTTP server |
| `k8s/hpa-fractal-aware.yaml` | Kubernetes HPA manifest |
| `tests/fractal_hpa_test.go` | Integration & unit tests |

---

## Architecture

```
Training Pod (N replicas)
  │
  ├── computes D₀ per batch via ComputeBoxCountingDimension()
  │
  └── FractalExporter.RecordBatch()
        │
        ├── rolling variance window (default 30 observations)
        │
        └── /metrics  (port 9101)
              │
              └── Prometheus scrape
                    │
                    └── custom-metrics-apiserver adapter
                          │
                          └── HPA controller
                                ├── scale-up  (+4 pods) when variance > 0.5
                                └── scale-down (−2 pods) when variance ≤ 0.5
```

---

## Metric

**Name:** `logostalisman_fractal_complexity_metric`

**Labels exposed on `/metrics`:**

| Label | Description |
|-------|-------------|
| `d0_variance` | Rolling sample variance of D₀ across the observation window |
| `d0_current` | Most recently computed Box-Counting Dimension |
| `observation_count` | Number of observations currently in the window |

**HPA target:** `averageValue: 500m` (i.e. 0.5, expressed as a Kubernetes
milli-quantity). Pods are scaled up when the per-pod average exceeds this value.

---

## Scaling Policy

| Parameter | Value |
|-----------|-------|
| Min replicas | 8 |
| Max replicas | 64 |
| Scale-up step | +4 pods |
| Scale-down step | −2 pods |
| Cooldown (stabilisation window) | 60 seconds |
| Trigger condition | D₀ variance > 0.5 |

---

## Box-Counting Dimension Algorithm

The BCD quantifies the fractal dimension of the latent-code batch by counting
how the number of occupied grid boxes scales with box size:

1. Normalise latent codes to `[0, 1]^d`.
2. For each level `k = 1 … 8`, set box size `r_k = 2^(-k)`.
3. Count `N(r_k)` = number of `d`-dimensional boxes containing ≥ 1 point.
4. Estimate `D₀ = -d(log N(r)) / d(log r)` via OLS regression of
   `log N(r_k)` against `log(1/r_k)`.

Complexity: `O(K × n × d)` where `K = 8`, `n = batch_size`, `d = latent_dim`.

---

## Deployment

### 1. Deploy the exporter sidecar

Add the metrics exporter as a sidecar to the `logostalisman-training` Deployment,
or embed it in the training pod and expose port `9101`.

```yaml
# excerpt – add to your training Deployment's container spec
ports:
  - name: fractal-metrics
    containerPort: 9101
    protocol: TCP
```

### 2. Create a Prometheus ServiceMonitor (or scrape config)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: logostalisman-fractal-metrics
  namespace: logostalisman
spec:
  selector:
    matchLabels:
      app: logostalisman
  endpoints:
    - port: fractal-metrics
      interval: 15s
      path: /metrics
```

### 3. Configure the custom-metrics API adapter

Follow the
[Kubernetes custom metrics API documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#scaling-on-custom-metrics)
to bridge the Prometheus metric into the `custom.metrics.k8s.io` API group
(e.g. using the
[prometheus-adapter](https://github.com/kubernetes-sigs/prometheus-adapter)).

Example `prometheus-adapter` rule:

```yaml
rules:
  - seriesQuery: 'logostalisman_fractal_complexity_metric{dimension="d0_variance"}'
    resources:
      overrides:
        namespace:
          resource: namespace
        pod:
          resource: pod
    name:
      matches: "logostalisman_fractal_complexity_metric"
      as: "logostalisman_fractal_complexity_metric"
    metricsQuery: 'avg_over_time(logostalisman_fractal_complexity_metric{dimension="d0_variance",<<.LabelMatchers>>}[1m])'
```

### 4. Apply the HPA manifest

```bash
kubectl apply -f k8s/hpa-fractal-aware.yaml
```

---

## Success Criteria

| Criterion | Target |
|-----------|--------|
| Scale-up when D₀ variance > 0.5 | ✓ verified by `TestFractalExporter_ScaleUpTriggered` |
| Scaling event completes within 90 s | depends on cluster; cooldown set to 60 s |
| No unnecessary oscillation | ✓ verified by `TestFractalExporter_StabilityNearThreshold` |
| Resource utilisation 60–80% during scaling | monitored via existing Prometheus/Grafana stack |

---

## Running Tests

```bash
go test ./tests/... -run TestFractal -v
```

---

## References

- [Kubernetes HPA documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [Prometheus Adapter](https://github.com/kubernetes-sigs/prometheus-adapter)
- LogosTalisman Architecture: `docs/architecture/distributed-orchestration.md`
- Statistical Type Inference (D₀ provider): `engine/stat_inference.go`
