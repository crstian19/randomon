package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// defaultImages is the pool randomon picks from when RANDOMON_IMAGES is unset.
//
// Everything is self-hosted on the R2 bucket behind gif.crstian.me: Giphy URLs
// carried a session tracking blob that can rot, and every README view leaked a
// request to a third party.
//
// gif.crstian.me and randomon.crstian.me are sibling subdomains, so a typo here
// is easy to make and turns a redirect into a loop pointing at this service.
// validateImages plus TestDefaultImagesAreAbsoluteAndNotSelfReferential exist
// because that bug already shipped once.
var defaultImages = []string{
	"https://gif.crstian.me/surprised-pikachu.webp",
	"https://gif.crstian.me/pikachu-flotador.webp",
	"https://gif.crstian.me/charmander-risas-hombros.webp",
}

// loadImages returns the redirect pool. RANDOMON_IMAGES overrides the built-in
// list with a whitespace-separated set of URLs, so the pool can be changed
// without rebuilding the binary.
//
// Every entry is validated, because the env override is the one path that can
// reintroduce at runtime the bug this service already shipped once: an absolute
// URL pointing back at randomon turns a redirect into a loop, and a relative
// one produces a Location the browser resolves against randomon itself.
func loadImages() ([]string, error) {
	raw := strings.TrimSpace(os.Getenv("RANDOMON_IMAGES"))
	if raw == "" {
		return validateImages(defaultImages)
	}
	return validateImages(strings.Fields(raw))
}

func validateImages(images []string) ([]string, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("image pool is empty")
	}
	for _, raw := range images {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("image %q: %w", raw, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return nil, fmt.Errorf("image %q: scheme must be http or https, got %q", raw, u.Scheme)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("image %q: missing host (a relative URL would redirect back to randomon)", raw)
		}
	}
	return images, nil
}
