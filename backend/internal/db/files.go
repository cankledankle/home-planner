package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type FileRow struct {
	ID         string
	PlanID     string
	Category   string
	Slot       *string
	Filename   string
	StorageKey string
	FileType   string
	SizeBytes  int64
	UploadedAt time.Time
	UploadedBy *string
}

type UserInfo struct {
	ID   string
	Name string
}

type FileWithUploader struct {
	FileRow
	UploadedByUser *UserInfo
}

func (s *Store) GetFilesByPlanID(ctx context.Context, planID string) (map[string][]FileWithUploader, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT f.id, f.plan_id, f.category, f.slot, f.filename, f.storage_key, 
		       f.file_type, f.size_bytes, f.uploaded_at, f.uploaded_by,
		       u.id, u.name
		FROM files f
		LEFT JOIN users u ON f.uploaded_by = u.id
		WHERE f.plan_id = $1
		ORDER BY f.uploaded_at DESC
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]FileWithUploader)
	result["website"] = []FileWithUploader{}
	result["reference"] = []FileWithUploader{}
	result["technical"] = []FileWithUploader{}
	result["3d"] = []FileWithUploader{}
	result["other"] = []FileWithUploader{}

	for rows.Next() {
		var f FileWithUploader
		var uploadedByID, uploadedByName *string

		err := rows.Scan(
			&f.ID, &f.PlanID, &f.Category, &f.Slot, &f.Filename, &f.StorageKey,
			&f.FileType, &f.SizeBytes, &f.UploadedAt, &f.UploadedBy,
			&uploadedByID, &uploadedByName,
		)
		if err != nil {
			return nil, err
		}

		if uploadedByID != nil && uploadedByName != nil {
			f.UploadedByUser = &UserInfo{
				ID:   *uploadedByID,
				Name: *uploadedByName,
			}
		}

		result[f.Category] = append(result[f.Category], f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Store) GetFileByID(ctx context.Context, fileID string) (*FileRow, error) {
	var f FileRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, plan_id, category, slot, filename, storage_key, 
		       file_type, size_bytes, uploaded_at, uploaded_by
		FROM files
		WHERE id = $1
	`, fileID).Scan(
		&f.ID, &f.PlanID, &f.Category, &f.Slot, &f.Filename, &f.StorageKey,
		&f.FileType, &f.SizeBytes, &f.UploadedAt, &f.UploadedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (s *Store) CreateFile(ctx context.Context, planID, category, slot, filename, storageKey, fileType string, sizeBytes int64, uploadedBy string) (*FileRow, error) {
	var f FileRow

	var slotPtr *string
	if slot != "" {
		slotPtr = &slot
	}

	var uploadedByPtr *string
	if uploadedBy != "" {
		uploadedByPtr = &uploadedBy
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO files (plan_id, category, slot, filename, storage_key, file_type, size_bytes, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, plan_id, category, slot, filename, storage_key, file_type, size_bytes, uploaded_at, uploaded_by
	`, planID, category, slotPtr, filename, storageKey, fileType, sizeBytes, uploadedByPtr).Scan(
		&f.ID, &f.PlanID, &f.Category, &f.Slot, &f.Filename, &f.StorageKey,
		&f.FileType, &f.SizeBytes, &f.UploadedAt, &f.UploadedBy,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *Store) DeleteFileByID(ctx context.Context, fileID string) error {
	result, err := s.pool.Exec(ctx, "DELETE FROM files WHERE id = $1", fileID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetFileByPlanAndSlot(ctx context.Context, planID, slot string) (*FileRow, error) {
	var f FileRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, plan_id, category, slot, filename, storage_key, 
		       file_type, size_bytes, uploaded_at, uploaded_by
		FROM files
		WHERE plan_id = $1 AND category = 'website' AND slot = $2
	`, planID, slot).Scan(
		&f.ID, &f.PlanID, &f.Category, &f.Slot, &f.Filename, &f.StorageKey,
		&f.FileType, &f.SizeBytes, &f.UploadedAt, &f.UploadedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (s *Store) UpsertWebsiteFile(ctx context.Context, planID, slot, filename, storageKey, fileType string, sizeBytes int64, uploadedBy string) (*FileRow, error) {
	var f FileRow

	var uploadedByPtr *string
	if uploadedBy != "" {
		uploadedByPtr = &uploadedBy
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO files (plan_id, category, slot, filename, storage_key, file_type, size_bytes, uploaded_by)
		VALUES ($1, 'website', $2, $3, $4, $5, $6, $7)
		ON CONFLICT (plan_id, slot) WHERE category = 'website' AND slot IS NOT NULL
		DO UPDATE SET
			filename = EXCLUDED.filename,
			storage_key = EXCLUDED.storage_key,
			file_type = EXCLUDED.file_type,
			size_bytes = EXCLUDED.size_bytes,
			uploaded_at = NOW(),
			uploaded_by = EXCLUDED.uploaded_by
		RETURNING id, plan_id, category, slot, filename, storage_key, file_type, size_bytes, uploaded_at, uploaded_by
	`, planID, slot, filename, storageKey, fileType, sizeBytes, uploadedByPtr).Scan(
		&f.ID, &f.PlanID, &f.Category, &f.Slot, &f.Filename, &f.StorageKey,
		&f.FileType, &f.SizeBytes, &f.UploadedAt, &f.UploadedBy,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (s *Store) DeleteFilesByPlanID(ctx context.Context, planID string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM files WHERE plan_id = $1", planID)
	return err
}
