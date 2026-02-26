package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID             uuid.UUID  `json:"id"`
	PlanID         uuid.UUID  `json:"plan_id"`
	Category       string     `json:"category"`
	Slot           *string    `json:"slot"`
	Filename       string     `json:"filename"`
	StorageKey     string     `json:"storage_key"`
	FileType       string     `json:"file_type"`
	SizeBytes      int64      `json:"size_bytes"`
	UploadedAt     time.Time  `json:"uploaded_at"`
	UploadedBy     *uuid.UUID `json:"uploaded_by"`
	UploadedByUser *User      `json:"uploaded_by_user,omitempty"`
}
