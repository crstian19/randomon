# Randomon

![Docker](https://img.shields.io/badge/docker-ghcr.io-blue?style=for-the-badge&logo=docker&logoColor=white)
![Go Version](https://img.shields.io/github/go-mod/go-version/crstian19/randomon?style=for-the-badge&logo=go&logoColor=white&label=go)
![License](https://img.shields.io/github/license/crstian19/randomon?style=for-the-badge&logo=unlicense&logoColor=white)

<img src="https://gif.crstian.me/randomon-nobg.png" alt="Randomon Repository Image" width="800"/>


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

The latest Docker image is automatically built and available at:
```
ghcr.io/crstian19/randomon:latest
```

Run it with:
```bash
docker run -p 3000:3000 ghcr.io/crstian19/randomon:latest
```

### Building from source

1. Clone the repository:
```bash
git clone https://github.com/crstian19/randomon.git
cd randomon
```

2. Build and run:
```bash
go build -o randomon main.go
./randomon
```

The service will start on port 3000.

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

Currently, the images are hardcoded in the `main.go` file. To customize the image collection:

1. Edit the `images` slice in `main.go`
2. Rebuild the application
3. Deploy your updated version

## Contributing

Contributions are welcome! Feel free to:

- Add more image sources
- Improve the randomization algorithm
- Add configuration options
- Enhance documentation

Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Images are sourced from various providers including [Giphy](https://giphy.com/)
- Built with Go's standard library
- Inspired by simple redirect utilities