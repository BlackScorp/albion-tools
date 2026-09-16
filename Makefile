.DEFAULT_GOAL := help

IMAGE := albion-helper-builder:go1.25.1
GO_CACHE := albion-helper-go-cache
GO_BUILD_CACHE := albion-helper-go-build-cache

.PHONY: help test test-race build-windows verify

help:
	@printf '%s\n' \
	  'Albion Helper – verfügbare Befehle:' \
	  '  make test          Go-Tests im Docker-Container ausführen' \
	  '  make test-race     Go-Tests mit Race Detector ausführen' \
	  '  make build-windows Windows-amd64-Programm nach dist/albion-helper.exe bauen' \
	  '  make verify        Formatierung, go vet und alle Tests ausführen' \
	  '  make help          Diese Übersicht anzeigen'

define run_in_docker
@log=$$(mktemp); \
if docker image inspect $(IMAGE) >/dev/null 2>&1 || \
   docker build --quiet --tag $(IMAGE) . >"$$log" 2>&1; then \
  if \
   docker run --rm --volume "$(CURDIR):/workspace" \
     --volume "$(GO_CACHE):/go/pkg" \
     --volume "$(GO_BUILD_CACHE):/root/.cache/go-build" \
     --workdir /workspace $(IMAGE) \
     /bin/sh -c 'go mod tidy && $(1)' >>"$$log" 2>&1; then \
    printf 'Success\n'; \
  else \
    cat "$$log"; rm -f "$$log"; exit 1; \
  fi; \
else \
  cat "$$log"; rm -f "$$log"; exit 1; \
fi; \
rm -f "$$log"
endef

test:
	$(call run_in_docker,go test ./...)

test-race:
	$(call run_in_docker,go test -race ./...)

build-windows:
	@mkdir -p dist
	$(call run_in_docker,GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildvcs=false -o dist/albion-helper.exe ./cmd/albion-helper)

verify:
	$(call run_in_docker,gofmt -w cmd internal && go vet ./... && go test ./...)
