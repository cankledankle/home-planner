package db

import (
	"context"
	"fmt"
	"strings"
)

type FileExportRow struct {
	*FileRow
	PlanName string
}

func GetPlansForExport(ctx context.Context, planIDs []string) ([]*PlanRow, error) {
	var query string
	var args []interface{}

	if len(planIDs) > 0 {
		placeholders := make([]string, len(planIDs))
		for i, id := range planIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args = append(args, id)
		}
		query = fmt.Sprintf(`
			SELECT id, name, slug, type, style, status, beds, baths, half_baths,
			       main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
			       garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
			       notes, deleted_at, created_at, updated_at, created_by, updated_by
			FROM plans
			WHERE deleted_at IS NULL AND id IN (%s)
			ORDER BY name
		`, strings.Join(placeholders, ", "))
	} else {
		query = `
			SELECT id, name, slug, type, style, status, beds, baths, half_baths,
			       main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
			       garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
			       notes, deleted_at, created_at, updated_at, created_by, updated_by
			FROM plans
			WHERE deleted_at IS NULL
			ORDER BY name
		`
	}

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*PlanRow
	for rows.Next() {
		var p PlanRow
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Type, &p.Style, &p.Status, &p.Beds, &p.Baths, &p.HalfBaths,
			&p.MainSF, &p.UpperSF, &p.LowerSF, &p.PorchDeckSF, &p.GarageSF,
			&p.GarageApartmentSF, &p.UnfinishedSF, &p.HeatedSF, &p.TotalSF,
			&p.Notes, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func GetFilesForExport(ctx context.Context, planIDs []string, categories []string) ([]*FileExportRow, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "f.plan_id = p.id")
	conditions = append(conditions, "p.deleted_at IS NULL")

	if len(planIDs) > 0 {
		placeholders := make([]string, len(planIDs))
		for i, id := range planIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("p.id IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(categories) > 0 {
		placeholders := make([]string, len(categories))
		for i, cat := range categories {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, cat)
			argIdx++
		}
		conditions = append(conditions, fmt.Sprintf("f.category IN (%s)", strings.Join(placeholders, ", ")))
	}

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT f.id, f.plan_id, f.category, f.slot, f.filename, f.storage_key,
		       f.file_type, f.size_bytes, f.uploaded_at, f.uploaded_by,
		       p.name
		FROM files f
		JOIN plans p ON f.plan_id = p.id
		WHERE %s
		ORDER BY p.name, f.category, f.filename
	`, whereClause)

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*FileExportRow
	for rows.Next() {
		var f FileRow
		var planName string
		err := rows.Scan(
			&f.ID, &f.PlanID, &f.Category, &f.Slot, &f.Filename, &f.StorageKey,
			&f.FileType, &f.SizeBytes, &f.UploadedAt, &f.UploadedBy,
			&planName,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, &FileExportRow{
			FileRow:  &f,
			PlanName: planName,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
