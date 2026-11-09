# AGENTS.md

Orientation for agents and humans landing in this repo.

## What this is

`randomon` is a single-purpose HTTP service: every request gets a `302` to a
randomly chosen image URL. It backs `randomon.crstian.me`, which is used as an
`<img src>` in READMEs.

## Commands

All commands live in the `justfile` — run `just` to list them.

| Command | What it does |
|---|---|
| `just setup` | Install the pinned toolchain (mise) and git hooks (lefthook) |
| `just dev` | Run locally on `:3000` |
| `just test` | `go test -race ./...` |
| `just lint` | golangci-lint run + format check |
| `just image` | Build the container image locally with ko |

## Layout

One file per responsibility — the repo is small enough not to need packages.

| File | Responsibility |
|---|---|
| `main.go` | Server lifecycle: config, timeouts, signals, graceful shutdown |
| `handler.go` | Routes, handlers, Prometheus metric |
| `images.go` | The redirect pool and how it is loaded |

## Conventions

- Go stdlib `net/http`, no router or HTTP wrapper — this service has three routes.
- `golangci-lint` v2 with the config in `.golangci.yml`. `//nolint` always carries a written reason.
- Commits: Conventional Commits, in English. Enforced by `lefthook.yml`.
- The image is built by **ko** (distroless, non-root, multi-arch, digest-pinned base).
  There is no Dockerfile on purpose.

## Config

| Env var | Default | Meaning |
|---|---|---|
| `RANDOMON_ADDR` | `:3000` | Public listen address |
| `RANDOMON_METRICS_ADDR` | `:9090` | Internal metrics listen address |
| `LOG_LEVEL` | `INFO` | `DEBUG`, `INFO`, `WARN` or `ERROR` |
| `RANDOMON_IMAGES` | built-in list | Whitespace-separated URLs, overrides the pool. Validated at startup: absolute http(s) only. |

## Endpoints

| Path | Purpose |
|---|---|
| `/` (and anything else) | `302` to a random image |
| `/healthz` | Liveness/readiness probe |
| `/metrics` | Prometheus metrics — **port 9090 only**, never the public port |

## Releases

[releaser-pleaser](https://github.com/apricote/releaser-pleaser) keeps a release PR
open with the next version and `CHANGELOG.md`, computed from the Conventional
Commits on `main`. Merging it tags `vX.Y.Z`, and the tag publishes the image.
`feat` bumps the minor, `fix` the patch, `!` the major; `chore`, `ci`, `docs`
and the rest do not cut a release on their own.

Renovate commits Go module and toolchain updates as `fix(deps)`, so they ship
in the next patch release; CI and tooling updates are `chore(deps)`.

The workflow authenticates as the `crstian19-releaser` GitHub App (secrets
`RELEASER_APP_CLIENT_ID` and `RELEASER_APP_PRIVATE_KEY`): a tag pushed with
`GITHUB_TOKEN` would not trigger CI.

## Deployment

This repo only builds and publishes the container image to GHCR: `main` and
`sha-<commit>` on every push to `main`, and the bare version (`1.2.3`) on every
release. Rolling it out is handled separately by GitOps from a private
infrastructure repo, so there is nothing to deploy from here.
