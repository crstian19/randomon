# randomon — command index

default:
    @just --list

# Install toolchain and git hooks
setup:
    mise install
    lefthook install

# Run locally on :3000
dev:
    go run .

test:
    go test -race ./...

lint:
    golangci-lint run ./...
    golangci-lint fmt --diff

fmt:
    golangci-lint fmt

build:
    go build -trimpath -ldflags="-s -w" -o bin/randomon .

# Build the container image locally (no push)
image:
    KO_DOCKER_REPO=ko.local ko build --bare .

# Build and push to GHCR
publish:
    KO_DOCKER_REPO=ghcr.io/crstian19 ko build --bare --platform=linux/amd64,linux/arm64 .
