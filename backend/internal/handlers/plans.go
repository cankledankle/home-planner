package handlers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/models"
	"github.com/cankledankle/home-planner/internal/processing"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreatePlanRequest struct {
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

type UpdatePlanRequest struct {
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

type DuplicatePlanRequest struct {
	Name string `json:"name"`
}

type PlanHandler struct{}

func NewPlanHandler() *PlanHandler {
	return &PlanHandler{}
}

func (h *PlanHandler) List(c *fiber.Ctx) error {
	filters := db.PlanListFilters{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Type:   c.Query("type"),
		Style:  c.Query("style"),
		Sort:   c.Query("sort"),
		Order:  c.Query("order"),
	}

	if bedsMin := c.Query("beds_min"); bedsMin != "" {
		if val, err := strconv.Atoi(bedsMin); err == nil {
			filters.BedsMin = &val
		}
	}
	if bedsMax := c.Query("beds_max"); bedsMax != "" {
		if val, err := strconv.Atoi(bedsMax); err == nil {
			filters.BedsMax = &val
		}
	}
	if bathsMin := c.Query("baths_min"); bathsMin != "" {
		if val, err := strconv.Atoi(bathsMin); err == nil {
			filters.BathsMin = &val
		}
	}
	if bathsMax := c.Query("baths_max"); bathsMax != "" {
		if val, err := strconv.Atoi(bathsMax); err == nil {
			filters.BathsMax = &val
		}
	}
	if heatedSFMin := c.Query("heated_sf_min"); heatedSFMin != "" {
		if val, err := strconv.Atoi(heatedSFMin); err == nil {
			filters.HeatedSFMin = &val
		}
	}
	if heatedSFMax := c.Query("heated_sf_max"); heatedSFMax != "" {
		if val, err := strconv.Atoi(heatedSFMax); err == nil {
			filters.HeatedSFMax = &val
		}
	}
	filters.MissingSlot = c.Query("missing_slot")

	if page := c.Query("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil && val > 0 {
			filters.Page = val
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 && val <= 100 {
			filters.Limit = val
		}
	}

	result, err := db.ListPlans(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch plans",
			},
		})
	}

	plans := make([]models.PlanListResponse, len(result.Plans))
	for i, p := range result.Plans {
		planUUID, _ := uuid.Parse(p.ID)
		plans[i] = models.PlanListResponse{
			ID:        planUUID,
			Name:      p.Name,
			Slug:      p.Slug,
			Type:      p.Type,
			Style:     p.Style,
			Status:    p.Status,
			Beds:      p.Beds,
			Baths:     p.Baths,
			HalfBaths: p.HalfBaths,
			HeatedSF:  p.HeatedSF,
			TotalSF:   p.TotalSF,
			UpdatedAt: p.UpdatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"data": plans,
		"meta": fiber.Map{
			"page":        result.Page,
			"limit":       result.Limit,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
	})
}

func (h *PlanHandler) Get(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	plan, err := db.GetPlanByID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch plan",
			},
		})
	}

	if plan == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "Plan not found",
			},
		})
	}

	files, err := db.GetFilesByPlanID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch plan files",
			},
		})
	}

	planUUID, _ := uuid.Parse(plan.ID)
	createdByUUID := parseOptionalUUID(plan.CreatedBy)
	updatedByUUID := parseOptionalUUID(plan.UpdatedBy)

	response := fiber.Map{
		"id":                  planUUID,
		"name":                plan.Name,
		"slug":                plan.Slug,
		"type":                plan.Type,
		"style":               plan.Style,
		"status":              plan.Status,
		"beds":                plan.Beds,
		"baths":               plan.Baths,
		"half_baths":          plan.HalfBaths,
		"main_sf":             plan.MainSF,
		"upper_sf":            plan.UpperSF,
		"lower_sf":            plan.LowerSF,
		"porch_deck_sf":       plan.PorchDeckSF,
		"garage_sf":           plan.GarageSF,
		"garage_apartment_sf": plan.GarageApartmentSF,
		"unfinished_sf":       plan.UnfinishedSF,
		"heated_sf":           plan.HeatedSF,
		"total_sf":            plan.TotalSF,
		"notes":               plan.Notes,
		"created_at":          plan.CreatedAt,
		"updated_at":          plan.UpdatedAt,
		"created_by":          createdByUUID,
		"updated_by":          updatedByUUID,
		"files":               formatFilesResponse(files),
	}

	return c.JSON(fiber.Map{"data": response})
}

func (h *PlanHandler) Create(c *fiber.Ctx) error {
	var req CreatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan name is required",
			},
		})
	}

	userID := c.Locals("userID").(string)

	input := db.CreatePlanInput{
		Name:              req.Name,
		Type:              req.Type,
		Style:             req.Style,
		Beds:              req.Beds,
		Baths:             req.Baths,
		HalfBaths:         req.HalfBaths,
		MainSF:            req.MainSF,
		UpperSF:           req.UpperSF,
		LowerSF:           req.LowerSF,
		PorchDeckSF:       req.PorchDeckSF,
		GarageSF:          req.GarageSF,
		GarageApartmentSF: req.GarageApartmentSF,
		UnfinishedSF:      req.UnfinishedSF,
		HeatedSF:          req.HeatedSF,
		TotalSF:           req.TotalSF,
		Notes:             req.Notes,
		CreatedBy:         userID,
	}

	plan, err := db.CreatePlan(c.Context(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create plan",
			},
		})
	}

	userUUID, _ := uuid.Parse(userID)
	db.LogActivity(c.Context(), &userUUID, strPtr(plan.ID), "plan.created", map[string]interface{}{
		"name": plan.Name,
	})

	planUUID, _ := uuid.Parse(plan.ID)
	createdByUUID := parseOptionalUUID(plan.CreatedBy)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":                  planUUID,
			"name":                plan.Name,
			"slug":                plan.Slug,
			"type":                plan.Type,
			"style":               plan.Style,
			"status":              plan.Status,
			"beds":                plan.Beds,
			"baths":               plan.Baths,
			"half_baths":          plan.HalfBaths,
			"main_sf":             plan.MainSF,
			"upper_sf":            plan.UpperSF,
			"lower_sf":            plan.LowerSF,
			"porch_deck_sf":       plan.PorchDeckSF,
			"garage_sf":           plan.GarageSF,
			"garage_apartment_sf": plan.GarageApartmentSF,
			"unfinished_sf":       plan.UnfinishedSF,
			"heated_sf":           plan.HeatedSF,
			"total_sf":            plan.TotalSF,
			"notes":               plan.Notes,
			"created_at":          plan.CreatedAt,
			"updated_at":          plan.UpdatedAt,
			"created_by":          createdByUUID,
		},
	})
}

func (h *PlanHandler) Update(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	var req UpdatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan name is required",
			},
		})
	}

	userID := c.Locals("userID").(string)

	input := db.UpdatePlanInput{
		Name:              req.Name,
		Type:              req.Type,
		Style:             req.Style,
		Beds:              req.Beds,
		Baths:             req.Baths,
		HalfBaths:         req.HalfBaths,
		MainSF:            req.MainSF,
		UpperSF:           req.UpperSF,
		LowerSF:           req.LowerSF,
		PorchDeckSF:       req.PorchDeckSF,
		GarageSF:          req.GarageSF,
		GarageApartmentSF: req.GarageApartmentSF,
		UnfinishedSF:      req.UnfinishedSF,
		HeatedSF:          req.HeatedSF,
		TotalSF:           req.TotalSF,
		Notes:             req.Notes,
		UpdatedBy:         userID,
	}

	plan, err := db.UpdatePlan(c.Context(), planID, input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update plan",
			},
		})
	}

	if plan == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "Plan not found",
			},
		})
	}

	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(plan.ID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "plan.updated", map[string]interface{}{
		"name": plan.Name,
	})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":                  planUUID,
			"name":                plan.Name,
			"slug":                plan.Slug,
			"type":                plan.Type,
			"style":               plan.Style,
			"status":              plan.Status,
			"beds":                plan.Beds,
			"baths":               plan.Baths,
			"half_baths":          plan.HalfBaths,
			"main_sf":             plan.MainSF,
			"upper_sf":            plan.UpperSF,
			"lower_sf":            plan.LowerSF,
			"porch_deck_sf":       plan.PorchDeckSF,
			"garage_sf":           plan.GarageSF,
			"garage_apartment_sf": plan.GarageApartmentSF,
			"unfinished_sf":       plan.UnfinishedSF,
			"heated_sf":           plan.HeatedSF,
			"total_sf":            plan.TotalSF,
			"notes":               plan.Notes,
			"updated_at":          plan.UpdatedAt,
		},
	})
}

func (h *PlanHandler) Delete(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	err := db.SoftDeletePlan(c.Context(), planID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "Plan not found",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to delete plan",
			},
		})
	}

	userID := c.Locals("userID").(string)
	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(planID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "plan.deleted", map[string]interface{}{})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Plan deleted",
		},
	})
}

func (h *PlanHandler) Restore(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	err := db.RestorePlan(c.Context(), planID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "Plan not found",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to restore plan",
			},
		})
	}

	userID := c.Locals("userID").(string)
	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(planID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "plan.restored", map[string]interface{}{})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Plan restored",
		},
	})
}

func (h *PlanHandler) Duplicate(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	var req DuplicatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "New plan name is required",
			},
		})
	}

	userID := c.Locals("userID").(string)

	plan, err := db.DuplicatePlan(c.Context(), planID, req.Name, userID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "Source plan not found",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to duplicate plan",
			},
		})
	}

	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(plan.ID)
	sourceUUID, _ := uuid.Parse(planID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "plan.duplicated", map[string]interface{}{
		"source_plan_id": sourceUUID,
		"new_name":       plan.Name,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":                  planUUID,
			"name":                plan.Name,
			"slug":                plan.Slug,
			"type":                plan.Type,
			"style":               plan.Style,
			"status":              plan.Status,
			"beds":                plan.Beds,
			"baths":               plan.Baths,
			"half_baths":          plan.HalfBaths,
			"main_sf":             plan.MainSF,
			"upper_sf":            plan.UpperSF,
			"lower_sf":            plan.LowerSF,
			"porch_deck_sf":       plan.PorchDeckSF,
			"garage_sf":           plan.GarageSF,
			"garage_apartment_sf": plan.GarageApartmentSF,
			"unfinished_sf":       plan.UnfinishedSF,
			"heated_sf":           plan.HeatedSF,
			"total_sf":            plan.TotalSF,
			"notes":               plan.Notes,
			"created_at":          plan.CreatedAt,
			"updated_at":          plan.UpdatedAt,
		},
	})
}

func (h *PlanHandler) Flag(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	err := db.FlagPlan(c.Context(), planID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "Plan not found",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to flag plan",
			},
		})
	}

	userID := c.Locals("userID").(string)
	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(planID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "plan.flagged", map[string]interface{}{})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Plan flagged",
		},
	})
}

func (h *PlanHandler) Unflag(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	err := db.UnflagPlan(c.Context(), planID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "Plan not found or not flagged",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to unflag plan",
			},
		})
	}

	userID := c.Locals("userID").(string)
	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(planID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "plan.unflagged", map[string]interface{}{})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Plan unflagged",
		},
	})
}

func parseOptionalUUID(s *string) *uuid.UUID {
	if s == nil {
		return nil
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &u
}

func strPtr(s string) *uuid.UUID {
	u, _ := uuid.Parse(s)
	return &u
}

func formatFilesResponse(files map[string][]db.FileWithUploader) fiber.Map {
	websiteSlots := fiber.Map{
		"render-front":     nil,
		"elevation-front":  nil,
		"elevation-left":   nil,
		"elevation-rear":   nil,
		"elevation-right":  nil,
		"floor-plan-main":  nil,
		"floor-plan-upper": nil,
		"floor-plan-lower": nil,
		"poster":           nil,
	}

	for _, file := range files["website"] {
		if file.Slot != nil {
			f := file
			websiteSlots[*file.Slot] = formatFileResponse(&f)
		}
	}

	return fiber.Map{
		"website":   websiteSlots,
		"reference": formatFileList(files["reference"]),
		"technical": formatFileList(files["technical"]),
		"3d":        formatFileList(files["3d"]),
		"other":     formatFileList(files["other"]),
	}
}

func formatFileResponse(f *db.FileWithUploader) fiber.Map {
	if f == nil {
		return nil
	}

	fileUUID, _ := uuid.Parse(f.ID)
	planUUID, _ := uuid.Parse(f.PlanID)

	response := fiber.Map{
		"id":          fileUUID,
		"plan_id":     planUUID,
		"category":    f.Category,
		"filename":    f.Filename,
		"storage_key": f.StorageKey,
		"file_type":   f.FileType,
		"size_bytes":  f.SizeBytes,
		"uploaded_at": f.UploadedAt,
	}

	if f.Slot != nil {
		response["slot"] = *f.Slot
	}

	if f.UploadedByUser != nil {
		uploaderUUID, _ := uuid.Parse(f.UploadedByUser.ID)
		response["uploaded_by"] = fiber.Map{
			"id":   uploaderUUID,
			"name": f.UploadedByUser.Name,
		}
	}

	return response
}

func formatFileList(files []db.FileWithUploader) []fiber.Map {
	result := make([]fiber.Map, len(files))
	for i, f := range files {
		result[i] = formatFileResponse(&f)
	}
	return result
}

func (h *PlanHandler) GetStats(c *fiber.Ctx) error {
	stats, err := db.GetDashboardStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch dashboard stats",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": stats,
	})
}

func (h *PlanHandler) GetRecent(c *fiber.Ctx) error {
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil && val > 0 && val <= 50 {
			limit = val
		}
	}

	plans, err := db.GetRecentPlans(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch recent plans",
			},
		})
	}

	result := make([]models.PlanListResponse, len(plans))
	for i, p := range plans {
		planUUID, _ := uuid.Parse(p.ID)
		result[i] = models.PlanListResponse{
			ID:        planUUID,
			Name:      p.Name,
			Slug:      p.Slug,
			Type:      p.Type,
			Style:     p.Style,
			Status:    p.Status,
			Beds:      p.Beds,
			Baths:     p.Baths,
			HalfBaths: p.HalfBaths,
			HeatedSF:  p.HeatedSF,
			TotalSF:   p.TotalSF,
			UpdatedAt: p.UpdatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

type BulkFileUploadRequest struct {
	PlanID string `json:"plan_id"`
	Slot   string `json:"slot"`
}

type BulkFileUploadResult struct {
	Success  bool   `json:"success"`
	PlanID   string `json:"plan_id"`
	Slot     string `json:"slot"`
	Filename string `json:"filename"`
	Message  string `json:"message,omitempty"`
}

func (h *PlanHandler) BulkUploadFiles(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Failed to parse multipart form",
			},
		})
	}

	metadataJSON := form.Value["metadata"]
	if len(metadataJSON) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Metadata is required",
			},
		})
	}

	var requests []BulkFileUploadRequest
	if err := json.Unmarshal([]byte(metadataJSON[0]), &requests); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid metadata JSON",
			},
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "No files provided",
			},
		})
	}

	if len(files) != len(requests) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Number of files must match number of metadata entries",
			},
		})
	}

	results := make([]BulkFileUploadResult, len(files))

	for i, fileHeader := range files {
		req := requests[i]
		result := BulkFileUploadResult{
			PlanID:   req.PlanID,
			Slot:     req.Slot,
			Filename: fileHeader.Filename,
		}

		if !validWebsiteSlots[req.Slot] {
			result.Success = false
			result.Message = "Invalid slot name"
			results[i] = result
			continue
		}

		plan, err := db.GetPlanByID(c.Context(), req.PlanID)
		if err != nil {
			result.Success = false
			result.Message = "Database error"
			results[i] = result
			continue
		}
		if plan == nil {
			result.Success = false
			result.Message = "Plan not found"
			results[i] = result
			continue
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			switch ext {
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			}
		}

		if !validImageTypes[contentType] {
			result.Success = false
			result.Message = "Only JPEG and PNG images are allowed"
			results[i] = result
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			result.Success = false
			result.Message = "Failed to read file"
			results[i] = result
			continue
		}

		fileData := make([]byte, fileHeader.Size)
		_, err = file.Read(fileData)
		file.Close()
		if err != nil {
			result.Success = false
			result.Message = "Failed to read file data"
			results[i] = result
			continue
		}

		// Process the image
		isPoster := req.Slot == "poster"
		processResult, err := processing.ProcessWebsiteImage(fileData, contentType, isPoster)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Image processing failed: %v", err)
			results[i] = result
			continue
		}

		// Use standardized filename and storage key
		storageKey := processing.GenerateStorageKey(plan.Slug, req.Slot)
		processedFilename := processing.StandardizeFilenameForSlot(fileHeader.Filename, plan.Slug, req.Slot)

		_, err = db.UpsertWebsiteFile(c.Context(), req.PlanID, req.Slot, processedFilename, storageKey, "image/jpeg", processResult.SizeBytes, userID)
		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("Failed to save file: %v", err)
			results[i] = result
			continue
		}

		db.RecalculatePlanStatus(c.Context(), req.PlanID)

		userUUID, _ := uuid.Parse(userID)
		planUUID, _ := uuid.Parse(req.PlanID)
		db.LogActivity(c.Context(), &userUUID, &planUUID, "file.uploaded", map[string]interface{}{
			"filename":       processedFilename,
			"slot":           req.Slot,
			"category":       "website",
			"original_size":  processResult.OriginalSize,
			"processed_size": processResult.SizeBytes,
		})

		result.Success = true
		result.Message = fmt.Sprintf("Processed: %dKB → %dKB", processResult.OriginalSize/1024, processResult.SizeBytes/1024)
		results[i] = result
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"results": results,
			"summary": fiber.Map{
				"total":   len(results),
				"success": successCount,
				"failed":  len(results) - successCount,
			},
		},
	})
}
