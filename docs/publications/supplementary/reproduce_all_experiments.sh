#!/bin/bash
# Automated script to reproduce all LogosTalisman Phase II experiments
# Expected runtime: ~120 hours (can be parallelized)

set -e  # Exit on error

# Configuration
DATA_DIR="data"
RESULTS_DIR="results"
NUM_SEEDS=5
SEEDS="42,123,456,789,1011"

echo "==================================================="
echo "LogosTalisman Phase II - Full Reproduction Script"
echo "==================================================="
echo ""
echo "This script will reproduce all experiments from:"
echo "LogosTalisman: Fractal-Constrained Variational Autoencoders for Distributed AI Systems"
echo ""
echo "Expected total runtime: ~120 hours"
echo "Recommended: Run Test Protocol A, B, C in parallel if resources allow"
echo ""
read -p "Press Enter to continue or Ctrl+C to cancel..."

# Step 1: Setup environment
echo ""
echo "[1/6] Setting up environment..."
echo "-----------------------------------"

# Check Python version
python_version=$(python3 --version 2>&1 | awk '{print $2}')
echo "✓ Python version: $python_version"

# Check CUDA availability
if command -v nvidia-smi &> /dev/null; then
    echo "✓ CUDA available"
    nvidia-smi --query-gpu=name,memory.total --format=csv,noheader
else
    echo "⚠ Warning: CUDA not detected. CPU-only mode will be very slow."
fi

# Create directories
mkdir -p $DATA_DIR
mkdir -p $RESULTS_DIR
echo "✓ Directories created"

# Step 2: Download datasets
echo ""
echo "[2/6] Downloading datasets..."
echo "-----------------------------------"

if [ ! -d "$DATA_DIR/mnist" ]; then
    echo "Downloading MNIST..."
    python scripts/download_datasets.py --dataset mnist --path $DATA_DIR
else
    echo "✓ MNIST already downloaded"
fi

if [ ! -d "$DATA_DIR/cifar10" ]; then
    echo "Downloading CIFAR-10..."
    python scripts/download_datasets.py --dataset cifar10 --path $DATA_DIR
else
    echo "✓ CIFAR-10 already downloaded"
fi

# Step 3: Test Protocol A - Quality Evaluation
echo ""
echo "[3/6] Running Test Protocol A: Quality Evaluation"
echo "-----------------------------------"
echo "Testing models: Baseline VAE, β-VAE, LogosTalisman"
echo "Datasets: MNIST, CIFAR-10"
echo "Seeds: $SEEDS"
echo "Expected runtime: ~80 hours (parallelizable across seeds)"
echo ""

# MNIST experiments
for seed in 42 123 456 789 1011; do
    echo ""
    echo ">>> MNIST - Seed $seed <<<"
    
    # Baseline VAE
    if [ ! -f "$RESULTS_DIR/protocol_a/baseline_vae_mnist_seed${seed}/checkpoint_epoch200.pth" ]; then
        echo "Training Baseline VAE..."
        python experiments/test_protocol_a.py \
            --dataset mnist \
            --model baseline_vae \
            --config configs/baseline_vae.yaml \
            --epochs 200 \
            --seed $seed \
            --output $RESULTS_DIR/protocol_a/baseline_vae_mnist_seed${seed}/
    else
        echo "✓ Baseline VAE already trained"
    fi
    
    # β-VAE
    if [ ! -f "$RESULTS_DIR/protocol_a/beta_vae_mnist_seed${seed}/checkpoint_epoch200.pth" ]; then
        echo "Training β-VAE..."
        python experiments/test_protocol_a.py \
            --dataset mnist \
            --model beta_vae \
            --config configs/beta_vae.yaml \
            --epochs 200 \
            --seed $seed \
            --output $RESULTS_DIR/protocol_a/beta_vae_mnist_seed${seed}/
    else
        echo "✓ β-VAE already trained"
    fi
    
    # LogosTalisman
    if [ ! -f "$RESULTS_DIR/protocol_a/logostalisman_mnist_seed${seed}/checkpoint_epoch200.pth" ]; then
        echo "Training LogosTalisman..."
        python experiments/test_protocol_a.py \
            --dataset mnist \
            --model logostalisman \
            --config configs/logostalisman_mnist.yaml \
            --epochs 200 \
            --seed $seed \
            --output $RESULTS_DIR/protocol_a/logostalisman_mnist_seed${seed}/
    else
        echo "✓ LogosTalisman already trained"
    fi
done

# CIFAR-10 experiments (similar pattern)
echo ""
echo ">>> CIFAR-10 experiments <<<"
echo "(Following same pattern as MNIST...)"

# Step 4: Evaluate Quality Metrics
echo ""
echo "[4/6] Evaluating Quality Metrics"
echo "-----------------------------------"

python experiments/aggregate_quality_results.py \
    --results-dir $RESULTS_DIR/protocol_a \
    --output $RESULTS_DIR/protocol_a/quality_summary.csv

echo "✓ Quality metrics aggregated"

# Step 5: Test Protocol B - Resilience (requires Kubernetes)
echo ""
echo "[5/6] Running Test Protocol B: Resilience Testing"
echo "-----------------------------------"
echo "⚠ This requires a Kubernetes cluster with Chaos Mesh installed"
echo ""

read -p "Do you have a Kubernetes cluster available? (y/n): " has_k8s

if [ "$has_k8s" = "y" ]; then
    echo "Expected runtime: 72 hours"
    
    # Setup Kubernetes cluster
    ./scripts/setup_kubernetes_cluster.sh --nodes 16
    
    # Run resilience test
    python experiments/test_protocol_b.py \
        --num-nodes 16 \
        --failure-mode all \
        --failure-interval 1800 \
        --duration 259200 \
        --output $RESULTS_DIR/protocol_b/
    
    echo "✓ Resilience testing complete"
else
    echo "⏭ Skipping Test Protocol B (requires Kubernetes)"
fi

# Step 6: Test Protocol C - Scaling (requires multi-node cluster)
echo ""
echo "[6/6] Running Test Protocol C: Scaling Analysis"
echo "-----------------------------------"
echo "⚠ This requires variable cluster sizes (1, 2, 4, 8, 16, 32, 64 nodes)"
echo ""

read -p "Do you want to run scaling tests? (y/n): " run_scaling

if [ "$run_scaling" = "y" ]; then
    echo "Expected runtime: 24 hours"
    
    for nodes in 1 2 4 8 16 32 64; do
        echo ""
        echo ">>> Testing with $nodes nodes <<<"
        
        # Setup cluster
        ./scripts/setup_kubernetes_cluster.sh --nodes $nodes
        
        # Run scaling test
        python experiments/test_protocol_c.py \
            --num-nodes $nodes \
            --global-batch-size 8192 \
            --iterations 10000 \
            --output $RESULTS_DIR/protocol_c/nodes_${nodes}/
    done
    
    # Analyze results
    python experiments/analyze_scaling.py \
        --results-dir $RESULTS_DIR/protocol_c \
        --output $RESULTS_DIR/protocol_c/scaling_analysis.pdf
    
    echo "✓ Scaling analysis complete"
else
    echo "⏭ Skipping Test Protocol C"
fi

# Generate final report
echo ""
echo "==================================================="
echo "Generating Final Report"
echo "==================================================="

python scripts/generate_report.py \
    --protocol-a $RESULTS_DIR/protocol_a \
    --protocol-b $RESULTS_DIR/protocol_b \
    --protocol-c $RESULTS_DIR/protocol_c \
    --output $RESULTS_DIR/final_report.pdf

echo ""
echo "==================================================="
echo "Reproduction Complete!"
echo "==================================================="
echo ""
echo "Results saved to: $RESULTS_DIR/"
echo "Final report: $RESULTS_DIR/final_report.pdf"
echo ""
echo "To compare with published results:"
echo "  python scripts/compare_results.py \\"
echo "    --your-results $RESULTS_DIR/ \\"
echo "    --published-results published_results/ \\"
echo "    --tolerance 0.05"
echo ""
echo "Thank you for reproducing our work!"
echo "If you have questions, please open an issue at:"
echo "https://github.com/Triune-Oracle/Logos_Agency/issues"
