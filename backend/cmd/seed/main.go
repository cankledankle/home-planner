package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlanInfo holds info about each plan from the images
var planData = map[string]struct {
	Name      string
	Type      string
	Style     string
	Beds      int
	Baths     int
	HalfBaths int
	TotalSF   int
	MainSF    int
	UpperSF   int
	LowerSF   int
	Notes     string
}{
	"abilene": {
		Name:      "Abilene",
		Type:      "multi_level",
		Style:     "ranch",
		Beds:      3,
		Baths:     2,
		HalfBaths: 1,
		TotalSF:   2450,
		MainSF:    1650,
		UpperSF:   800,
		Notes:     "Beautiful ranch-style home with open concept living space and spacious master suite.",
	},
	"angler": {
		Name:      "Angler",
		Type:      "multi_level",
		Style:     "cabin",
		Beds:      4,
		Baths:     3,
		HalfBaths: 0,
		TotalSF:   3200,
		MainSF:    2000,
		UpperSF:   1200,
		Notes:     "Stunning cabin design perfect for mountain retreats with vaulted ceilings and large windows.",
	},
	"armadillo-ranch": {
		Name:      "Armadillo Ranch",
		Type:      "single_level",
		Style:     "ranch",
		Beds:      3,
		Baths:     2,
		HalfBaths: 1,
		TotalSF:   2100,
		MainSF:    2100,
		Notes:     "Single-story ranch with open floor plan and covered patio. Great for families.",
	},
	"arrowhead-lodge": {
		Name:      "Arrowhead Lodge",
		Type:      "multi_level",
		Style:     "lodge",
		Beds:      5,
		Baths:     4,
		HalfBaths: 1,
		TotalSF:   4500,
		MainSF:    2500,
		UpperSF:   1500,
		LowerSF:   500,
		Notes:     "Luxurious lodge-style home with rustic charm and modern amenities. Perfect for entertaining.",
	},
}

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

func main() {
	ctx := context.Background()

	// Get database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/homeplanner?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Path to images (relative to project root)
	imageDir := "../src/lib/home-plans"

	// Get admin user ID
	var adminID string
	err = pool.QueryRow(ctx, "SELECT id FROM users WHERE role = 'admin' LIMIT 1").Scan(&adminID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No admin user found: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using admin user: %s\n", adminID)

	// Process each plan
	for planSlug, planInfo := range planData {
		fmt.Printf("\nProcessing plan: %s\n", planInfo.Name)

		// Check if plan exists
		var planID string
		err = pool.QueryRow(ctx,
			"SELECT id FROM plans WHERE slug = $1",
			planSlug,
		).Scan(&planID)

		if err == pgx.ErrNoRows {
			// Create plan
			err = pool.QueryRow(ctx, `
				INSERT INTO plans (name, slug, type, style, status, beds, baths, half_baths, 
					total_sf, main_sf, upper_sf, lower_sf, notes, created_by, updated_by)
				VALUES ($1, $2, $3, $4, 'incomplete', $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
				RETURNING id
			`,
				planInfo.Name,
				planSlug,
				planInfo.Type,
				planInfo.Style,
				planInfo.Beds,
				planInfo.Baths,
				planInfo.HalfBaths,
				planInfo.TotalSF,
				planInfo.MainSF,
				planInfo.UpperSF,
				planInfo.LowerSF,
				planInfo.Notes,
				adminID,
			).Scan(&planID)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create plan %s: %v\n", planInfo.Name, err)
				continue
			}
			fmt.Printf("  Created plan with ID: %s\n", planID)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check plan %s: %v\n", planInfo.Name, err)
			continue
		} else {
			fmt.Printf("  Plan already exists with ID: %s\n", planID)
		}

		// Find and process images for this plan
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

			// Check if file already exists
			var existingID string
			err = pool.QueryRow(ctx,
				"SELECT id FROM files WHERE plan_id = $1 AND category = 'website' AND slot = $2",
				planID, slot,
			).Scan(&existingID)

			if err == nil {
				fmt.Printf("  File already exists for slot %s\n", slot)
				continue
			}

			// Create file record (without actual storage - just database entry)
			// Storage key follows pattern: plans/{slug}/website/{slot}.jpg
			ext := filepath.Ext(filename)
			storageKey := fmt.Sprintf("plans/%s/website/%s%s", planSlug, slot, ext)
			filePath := filepath.Join(imageDir, filename)

			// Get file info
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to stat file %s: %v\n", filename, err)
				continue
			}

			_, err = pool.Exec(ctx, `
				INSERT INTO files (plan_id, category, slot, filename, storage_key, file_type, size_bytes, uploaded_by)
				VALUES ($1, 'website', $2, $3, $4, 'image/jpeg', $5, $6)
			`,
				planID,
				slot,
				filename,
				storageKey,
				fileInfo.Size(),
				adminID,
			)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create file record for %s: %v\n", filename, err)
				continue
			}

			fmt.Printf("  Added file: %s (slot: %s, size: %d bytes)\n", filename, slot, fileInfo.Size())
		}

		// Log activity
		_, err = pool.Exec(ctx, `
			INSERT INTO activity_log (user_id, plan_id, action, detail)
			VALUES ($1, $2, 'plan_created', '{"source": "seed_data"}'::jsonb)
		`, adminID, planID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to log activity for %s: %v\n", planInfo.Name, err)
		}
	}

	fmt.Println("\n✅ Seed data created successfully!")
	fmt.Println("\nNote: File records were created in the database but the actual files need to be uploaded to R2.")
	fmt.Println("The storage_key references: plans/{slug}/website/{slot}.jpg")
}
