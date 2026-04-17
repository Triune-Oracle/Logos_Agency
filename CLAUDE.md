# CLAUDE.md

## Purpose
Logos Agency - Creative and strategic services for branding, marketing, and digital presence management.
Narrative architecture firm translating deep technical innovation and philosophical complexity into market-ready positioning.

## Stack
- **Go 1.24** — Core statistical inference engine, heuristic CSV scanner, circuit breaker, resilience packages
- **Python 3 (asyncio)** — Orchestration layer (`services/orchestrator/`): scroll ingestion, Memory Core, Mind Nexus stubs
- **TypeScript / React** — Frontend web app (`apps/web/`): Vite, Vitest, wouter router, Shadcn UI, Tailwind CSS
- **Node.js** — Workspace tooling, Turbo task runner
- **Docker** — Multi-stage build (golang:1.24 → distroless), docker-compose for local services

## Structure
```
/
├── apps/
│   └── web/              # React/TS frontend (Vite)
├── services/
│   └── orchestrator/     # Python async orchestration service
├── engine/               # Go: Bayesian inference, SIMD, effect size calculations
├── pkg/
│   ├── resilience/       # Go: Circuit breaker pattern
│   ├── fractal/          # Go: Fractal loss functions for VAE training
│   ├── distributed/      # Go: NCCL/RDMA communication wrapper
│   └── k8s/metrics/      # Go: Prometheus/fractal HPA exporter
├── packages/
│   └── tsconfig/         # Shared TypeScript base configuration
├── tests/                # Go unit & benchmark test suite (+ fixtures/)
├── benchmarks/           # Python benchmark scripts
├── config/               # UCX/RDMA transport config
├── docs/                 # Architecture, publications, performance docs
├── k8s/                  # Kubernetes HPA manifests
├── scripts/              # Deployment and ship scripts
├── CODEx/                # Scroll/record payloads for orchestrator
└── .github/workflows/    # CI: Go tests + Node/Turbo build pipeline
```

## Build Commands
```bash
# Go
make build          # compile binary → bin/logos_agency
make test           # run Go unit tests (80% coverage minimum)
make bench          # run Go benchmarks (separate from test)
make cover          # generate coverage.out
make ci             # fmt + test + cover + check-coverage
make docker-build   # build Docker image

# Frontend (apps/web/)
npm install         # install all workspace dependencies (from root)
npm run dev         # start Vite dev server
npm run build       # production build → apps/web/dist/
npm run test        # Vitest unit tests

# Monorepo (TurboRepo)
npx turbo run build        # build all apps/packages in pipeline order
npx turbo run test         # run all tests (Go via make, JS via vitest)
npx turbo run lint         # lint all workspaces

# Python orchestrator
cd services/orchestrator && python -m pytest   # run orchestrator tests
```

## Conventions
- Follow standard coding patterns for the identified stack.
- Maintain consistent naming: camelCase for JS/TS, snake_case for Python, PascalCase for Go exported types.
- Use functional components and hooks for React projects.
- Go packages use `pkg/` for shared libraries, `engine/` for core domain logic.
- Deterministic RNG: default seed 42 in Bayesian inference engine.

## Testing
- Go: `tests/` directory — `go test ./...` or `make test`. Fixtures in `tests/fixtures/`.
- TypeScript: `apps/web/src/__tests__/` — run with `vitest` (not Jest).
- Python: `services/orchestrator/tests/` — run with `pytest`.
- Coverage minimum: 80% for Go (enforced by `make check-coverage`).
- Benchmarks: run via `make bench` only, not included in standard `make test`.

## MCP Connections
- External services: Supabase, Vercel, GitHub.
- Supabase project: `sipaqppmhsshycmbgllq` (see TSCP Protocol below).

## Monorepo Migration
- TurboRepo configuration: `turbo.json` (root)
- Vercel deployment: `vercel.json` (root) — builds `apps/web/dist/`, rewrites `/api/*` to Go backend
- Workspace definition: root `package.json` → `"workspaces": ["apps/*", "packages/*"]`
- CI pipeline: `.github/workflows/ci.yml` runs both Go tests and `turbo run build test lint`

## Protected
- DO NOT MODIFY: `.env`, `prisma/migrations`, `PM2` configs, `Stripe` webhook configs, `*.key`, `*.pem`.

## TSCP Protocol
- Log significant changes to `partyline_entries` table in Supabase project `sipaqppmhsshycmbgllq`.
- Escalate to seansouthwick77@yahoo.com for:
  - `.env` modifications
  - Database migrations
  - Dependency major version upgrades
  - PM2 service restarts
