# Randomon

![Docker](https://img.shields.io/badge/docker-ghcr.io-blue?style=for-the-badge&logo=docker&logoColor=white)
![Go Version](https://img.shields.io/github/go-mod/go-version/crstian19/randomon?style=for-the-badge&logo=go&logoColor=white&label=go)
![License](https://img.shields.io/github/license/crstian19/randomon?style=for-the-badge&logo=unlicense&logoColor=white)

<img src="https://gif.crstian.me/randomon-nobg.png" alt="Randomon Repository Image" width="400"/>


A simple HTTP service that redirects to random images.

## What it does

Randomon is a lightweight web service that, when accessed, redirects you to a random image from a predefined collection. It's perfect for:

- Adding variety to your automated workflows
- Testing redirect behavior
- Simple random content delivery
- Fun experiments with HTTP services

## How it works

1. Make an HTTP request to the service
2. The service randomly selects an image URL from its collection
3. Returns an HTTP 302 redirect to the selected image
4. Images are served with cache-control headers to prevent caching

## Deployment

### Using Docker (Recommended)

The image is built by [ko](https://ko.build) on every push to `main` and published to GHCR:

```
ghcr.io/crstian19/randomon:main          # moving tag, follows main
ghcr.io/crstian19/randomon:sha-<commit>  # immutable, pin this in production
```

Run it with:
```bash
docker run -p 3000:3000 ghcr.io/crstian19/randomon:main
```

The image is distroless, runs as non-root (uid 65532) and ships for both
`linux/amd64` and `linux/arm64`.

### Building from source

1. Clone the repository:
```bash
git clone https://github.com/crstian19/randomon.git
cd randomon
```

2. Build and run:
```bash
go build -o randomon .
./randomon
```

The service will start on port 3000. With [just](https://github.com/casey/just)
installed, `just dev` does the same thing.

## Usage examples

### Simple curl

```bash
curl -I http://localhost:3000
```

This will return an HTTP 302 response with a `Location` header pointing to a random image.

### In HTML

```html
<img src="http://localhost:3000" alt="Random image" />
```

Each page load will show a different random image.

### In Go code

```go
resp, err := http.Get("http://localhost:3000")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

// Follow the redirect
finalURL := resp.Request.URL
fmt.Println("Random image URL:", finalURL)
```

## Configuration

All configuration is done through environment variables.

| Variable | Default | Description |
|---|---|---|
| `RANDOMON_ADDR` | `:3000` | Address the public HTTP server listens on |
| `RANDOMON_METRICS_ADDR` | `:9090` | Address the internal metrics server listens on |
| `LOG_LEVEL` | `INFO` | Log verbosity: `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `RANDOMON_IMAGES` | built-in list | Whitespace-separated image URLs. Overrides the built-in pool without rebuilding. Each entry must be an absolute `http`/`https` URL — the service refuses to start otherwise. |

```bash
RANDOMON_IMAGES="https://example.com/a.gif https://example.com/b.gif" ./randomon
```

To change the built-in pool instead, edit `defaultImages` in `images.go` and rebuild.

## Endpoints

| Path | Description |
|---|---|
| `/` | `302` redirect to a random image (any other path behaves the same) |
| `/healthz` | Liveness/readiness probe, returns `200 ok` |
| `/metrics` | Prometheus metrics, including `randomon_redirects_total`. Served on `METRICS_ADDR` (`:9090`), **not** on the public port. |

## Contributing

Contributions are welcome! Feel free to:

- Add more image sources
- Improve the randomization algorithm
- Add configuration options
- Enhance documentation

Please open an issue or submit a pull request.

Before committing, run `just setup` once to install the git hooks. `just lint`
and `just test` are what CI runs.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Images are sourced from various providers including [Giphy](https://giphy.com/)
- Built with Go's standard library
- Inspired by simple redirect utilities