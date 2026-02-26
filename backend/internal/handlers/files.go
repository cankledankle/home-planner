package handlers

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/cankledankle/home-planner/internal/processing"
	"github.com/cankledankle/home-planner/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var validWebsiteSlots = map[string]bool{
	"render-front":     true,
	"elevation-front":  true,
	"elevation-left":   true,
	"elevation-rear":   true,
	"elevation-right":  true,
	"floor-plan-main":  true,
	"floor-plan-upper": true,
	"floor-plan-lower": true,
	"poster":           true,
}

var validImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

var validCategories = map[string]bool{
	"reference": true,
	"technical": true,
	"3d":        true,
	"other":     true,
}

const (
	maxImageSize = processing.MaxWebsiteImageSize  // 5MB
	maxFileSize  = processing.MaxReferenceFileSize // 50MB
)

type FileHandler struct {
	r2Client *storage.R2Client
}

func NewFileHandler(r2Client *storage.R2Client) *FileHandler {
	return &FileHandler{r2Client: r2Client}
}

func (h *FileHandler) UploadWebsite(c *fiber.Ctx) error {
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

	slot := c.FormValue("slot")
	if slot == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Slot is required",
			},
		})
	}

	if !validWebsiteSlots[slot] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid slot name",
			},
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "File is required",
			},
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Only JPEG and PNG images are allowed for website uploads",
			},
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to read file",
			},
		})
	}
	defer file.Close()

	fileData := make([]byte, fileHeader.Size)
	_, err = file.Read(fileData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to read file data",
			},
		})
	}

	// Process the image
	isPoster := slot == "poster"
	processResult, err := processing.ProcessWebsiteImage(fileData, contentType, isPoster)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "PROCESSING_ERROR",
				"message": fmt.Sprintf("Image processing failed: %v", err),
			},
		})
	}

	// Use standardized filename and storage key
	storageKey := processing.GenerateStorageKey(plan.Slug, slot)
	processedFilename := processing.StandardizeFilenameForSlot(fileHeader.Filename, plan.Slug, slot)

	if h.r2Client != nil {
		err = h.r2Client.UploadFile(c.Context(), storageKey, processResult.Data, "image/jpeg")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": fmt.Sprintf("Failed to upload to storage: %v", err),
				},
			})
		}
	}

	userID := c.Locals("userID").(string)

	fileRow, err := db.UpsertWebsiteFile(c.Context(), planID, slot, processedFilename, storageKey, "image/jpeg", processResult.SizeBytes, userID)
	if err != nil {
		if h.r2Client != nil {
			h.r2Client.DeleteFile(c.Context(), storageKey)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to save file record: %v", err),
			},
		})
	}

	_, err = db.RecalculatePlanStatus(c.Context(), planID)
	if err != nil {
		fmt.Printf("Failed to recalculate plan status: %v\n", err)
	}

	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(planID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "file.uploaded", map[string]interface{}{
		"filename":       processedFilename,
		"slot":           slot,
		"category":       "website",
		"original_size":  processResult.OriginalSize,
		"processed_size": processResult.SizeBytes,
	})

	response := formatFileResponseFromRow(fileRow)
	response["processing"] = fiber.Map{
		"original_size":  processResult.OriginalSize,
		"processed_size": processResult.SizeBytes,
		"width":          processResult.Width,
		"height":         processResult.Height,
		"was_resized":    processResult.WasResized,
		"was_converted":  processResult.WasConverted,
		"warnings":       processResult.Warnings,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": response,
	})
}

func (h *FileHandler) Upload(c *fiber.Ctx) error {
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

	category := c.FormValue("category")
	if category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Category is required",
			},
		})
	}

	if category == "website" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Use /api/plans/:id/files/website for website uploads",
			},
		})
	}

	if !validCategories[category] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid category",
			},
		})
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Failed to parse multipart form",
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

	userID := c.Locals("userID").(string)
	var uploadedFiles []fiber.Map

	for _, fileHeader := range files {
		if fileHeader.Size > maxFileSize {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "VALIDATION_ERROR",
					"message": fmt.Sprintf("File %s exceeds 50MB limit", fileHeader.Filename),
				},
			})
		}

		file, err := fileHeader.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": fmt.Sprintf("Failed to read file %s", fileHeader.Filename),
				},
			})
		}

		fileData := make([]byte, fileHeader.Size)
		_, err = file.Read(fileData)
		file.Close()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": fmt.Sprintf("Failed to read file data for %s", fileHeader.Filename),
				},
			})
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		ext := filepath.Ext(fileHeader.Filename)
		storageKey := fmt.Sprintf("plans/%s/%s/%s%s", plan.Slug, category, generateSafeFilename(fileHeader.Filename), ext)

		if h.r2Client != nil {
			err = h.r2Client.UploadFile(c.Context(), storageKey, fileData, contentType)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fiber.Map{
						"code":    "INTERNAL_ERROR",
						"message": fmt.Sprintf("Failed to upload %s to storage: %v", fileHeader.Filename, err),
					},
				})
			}
		}

		fileRow, err := db.CreateFile(c.Context(), planID, category, "", fileHeader.Filename, storageKey, contentType, fileHeader.Size, userID)
		if err != nil {
			if h.r2Client != nil {
				h.r2Client.DeleteFile(c.Context(), storageKey)
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": fmt.Sprintf("Failed to save file record for %s: %v", fileHeader.Filename, err),
				},
			})
		}

		uploadedFiles = append(uploadedFiles, formatFileResponseFromRow(fileRow))
	}

	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(planID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "file.uploaded", map[string]interface{}{
		"count":    len(uploadedFiles),
		"category": category,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": uploadedFiles,
	})
}

func (h *FileHandler) List(c *fiber.Ctx) error {
	planID := c.Params("id")
	if planID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "Plan ID is required",
			},
		})
	}

	files, err := db.GetFilesByPlanID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch files",
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": formatFilesResponseForList(files),
	})
}

func (h *FileHandler) GetURL(c *fiber.Ctx) error {
	fileID := c.Params("id")
	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "File ID is required",
			},
		})
	}

	file, err := db.GetFileByID(c.Context(), fileID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch file",
			},
		})
	}

	if file == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "File not found",
			},
		})
	}

	if h.r2Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "SERVICE_UNAVAILABLE",
				"message": "Storage not configured",
			},
		})
	}

	url, err := h.r2Client.GetPresignedURL(c.Context(), file.StorageKey, 60)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to generate presigned URL: %v", err),
			},
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url":        url,
			"expires_at": "60 minutes",
		},
	})
}

func (h *FileHandler) Delete(c *fiber.Ctx) error {
	fileID := c.Params("id")
	if fileID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "VALIDATION_ERROR",
				"message": "File ID is required",
			},
		})
	}

	file, err := db.GetFileByID(c.Context(), fileID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to fetch file",
			},
		})
	}

	if file == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "NOT_FOUND",
				"message": "File not found",
			},
		})
	}

	if h.r2Client != nil {
		err = h.r2Client.DeleteFile(c.Context(), file.StorageKey)
		if err != nil {
			// Log error but continue to delete from DB
			fmt.Printf("Failed to delete file from storage: %v\n", err)
		}
	}

	err = db.DeleteFileByID(c.Context(), fileID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to delete file record",
			},
		})
	}

	// Recalculate plan status if it was a website file
	if file.Category == "website" && file.Slot != nil {
		_, err = db.RecalculatePlanStatus(c.Context(), file.PlanID)
		if err != nil {
			fmt.Printf("Failed to recalculate plan status: %v\n", err)
		}
	}

	userID := c.Locals("userID").(string)
	userUUID, _ := uuid.Parse(userID)
	planUUID, _ := uuid.Parse(file.PlanID)
	db.LogActivity(c.Context(), &userUUID, &planUUID, "file.deleted", map[string]interface{}{
		"filename": file.Filename,
		"category": file.Category,
	})

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "File deleted",
		},
	})
}

func generateSafeFilename(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Replace unsafe characters
	safe := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"\"", "-",
		"<", "-",
		">", "-",
		"|", "-",
		" ", "-",
	).Replace(name)

	return safe
}

func formatFileResponseFromRow(file *db.FileRow) fiber.Map {
	fileUUID, _ := uuid.Parse(file.ID)
	planUUID, _ := uuid.Parse(file.PlanID)

	response := fiber.Map{
		"id":          fileUUID,
		"plan_id":     planUUID,
		"category":    file.Category,
		"filename":    file.Filename,
		"storage_key": file.StorageKey,
		"file_type":   file.FileType,
		"size_bytes":  file.SizeBytes,
		"uploaded_at": file.UploadedAt,
	}

	if file.Slot != nil {
		response["slot"] = *file.Slot
	}

	if file.UploadedBy != nil {
		uploaderUUID, _ := uuid.Parse(*file.UploadedBy)
		response["uploaded_by"] = fiber.Map{
			"id": uploaderUUID,
		}
	}

	return response
}

func formatFilesResponseForList(files map[string][]db.FileWithUploader) fiber.Map {
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
			websiteSlots[*file.Slot] = formatFileWithUploader(&file)
		}
	}

	return fiber.Map{
		"website":   websiteSlots,
		"reference": formatFileListWithUploader(files["reference"]),
		"technical": formatFileListWithUploader(files["technical"]),
		"3d":        formatFileListWithUploader(files["3d"]),
		"other":     formatFileListWithUploader(files["other"]),
	}
}

func formatFileWithUploader(f *db.FileWithUploader) fiber.Map {
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

func formatFileListWithUploader(files []db.FileWithUploader) []fiber.Map {
	result := make([]fiber.Map, len(files))
	for i, f := range files {
		result[i] = formatFileWithUploader(&f)
	}
	return result
}
