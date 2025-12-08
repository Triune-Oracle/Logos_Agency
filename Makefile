# Makefile - Logos_Agency
BINARY := logos_agency
PKG := ./...
GO := go
COVER_MIN := 80.0

.PHONY: all build test bench cover check-coverage ci fmt lint docker-build docker-run clean

all: build

build:
	@echo ">> building binary..."
	$(GO) build -o bin/$(BINARY) .

test:
	@echo ">> running unit tests..."
	$(GO) test $(PKG)

bench:
	@echo ">> running benchmarks..."
	$(GO) test -bench=. -benchmem $(PKG)

cover:
	@echo ">> generating coverage..."
	$(GO) test -coverprofile=coverage.out $(PKG)
	@echo "coverage.out generated"

check-coverage: cover
	@echo ">> verifying coverage threshold ($(COVER_MIN)%)..."
	@COVER=$$(go tool cover -func=coverage.out | awk '/total:/ {print $$3}' | sed 's/%//'); \
	echo "Observed coverage: $$COVER%"; \
	python3 - <<PY
import sys
try:
    cov = float("$$COVER")
except:
    print("Could not parse coverage")
    sys.exit(2)
if cov < float("$(COVER_MIN)"):
    print(f"Coverage check failed: {cov:.2f}% < {float('$(COVER_MIN)'):.2f}%")
    sys.exit(1)
print(f"Coverage OK: {cov:.2f}% >= {float('$(COVER_MIN)'):.2f}%")
PY

ci: fmt test cover check-coverage
	@echo ">> CI checks passed."

fmt:
	@echo ">> gofmt"
	$(GO) fmt ./...

lint:
	@echo ">> go vet"
	$(GO) vet ./...

docker-build:
	docker build -t triumvirate/logos_agency:latest .

docker-run:
	docker run --rm --name logos_agency_run -p 8080:8080 triumvirate/logos_agency:latest

clean:
	-rm -rf bin/
	-rm -f coverage.out
