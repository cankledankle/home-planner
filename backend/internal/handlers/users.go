package handlers

import (
	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/models"
	"github.com/cankledankle/home-planner/internal/sftpgo"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

type UserHandler struct {
	sftpService *sftpgo.Service
}

func NewUserHandler() *UserHandler {
	sftpService, err := sftpgo.NewService()
	if err != nil {
		// SFTPGo is optional, log but don't fail
		return &UserHandler{sftpService: nil}
	}
	return &UserHandler{sftpService: sftpService}
}

func (h *UserHandler) List(c *fiber.Ctx) error {
	users, err := db.GetAllUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch users",
			},
		})
	}

	response := make([]models.UserResponse, len(users))
	for i, user := range users {
		userUUID, _ := uuid.Parse(user.ID)
		response[i] = models.UserResponse{
			ID:        userUUID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Name, email, password, and role are required",
			},
		})
	}

	if req.Role != "admin" && req.Role != "editor" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Role must be 'admin' or 'editor'",
			},
		})
	}

	taken, err := db.IsEmailTaken(c.Context(), req.Email, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to check email availability",
			},
		})
	}

	if taken {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "CONFLICT",
				"message": "Email already in use",
			},
		})
	}

	user, err := db.CreateUser(c.Context(), req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create user",
			},
		})
	}

	// Create SFTPGo user if service is configured
	var sftpCredentials *sftpgo.UserCredentials
	if h.sftpService != nil && h.sftpService.IsConfigured() {
		sftpCredentials, err = h.sftpService.CreateUser(c.Context(), sftpgo.CreateUserRequest{
			UserID:     user.ID,
			Email:      user.Email,
			Name:       user.Name,
			Permission: "readwrite", // Default to read+write for new users
		})
		if err != nil {
			// Log error but don't fail user creation
			// User was created successfully in the app
			// SFTP credentials can be generated later
			c.Context().Logger().Printf("Warning: Failed to create SFTPGo user: %v", err)
		}
	}

	userUUID, _ := uuid.Parse(user.ID)
	response := fiber.Map{
		"data": models.UserResponse{
			ID:        userUUID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	// Include SFTP credentials in response if they were created
	if sftpCredentials != nil {
		response["sftp_credentials"] = sftpCredentials
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Name == "" || req.Email == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Name, email, and role are required",
			},
		})
	}

	if req.Role != "admin" && req.Role != "editor" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Role must be 'admin' or 'editor'",
			},
		})
	}

	taken, err := db.IsEmailTaken(c.Context(), req.Email, &userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to check email availability",
			},
		})
	}

	if taken {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "CONFLICT",
				"message": "Email already in use",
			},
		})
	}

	user, err := db.UpdateUser(c.Context(), userID, req.Name, req.Email, req.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update user",
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

	userUUID, _ := uuid.Parse(user.ID)
	return c.JSON(fiber.Map{
		"data": models.UserResponse{
			ID:        userUUID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	var req UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Password is required",
			},
		})
	}

	err := db.UpdateUserPassword(c.Context(), userID, req.Password)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "User not found",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to update password",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Password updated",
		},
	})
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "User ID is required",
			},
		})
	}

	currentUserID := c.Locals("userID")
	if currentUserID == userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "FORBIDDEN",
				"message": "Cannot delete your own account",
			},
		})
	}

	// Get user email before deletion for SFTP cleanup
	user, err := db.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch user",
			},
		})
	}

	err = db.DeleteUser(c.Context(), userID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "NOT_FOUND",
					"message": "User not found",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to delete user",
			},
		})
	}

	// Delete SFTPGo user if service is configured
	if h.sftpService != nil && h.sftpService.IsConfigured() && user != nil {
		if err := h.sftpService.DeleteUser(c.Context(), user.Email); err != nil {
			// Log error but don't fail the deletion
			c.Context().Logger().Printf("Warning: Failed to delete SFTPGo user: %v", err)
		}
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "User deleted",
		},
	})
}
