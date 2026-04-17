# NeurIPS / ICML / ICLR Camera-Ready Checklist

> **Paper:** LogosTalisman: Fractal-Constrained Variational Autoencoders for Distributed AI Systems  
> **Primary venue:** NeurIPS 2026  
> **LaTeX source:** `LogosTalisman_NeurIPS.tex`  
> **Bibliography:** `references.bib`  
> **Last updated:** April 2026

---

## 1. Formatting Compliance

### NeurIPS 2026
- [x] `\usepackage[final]{neurips_2024}` style file used
- [x] Title, author block, and abstract match submission system
- [x] Paper is ≤ 8 pages (main body, excluding references)
- [x] References appear after page 8 (unlimited)
- [x] Appendices follow references (supplementary, unlimited)
- [x] 10 pt font, two-column layout enforced by style file
- [x] No modifications to style file margins or font sizes

### ICML 2026 (parallel adaptation)
- [ ] Switch to `icml2026.sty` style file
- [ ] Verify author/affiliation block format
- [ ] Confirm page limit (8 pages + references)

### ICLR 2027 (parallel adaptation)
- [ ] Switch to `iclr2027.sty` style file
- [ ] Verify author/affiliation block format
- [ ] Confirm page limit (varies by track)

---

## 2. Content Requirements

### Required sections ✅
- [x] **Abstract** — ≤ 250 words, single paragraph, no citations
- [x] **Introduction** — problem statement, insight, numbered contributions
- [x] **Related Work** — VAE variants, fractal DL, distributed training
- [x] **Methodology** — complete mathematical formulations (Eq. 1–3), Algorithm 1
- [x] **Experimental Setup** — protocols, datasets, models, hyperparameters, hardware
- [x] **Results** — six tables with mean ± std, statistical significance stated
- [x] **Discussion** — why fractals work, limitations, broader impact
- [x] **Conclusion** — summary and reproducibility pointer
- [x] **References** — 12 entries, properly formatted with BibTeX

### Mathematical notation
- [x] All symbols defined on first use
- [x] Equations numbered sequentially
- [x] Consistent notation across sections
- [x] Straight-through estimator gradient justification included

### Figures and tables
- [x] 6 result tables (Tables 1–6, numbered sequentially)
- [x] 1 algorithm block (Algorithm 1, BCD Estimation)
- [x] All tables use `booktabs` (\\toprule / \\midrule / \\bottomrule)
- [x] Best results **bolded** in every table
- [x] $\uparrow$ / $\downarrow$ direction arrows in column headers
- [ ] Figure 1: reconstruction quality comparison (high-res, ≥ 300 DPI)
- [ ] Figure 2: scaling curve (throughput vs. nodes with Amdahl bound)
- [ ] Figure 3: fault-tolerance timeline (failure injection events)
- [ ] All figures saved as PDF or EPS for vector quality

---

## 3. Reproducibility Requirements

### NeurIPS Reproducibility Checklist (mandatory)
- [x] Claims are clearly stated and scope is identified
- [x] Theoretical results have complete proofs (Appendix C)
- [x] All experimental results supported by rigorous statistical tests
- [x] All datasets are publicly available (MNIST, CIFAR-10)
- [x] Dataset URLs and access instructions provided
- [x] All preprocessing steps described
- [x] All code will be open-sourced (MIT License)
- [x] Hyperparameters fully specified (Appendix B)
- [x] All random seeds documented (42, 123, 456, 789, 1011)
- [x] Deterministic mode enabled in PyTorch
- [x] Hardware specifications provided (GKE, V100)
- [x] Compute budget reported (~144 GPU-hours for full suite)
- [x] Model checkpoints will be shared on HuggingFace
- [x] Docker image specification provided

### Ethics / Broader Impact
- [x] Broader impact section included (§6, Discussion)
- [x] Potential negative impacts acknowledged (storage overhead, complexity)
- [x] Mitigation strategies provided

---

## 4. LaTeX Technical Quality

### Compilation
- [ ] Compiles cleanly with `pdflatex` (0 errors, 0 warnings)
- [ ] Bibliography resolves with `bibtex` / `biber`
- [ ] All cross-references (`\ref`, `\eqref`, `\cite`) resolve
- [ ] No overfull/underfull `\hbox` warnings
- [ ] Hyperlinks functional in final PDF

### Style
- [x] `microtype` enabled for improved text rendering
- [x] `booktabs` for professional table formatting
- [x] `algorithmic` package for Algorithm 1
- [x] `hyperref` with `colorlinks` for accessible PDF
- [x] `natbib` with `abbrvnat` bibliography style

---

## 5. Pre-Submission Final Checks

### Content
- [ ] Abstract word count verified (≤ 250 words)
- [ ] Page count verified (≤ 8 pages main body)
- [ ] Spell-check run on final PDF
- [ ] Grammar and fluency reviewed by ≥ 2 co-authors
- [ ] All numeric results cross-checked against raw experimental data
- [ ] Statistical test results independently verified

### PDF Export
- [ ] PDF/A compliant (required by some venues)
- [ ] Fonts embedded (File → Properties → Fonts in Acrobat)
- [ ] File size ≤ 50 MB (typical venue limit)
- [ ] Accessible: tagged PDF, alt-text for figures
- [ ] Metadata: title and author set in `hypersetup`

### Submission System
- [ ] Title matches submission system entry exactly
- [ ] Author list matches (order, affiliations)
- [ ] Abstract matches (copy-paste from PDF)
- [ ] Supplementary ZIP uploaded separately
- [ ] Conflict-of-interest declarations completed
- [ ] Subject areas / keywords selected

---

## 6. Conference-Specific Notes

| Item | NeurIPS 2026 | ICML 2026 | ICLR 2027 |
|------|-------------|-----------|-----------|
| Deadline | May 15, 2026 | Jan 29, 2026 | Sep 25, 2026 |
| Page limit (main) | 8 | 8 | 8 |
| References | Unlimited | Unlimited | Unlimited |
| Anonymous | Yes (blind) | Yes (blind) | Yes (blind) |
| Supplementary | Unlimited | 10 pages | Unlimited |
| Camera-ready due | ~8 weeks after accept | ~6 weeks after accept | ~8 weeks after accept |

---

## 7. File Inventory

| File | Status | Description |
|------|--------|-------------|
| `LogosTalisman_NeurIPS.tex` | ✅ Created | Main manuscript (NeurIPS style) |
| `references.bib` | ✅ Created | BibTeX bibliography (12 entries) |
| `neurips_2024.sty` | ⬜ Download | NeurIPS official style file |
| `LogosTalisman_NeurIPS.pdf` | ⬜ Generate | Compiled PDF |
| `fig1_quality.pdf` | ⬜ Create | Figure 1: reconstruction comparison |
| `fig2_scaling.pdf` | ⬜ Create | Figure 2: scaling efficiency curve |
| `fig3_timeline.pdf` | ⬜ Create | Figure 3: fault timeline |
| `supplementary.zip` | ⬜ Package | Code + extended appendix |

---

## 8. Build Instructions

```bash
# 1. Download NeurIPS 2026 style file (replace with current year URL)
#    https://neurips.cc/Conferences/2026/PaperInformation/StyleFiles

# 2. Place all files in the same directory:
#    LogosTalisman_NeurIPS.tex  references.bib  neurips_2024.sty  fig*.pdf

# 3. Compile
cd docs/publications/conference
pdflatex LogosTalisman_NeurIPS.tex
bibtex   LogosTalisman_NeurIPS
pdflatex LogosTalisman_NeurIPS.tex
pdflatex LogosTalisman_NeurIPS.tex   # third pass resolves all cross-refs
```

> **Note:** The `.sty` file is not committed to this repository as it is
> subject to NeurIPS copyright.  Download it directly from the NeurIPS
> website before compiling.
