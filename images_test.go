package main

import (
	"slices"
	"testing"
)

func TestLoadImagesFromEnv(t *testing.T) {
	t.Setenv("RANDOMON_IMAGES", "  https://example.test/1.gif\n https://example.test/2.gif ")

	got, err := loadImages()
	if err != nil {
		t.Fatalf("loadImages() error = %v", err)
	}
	want := []string{"https://example.test/1.gif", "https://example.test/2.gif"}
	if !slices.Equal(got, want) {
		t.Errorf("loadImages() = %v, want %v", got, want)
	}
}

func TestLoadImagesFallsBackToDefaults(t *testing.T) {
	t.Setenv("RANDOMON_IMAGES", "   ")

	got, err := loadImages()
	if err != nil {
		t.Fatalf("loadImages() error = %v", err)
	}
	if len(got) != len(defaultImages) {
		t.Errorf("loadImages() returned %d images, want the %d built-in ones", len(got), len(defaultImages))
	}
}

// TestLoadImagesRejectsBadOverrides covers the one surface that can reintroduce
// the original bug at runtime without code review.
func TestLoadImagesRejectsBadOverrides(t *testing.T) {
	for name, value := range map[string]string{
		"relative URL":     "/chibimon_angry.gif",
		"scheme-relative":  "//evil.test/x.gif",
		"javascript":       "javascript:alert(1)",
		"data URI":         "data:image/gif;base64,AAAA",
		"one bad in a set": "https://ok.test/a.gif ftp://bad.test/b.gif",
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("RANDOMON_IMAGES", value)
			if _, err := loadImages(); err == nil {
				t.Errorf("loadImages() accepted %q, want an error", value)
			}
		})
	}
}

func TestValidateImagesRejectsEmptyPool(t *testing.T) {
	if _, err := validateImages(nil); err == nil {
		t.Error("validateImages(nil) = nil error, want an error")
	}
}
