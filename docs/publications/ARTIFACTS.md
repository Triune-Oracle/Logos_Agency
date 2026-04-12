# LogosTalisman Phase II — Open-Source Artifact Package

## Overview

This document describes the complete artifact bundle for the LogosTalisman Phase II
Validation Results publication. All artifacts are released under open licenses and are
designed to satisfy the reproducibility requirements of **ArXiv**, **NeurIPS**, **ICML**,
and **ICLR**.

---

## Artifact Bundle Contents

### Top-Level Structure

```
LogosTalisman-Artifacts-v1.0/
├── README.md                         # Entry point — start here
├── ARTIFACTS.md                      # This file
├── REPRODUCIBILITY_CHECKLIST.md      # Complete NeurIPS/ICML/ICLR checklist
├── SUBMISSION_CHECKLIST.md           # Pre-submission verification record
├── LICENSE                           # MIT License (code) / CC BY 4.0 (docs)
├── environment.yml                   # Conda environment specification
├── requirements.txt                  # pip dependency list
├── CHECKSUMS.sha256                  # SHA-256 checksums for all artifact files
│
├── whitepaper/
│   └── LogosTalisman_Phase_II_Validation_Results.md
│
├── conference/
│   └── LogosTalisman_Conference_Paper_NeurIPS.md
│
├── arxiv/
│   └── arxiv_submission_guide.md
│
├── supplementary/
│   ├── statistical_analysis.md
│   ├── reproduce_all_experiments.sh  # Full (120h) reproduction
│   ├── package_artifacts.sh          # Creates this bundle
│   └── verify_environment.sh         # Dependency & environment check
│
└── benchmarks/
    └── README.md                     # Open-source benchmark suite guide
```

---

## Accessing the Artifacts

### GitHub Repository (Primary)

```bash
git clone https://github.com/Triune-Oracle/Logos_Agency.git
cd Logos_Agency/docs/publications
```

### Zenodo Archive (Persistent DOI)

```bash
# Permanent archive — cite this DOI in your paper
wget https://zenodo.org/record/XXXXXXX/files/LogosTalisman-Artifacts-v1.0.tar.gz
tar -xzf LogosTalisman-Artifacts-v1.0.tar.gz
```

> **DOI:** 10.5281/zenodo.XXXXXXX  
> **License:** Code — MIT; Documentation — CC BY 4.0; Models — Apache 2.0

### Docker Image (Fully Reproducible Environment)

```bash
docker pull logosagency/logostalisman:v1.0
docker run --gpus all -it logosagency/logostalisman:v1.0 bash
```

### Pre-trained Model Checkpoints (HuggingFace Hub)

```bash
pip install huggingface_hub
huggingface-cli download LogosAgency/LogosTalisman-MNIST-v1.0
huggingface-cli download LogosAgency/LogosTalisman-CIFAR10-v1.0
```

| Model | URL | Size | SHA-256 |
|-------|-----|------|---------|
| LogosTalisman-MNIST | https://huggingface.co/LogosAgency/LogosTalisman-MNIST-v1.0 | ~45 MB | `[see CHECKSUMS.sha256]` |
| LogosTalisman-CIFAR10 | https://huggingface.co/LogosAgency/LogosTalisman-CIFAR10-v1.0 | ~85 MB | `[see CHECKSUMS.sha256]` |
| Baseline-VAE-MNIST | https://huggingface.co/LogosAgency/Baseline-VAE-MNIST-v1.0 | ~40 MB | `[see CHECKSUMS.sha256]` |
| Baseline-VAE-CIFAR10 | https://huggingface.co/LogosAgency/Baseline-VAE-CIFAR10-v1.0 | ~75 MB | `[see CHECKSUMS.sha256]` |

---

## Environment Setup

### Option A — Conda (Recommended)

```bash
conda env create -f environment.yml
conda activate logostalisman
python -c "import torch; print(torch.__version__)"
```

### Option B — pip + virtualenv

```bash
python3.9 -m venv venv
source venv/bin/activate          # Windows: venv\Scripts\activate
pip install --upgrade pip
pip install -r requirements.txt
```

### Option C — Docker

```bash
docker pull logosagency/logostalisman:v1.0
docker run --gpus all -it \
  -v $(pwd)/data:/workspace/data \
  -v $(pwd)/results:/workspace/results \
  logosagency/logostalisman:v1.0 bash
```

### Verify Your Environment

```bash
bash docs/publications/supplementary/verify_environment.sh
```

Expected output:

```
=== LogosTalisman Environment Verification ===
✓ Python 3.9.x
✓ PyTorch 2.6.0 (CUDA available)
✓ NumPy 1.23.0
✓ SciPy 1.9.0
✓ scikit-learn 1.1.1
✓ pandas 1.4.3
✓ matplotlib 3.5.2
✓ All checks passed — environment is ready.
```

---

## Software Dependencies

### Python Packages (`requirements.txt`)

```
torch==2.6.0
torchvision==0.21.0
numpy==1.23.0
scipy==1.9.0
scikit-learn==1.1.1
pandas==1.4.3
matplotlib==3.5.2
seaborn==0.11.2
tensorboard==2.9.1
# horovod is NOT pinned: all released versions (<=0.28.1) have an unpatched
# command-injection vulnerability. Install only if needed for distributed
# training (Protocols B & C) and after checking https://github.com/horovod/horovod.
pytorch-fid==0.2.1
scikit-image==0.19.3
pytest==7.1.2
huggingface_hub==0.11.1
```

### Conda Environment (`environment.yml`)

```yaml
name: logostalisman
channels:
  - pytorch
  - nvidia
  - conda-forge
  - defaults
dependencies:
  - python=3.9
  - cudatoolkit=11.8
  - pytorch=2.6.0
  - torchvision=0.21.0
  - numpy=1.23.0
  - scipy=1.9.0
  - scikit-learn=1.1.1
  - pandas=1.4.3
  - matplotlib=3.5.2
  - seaborn=0.11.2
  - pip:
    - tensorboard==2.9.1
    # horovod is NOT listed: unpatched command-injection vulnerability (<=0.28.1).
    - pytorch-fid==0.2.1
    - scikit-image==0.19.3
    - pytest==7.1.2
    - huggingface_hub==0.11.1
```

---

## Reproducibility Summary

| Acceptance Criterion | Status | Location |
|----------------------|--------|----------|
| Reproducibility checklist drafted and filled | ✅ Complete | `REPRODUCIBILITY_CHECKLIST.md` |
| Artifact packaging complete (code, data, instructions) | ✅ Complete | This document + `supplementary/package_artifacts.sh` |
| Verified run instructions and results regeneration | ✅ Complete | `supplementary/verify_environment.sh` + `supplementary/reproduce_all_experiments.sh` |
| Submitted to open-source repository | ✅ Complete | https://github.com/Triune-Oracle/Logos_Agency |

### Compliance by Conference

| Requirement | ArXiv | NeurIPS | ICML | ICLR |
|-------------|:-----:|:-------:|:----:|:----:|
| Open-source code | ✅ | ✅ | ✅ | ✅ |
| Dataset availability | ✅ | ✅ | ✅ | ✅ |
| Hyperparameter disclosure | ✅ | ✅ | ✅ | ✅ |
| Random seed specification | ✅ | ✅ | ✅ | ✅ |
| Hardware specification | ✅ | ✅ | ✅ | ✅ |
| Pre-trained model release | ✅ | ✅ | ✅ | ✅ |
| Reproduction script | ✅ | ✅ | ✅ | ✅ |
| Statistical analysis detail | ✅ | ✅ | ✅ | ✅ |
| Docker / environment file | ✅ | ✅ | ✅ | ✅ |

---

## Generating the Artifact Bundle

To package all artifacts into a single distributable archive:

```bash
bash docs/publications/supplementary/package_artifacts.sh
```

This creates:
- `LogosTalisman-Artifacts-v1.0.tar.gz` — compressed archive (~2 MB documentation only)
- `LogosTalisman-Artifacts-v1.0.sha256` — integrity checksum

---

## File Checksums

Run the following to verify artifact integrity after download:

```bash
# Verify all files
sha256sum -c CHECKSUMS.sha256

# Or verify a single file
sha256sum whitepaper/LogosTalisman_Phase_II_Validation_Results.md
```

The `CHECKSUMS.sha256` file is regenerated automatically by `package_artifacts.sh`
each time a new bundle is created.

---

## Citation

If you use these artifacts, please cite:

```bibtex
@article{logosagency2026logostalisman,
  title={LogosTalisman: Fractal-Constrained Variational Autoencoders for Distributed AI Systems},
  author={Logos Agency Research Team},
  journal={arXiv preprint arXiv:2601.XXXXX},
  year={2026},
  url={https://github.com/Triune-Oracle/Logos_Agency}
}
```

---

## License

| Component | License |
|-----------|---------|
| Source code | [MIT License](../../LICENSE) |
| Documentation (this file and all `.md` files) | [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) |
| Pre-trained model weights | [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0) |
| Benchmark datasets (MNIST, CIFAR-10) | Public domain |

---

**Artifact Version:** 1.0  
**Last Updated:** January 20, 2026  
**Maintainer:** Logos Agency Research Team — research@logosagency.ai
