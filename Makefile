.PHONY: help build run test test-unit test-docker lint fmt fmt-check check devc devc-up devc-down devc-restart devc-reset devc-status install dev-skillshare clean

help:
	@echo "Common tasks:"
	@echo "  make build          # build binary"
	@echo "  make run            # run binary help"
	@echo "  make test           # all tests"
	@echo "  make test-unit      # unit tests only"
	@echo "  make test-docker    # tests in docker sandbox"
	@echo "  make lint           # go vet"
	@echo "  make fmt            # format Go files"
	@echo "  make check          # fmt-check + lint + test"
	@echo "  make devc           # start devcontainer + enter shell"
	@echo "  make devc-up        # start devcontainer"
	@echo "  make devc-down      # stop devcontainer"
	@echo "  make devc-restart   # restart devcontainer"
	@echo "  make devc-reset     # full reset (remove volumes)"
	@echo "  make devc-status    # show devcontainer status"
	@echo "  make clean          # remove build artifacts"
	@echo "  make dev-skillshare # cross-compile Linux binary to ../skillshare/bin/"

build:
	mkdir -p bin && go build -o bin/mdproof ./cmd/mdproof

run: build
	./bin/mdproof --help

test: build
	./scripts/test.sh

test-unit:
	./scripts/test.sh --unit

test-docker:
	docker compose -f docker-compose.sandbox.yml --profile offline up --build --abort-on-container-exit --exit-code-from sandbox-offline
	docker compose -f docker-compose.sandbox.yml --profile offline down

lint:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

fmt-check:
	test -z "$$(gofmt -l ./cmd ./internal)"

check: fmt-check lint test

devc:
	./scripts/devc.sh up && ./scripts/devc.sh shell

devc-up:
	./scripts/devc.sh up

devc-down:
	./scripts/devc.sh down

devc-restart:
	./scripts/devc.sh restart

devc-reset:
	./scripts/devc.sh reset

devc-status:
	./scripts/devc.sh status

install:
	go install ./cmd/mdproof

dev-skillshare:
	@mkdir -p ../skillshare/bin
	GOOS=linux GOARCH=$$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') CGO_ENABLED=0 go build -o ../skillshare/bin/mdproof ./cmd/mdproof
	@echo "Installed Linux binary → ../skillshare/bin/mdproof"

clean:
	rm -rf bin coverage.out
