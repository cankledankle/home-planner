package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ImageSlotMap maps image patterns to slot names
var slotMap = map[string]string{
	"render--front":     "render-front",
	"elevation--front":  "elevation-front",
	"elevation--left":   "elevation-left",
	"elevation--rear":   "elevation-rear",
	"elevation--right":  "elevation-right",
	"floor-plan--main":  "floor-plan-main",
	"floor-plan--upper": "floor-plan-upper",
	"floor-plan--lower": "floor-plan-lower",
	"poster":            "poster",
}

// Plan slugs to process
var planSlugs = []string{"abilene", "angler", "armadillo-ranch", "arrowhead-lodge"}

func main() {
	ctx := context.Background()

	// Get R2 credentials from env
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")

	if accountID == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		fmt.Println("Error: R2 credentials not configured")
		os.Exit(1)
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
				Source:            aws.EndpointSourceCustom,
			}, nil
		})),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load AWS config: %v\n", err)
		os.Exit(1)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Path to images (relative to project root)
	imageDir := "../src/lib/home-plans"

	// Process each plan
	for _, planSlug := range planSlugs {
		fmt.Printf("\nProcessing plan: %s\n", planSlug)

		// Find and upload images for this plan
		entries, err := os.ReadDir(imageDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read image directory: %v\n", err)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			filename := entry.Name()
			if !strings.HasPrefix(filename, planSlug+"--") {
				continue
			}

			// Parse filename: plan-slug--slot-name.jpg
			parts := strings.Split(filename, "--")
			if len(parts) < 2 {
				continue
			}

			// Get slot from remaining parts
			slotPart := strings.TrimSuffix(strings.Join(parts[1:], "--"), filepath.Ext(filename))
			slot, ok := slotMap[slotPart]
			if !ok {
				fmt.Printf("  Unknown slot: %s\n", slotPart)
				continue
			}

			// Read file
			filePath := filepath.Join(imageDir, filename)
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Failed to read file %s: %v\n", filename, err)
				continue
			}

			// Upload to R2
			ext := filepath.Ext(filename)
			storageKey := fmt.Sprintf("plans/%s/website/%s%s", planSlug, slot, ext)
			contentType := "image/jpeg"
			if ext == ".png" {
				contentType = "image/png"
			}

			_, err = client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      aws.String(bucketName),
				Key:         aws.String(storageKey),
				Body:        strings.NewReader(string(fileData)),
				ContentType: aws.String(contentType),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Failed to upload %s: %v\n", filename, err)
				continue
			}

			fmt.Printf("  Uploaded: %s -> %s\n", filename, storageKey)
		}
	}

	fmt.Println("\n✅ All files uploaded to R2!")
}
