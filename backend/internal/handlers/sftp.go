package handlers

import (
	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/sftpgo"
	"github.com/gofiber/fiber/v2"
)

type SFTPHandler struct {
	service *sftpgo.Service
}

type GenerateCredentialsRequest struct {
	UserID     string `json:"user_id"`
	Permission string `json:"permission"` // "read" or "readwrite"
}

type UpdatePermissionRequest struct {
	Permission string `json:"permission"` // "read" or "readwrite"
}

func NewSFTPHandler() *SFTPHandler {
	service, err := sftpgo.NewService()
	if err != nil {
		return &SFTPHandler{service: nil}
	}
	return &SFTPHandler{service: service}
}

func (h *SFTPHandler) GetStatus(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"configured": false,
				"message":    "SFTP service not configured",
			},
		})
	}

	if err := h.service.CheckHealth(c.Context()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"data": fiber.Map{
				"configured": false,
				"healthy":    false,
				"message":    err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"configured": true,
			"healthy":    true,
		},
	})
}

func (h *SFTPHandler) GetUserCredentials(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_CONFIGURED",
				"message": "SFTP service not configured",
			},
		})
	}

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	user, err := db.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch user",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	credentials, err := h.service.GetCredentials(c.Context(), user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get SFTP credentials",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": credentials,
	})
}

func (h *SFTPHandler) GenerateCredentials(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_CONFIGURED",
				"message": "SFTP service not configured",
			},
		})
	}

	var req GenerateCredentialsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	if req.Permission != "read" && req.Permission != "readwrite" {
		req.Permission = "readwrite" // Default
	}

	user, err := db.GetUserByID(c.Context(), req.UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch user",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	credentials, err := h.service.CreateUser(c.Context(), sftpgo.CreateUserRequest{
		UserID:     req.UserID,
		Email:      user.Email,
		Name:       user.Name,
		Permission: req.Permission,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create SFTP credentials",
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": credentials,
	})
}

func (h *SFTPHandler) RotateCredentials(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_CONFIGURED",
				"message": "SFTP service not configured",
			},
		})
	}

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	user, err := db.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch user",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	credentials, err := h.service.RotateCredentials(c.Context(), user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to rotate SFTP credentials",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": credentials,
	})
}

func (h *SFTPHandler) RevokeCredentials(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_CONFIGURED",
				"message": "SFTP service not configured",
			},
		})
	}

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	user, err := db.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch user",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	if err := h.service.RevokeAccess(c.Context(), user.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to revoke SFTP credentials",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "SFTP access revoked",
		},
	})
}

func (h *SFTPHandler) DeleteCredentials(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_CONFIGURED",
				"message": "SFTP service not configured",
			},
		})
	}

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	user, err := db.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch user",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	if err := h.service.DeleteUser(c.Context(), user.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to delete SFTP credentials",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "SFTP credentials deleted",
		},
	})
}

func (h *SFTPHandler) UpdatePermission(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_CONFIGURED",
				"message": "SFTP service not configured",
			},
		})
	}

	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	var req UpdatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Permission != "read" && req.Permission != "readwrite" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Permission must be 'read' or 'readwrite'",
			},
		})
	}

	user, err := db.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch user",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	if err := h.service.UpdatePermission(c.Context(), user.Email, sftpgo.UpdatePermissionRequest{
		Permission: req.Permission,
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update SFTP permissions",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "SFTP permissions updated",
		},
	})
}

func (h *SFTPHandler) ListAllCredentials(c *fiber.Ctx) error {
	if h.service == nil || !h.service.IsConfigured() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_CONFIGURED",
				"message": "SFTP service not configured",
			},
		})
	}

	credentials, err := h.service.ListAllUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to list SFTP credentials",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": credentials,
	})
}
