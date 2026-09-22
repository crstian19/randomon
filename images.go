package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// defaultImages is the pool randomon picks from when RANDOMON_IMAGES is unset.
//
// The first entry used to point at randomon itself (randomon.crstian.me), which
// made the service redirect back into its own catch-all handler instead of
// serving the GIF. The asset lives on the R2 bucket behind gif.crstian.me.
var defaultImages = []string{
	"https://gif.crstian.me/chibimon_angry.gif",
	"https://media4.giphy.com/media/v1.Y2lkPTc5MGI3NjExanZzY3RjdHdmNjNyMzZjOWd5bmRiYWdrc2swM2tpbmt0dWIzank1aiZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/lEc7pLgBpipbi/giphy.webp",
	"https://media0.giphy.com/media/v1.Y2lkPTc5MGI3NjExbHk1d205OWI2aWg2ejc4NGt0d3A0YnZ4bzh2cXRja3BqdG9tZncxaCZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/qgkrtsDy4MhLq/giphy.webp",
	"https://media2.giphy.com/media/slVWEctHZKvWU/giphy.gif?cid=ecf05e476ptp3hv2w3xzo3pbiupfyinr0o0u1jmzjhshvq89&rid=giphy.gif&ct=g",
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
