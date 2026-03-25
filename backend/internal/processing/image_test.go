package processing_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/cankledankle/home-planner/internal/processing"
)

// makeJPEG creates a minimal valid JPEG of the given dimensions.
func makeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{200, 100, 50, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("makeJPEG: %v", err)
	}
	return buf.Bytes()
}

// makeOpaquePNG creates a minimal valid PNG with no transparency.
func makeOpaquePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("makeOpaquePNG: %v", err)
	}
	return buf.Bytes()
}

// makeTransparentPNG creates a PNG with at least one transparent pixel.
func makeTransparentPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}
	// Set one pixel fully transparent
	img.Set(0, 0, color.RGBA{0, 0, 0, 0})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("makeTransparentPNG: %v", err)
	}
	return buf.Bytes()
}

func TestProcessWebsiteImage_JPEG(t *testing.T) {
	data := makeJPEG(t, 800, 600)
	result, err := processing.ProcessWebsiteImage(data, "image/jpeg", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ContentType != "image/jpeg" {
		t.Fatalf("expected image/jpeg, got %q", result.ContentType)
	}
	if len(result.Data) == 0 {
		t.Fatal("expected non-empty output data")
	}
}

func TestProcessWebsiteImage_OpaquePNG_Converted(t *testing.T) {
	data := makeOpaquePNG(t, 400, 300)
	result, err := processing.ProcessWebsiteImage(data, "image/png", false)
	if err != nil {
		t.Fatalf("unexpected error for opaque PNG: %v", err)
	}
	if !result.WasConverted {
		t.Fatal("expected WasConverted=true for PNG input")
	}
	if result.ContentType != "image/jpeg" {
		t.Fatalf("expected output content type image/jpeg, got %q", result.ContentType)
	}
}

func TestProcessWebsiteImage_TransparentPNG_Rejected(t *testing.T) {
	data := makeTransparentPNG(t, 400, 300)
	_, err := processing.ProcessWebsiteImage(data, "image/png", false)
	if err == nil {
		t.Fatal("expected error for transparent PNG, got nil")
	}
}

func TestProcessWebsiteImage_OversizedDimensions_Resized(t *testing.T) {
	// Create an image larger than MaxImageDimension (4000px)
	oversize := processing.MaxImageDimension + 100
	data := makeJPEG(t, oversize, oversize)

	result, err := processing.ProcessWebsiteImage(data, "image/jpeg", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Width > processing.MaxImageDimension || result.Height > processing.MaxImageDimension {
		t.Fatalf("expected dimensions <= %d, got %dx%d",
			processing.MaxImageDimension, result.Width, result.Height)
	}
}

func TestProcessWebsiteImage_InvalidData_Error(t *testing.T) {
	_, err := processing.ProcessWebsiteImage([]byte("not an image"), "image/jpeg", false)
	if err == nil {
		t.Fatal("expected error for invalid image data, got nil")
	}
}

func TestProcessWebsiteImage_Poster(t *testing.T) {
	data := makeJPEG(t, 1000, 1500)
	result, err := processing.ProcessWebsiteImage(data, "image/jpeg", true)
	if err != nil {
		t.Fatalf("unexpected error for poster: %v", err)
	}
	if len(result.Data) == 0 {
		t.Fatal("expected non-empty poster output")
	}
}
