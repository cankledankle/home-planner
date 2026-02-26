package processing

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
)

const (
	MaxWebsiteImageSize  = 5 * 1024 * 1024  // 5MB
	MaxReferenceFileSize = 50 * 1024 * 1024 // 50MB
	MaxImageDimension    = 4000             // 4000px max width/height
	JPEGQuality          = 90
	MaxFilenameLength    = 100
	PosterTargetWidth    = 2400 // 8" @ 300dpi
	PosterTargetHeight   = 3600 // 12" @ 300dpi
)

var (
	ValidImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}

	ValidExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	// Filename pattern: plan-slug--slot-type--view.ext or plan-slug--slot.ext
	// Allowed characters: a-z, 0-9, hyphen, underscore
	ValidFilenamePattern = regexp.MustCompile(`^[a-z0-9_-]+--[a-z0-9_-]+(--[a-z0-9_-]+)?\.(jpg|jpeg|png)$`)
)

// ProcessingResult contains the result of image processing
type ProcessingResult struct {
	Data         []byte
	ContentType  string
	SizeBytes    int64
	Width        int
	Height       int
	WasResized   bool
	WasConverted bool
	OriginalSize int64
	Warnings     []string
}

// ValidateFilename checks if a filename matches the required pattern
func ValidateFilename(filename string) error {
	if len(filename) > MaxFilenameLength {
		return fmt.Errorf("filename exceeds maximum length of %d characters", MaxFilenameLength)
	}

	lowerFilename := strings.ToLower(filename)

	if !ValidFilenamePattern.MatchString(lowerFilename) {
		return fmt.Errorf("filename must match pattern: {plan-slug}--{slot}[--{view}].{ext} (lowercase, no spaces)")
	}

	return nil
}

// StandardizeFilename converts filename to lowercase and validates it
func StandardizeFilename(filename string) (string, error) {
	// Convert to lowercase
	standardized := strings.ToLower(filename)

	// Replace spaces with hyphens
	standardized = strings.ReplaceAll(standardized, " ", "-")

	if err := ValidateFilename(standardized); err != nil {
		return "", err
	}

	return standardized, nil
}

// StandardizeFilenameForSlot generates a standardized filename for a specific plan and slot
func StandardizeFilenameForSlot(originalFilename, planSlug, slot string) string {
	// Use the original extension
	ext := filepath.Ext(originalFilename)

	// Create standardized name: plan-slug--slot.jpg
	standardized := fmt.Sprintf("%s--%s.jpg", planSlug, slot)

	// If original had a different extension, we still use .jpg for output
	_ = ext // Keep for reference

	return standardized
}

// GenerateStorageKey creates a storage key from plan slug and slot
func GenerateStorageKey(planSlug, slot string) string {
	return fmt.Sprintf("plans/%s/website/%s.jpg", planSlug, slot)
}

// DetectTransparency checks if an image has transparent pixels
func DetectTransparency(img image.Image) bool {
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			// RGBA values are in range [0, 65535]
			if a < 65535 {
				return true
			}
		}
	}

	return false
}

// DetectTransparencyFromBytes checks if image data has transparency
func DetectTransparencyFromBytes(data []byte) (bool, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false, fmt.Errorf("failed to decode image: %w", err)
	}

	return DetectTransparency(img), nil
}

// ConvertPNGToJPEG converts a PNG image to JPEG with white background
func ConvertPNGToJPEG(img image.Image) ([]byte, error) {
	bounds := img.Bounds()

	// Create a new image with white background
	whiteBg := image.NewRGBA(bounds)

	// Fill with white
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			whiteBg.Set(x, y, color.White)
		}
	}

	// Draw the original image on top (preserving transparency blending)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Blend with white background
			alpha := float64(a) / 65535.0

			// Convert from 16-bit to 8-bit and blend
			newR := uint8((float64(r)*alpha + 65535.0*(1-alpha)) / 256.0)
			newG := uint8((float64(g)*alpha + 65535.0*(1-alpha)) / 256.0)
			newB := uint8((float64(b)*alpha + 65535.0*(1-alpha)) / 256.0)

			whiteBg.Set(x, y, color.RGBA{newR, newG, newB, 255})
		}
	}

	// Encode as JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, whiteBg, &jpeg.Options{Quality: JPEGQuality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

// ProcessWebsiteImage processes an image for website use
// Returns processed JPEG data and metadata
func ProcessWebsiteImage(data []byte, contentType string, isPoster bool) (*ProcessingResult, error) {
	result := &ProcessingResult{
		OriginalSize: int64(len(data)),
		ContentType:  "image/jpeg",
		Warnings:     []string{},
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	result.Width = bounds.Dx()
	result.Height = bounds.Dy()

	// Handle PNG files
	if format == "png" || contentType == "image/png" {
		hasTransparency := DetectTransparency(img)

		if hasTransparency {
			return nil, fmt.Errorf("PNG image contains transparency and cannot be used for website images")
		}

		// Convert PNG to JPEG
		jpegData, err := ConvertPNGToJPEG(img)
		if err != nil {
			return nil, fmt.Errorf("failed to convert PNG to JPEG: %w", err)
		}

		data = jpegData
		result.WasConverted = true
		result.Warnings = append(result.Warnings, "PNG converted to JPEG with white background")

		// Re-decode the converted JPEG for further processing
		img, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode converted image: %w", err)
		}
	} else if format == "jpeg" || contentType == "image/jpeg" {
		// Re-encode JPEG at specified quality to strip metadata
		var buf bytes.Buffer
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality})
		if err != nil {
			return nil, fmt.Errorf("failed to re-encode JPEG: %w", err)
		}
		data = buf.Bytes()
		result.WasConverted = true
	}

	// Resize if needed
	if isPoster {
		data, err = processPosterImage(img)
		if err != nil {
			return nil, err
		}
	} else {
		data, err = resizeIfNeeded(img)
		if err != nil {
			return nil, err
		}
	}

	if len(data) != int(result.OriginalSize) {
		result.WasResized = true
	}

	// Validate output size
	if int64(len(data)) > MaxWebsiteImageSize {
		return nil, fmt.Errorf("processed image exceeds %dMB limit", MaxWebsiteImageSize/(1024*1024))
	}

	result.Data = data
	result.SizeBytes = int64(len(data))

	// Update dimensions after processing
	finalImg, _, err := image.Decode(bytes.NewReader(data))
	if err == nil {
		bounds := finalImg.Bounds()
		result.Width = bounds.Dx()
		result.Height = bounds.Dy()
	}

	return result, nil
}

// processPosterImage processes a poster image with specific dimensions
func processPosterImage(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Resize if larger than max dimensions
	if width > MaxImageDimension || height > MaxImageDimension {
		img = imaging.Fit(img, MaxImageDimension, MaxImageDimension, imaging.Lanczos)
	}

	// For posters, we want to fit within target dimensions while maintaining aspect ratio
	img = imaging.Fit(img, PosterTargetWidth, PosterTargetHeight, imaging.Lanczos)

	// Encode as JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode poster image: %w", err)
	}

	return buf.Bytes(), nil
}

// resizeIfNeeded resizes an image if it exceeds max dimensions
func resizeIfNeeded(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check if resize is needed
	if width <= MaxImageDimension && height <= MaxImageDimension {
		// Just re-encode to strip metadata
		var buf bytes.Buffer
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}
		return buf.Bytes(), nil
	}

	// Resize using Lanczos resampling for quality
	resized := imaging.Fit(img, MaxImageDimension, MaxImageDimension, imaging.Lanczos)

	// Encode as JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: JPEGQuality})
	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return buf.Bytes(), nil
}

// ProcessReferenceFile validates a reference file without image processing
func ProcessReferenceFile(data []byte, filename string) error {
	if int64(len(data)) > MaxReferenceFileSize {
		return fmt.Errorf("file size exceeds %dMB limit", MaxReferenceFileSize/(1024*1024))
	}

	return nil
}

// ParseFilename extracts plan slug and slot from standardized filename
// Expected format: {plan-slug}--{slot}[--{view}].{ext}
func ParseFilename(filename string) (planSlug, slot, view string, err error) {
	// Remove extension
	ext := strings.ToLower(filepath.Ext(filename))
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Split by --
	parts := strings.Split(nameWithoutExt, "--")

	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("filename must contain at least plan-slug and slot separated by --")
	}

	planSlug = parts[0]
	slot = parts[1]

	if len(parts) >= 3 {
		view = parts[2]
	}

	return planSlug, slot, view, nil
}

// IsValidImageType checks if content type is a valid image type
func IsValidImageType(contentType string) bool {
	return ValidImageTypes[contentType]
}

// GetImageDimensions returns the width and height of an image
func GetImageDimensions(data []byte) (width, height int, err error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}

	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}

// StripMetadata re-encodes an image to remove all EXIF/metadata
func StripMetadata(data []byte, contentType string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	switch contentType {
	case "image/jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality})
	case "image/png":
		err = png.Encode(&buf, img)
	default:
		return nil, fmt.Errorf("unsupported content type for metadata stripping: %s", contentType)
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
