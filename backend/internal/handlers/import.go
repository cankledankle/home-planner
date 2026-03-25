package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/storage"
	"github.com/gofiber/fiber/v2"
)

type ImportHandler struct {
	store    *db.Store
	r2Client *storage.R2Client
}

func NewImportHandler(store *db.Store, r2Client *storage.R2Client) *ImportHandler {
	return &ImportHandler{store: store, r2Client: r2Client}
}

func (h *ImportHandler) PreviewCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "CSV file is required",
			},
		})
	}

	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to open CSV file",
			},
		})
	}
	defer fileContent.Close()

	reader := csv.NewReader(fileContent)
	reader.FieldsPerRecord = -1

	columns, err := reader.Read()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Failed to read CSV header",
			},
		})
	}

	var previewRows []map[string]string
	rowCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		rowCount++

		if len(previewRows) < 5 {
			row := make(map[string]string)
			for i, col := range columns {
				if i < len(record) {
					row[col] = record[i]
				} else {
					row[col] = ""
				}
			}
			previewRows = append(previewRows, row)
		}
	}

	suggestedMapping := suggestColumnMapping(columns)

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"total_rows":         rowCount,
			"columns":           columns,
			"preview":      previewRows,
			"suggested_mapping": suggestedMapping,
		},
	})
}

func suggestColumnMapping(columns []string) map[string]string {
	mapping := make(map[string]string)
	fieldAliases := map[string][]string{
		"name":                {"name", "plan name", "title", "plan"},
		"type":                {"type", "plan type", "home type"},
		"style":               {"style", "home style", "architecture"},
		"beds":                {"beds", "bedrooms", "bed", "bedroom"},
		"baths":               {"baths", "bathrooms", "bath", "bathroom", "full baths"},
		"half_baths":          {"half baths", "half_baths", "half baths", "1/2 baths", "half bath"},
		"main_sf":             {"main sf", "main_sf", "main level sf", "main sqft", "main square feet"},
		"upper_sf":            {"upper sf", "upper_sf", "upper level sf", "upper sqft", "second floor sf"},
		"lower_sf":            {"lower sf", "lower_sf", "lower level sf", "lower sqft", "basement sf"},
		"porch_deck_sf":       {"porch deck sf", "porch_deck_sf", "porch sf", "deck sf", "outdoor sf"},
		"garage_sf":           {"garage sf", "garage_sf", "garage square feet"},
		"garage_apartment_sf": {"garage apartment sf", "garage_apartment_sf", "garage apt sf", "apartment sf"},
		"unfinished_sf":       {"unfinished sf", "unfinished_sf", "unfinished square feet"},
		"heated_sf":           {"heated sf", "heated_sf", "heated square feet", "heated sqft", "living area"},
		"total_sf":            {"total sf", "total_sf", "total square feet", "total sqft"},
		"notes":               {"notes", "description", "comments"},
	}

	for _, col := range columns {
		colLower := strings.ToLower(strings.TrimSpace(col))
		for field, aliases := range fieldAliases {
			for _, alias := range aliases {
				if colLower == alias {
					mapping[col] = field
					break
				}
			}
		}
	}

	return mapping
}

type ImportCSVRequest struct {
	Mapping json.RawMessage `json:"mapping"`
	Mode    string          `json:"mode"`
}

type ImportResult struct {
	Created int           `json:"created"`
	Updated int           `json:"updated"`
	Skipped int           `json:"skipped"`
	Errors  []ImportError `json:"errors"`
	PlanIDs []string      `json:"plan_ids"`
}

type ImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

func (h *ImportHandler) ImportCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "CSV file is required",
			},
		})
	}

	mappingJSON := c.FormValue("mapping")
	if mappingJSON == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Mapping is required",
			},
		})
	}

	var mapping map[string]string
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid mapping JSON",
			},
		})
	}

	mode := c.FormValue("mode", "upsert")
	// Normalize frontend shorthand values
	if mode == "create" {
		mode = "create_only"
	} else if mode == "update" {
		mode = "update_only"
	}
	if mode != "create_only" && mode != "update_only" && mode != "upsert" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid mode. Use: create_only, update_only, or upsert",
			},
		})
	}

	userID := c.Locals("userID").(string)

	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to open CSV file",
			},
		})
	}
	defer fileContent.Close()

	reader := csv.NewReader(fileContent)
	reader.FieldsPerRecord = -1

	columns, err := reader.Read()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Failed to read CSV header",
			},
		})
	}

	result := ImportResult{
		Created: 0,
		Updated: 0,
		Skipped: 0,
		Errors:  []ImportError{},
		PlanIDs: []string{},
	}

	ctx := c.Context()
	rowNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum,
				Message: fmt.Sprintf("Failed to read row: %v", err),
			})
			rowNum++
			continue
		}
		rowNum++

		rowData := make(map[string]string)
		for i, col := range columns {
			if i < len(record) {
				rowData[col] = record[i]
			} else {
				rowData[col] = ""
			}
		}

		planData := extractPlanData(rowData, mapping)

		if planData.Name == "" {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum - 1,
				Message: "Plan name is required",
			})
			continue
		}

		existingPlan, err := h.store.GetPlanBySlug(ctx, planData.Slug)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum - 1,
				Message: fmt.Sprintf("Database error: %v", err),
			})
			continue
		}

		if existingPlan != nil {
			if mode == "create_only" {
				result.Skipped++
				continue
			}

			updateInput := db.UpdatePlanInput{
				Name:              planData.Name,
				Type:              planData.Type,
				Style:             planData.Style,
				Beds:              planData.Beds,
				Baths:             planData.Baths,
				HalfBaths:         planData.HalfBaths,
				MainSF:            planData.MainSF,
				UpperSF:           planData.UpperSF,
				LowerSF:           planData.LowerSF,
				PorchDeckSF:       planData.PorchDeckSF,
				GarageSF:          planData.GarageSF,
				GarageApartmentSF: planData.GarageApartmentSF,
				UnfinishedSF:      planData.UnfinishedSF,
				HeatedSF:          planData.HeatedSF,
				TotalSF:           planData.TotalSF,
				Notes:             planData.Notes,
				UpdatedBy:         userID,
			}

			_, err := h.store.UpdatePlan(ctx, existingPlan.ID, updateInput)
			if err != nil {
				result.Errors = append(result.Errors, ImportError{
					Row:     rowNum - 1,
					Message: fmt.Sprintf("Failed to update plan: %v", err),
				})
				continue
			}
			result.Updated++
			h.linkPlanImages(ctx, existingPlan.ID, existingPlan.Slug, planData.Images, userID)
		} else {
			if mode == "update_only" {
				result.Skipped++
				continue
			}

			createInput := db.CreatePlanInput{
				Name:              planData.Name,
				Type:              planData.Type,
				Style:             planData.Style,
				Beds:              planData.Beds,
				Baths:             planData.Baths,
				HalfBaths:         planData.HalfBaths,
				MainSF:            planData.MainSF,
				UpperSF:           planData.UpperSF,
				LowerSF:           planData.LowerSF,
				PorchDeckSF:       planData.PorchDeckSF,
				GarageSF:          planData.GarageSF,
				GarageApartmentSF: planData.GarageApartmentSF,
				UnfinishedSF:      planData.UnfinishedSF,
				HeatedSF:          planData.HeatedSF,
				TotalSF:           planData.TotalSF,
				Notes:             planData.Notes,
				CreatedBy:         userID,
			}

			createdPlan, err := h.store.CreatePlan(ctx, createInput)
			if err != nil {
				result.Errors = append(result.Errors, ImportError{
					Row:     rowNum - 1,
					Message: fmt.Sprintf("Failed to create plan: %v", err),
				})
				continue
			}
			result.Created++
			result.PlanIDs = append(result.PlanIDs, createdPlan.ID)
			h.linkPlanImages(ctx, createdPlan.ID, createdPlan.Slug, planData.Images, userID)
		}
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

type ImportPlanData struct {
	Name              string
	Slug              string
	Type              *string
	Style             *string
	Beds              *int
	Baths             *int
	HalfBaths         *int
	MainSF            *int
	UpperSF           *int
	LowerSF           *int
	PorchDeckSF       *int
	GarageSF          *int
	GarageApartmentSF *int
	UnfinishedSF      *int
	HeatedSF          *int
	TotalSF           *int
	Notes             *string
	Images            map[string]string // slot → csv filename value
}

var imgFieldToSlot = map[string]string{
	"img_render_front":    "render-front",
	"img_elevation_front": "elevation-front",
	"img_elevation_left":  "elevation-left",
	"img_elevation_rear":  "elevation-rear",
	"img_elevation_right": "elevation-right",
	"img_floor_plan_main": "floor-plan-main",
	"img_floor_plan_upper": "floor-plan-upper",
	"img_floor_plan_lower": "floor-plan-lower",
	"img_poster":          "poster",
}

func extractPlanData(rowData map[string]string, mapping map[string]string) ImportPlanData {
	data := ImportPlanData{Images: make(map[string]string)}

	for col, field := range mapping {
		value := strings.TrimSpace(rowData[col])
		switch field {
		case "name":
			data.Name = value
			if data.Name != "" {
				data.Slug = generateSlug(data.Name)
			}
		case "type":
			if value != "" {
				data.Type = &value
			}
		case "style":
			if value != "" {
				data.Style = &value
			}
		case "beds":
			if v, err := strconv.Atoi(value); err == nil {
				data.Beds = &v
			}
		case "baths":
			if v, err := strconv.Atoi(value); err == nil {
				data.Baths = &v
			}
		case "half_baths":
			if v, err := strconv.Atoi(value); err == nil {
				data.HalfBaths = &v
			}
		case "main_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.MainSF = &v
			}
		case "upper_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.UpperSF = &v
			}
		case "lower_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.LowerSF = &v
			}
		case "porch_deck_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.PorchDeckSF = &v
			}
		case "garage_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.GarageSF = &v
			}
		case "garage_apartment_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.GarageApartmentSF = &v
			}
		case "unfinished_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.UnfinishedSF = &v
			}
		case "heated_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.HeatedSF = &v
			}
		case "total_sf":
			if v, err := strconv.Atoi(value); err == nil {
				data.TotalSF = &v
			}
		case "notes":
			if value != "" {
				data.Notes = &value
			}
		default:
			if slot, ok := imgFieldToSlot[field]; ok && value != "" {
				data.Images[slot] = value
			}
		}
	}

	return data
}

func (h *ImportHandler) linkPlanImages(ctx context.Context, planID, planSlug string, images map[string]string, userID string) {
	if len(images) == 0 {
		return
	}
	linked := 0
	for slot, csvFilename := range images {
		if csvFilename == "" {
			continue
		}
		storageKey := fmt.Sprintf("plans/%s/website/%s.jpg", planSlug, slot)
		filename := fmt.Sprintf("%s--%s.jpg", planSlug, slot)
		_, err := h.store.UpsertWebsiteFile(ctx, planID, slot, filename, storageKey, "image/jpeg", 0, userID)
		if err != nil {
			fmt.Printf("Failed to link image for plan %s slot %s: %v\n", planSlug, slot, err)
			continue
		}
		linked++
	}
	if linked > 0 {
		h.store.RecalculatePlanStatus(ctx, planID)
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (h *ImportHandler) GetRecentImports(c *fiber.Ctx) error {
	ctx := c.Context()

	plans, err := h.store.GetRecentlyImportedPlans(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to fetch recent imports: %v", err),
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": plans,
	})
}
