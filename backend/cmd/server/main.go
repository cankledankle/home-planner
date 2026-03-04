package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/handlers"
	"github.com/cankledankle/home-planner/internal/middleware"
	"github.com/cankledankle/home-planner/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env from multiple locations
	// 1. Current directory
	// 2. Parent directory (for running from backend/ subdir)
	loaded := false
	if err := godotenv.Load(); err == nil {
		loaded = true
	} else if err := godotenv.Load("../.env"); err == nil {
		loaded = true
	}

	if !loaded {
		log.Println("No .env file found in current or parent directory")
	}

	// Connect to database
	if err := db.Connect(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied")

	// Seed admin user if env vars are set and no users exist
	if err := db.SeedAdminUser(); err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	// Cleanup expired refresh tokens on startup
	if err := db.CleanupExpiredRefreshTokens(); err != nil {
		log.Printf("Warning: failed to cleanup expired tokens: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "Home Planner API",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	// Middleware
	app.Use(fiberlogger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     os.Getenv("FRONTEND_URL"),
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Rate limiting for API routes
	app.Use("/api", limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"status":  "error",
				"message": "Rate limit exceeded. Please try again later.",
			})
		},
	}))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	userHandler := handlers.NewUserHandler()
	planHandler := handlers.NewPlanHandler()
	activityHandler := handlers.NewActivityHandler()

	// Initialize R2 client (will return nil if not configured)
	r2Client, r2Err := storage.NewR2Client()
	if r2Err != nil {
		log.Printf("Warning: R2 storage not configured: %v", r2Err)
	} else {
		log.Println("R2 storage connected")
	}
	exportHandler := handlers.NewExportHandler(r2Client)
	importHandler := handlers.NewImportHandler()
	fileHandler := handlers.NewFileHandler(r2Client)

	// Health check endpoint - checks database connectivity
	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database connection
		dbErr := db.Ping()
		if dbErr != nil {
			return c.Status(503).JSON(fiber.Map{
				"status":  "error",
				"message": "Database connection failed",
			})
		}

		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// Auth routes (public)
	app.Post("/api/auth/login", authHandler.Login)
	app.Post("/api/auth/refresh", authHandler.Refresh)

	// Protected routes
	api := app.Group("/api")
	api.Use(middleware.AuthMiddleware)

	// Auth routes (protected)
	api.Post("/auth/logout", authHandler.Logout)
	api.Get("/auth/me", authHandler.Me)

	// Activity routes (protected - all authenticated users)
	api.Get("/activity", activityHandler.List)
	api.Get("/plans/:id/activity", activityHandler.ListForPlan)

	// Export routes (protected - all authenticated users)
	api.Get("/export/csv", exportHandler.ExportCSV)
	api.Get("/export/zip", exportHandler.ExportZIP)

	// File routes (protected - all authenticated users)
	api.Get("/plans/:id/files", fileHandler.List)
	api.Post("/plans/:id/files/website", fileHandler.UploadWebsite)
	api.Post("/plans/:id/files", fileHandler.Upload)
	api.Get("/files/:id/url", fileHandler.GetURL)
	api.Delete("/files/:id", fileHandler.Delete)

	// Plans routes (protected - all authenticated users)
	api.Get("/plans/stats", planHandler.GetStats)
	api.Get("/plans/recent", planHandler.GetRecent)
	api.Get("/plans", planHandler.List)
	api.Get("/plans/:id", planHandler.Get)
	api.Post("/plans", planHandler.Create)
	api.Put("/plans/:id", planHandler.Update)
	api.Post("/plans/:id/duplicate", planHandler.Duplicate)
	api.Put("/plans/:id/flag", planHandler.Flag)
	api.Put("/plans/:id/unflag", planHandler.Unflag)

	// Admin-only routes
	admin := api.Group("/")
	admin.Use(middleware.AdminMiddleware)

	// Users routes (admin-only)
	admin.Get("/users", userHandler.List)
	admin.Post("/users", userHandler.Create)
	admin.Put("/users/:id", userHandler.Update)
	admin.Put("/users/:id/password", userHandler.UpdatePassword)
	admin.Delete("/users/:id", userHandler.Delete)

	// Plan admin routes (admin-only)
	admin.Delete("/plans/:id", planHandler.Delete)
	admin.Post("/plans/:id/restore", planHandler.Restore)

	// Import routes (admin-only)
	admin.Post("/import/csv/preview", importHandler.PreviewCSV)
	admin.Post("/import/csv", importHandler.ImportCSV)
	admin.Get("/import/recent", importHandler.GetRecentImports)

	// Bulk file upload (admin-only)
	admin.Post("/plans/bulk-files", planHandler.BulkUploadFiles)

	// Serve static files from the frontend build
	// The static files are expected to be at /app/static (in production container)
	staticPath := os.Getenv("STATIC_PATH")
	if staticPath == "" {
		staticPath = "./static" // Development fallback
	}

	// Serve immutable assets with long cache headers (1 year)
	app.Use("/_app/immutable", filesystem.New(filesystem.Config{
		Root:   http.Dir(staticPath + "/_app/immutable"),
		Browse: false,
		MaxAge: 31536000, // 1 year in seconds
	}))

	// Serve all other static files with no-cache headers
	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.Dir(staticPath),
		Index:        "index.html",
		Browse:       false,
		MaxAge:       0,            // No cache for SPA routes
		NotFoundFile: "index.html", // For SPA routing
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	log.Printf("Server started on port %s", port)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
