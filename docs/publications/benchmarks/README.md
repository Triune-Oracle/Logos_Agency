# LogosTalisman Benchmark Suite

## Overview

This benchmark suite provides standardized evaluation protocols for LogosTalisman and comparison with baseline VAE variants. It includes dataset loaders, evaluation metrics, automated testing scripts, and visualization tools.

---

## Directory Structure

```
benchmarks/
├── README.md                           # This file
├── setup.py                           # Package installation
├── requirements.txt                   # Dependencies
├── configs/                           # Configuration files
│   ├── logostalisman_mnist.yaml
│   ├── logostalisman_cifar10.yaml
│   ├── baseline_vae.yaml
│   └── beta_vae.yaml
├── data/                              # Dataset management
│   ├── __init__.py
│   ├── download.py                    # Dataset downloaders
│   ├── loaders.py                     # PyTorch DataLoaders
│   └── preprocessing.py               # Data preprocessing
├── models/                            # Model implementations
│   ├── __init__.py
│   ├── logostalisman.py               # LogosTalisman VAE
│   ├── baseline_vae.py                # Standard VAE
│   ├── beta_vae.py                    # β-VAE
│   └── components.py                  # Shared components
├── metrics/                           # Evaluation metrics
│   ├── __init__.py
│   ├── quality.py                     # PSNR, SSIM, FID
│   ├── fractal.py                     # Box-Counting Dimension
│   ├── resilience.py                  # Fault tolerance metrics
│   └── efficiency.py                  # Scaling metrics
├── experiments/                       # Test protocols
│   ├── __init__.py
│   ├── test_protocol_a.py             # Quality evaluation
│   ├── test_protocol_b.py             # Resilience testing
│   ├── test_protocol_c.py             # Scaling analysis
│   ├── evaluate_quality.py            # Quality metric computation
│   ├── monitor_resilience.py          # Resilience monitoring
│   └── analyze_scaling.py             # Scaling analysis
├── distributed/                       # Distributed training
│   ├── __init__.py
│   ├── circuit_breaker.py             # Circuit breaker mechanism
│   ├── checkpoint.py                  # Checkpoint management
│   ├── failure_injection.py           # Chaos engineering
│   └── kubernetes_utils.py            # K8s utilities
├── visualization/                     # Plotting and visualization
│   ├── __init__.py
│   ├── quality_plots.py               # Quality comparison plots
│   ├── latent_space.py                # t-SNE, PCA visualizations
│   ├── resilience_timeline.py         # Failure timeline plots
│   └── scaling_curves.py              # Scaling efficiency plots
├── scripts/                           # Automation scripts
│   ├── download_datasets.py
│   ├── reproduce_all_experiments.sh
│   ├── setup_kubernetes_cluster.sh
│   └── compare_results.py
└── tests/                             # Unit tests
    ├── test_models.py
    ├── test_metrics.py
    ├── test_circuit_breaker.py
    └── test_fractal.py
```

---

## Installation

### Option 1: pip install

```bash
# Clone repository
git clone https://github.com/Triune-Oracle/LogosTalisman-Benchmarks.git
cd LogosTalisman-Benchmarks

# Install package
pip install -e .

# Verify installation
python -c "import logostalisman_benchmarks; print('Installation successful!')"
```

### Option 2: Docker

```bash
# Pull Docker image
docker pull logosagency/logostalisman-benchmarks:v1.0

# Run container
docker run --gpus all -v $(pwd)/data:/data -v $(pwd)/results:/results \
  -it logosagency/logostalisman-benchmarks:v1.0 bash
```

### Option 3: Conda

```bash
# Create environment
conda env create -f environment.yml
conda activate logostalisman-benchmarks

# Install package
pip install -e .
```

---

## Quick Start

### 1. Download Datasets

```bash
# Download MNIST
python scripts/download_datasets.py --dataset mnist --path data/mnist

# Download CIFAR-10
python scripts/download_datasets.py --dataset cifar10 --path data/cifar10

# Download all
python scripts/download_datasets.py --dataset all --path data/
```

### 2. Train LogosTalisman (Single-Node)

```bash
# MNIST
python experiments/test_protocol_a.py \
  --dataset mnist \
  --model logostalisman \
  --config configs/logostalisman_mnist.yaml \
  --epochs 200 \
  --seed 42 \
  --output results/mnist_single/

# CIFAR-10
python experiments/test_protocol_a.py \
  --dataset cifar10 \
  --model logostalisman \
  --config configs/logostalisman_cifar10.yaml \
  --epochs 200 \
  --seed 42 \
  --output results/cifar10_single/
```

### 3. Evaluate Quality

```bash
# Evaluate trained model
python experiments/evaluate_quality.py \
  --checkpoint results/mnist_single/checkpoints/epoch200.pth \
  --dataset mnist \
  --metrics psnr ssim fid bcd \
  --output results/mnist_single/evaluation.json

# Generate comparison plots
python visualization/quality_plots.py \
  --baseline results/baseline_vae/evaluation.json \
  --beta-vae results/beta_vae/evaluation.json \
  --logostalisman results/mnist_single/evaluation.json \
  --output figures/quality_comparison.pdf
```

### 4. Multi-Node Training (Kubernetes)

```bash
# Setup 16-node cluster
./scripts/setup_kubernetes_cluster.sh --nodes 16

# Run resilience test
python experiments/test_protocol_b.py \
  --num-nodes 16 \
  --failure-mode random \
  --failure-interval 1800 \
  --duration 259200 \
  --output results/resilience_test/

# Monitor in real-time
python experiments/monitor_resilience.py \
  --log-dir results/resilience_test/logs/
```

---

## Benchmark Protocols

### Test Protocol A: Quality Evaluation

**Objective:** Measure reconstruction quality across models and datasets

**Models Evaluated:**
- Baseline VAE (Gaussian prior, β=1.0)
- β-VAE (Gaussian prior, β=4.0)
- LogosTalisman (Fractal prior, λ=0.1)

**Metrics:**
- PSNR (Peak Signal-to-Noise Ratio)
- SSIM (Structural Similarity Index)
- FID (Fréchet Inception Distance)
- BCD (Box-Counting Dimension)

**Command:**
```bash
# Run full quality benchmark (all models, all datasets, 5 seeds)
python experiments/test_protocol_a.py \
  --models baseline,beta-vae,logostalisman \
  --datasets mnist,cifar10 \
  --seeds 42,123,456,789,1011 \
  --epochs 200 \
  --output results/protocol_a/

# Expected runtime: ~80 hours (parallelizable across seeds)
```

**Outputs:**
- `results/protocol_a/metrics_summary.csv`
- `results/protocol_a/statistical_tests.json`
- `results/protocol_a/checkpoints/`
- `results/protocol_a/samples/`

### Test Protocol B: Resilience Testing

**Objective:** Validate fault tolerance during distributed training

**Failure Scenarios:**
1. Single node crash (random pod termination)
2. Network partition (25% nodes isolated)
3. Cascading failure (3 sequential failures)
4. Byzantine fault (corrupted gradients)

**Metrics:**
- Training continuity (%)
- Recovery time (seconds)
- Loss spike (ratio)
- Checkpoint overhead (GB)

**Command:**
```bash
# Run resilience benchmark
python experiments/test_protocol_b.py \
  --num-nodes 16 \
  --failure-scenarios all \
  --failure-interval 1800 \
  --duration 259200 \
  --chaos-mesh-enabled \
  --output results/protocol_b/

# Expected runtime: 72 hours
```

**Outputs:**
- `results/protocol_b/continuity_metrics.json`
- `results/protocol_b/recovery_times.csv`
- `results/protocol_b/failure_events.log`
- `results/protocol_b/timeline_plot.pdf`

### Test Protocol C: Scaling Analysis

**Objective:** Measure throughput and efficiency across cluster sizes

**Configurations:**
- Node counts: 1, 2, 4, 8, 16, 32, 64
- Global batch size: 8192 (fixed)
- Iterations: 10,000 per configuration

**Metrics:**
- Throughput (samples/second)
- Speedup S(N) = Throughput(N) / Throughput(1)
- Scaling efficiency E(N) = S(N) / N × 100%
- Communication overhead (ms)

**Command:**
```bash
# Run scaling benchmark (automated cluster resizing)
python experiments/test_protocol_c.py \
  --node-configs 1,2,4,8,16,32,64 \
  --global-batch-size 8192 \
  --iterations 10000 \
  --profile-communication \
  --output results/protocol_c/

# Expected runtime: 24 hours (sequential)
```

**Outputs:**
- `results/protocol_c/throughput.csv`
- `results/protocol_c/scaling_efficiency.pdf`
- `results/protocol_c/communication_breakdown.csv`
- `results/protocol_c/gpu_utilization.csv`

---

## API Reference

### Data Loading

```python
from logostalisman_benchmarks.data import get_dataloader

# MNIST
train_loader, test_loader = get_dataloader(
    dataset='mnist',
    batch_size=128,
    shuffle=True,
    num_workers=4
)

# CIFAR-10
train_loader, test_loader = get_dataloader(
    dataset='cifar10',
    batch_size=128,
    shuffle=True,
    num_workers=4
)
```

### Model Creation

```python
from logostalisman_benchmarks.models import (
    LogosTalisman,
    BaselineVAE,
    BetaVAE
)

# LogosTalisman
model = LogosTalisman(
    input_channels=1,  # 1 for MNIST, 3 for CIFAR-10
    latent_dim=32,
    lambda_fractal=0.1,
    gamma_circuit=0.05,
    target_bcd=1.7,  # 1.7 for MNIST, 2.3 for CIFAR-10
    circuit_epsilon=2.0
)

# Baseline VAE
baseline = BaselineVAE(
    input_channels=1,
    latent_dim=32,
    beta=1.0
)

# β-VAE
beta_vae = BetaVAE(
    input_channels=1,
    latent_dim=32,
    beta=4.0
)
```

### Training

```python
from logostalisman_benchmarks.training import train_epoch, evaluate

# Training loop
for epoch in range(num_epochs):
    # Train
    train_metrics = train_epoch(
        model=model,
        dataloader=train_loader,
        optimizer=optimizer,
        device='cuda',
        epoch=epoch
    )
    
    # Evaluate
    eval_metrics = evaluate(
        model=model,
        dataloader=test_loader,
        device='cuda'
    )
    
    print(f"Epoch {epoch}: Loss={train_metrics['loss']:.4f}, "
          f"PSNR={eval_metrics['psnr']:.2f}, "
          f"BCD={eval_metrics['bcd']:.2f}")
```

### Quality Metrics

```python
from logostalisman_benchmarks.metrics import (
    compute_psnr,
    compute_ssim,
    compute_fid,
    compute_box_counting_dimension
)

# PSNR
psnr = compute_psnr(original_images, reconstructed_images)

# SSIM
ssim = compute_ssim(original_images, reconstructed_images)

# FID
fid = compute_fid(real_images, generated_images, device='cuda')

# Box-Counting Dimension
bcd = compute_box_counting_dimension(latent_codes, num_scales=8)
```

### Circuit Breaker

```python
from logostalisman_benchmarks.distributed import CircuitBreaker

# Initialize circuit breaker
circuit_breaker = CircuitBreaker(
    checkpoint_interval=100,
    replication_factor=3,
    epsilon=2.0,
    failure_timeout=30,
    gradient_anomaly_threshold=10.0
)

# Training with circuit breaker
for batch_idx, (data, _) in enumerate(dataloader):
    # Forward pass
    loss, latent_codes = model(data)
    
    # Add circuit breaker loss
    if circuit_breaker.has_checkpoints():
        circuit_loss = circuit_breaker.compute_loss(latent_codes)
        loss = loss + gamma * circuit_loss
    
    # Backward pass
    loss.backward()
    optimizer.step()
    
    # Update checkpoint
    if batch_idx % 100 == 0:
        circuit_breaker.save_checkpoint(
            model=model,
            latent_codes=latent_codes,
            batch_idx=batch_idx
        )
    
    # Check for failures
    if circuit_breaker.detect_failure(gradients):
        circuit_breaker.recover(model)
```

### Visualization

```python
from logostalisman_benchmarks.visualization import (
    plot_quality_comparison,
    plot_latent_space,
    plot_resilience_timeline,
    plot_scaling_curves
)

# Quality comparison
plot_quality_comparison(
    results_dict={
        'Baseline VAE': baseline_results,
        'β-VAE': beta_vae_results,
        'LogosTalisman': logostalisman_results
    },
    metrics=['psnr', 'ssim', 'fid'],
    output='figures/quality_comparison.pdf'
)

# Latent space visualization
plot_latent_space(
    latent_codes=latent_codes,
    labels=labels,
    method='tsne',  # or 'pca'
    output='figures/latent_space_tsne.pdf'
)

# Resilience timeline
plot_resilience_timeline(
    loss_history=loss_history,
    failure_events=failure_events,
    recovery_times=recovery_times,
    output='figures/resilience_timeline.pdf'
)

# Scaling curves
plot_scaling_curves(
    node_counts=[1, 2, 4, 8, 16, 32, 64],
    throughputs=throughputs,
    output='figures/scaling_curves.pdf'
)
```

---

## Configuration Files

### LogosTalisman (MNIST)

```yaml
# configs/logostalisman_mnist.yaml

model:
  name: logostalisman
  input_channels: 1
  latent_dim: 32
  
  # Encoder architecture
  encoder:
    - {type: conv2d, out_channels: 32, kernel: 3, stride: 2, padding: 1}
    - {type: relu}
    - {type: conv2d, out_channels: 64, kernel: 3, stride: 2, padding: 1}
    - {type: relu}
    - {type: flatten}
    - {type: linear, out_features: 256}
    - {type: relu}
    - {type: linear, out_features: 64}  # mu and logvar
  
  # Decoder architecture
  decoder:
    - {type: linear, out_features: 256}
    - {type: relu}
    - {type: linear, out_features: 3136}
    - {type: relu}
    - {type: reshape, shape: [-1, 64, 7, 7]}
    - {type: conv_transpose2d, out_channels: 32, kernel: 3, stride: 2, padding: 1, output_padding: 1}
    - {type: relu}
    - {type: conv_transpose2d, out_channels: 1, kernel: 3, stride: 2, padding: 1, output_padding: 1}
    - {type: sigmoid}

training:
  epochs: 200
  batch_size: 128
  learning_rate: 0.0001
  optimizer: adam
  optimizer_params:
    betas: [0.9, 0.999]
    eps: 1e-8
    weight_decay: 0.0
  
  scheduler: reduce_lr_on_plateau
  scheduler_params:
    mode: min
    factor: 0.5
    patience: 10
    min_lr: 1e-7
  
  loss_weights:
    beta: 1.0
    lambda: 0.1
    gamma: 0.05
  
  fractal:
    target_bcd: 1.7
    num_scales: 8
    min_box_size: 0.00390625
    max_box_size: 0.5
  
  circuit_breaker:
    epsilon: 2.0
    checkpoint_interval: 100
    replication_factor: 3
    failure_timeout: 30
    gradient_anomaly_threshold: 10.0
    recovery_gamma_start: 0.2
    recovery_gamma_end: 0.05
    recovery_steps: 500

data:
  dataset: mnist
  data_dir: data/mnist
  num_workers: 4
  pin_memory: true

logging:
  log_interval: 100
  save_interval: 10
  tensorboard: true
  wandb: false

device: cuda
seed: 42
deterministic: true
```

---

## Example Scripts

### Download Datasets

```python
# scripts/download_datasets.py
import argparse
from torchvision import datasets

def download_mnist(path):
    """Download MNIST dataset."""
    datasets.MNIST(path, train=True, download=True)
    datasets.MNIST(path, train=False, download=True)
    print(f"✓ MNIST downloaded to {path}")

def download_cifar10(path):
    """Download CIFAR-10 dataset."""
    datasets.CIFAR10(path, train=True, download=True)
    datasets.CIFAR10(path, train=False, download=True)
    print(f"✓ CIFAR-10 downloaded to {path}")

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--dataset', choices=['mnist', 'cifar10', 'all'], required=True)
    parser.add_argument('--path', default='data/')
    args = parser.parse_args()
    
    if args.dataset in ['mnist', 'all']:
        download_mnist(f"{args.path}/mnist")
    
    if args.dataset in ['cifar10', 'all']:
        download_cifar10(f"{args.path}/cifar10")

if __name__ == '__main__':
    main()
```

### Evaluate Quality

```python
# experiments/evaluate_quality.py
import argparse
import json
import torch
from logostalisman_benchmarks.models import load_model
from logostalisman_benchmarks.data import get_dataloader
from logostalisman_benchmarks.metrics import (
    compute_psnr, compute_ssim, compute_fid, compute_box_counting_dimension
)

def evaluate_quality(checkpoint_path, dataset, device='cuda'):
    """Evaluate model quality metrics."""
    # Load model
    model = load_model(checkpoint_path).to(device)
    model.eval()
    
    # Load data
    _, test_loader = get_dataloader(dataset, batch_size=128)
    
    # Collect samples
    original_images = []
    reconstructed_images = []
    latent_codes = []
    
    with torch.no_grad():
        for data, _ in test_loader:
            data = data.to(device)
            recon, mu, logvar, z = model(data)
            
            original_images.append(data.cpu())
            reconstructed_images.append(recon.cpu())
            latent_codes.append(z.cpu())
    
    # Concatenate
    original = torch.cat(original_images)
    reconstructed = torch.cat(reconstructed_images)
    latents = torch.cat(latent_codes)
    
    # Compute metrics
    metrics = {
        'psnr': compute_psnr(original, reconstructed),
        'ssim': compute_ssim(original, reconstructed),
        'fid': compute_fid(original, reconstructed, device),
        'bcd': compute_box_counting_dimension(latents, num_scales=8)
    }
    
    return metrics

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--checkpoint', required=True)
    parser.add_argument('--dataset', choices=['mnist', 'cifar10'], required=True)
    parser.add_argument('--metrics', nargs='+', default=['psnr', 'ssim', 'fid', 'bcd'])
    parser.add_argument('--output', default='evaluation.json')
    args = parser.parse_args()
    
    # Evaluate
    results = evaluate_quality(args.checkpoint, args.dataset)
    
    # Filter requested metrics
    results = {k: v for k, v in results.items() if k in args.metrics}
    
    # Save
    with open(args.output, 'w') as f:
        json.dump(results, f, indent=2)
    
    # Print
    print("Quality Metrics:")
    for metric, value in results.items():
        print(f"  {metric.upper()}: {value:.4f}")

if __name__ == '__main__':
    main()
```

---

## Performance Benchmarks

### Expected Training Times (Single V100)

| Dataset | Model | Epochs | Time per Epoch | Total Time |
|---------|-------|--------|----------------|------------|
| MNIST | Baseline VAE | 200 | 2 min | 6.7 hours |
| MNIST | β-VAE | 200 | 2 min | 6.7 hours |
| MNIST | LogosTalisman | 200 | 2.3 min | 7.7 hours |
| CIFAR-10 | Baseline VAE | 200 | 5 min | 16.7 hours |
| CIFAR-10 | β-VAE | 200 | 5 min | 16.7 hours |
| CIFAR-10 | LogosTalisman | 200 | 5.8 min | 19.3 hours |

**Overhead:** LogosTalisman adds ~15% training time due to BCD calculation.

### Expected Memory Usage

| Model | MNIST (1×28×28) | CIFAR-10 (3×32×32) |
|-------|-----------------|-------------------|
| Baseline VAE | 2.1 GB | 3.8 GB |
| β-VAE | 2.1 GB | 3.8 GB |
| LogosTalisman | 2.4 GB | 4.2 GB |

**Overhead:** LogosTalisman adds ~300 MB for checkpoint storage.

---

## Citation

If you use this benchmark suite, please cite:

```bibtex
@software{logostalisman_benchmarks2026,
  title={LogosTalisman Benchmark Suite},
  author={Logos Agency Research Team},
  year={2026},
  url={https://github.com/Triune-Oracle/LogosTalisman-Benchmarks}
}
```

---

## License

MIT License - see LICENSE file for details.

---

## Contributing

We welcome contributions! Please see CONTRIBUTING.md for guidelines.

**Priority areas:**
- Additional baseline models (VQ-VAE, WAE, RAE)
- Support for more datasets (CelebA, ImageNet, text datasets)
- Improved BCD calculation efficiency
- Distributed training optimizations

---

## Support

- **Documentation:** https://logostalisman-benchmarks.readthedocs.io
- **GitHub Issues:** https://github.com/Triune-Oracle/LogosTalisman-Benchmarks/issues
- **Email:** benchmarks@logosagency.ai
- **Discord:** https://discord.gg/logosagency

---

**Last Updated:** January 20, 2026  
**Version:** 1.0.0
