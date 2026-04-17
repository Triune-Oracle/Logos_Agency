# Conference Paper — LogosTalisman

**"LogosTalisman: Fractal-Constrained Variational Autoencoders for Distributed AI Systems"**

| Venue | Track | Deadline |
|-------|-------|----------|
| NeurIPS 2026 | Machine Learning Systems | May 15, 2026 |
| ICML 2026 | Deep Learning Applications | Jan 29, 2026 |
| ICLR 2027 | Representation Learning | Sep 25, 2026 |

---

## Files

| File | Description |
|------|-------------|
| `LogosTalisman_NeurIPS.tex` | Main LaTeX manuscript (NeurIPS 2026 format, 8 pages) |
| `references.bib` | BibTeX bibliography (12 references) |
| `camera_ready_checklist.md` | NeurIPS / ICML / ICLR camera-ready checklist |
| `LogosTalisman_Conference_Paper_NeurIPS.md` | Human-readable Markdown draft (source material) |

---

## Build Instructions

> **Prerequisites:** A standard TeX distribution (TeX Live 2023+ or MiKTeX) with
> `pdflatex` and `bibtex`.

```bash
# 1. Download the NeurIPS 2026 official style file
#    https://neurips.cc/Conferences/2026/PaperInformation/StyleFiles
#    Save as:  docs/publications/conference/neurips_2024.sty

# 2. Compile (three-pass for correct cross-references)
cd docs/publications/conference
pdflatex LogosTalisman_NeurIPS.tex
bibtex   LogosTalisman_NeurIPS
pdflatex LogosTalisman_NeurIPS.tex
pdflatex LogosTalisman_NeurIPS.tex
```

The final PDF is `LogosTalisman_NeurIPS.pdf`.

---

## Paper Structure

| Section | Pages | Content |
|---------|-------|---------|
| Abstract | 0.25 | 250-word summary |
| 1. Introduction | ~1.0 | Problem, insight, 4 contributions |
| 2. Related Work | ~0.75 | VAE variants, fractal DL, distributed training |
| 3. Methodology | ~1.5 | Equations 1-3, Algorithm 1, circuit breaker |
| 4. Experimental Setup | ~0.75 | Protocols, datasets, hardware |
| 5. Results | ~2.5 | 6 tables, statistical significance |
| 6. Discussion | ~0.5 | Why fractals work, limitations, impact |
| 7. Conclusion | ~0.25 | Summary + reproducibility pointer |
| References | (unlimited) | 12 entries via BibTeX |
| Appendix A–D | (unlimited) | Architecture, config, stats, chaos scripts |

---

## Key Results

- **28% PSNR improvement** over baseline VAE (p < 0.001, Cohen's d > 1.7)
- **100% training continuity** under node crashes / network partitions
- **89% scaling efficiency** on 64-node Kubernetes (86.6% measured vs. 87.2% predicted)
- **94% faster recovery** (86.3 s → 4.8 s)

---

## Status

See [`camera_ready_checklist.md`](camera_ready_checklist.md) for the full
camera-ready checklist and remaining action items (figure generation, PDF
export, spell-check, submission system upload).
