#!/bin/bash
# verify_environment.sh
# Checks that all required dependencies are installed and prints a summary.
# Run this before attempting to reproduce the LogosTalisman experiments.
#
# Usage:
#   bash docs/publications/supplementary/verify_environment.sh

set -euo pipefail

PASS=0
FAIL=0

ok()   { echo "  ✓  $*"; PASS=$((PASS + 1)); }
fail() { echo "  ✗  $*"; FAIL=$((FAIL + 1)); }
info() { echo "       $*"; }
sep()  { echo ""; }

echo ""
echo "=================================================="
echo "  LogosTalisman — Environment Verification"
echo "=================================================="
sep

# ---------------------------------------------------------------------------
# 1. Python version (require 3.8+)
# ---------------------------------------------------------------------------
echo "[ Python ]"
if command -v python3 &>/dev/null; then
    PY_VER=$(python3 -c "import sys; print('{}.{}.{}'.format(*sys.version_info[:3]))")
    PY_MAJOR=$(python3 -c "import sys; print(sys.version_info.major)")
    PY_MINOR=$(python3 -c "import sys; print(sys.version_info.minor)")
    if [ "$PY_MAJOR" -ge 3 ] && [ "$PY_MINOR" -ge 8 ]; then
        ok "Python ${PY_VER}"
    else
        fail "Python ${PY_VER} — need 3.8 or later"
    fi
else
    fail "python3 not found"
fi
sep

# ---------------------------------------------------------------------------
# 2. Core Python packages
# ---------------------------------------------------------------------------
echo "[ Python Packages ]"

check_pkg() {
    local pkg="$1"
    local required="$2"
    local import_name="${3:-$1}"

    actual=$(python3 -c "import ${import_name}; print(${import_name}.__version__)" 2>/dev/null || echo "NOT_INSTALLED")
    if [ "$actual" = "NOT_INSTALLED" ]; then
        fail "${pkg} — not installed (required: ${required})"
    else
        ok "${pkg} ${actual} (required: ${required})"
    fi
}

check_pkg "torch"        "1.12.0"
check_pkg "torchvision"  "0.13.0"
check_pkg "numpy"        "1.23.0"
check_pkg "scipy"        "1.9.0"
check_pkg "scikit-learn" "1.1.1"   "sklearn"
check_pkg "pandas"       "1.4.3"
check_pkg "matplotlib"   "3.5.2"
check_pkg "seaborn"      "0.11.2"
check_pkg "tensorboard"  "2.9.1"
check_pkg "scikit-image" "0.19.3"  "skimage"
check_pkg "pytest"       "7.1.2"
sep

# ---------------------------------------------------------------------------
# 3. CUDA / GPU availability
# ---------------------------------------------------------------------------
echo "[ GPU / CUDA ]"
if command -v nvidia-smi &>/dev/null; then
    GPU_INFO=$(nvidia-smi --query-gpu=name,memory.total --format=csv,noheader 2>/dev/null | head -1 || echo "unknown")
    ok "nvidia-smi available — ${GPU_INFO}"

    CUDA_AVAIL=$(python3 -c "import torch; print(torch.cuda.is_available())" 2>/dev/null || echo "False")
    if [ "$CUDA_AVAIL" = "True" ]; then
        CUDA_VER=$(python3 -c "import torch; print(torch.version.cuda)" 2>/dev/null || echo "unknown")
        ok "PyTorch CUDA available (CUDA ${CUDA_VER})"
    else
        fail "PyTorch CUDA not available — GPU experiments will not run"
        info "CPU-only mode is supported for quick verification (--quick flag)."
    fi
else
    fail "nvidia-smi not found — GPU experiments will not run"
    info "CPU-only mode is supported for quick verification (--quick flag)."
fi
sep

# ---------------------------------------------------------------------------
# 4. Container / orchestration tools (optional — needed for Protocols B & C)
# ---------------------------------------------------------------------------
echo "[ Container & Orchestration (optional) ]"

if command -v docker &>/dev/null; then
    DOCKER_VER=$(docker --version 2>/dev/null | awk '{print $3}' | tr -d ',')
    ok "Docker ${DOCKER_VER}"
else
    info "docker not found — required only for containerised reproduction"
fi

if command -v kubectl &>/dev/null; then
    K8S_VER=$(kubectl version --client --short 2>/dev/null | head -1 || echo "unknown")
    ok "kubectl — ${K8S_VER}"
else
    info "kubectl not found — required only for Test Protocol B/C (multi-node)"
fi

if python3 -c "import horovod" &>/dev/null; then
    HVOD_VER=$(python3 -c "import horovod; print(horovod.__version__)" 2>/dev/null || echo "unknown")
    ok "horovod ${HVOD_VER}"
else
    info "horovod not installed — required only for distributed training (Protocol B/C)"
fi
sep

# ---------------------------------------------------------------------------
# 5. Disk space (warn if < 200 GB available)
# ---------------------------------------------------------------------------
echo "[ Disk Space ]"
AVAIL_KB=$(df -k . 2>/dev/null | awk 'NR==2{print $4}' || echo "0")
AVAIL_GB=$((AVAIL_KB / 1024 / 1024))
if [ "$AVAIL_GB" -ge 200 ]; then
    ok "Available disk space: ${AVAIL_GB} GB (≥ 200 GB recommended for full reproduction)"
elif [ "$AVAIL_GB" -ge 50 ]; then
    ok "Available disk space: ${AVAIL_GB} GB (sufficient for Test Protocol A)"
    info "200 GB+ recommended for full reproduction including Protocols B & C"
else
    fail "Available disk space: ${AVAIL_GB} GB — at least 50 GB required"
fi
sep

# ---------------------------------------------------------------------------
# 6. Random-seed / determinism sanity check
# ---------------------------------------------------------------------------
echo "[ Determinism Check ]"
DETERM=$(python3 - <<'PYEOF' 2>/dev/null
import random, numpy as np
try:
    import torch
    random.seed(42)
    np.random.seed(42)
    torch.manual_seed(42)
    torch.backends.cudnn.deterministic = True
    torch.backends.cudnn.benchmark = False
    r = torch.randn(3).tolist()
    # Fixed-seed values for torch.randn(3) with seed 42
    expected = [0.3367, 0.1288, 0.2345]
    ok = all(abs(a - b) < 1e-3 for a, b in zip(r, expected))
    print("PASS" if ok else "WARN")
except Exception as e:
    print(f"WARN:{e}")
PYEOF
)
case "$DETERM" in
    PASS) ok "Deterministic seed check passed (torch.manual_seed(42))" ;;
    WARN*) info "Determinism check returned a warning: ${DETERM}" ;;
    *) info "Determinism check skipped (PyTorch not available)" ;;
esac
sep

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo "=================================================="
if [ "$FAIL" -eq 0 ]; then
    echo "  ✅  All checks passed (${PASS} passed, 0 failed)"
    echo "      Environment is ready for reproduction."
else
    echo "  ⚠️   ${FAIL} check(s) failed, ${PASS} passed."
    echo "      Install missing packages before running experiments."
    echo ""
    echo "  Quick fix:"
    echo "    pip install -r docs/publications/requirements.txt"
    echo "    # OR"
    echo "    conda env create -f docs/publications/environment.yml"
    echo "    conda activate logostalisman"
fi
echo "=================================================="
echo ""

# Exit non-zero if any hard checks failed
[ "$FAIL" -eq 0 ]
