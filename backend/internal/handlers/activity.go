package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ActivityListResponse struct {
	ID        string          `json:"id"`
	User      *ActivityUser   `json:"user,omitempty"`
	Plan      *ActivityPlan   `json:"plan,omitempty"`
	Action    string          `json:"action"`
	Detail    json.RawMessage `json:"detail,omitempty"`
	CreatedAt string          `json:"created_at"`
}

type ActivityUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ActivityPlan struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ActivityHandler struct {
	store *db.Store
}

func NewActivityHandler(store *db.Store) *ActivityHandler {
	return &ActivityHandler{store: store}
}

func (h *ActivityHandler) List(c *fiber.Ctx) error {
	filters := db.ActivityListFilters{
		Page:  1,
		Limit: 50,
	}

	if userID := c.Query("user_id"); userID != "" {
		filters.UserID = &userID
	}

	if planID := c.Query("plan_id"); planID != "" {
		filters.PlanID = &planID
	}

	filters.Action = c.Query("action")

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

	result, err := h.store.ListActivities(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch activity log",
			},
		})
	}

	activities := make([]ActivityListResponse, len(result.Activities))
	for i, a := range result.Activities {
		activityUUID, _ := uuid.Parse(a.ID)

		var user *ActivityUser
		if a.UserID != nil && a.UserName != nil {
			userUUID, _ := uuid.Parse(*a.UserID)
			user = &ActivityUser{
				ID:   userUUID.String(),
				Name: *a.UserName,
			}
		}

		var plan *ActivityPlan
		if a.PlanID != nil && a.PlanName != nil {
			planUUID, _ := uuid.Parse(*a.PlanID)
			plan = &ActivityPlan{
				ID:   planUUID.String(),
				Name: *a.PlanName,
			}
		}

		activities[i] = ActivityListResponse{
			ID:        activityUUID.String(),
			User:      user,
			Plan:      plan,
			Action:    a.Action,
			Detail:    a.Detail,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(fiber.Map{
		"data": activities,
		"meta": fiber.Map{
			"page":        result.Page,
			"limit":       result.Limit,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
	})
}

func (h *ActivityHandler) ListForPlan(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	page := 1
	limit := 50

	if pageParam := c.Query("page"); pageParam != "" {
		if val, err := strconv.Atoi(pageParam); err == nil && val > 0 {
			page = val
		}
	}

	if limitParam := c.Query("limit"); limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	result, err := h.store.ListActivitiesForPlan(c.Context(), planID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch activity log",
			},
		})
	}

	activities := make([]ActivityListResponse, len(result.Activities))
	for i, a := range result.Activities {
		activityUUID, _ := uuid.Parse(a.ID)

		var user *ActivityUser
		if a.UserID != nil && a.UserName != nil {
			userUUID, _ := uuid.Parse(*a.UserID)
			user = &ActivityUser{
				ID:   userUUID.String(),
				Name: *a.UserName,
			}
		}

		var plan *ActivityPlan
		if a.PlanID != nil && a.PlanName != nil {
			planUUID, _ := uuid.Parse(*a.PlanID)
			plan = &ActivityPlan{
				ID:   planUUID.String(),
				Name: *a.PlanName,
			}
		}

		activities[i] = ActivityListResponse{
			ID:        activityUUID.String(),
			User:      user,
			Plan:      plan,
			Action:    a.Action,
			Detail:    a.Detail,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(fiber.Map{
		"data": activities,
		"meta": fiber.Map{
			"page":        result.Page,
			"limit":       result.Limit,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
	})
}
