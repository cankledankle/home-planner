package handlers

import (
	"os"
	"strings"
	"time"

	"github.com/cankledankle/home-planner/internal/auth"
	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// cookieSecure returns true unless COOKIE_SECURE is explicitly set to "false" or "0".
// Defaults to true so production deployments are secure without extra configuration.
func cookieSecure() bool {
	val := os.Getenv("COOKIE_SECURE")
	if val == "" {
		return true
	}
	return strings.ToLower(val) != "false" && val != "0"
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	store *db.Store
}

func NewAuthHandler(store *db.Store) *AuthHandler {
	return &AuthHandler{store: store}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request body",
			},
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Email and password are required",
			},
		})
	}

	// Get user by email
	user, err := h.store.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to authenticate",
			},
		})
	}

	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_CREDENTIALS",
				"message": "Invalid email or password",
			},
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_CREDENTIALS",
				"message": "Invalid email or password",
			},
		})
	}

	// Generate tokens
	tokens, err := auth.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to generate tokens",
			},
		})
	}

	// Hash and store refresh token
	hashedRefreshToken := db.HashToken(tokens.RefreshToken)

	userUUID, _ := uuid.Parse(user.ID)
	if err := h.store.StoreRefreshToken(userUUID, hashedRefreshToken, tokens.RefreshExpiry); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to store session",
			},
		})
	}

	// Set cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Expires:  tokens.AccessExpiry,
		HTTPOnly: true,
		Secure:   cookieSecure(),
		SameSite: "Strict",
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  tokens.RefreshExpiry,
		HTTPOnly: true,
		Secure:   cookieSecure(),
		SameSite: "Strict",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"user": models.UserResponse{
				ID:    userUUID,
				Name:  user.Name,
				Email: user.Email,
				Role:  user.Role,
			},
		},
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_TOKEN",
				"message": "Refresh token required",
			},
		})
	}

	// Validate token against database using bcrypt comparison
	userID, err := h.store.ValidateRefreshToken(refreshToken)
	if err != nil {
		// Try with the raw token in case hash doesn't match (for comparison)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_TOKEN",
				"message": "Invalid refresh token",
			},
		})
	}

	// Get user details
	userUUID := userID.String()
	user, err := h.store.GetUserByID(c.Context(), userUUID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_TOKEN",
				"message": "User not found",
			},
		})
	}

	// Generate new access token
	tokens, err := auth.GenerateTokenPair(userUUID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to generate new token",
			},
		})
	}

	// Set new access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Expires:  tokens.AccessExpiry,
		HTTPOnly: true,
		Secure:   cookieSecure(),
		SameSite: "Strict",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Token refreshed",
		},
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	// Delete refresh token from database if present
	if refreshToken != "" {
		h.store.DeleteRefreshTokenByValue(refreshToken)
	}

	// Clear cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Logged out",
		},
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
	}

	userUUID, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Invalid user context",
			},
		})
	}

	user, err := h.store.GetUserByID(c.Context(), userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user",
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

	uuid, _ := uuid.Parse(user.ID)
	return c.JSON(fiber.Map{
		"data": models.UserResponse{
			ID:        uuid,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}
