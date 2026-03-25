package handlers

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/storage"
	"github.com/gofiber/fiber/v2"
)

type ExportHandler struct {
	store    *db.Store
	r2Client *storage.R2Client
}

func NewExportHandler(store *db.Store, r2Client *storage.R2Client) *ExportHandler {
	return &ExportHandler{store: store, r2Client: r2Client}
}

// Export preset constants - must match frontend contracts.ts
const (
	ExportPresetWPAllImport = "wp_all_import"
	ExportPresetGeneral     = "general"
	ExportPresetMinimal     = "minimal"
	ExportPresetCustom      = "custom"
)

// Valid export presets
var validExportPresets = map[string]bool{
	ExportPresetWPAllImport: true,
	ExportPresetGeneral:     true,
	ExportPresetMinimal:     true,
	ExportPresetCustom:      true,
}

var wpAllImportFields = []string{
	"name", "slug", "type", "style", "beds", "baths", "half_baths",
	"main_sf", "upper_sf", "lower_sf", "porch_deck_sf", "garage_sf",
	"garage_apartment_sf", "unfinished_sf", "heated_sf", "total_sf",
	"notes", "status",
	"render_front", "elevation_front", "elevation_left", "elevation_rear", "elevation_right",
	"floor_plan_main", "floor_plan_upper", "floor_plan_lower", "poster",
}

var generalFields = []string{
	"id", "name", "slug", "type", "style", "status", "beds", "baths", "half_baths",
	"main_sf", "upper_sf", "lower_sf", "porch_deck_sf", "garage_sf",
	"garage_apartment_sf", "unfinished_sf", "heated_sf", "total_sf",
	"notes", "created_at", "updated_at",
}

var allFields = []string{
	"id", "name", "slug", "type", "style", "status", "beds", "baths", "half_baths",
	"main_sf", "upper_sf", "lower_sf", "porch_deck_sf", "garage_sf",
	"garage_apartment_sf", "unfinished_sf", "heated_sf", "total_sf",
	"notes", "created_at", "updated_at", "created_by", "updated_by",
}

// validateExportPreset checks if the preset is valid
func validateExportPreset(preset string) error {
	if !validExportPresets[preset] {
		return fmt.Errorf("invalid preset: %s. Must be one of: wp_all_import, general, minimal, custom", preset)
	}
	return nil
}

func (h *ExportHandler) ExportCSV(c *fiber.Ctx) error {
	preset := c.Query("preset", "general")
	fieldsParam := c.Query("fields")
	idsParam := c.Query("ids")

	// Validate preset parameter
	if err := validateExportPreset(preset); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": err.Error(),
			},
		})
	}

	var fields []string
	switch preset {
	case ExportPresetWPAllImport:
		fields = wpAllImportFields
	case ExportPresetGeneral:
		fields = generalFields
	case ExportPresetCustom:
		if fieldsParam == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": "fields parameter required for custom preset",
				},
			})
		}
		fields = strings.Split(fieldsParam, ",")
	default:
		// This should never happen due to validateExportPreset, but just in case
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid preset. Use: wp_all_import, general, minimal, or custom",
			},
		})
	}

	var planIDs []string
	if idsParam != "" {
		planIDs = strings.Split(idsParam, ",")
	}

	ctx := c.Context()

	// WP All Import preset includes image slots, so we need files
	if preset == ExportPresetWPAllImport {
		plans, err := h.store.GetPlansWithFilesForExport(ctx, planIDs)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to fetch plans for export",
				},
			})
		}

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=\"home-plans.csv\"")

		writer := csv.NewWriter(c.Response().BodyWriter())
		defer writer.Flush()

		if err := writer.Write(fields); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to write CSV header",
				},
			})
		}

		for _, plan := range plans {
			row := make([]string, len(fields))
			for i, field := range fields {
				row[i] = getFieldValueWithFiles(plan, field)
			}
			if err := writer.Write(row); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fiber.Map{
						"code":    "INTERNAL_ERROR",
						"message": "Failed to write CSV row",
					},
				})
			}
		}
		return nil
	}

	// General and custom presets - no files needed
	plans, err := h.store.GetPlansForExport(ctx, planIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch plans for export",
			},
		})
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=\"home-plans.csv\"")

	writer := csv.NewWriter(c.Response().BodyWriter())
	defer writer.Flush()

	if err := writer.Write(fields); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to write CSV header",
			},
		})
	}

	for _, plan := range plans {
		row := make([]string, len(fields))
		for i, field := range fields {
			row[i] = getFieldValue(plan, field)
		}
		if err := writer.Write(row); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to write CSV row",
				},
			})
		}
	}

	return nil
}

func getFieldValue(plan *db.PlanRow, field string) string {
	switch field {
	case "id":
		return plan.ID
	case "name":
		return plan.Name
	case "slug":
		return plan.Slug
	case "type":
		return stringPtrToString(plan.Type)
	case "style":
		return stringPtrToString(plan.Style)
	case "status":
		return plan.Status
	case "beds":
		return intPtrToString(plan.Beds)
	case "baths":
		return intPtrToString(plan.Baths)
	case "half_baths":
		return intPtrToString(plan.HalfBaths)
	case "main_sf":
		return intPtrToString(plan.MainSF)
	case "upper_sf":
		return intPtrToString(plan.UpperSF)
	case "lower_sf":
		return intPtrToString(plan.LowerSF)
	case "porch_deck_sf":
		return intPtrToString(plan.PorchDeckSF)
	case "garage_sf":
		return intPtrToString(plan.GarageSF)
	case "garage_apartment_sf":
		return intPtrToString(plan.GarageApartmentSF)
	case "unfinished_sf":
		return intPtrToString(plan.UnfinishedSF)
	case "heated_sf":
		return intPtrToString(plan.HeatedSF)
	case "total_sf":
		return intPtrToString(plan.TotalSF)
	case "notes":
		return stringPtrToString(plan.Notes)
	case "created_at":
		return plan.CreatedAt.Format("2006-01-02 15:04:05")
	case "updated_at":
		return plan.UpdatedAt.Format("2006-01-02 15:04:05")
	case "created_by":
		return stringPtrToString(plan.CreatedBy)
	case "updated_by":
		return stringPtrToString(plan.UpdatedBy)
	default:
		return ""
	}
}

func getFieldValueWithFiles(plan *db.PlanWithFilesRow, field string) string {
	// Map field names to slot keys
	slotMapping := map[string]string{
		"render_front":     "render-front",
		"elevation_front":  "elevation-front",
		"elevation_left":   "elevation-left",
		"elevation_rear":   "elevation-rear",
		"elevation_right":  "elevation-right",
		"floor_plan_main":  "floor-plan-main",
		"floor_plan_upper": "floor-plan-upper",
		"floor_plan_lower": "floor-plan-lower",
		"poster":           "poster",
	}

	// Check if this is an image slot field
	if slot, ok := slotMapping[field]; ok {
		if filename, exists := plan.Files[slot]; exists && filename != "" {
			return filename
		}
		return ""
	}

	// Otherwise use standard field mapping
	return getFieldValue(plan.PlanRow, field)
}

func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intPtrToString(i *int) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}

func (h *ExportHandler) ExportZIP(c *fiber.Ctx) error {
	idsParam := c.Query("ids")
	categoriesParam := c.Query("categories")

	var planIDs []string
	if idsParam != "" {
		planIDs = strings.Split(idsParam, ",")
	}

	var categories []string
	if categoriesParam != "" {
		categories = strings.Split(categoriesParam, ",")
	} else {
		categories = []string{"website", "reference", "technical", "3d", "other"}
	}

	ctx := c.Context()
	files, err := h.store.GetFilesForExport(ctx, planIDs, categories)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch files for export",
			},
		})
	}

	if len(files) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "No files found for export",
			},
		})
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=\"home-plans-files.zip\"")

	zipWriter := zip.NewWriter(c.Response().BodyWriter())
	defer zipWriter.Close()

	type fileJob struct {
		file     *db.FileExportRow
		planName string
	}

	jobs := make(chan fileJob, len(files))
	results := make(chan struct {
		file     *db.FileExportRow
		planName string
		data     []byte
		err      error
	}, len(files))

	var wg sync.WaitGroup
	workerCount := 5
	if len(files) < workerCount {
		workerCount = len(files)
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if h.r2Client == nil {
					results <- struct {
						file     *db.FileExportRow
						planName string
						data     []byte
						err      error
					}{file: job.file, planName: job.planName, data: nil, err: fmt.Errorf("R2 client not initialized")}
					continue
				}

				url, err := h.r2Client.GetPresignedURL(ctx, job.file.StorageKey, 5)
				if err != nil {
					results <- struct {
						file     *db.FileExportRow
						planName string
						data     []byte
						err      error
					}{file: job.file, planName: job.planName, data: nil, err: err}
					continue
				}

				resp, err := http.Get(url)
				if err != nil {
					results <- struct {
						file     *db.FileExportRow
						planName string
						data     []byte
						err      error
					}{file: job.file, planName: job.planName, data: nil, err: err}
					continue
				}
				defer resp.Body.Close()

				data, err := io.ReadAll(resp.Body)
				if err != nil {
					results <- struct {
						file     *db.FileExportRow
						planName string
						data     []byte
						err      error
					}{file: job.file, planName: job.planName, data: nil, err: err}
					continue
				}

				results <- struct {
					file     *db.FileExportRow
					planName string
					data     []byte
					err      error
				}{file: job.file, planName: job.planName, data: data, err: nil}
			}
		}()
	}

	go func() {
		for _, file := range files {
			jobs <- fileJob{file: file, planName: file.PlanName}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var errors []string
	for result := range results {
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("Failed to fetch %s: %v", result.file.Filename, result.err))
			continue
		}

		safePlanName := sanitizeFileName(result.planName)
		zipPath := fmt.Sprintf("%s/%s/%s", safePlanName, result.file.Category, result.file.Filename)

		writer, err := zipWriter.Create(zipPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to create zip entry for %s: %v", result.file.Filename, err))
			continue
		}

		if _, err := writer.Write(result.data); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to write %s to zip: %v", result.file.Filename, err))
			continue
		}
	}

	if len(errors) > 0 && len(errors) == len(files) {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("All file fetches failed: %s", strings.Join(errors, "; ")),
			},
		})
	}

	return nil
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(name)
}

type ExportPlanInput struct {
	Name              string  `json:"name"`
	Type              *string `json:"type"`
	Style             *string `json:"style"`
	Beds              *int    `json:"beds"`
	Baths             *int    `json:"baths"`
	HalfBaths         *int    `json:"half_baths"`
	MainSF            *int    `json:"main_sf"`
	UpperSF           *int    `json:"upper_sf"`
	LowerSF           *int    `json:"lower_sf"`
	PorchDeckSF       *int    `json:"porch_deck_sf"`
	GarageSF          *int    `json:"garage_sf"`
	GarageApartmentSF *int    `json:"garage_apartment_sf"`
	UnfinishedSF      *int    `json:"unfinished_sf"`
	HeatedSF          *int    `json:"heated_sf"`
	TotalSF           *int    `json:"total_sf"`
	Notes             *string `json:"notes"`
}

type ExportFileInput struct {
	Category   string  `json:"category"`
	Slot       *string `json:"slot"`
	Filename   string  `json:"filename"`
	StorageKey string  `json:"storage_key"`
	FileType   string  `json:"file_type"`
	SizeBytes  int64   `json:"size_bytes"`
}
