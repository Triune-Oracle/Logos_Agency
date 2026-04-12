#!/bin/bash
# package_artifacts.sh
# Creates a distributable open-source artifact bundle for the LogosTalisman
# Phase II Validation Results publication.
#
# Usage:
#   bash docs/publications/supplementary/package_artifacts.sh [--output-dir DIR]
#
# Outputs:
#   LogosTalisman-Artifacts-v1.0.tar.gz   — compressed archive
#   LogosTalisman-Artifacts-v1.0.sha256   — SHA-256 integrity checksum

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
BUNDLE_NAME="LogosTalisman-Artifacts-v1.0"
PUBLICATIONS_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUTPUT_DIR="${1:-$(pwd)}"
BUNDLE_DIR="/tmp/${BUNDLE_NAME}"
ARCHIVE="${OUTPUT_DIR}/${BUNDLE_NAME}.tar.gz"
CHECKSUM_FILE="${OUTPUT_DIR}/${BUNDLE_NAME}.sha256"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
info()  { echo "  [INFO]  $*"; }
ok()    { echo "  [  OK ] $*"; }
warn()  { echo "  [WARN]  $*"; }
error() { echo "  [ERROR] $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Banner
# ---------------------------------------------------------------------------
echo ""
echo "============================================================"
echo "  LogosTalisman — Artifact Packaging Script"
echo "  Bundle: ${BUNDLE_NAME}"
echo "============================================================"
echo ""

# ---------------------------------------------------------------------------
# Verify required source files exist
# ---------------------------------------------------------------------------
info "Verifying source files..."

REQUIRED_FILES=(
    "ARTIFACTS.md"
    "REPRODUCIBILITY_CHECKLIST.md"
    "SUBMISSION_CHECKLIST.md"
    "INDEX.md"
    "README.md"
    "whitepaper/LogosTalisman_Phase_II_Validation_Results.md"
    "conference/LogosTalisman_Conference_Paper_NeurIPS.md"
    "arxiv/arxiv_submission_guide.md"
    "supplementary/statistical_analysis.md"
    "supplementary/reproduce_all_experiments.sh"
    "supplementary/verify_environment.sh"
    "benchmarks/README.md"
)

missing=0
for f in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "${PUBLICATIONS_DIR}/${f}" ]; then
        warn "Missing: ${f}"
        missing=$((missing + 1))
    fi
done

if [ "$missing" -gt 0 ]; then
    error "${missing} required file(s) missing. Aborting."
fi
ok "All required source files present."

# ---------------------------------------------------------------------------
# Create staging directory
# ---------------------------------------------------------------------------
info "Preparing staging directory: ${BUNDLE_DIR}"
rm -rf "${BUNDLE_DIR}"
mkdir -p "${BUNDLE_DIR}"

# ---------------------------------------------------------------------------
# Copy artifacts into staging area
# ---------------------------------------------------------------------------
info "Copying publication documents..."

# Root-level documents
for f in \
    ARTIFACTS.md \
    REPRODUCIBILITY_CHECKLIST.md \
    SUBMISSION_CHECKLIST.md \
    INDEX.md \
    README.md; do
    cp "${PUBLICATIONS_DIR}/${f}" "${BUNDLE_DIR}/"
done

# Sub-directories
for d in whitepaper conference arxiv supplementary benchmarks; do
    mkdir -p "${BUNDLE_DIR}/${d}"
    cp "${PUBLICATIONS_DIR}/${d}/"* "${BUNDLE_DIR}/${d}/" 2>/dev/null || true
done

# Repository-level LICENSE (go up two levels from docs/publications)
REPO_ROOT="$(cd "${PUBLICATIONS_DIR}/../.." && pwd)"
if [ -f "${REPO_ROOT}/LICENSE" ]; then
    cp "${REPO_ROOT}/LICENSE" "${BUNDLE_DIR}/LICENSE"
    ok "LICENSE copied."
else
    warn "LICENSE not found at repo root; skipping."
fi

ok "Artifacts staged."

# ---------------------------------------------------------------------------
# Generate environment files (if not already present at repo root)
# ---------------------------------------------------------------------------

# requirements.txt
if [ -f "${REPO_ROOT}/requirements.txt" ]; then
    cp "${REPO_ROOT}/requirements.txt" "${BUNDLE_DIR}/requirements.txt"
    ok "requirements.txt copied from repo root."
else
    info "Generating requirements.txt from checklist values..."
    cat > "${BUNDLE_DIR}/requirements.txt" <<'REQS'
torch==1.12.0
torchvision==0.13.0
numpy==1.23.0
scipy==1.9.0
scikit-learn==1.1.1
pandas==1.4.3
matplotlib==3.5.2
seaborn==0.11.2
tensorboard==2.9.1
horovod==0.27.0
pytorch-fid==0.2.1
scikit-image==0.19.3
pytest==7.1.2
huggingface_hub==0.11.1
REQS
    ok "requirements.txt generated."
fi

# environment.yml
if [ -f "${REPO_ROOT}/environment.yml" ]; then
    cp "${REPO_ROOT}/environment.yml" "${BUNDLE_DIR}/environment.yml"
    ok "environment.yml copied from repo root."
else
    info "Generating environment.yml from checklist values..."
    cat > "${BUNDLE_DIR}/environment.yml" <<'ENV'
name: logostalisman
channels:
  - pytorch
  - nvidia
  - conda-forge
  - defaults
dependencies:
  - python=3.9
  - cudatoolkit=11.6
  - pytorch=1.12.0
  - torchvision=0.13.0
  - numpy=1.23.0
  - scipy=1.9.0
  - scikit-learn=1.1.1
  - pandas=1.4.3
  - matplotlib=3.5.2
  - seaborn=0.11.2
  - pip:
    - tensorboard==2.9.1
    - horovod==0.27.0
    - pytorch-fid==0.2.1
    - scikit-image==0.19.3
    - pytest==7.1.2
    - huggingface_hub==0.11.1
ENV
    ok "environment.yml generated."
fi

# ---------------------------------------------------------------------------
# Generate SHA-256 checksums for every file in the bundle
# ---------------------------------------------------------------------------
info "Computing SHA-256 checksums..."
CHECKSUMS_FILE="${BUNDLE_DIR}/CHECKSUMS.sha256"

# Use subshell to keep relative paths in the checksum file
(
    cd "${BUNDLE_DIR}"
    find . -type f ! -name "CHECKSUMS.sha256" | sort | while read -r f; do
        sha256sum "$f"
    done
) > "${CHECKSUMS_FILE}"

ok "Checksums written to CHECKSUMS.sha256 ($(wc -l < "${CHECKSUMS_FILE}") files)."

# ---------------------------------------------------------------------------
# Create compressed archive
# ---------------------------------------------------------------------------
info "Creating archive: ${ARCHIVE}"
mkdir -p "${OUTPUT_DIR}"
tar -czf "${ARCHIVE}" -C "/tmp" "${BUNDLE_NAME}"
ok "Archive created: ${ARCHIVE} ($(du -sh "${ARCHIVE}" | cut -f1))."

# ---------------------------------------------------------------------------
# Compute archive checksum
# ---------------------------------------------------------------------------
info "Computing archive checksum..."
sha256sum "${ARCHIVE}" > "${CHECKSUM_FILE}"
ok "Archive checksum: ${CHECKSUM_FILE}"

# ---------------------------------------------------------------------------
# Cleanup staging area
# ---------------------------------------------------------------------------
rm -rf "${BUNDLE_DIR}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "============================================================"
echo "  Artifact Packaging Complete"
echo "============================================================"
echo ""
echo "  Archive  : ${ARCHIVE}"
echo "  Checksum : ${CHECKSUM_FILE}"
echo ""
echo "  To verify integrity after transfer:"
echo "    sha256sum -c ${BUNDLE_NAME}.sha256"
echo ""
echo "  To extract:"
echo "    tar -xzf ${BUNDLE_NAME}.tar.gz"
echo ""
echo "  Upload the archive and checksum file to:"
echo "    • Zenodo   : https://zenodo.org (for persistent DOI)"
echo "    • GitHub   : as a Release asset at"
echo "                 https://github.com/Triune-Oracle/Logos_Agency/releases"
echo ""
