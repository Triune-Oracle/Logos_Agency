# Logos Agency Website - TODO

## Core Pages & Sections
- [x] Hero Section with brand thesis and CTA
- [x] About/Philosophy Section (Mythological Coherence)
- [x] Services Section (The Three Teams: Architects, Alchemists, Ritualists)
- [x] Go-To-Market Strategy Section (Six Rituals of Traction)
- [x] Target Clientele Section
- [x] Contact/CTA Section
- [x] Footer with links and branding

## Design & Styling
- [x] Choose color palette and typography
- [x] Implement global theme and CSS variables
- [x] Create responsive layout for all screen sizes
- [x] Add visual elements (icons, illustrations, animations)

## Navigation & Routing
- [x] Top navigation bar with logo and menu
- [x] Smooth scrolling between sections
- [ ] Mobile-responsive hamburger menu

## Content & Copy
- [x] Write compelling hero copy
- [x] Develop detailed service descriptions
- [ ] Create case study/testimonial sections
- [x] Write footer content

## Deployment
- [x] Prepare project for deployment (TurboRepo + Vercel scaffolding added)
- [ ] Deploy to Vercel (run `vercel --prod` from repo root after CI passes)
- [ ] Obtain permanent live URL
- [ ] Test all pages and functionality post-deploy

## Polish & QA
- [ ] Mobile-responsive hamburger menu (nav hidden on mobile, needs toggle)
- [ ] Test responsiveness on mobile/tablet/desktop
- [ ] Verify all links and CTAs work
- [ ] Check accessibility and SEO basics (add meta tags, aria labels)
- [ ] Performance optimization (Lighthouse audit)

## TurboRepo Migration
- [x] Fix go.mod/go.sum merge conflict (testify v1.11.1 accepted)
- [x] Update CLAUDE.md with accurate stack, structure, build commands
- [x] Create apps/web/ with Vite + Vitest config and move React source
- [x] Move orchestrator/ → services/orchestrator/
- [x] Create packages/tsconfig/ shared base config
- [x] Update root package.json with workspaces + turbo scripts
- [x] Create turbo.json pipeline definition
- [x] Create vercel.json deployment config
- [x] Update .github/workflows/ci.yml with Node/Turbo steps
- [x] Delete orphaned files (pasted_content_*.txt, stray CSVs, duplicate CI workflows)
- [ ] Install turbo globally or via npx and run `turbo run build` end-to-end
- [ ] Configure Vercel project to use this repo (set VERCEL_PROJECT_ID in secrets)
- [ ] Add TURBO_TOKEN + TURBO_TEAM env vars to GitHub Actions for remote caching

## Test Improvements (Backlog)
- [ ] Add pytest setup for services/orchestrator/ (pytest.ini or pyproject.toml)
- [ ] Write test_supremehead.py, test_memory_core.py, test_mind_nexus.py
- [ ] Add Vitest tests for apps/web/src/App.tsx and Home.tsx
- [ ] Add Go tests for pkg/distributed/nccl_wrapper.go
- [ ] Add Go tests for pkg/fractal/fallback_loss.go
- [ ] Add Go tests for pkg/k8s/metrics/fractal_exporter.go
- [ ] Add main.go HTTP handler tests (health check, CSV endpoint)
- [ ] Consolidate scanner_test.go (root) into tests/heuristic_scanner_complete_test.go
- [ ] Add tests/fixtures/README.md documenting each fixture's purpose
- [ ] Add integration test skeleton between Go API and frontend
