package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PlanRow struct {
	ID                string
	Name              string
	Slug              string
	Type              *string
	Style             *string
	Status            string
	Beds              *int
	Baths             *int
	HalfBaths         *int
	MainSF            *int
	UpperSF           *int
	LowerSF           *int
	PorchDeckSF       *int
	GarageSF          *int
	GarageApartmentSF *int
	UnfinishedSF      *int
	HeatedSF          *int
	TotalSF           *int
	Notes             *string
	DeletedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CreatedBy         *string
	UpdatedBy         *string
}

// planColumns is the canonical SELECT/RETURNING column list for plans queries.
// Update this constant (and scanFields below) when adding or removing columns.
const planColumns = "id, name, slug, type, style, status, beds, baths, half_baths, " +
	"main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf, " +
	"garage_apartment_sf, unfinished_sf, heated_sf, total_sf, " +
	"notes, deleted_at, created_at, updated_at, created_by, updated_by"

// scanFields returns the scan destinations for a PlanRow in the same order as planColumns.
func (p *PlanRow) scanFields() []any {
	return []any{
		&p.ID, &p.Name, &p.Slug, &p.Type, &p.Style, &p.Status, &p.Beds, &p.Baths, &p.HalfBaths,
		&p.MainSF, &p.UpperSF, &p.LowerSF, &p.PorchDeckSF, &p.GarageSF,
		&p.GarageApartmentSF, &p.UnfinishedSF, &p.HeatedSF, &p.TotalSF,
		&p.Notes, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy,
	}
}

type PlanListFilters struct {
	Search      string
	Status      string
	Type        string
	Style       string
	BedsMin     *int
	BedsMax     *int
	BathsMin    *int
	BathsMax    *int
	HeatedSFMin *int
	HeatedSFMax *int
	MissingSlot string
	Sort        string
	Order       string
	Page        int
	Limit       int
}

type PaginatedPlans struct {
	Plans      []PlanRow
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (s *Store) makeUniqueSlug(ctx context.Context, baseSlug string) (string, error) {
	slug := baseSlug
	suffix := 0

	for {
		exists, err := s.CheckSlugExists(ctx, slug, nil)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		suffix++
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
	}
}

func (s *Store) CheckSlugExists(ctx context.Context, slug string, excludePlanID *string) (bool, error) {
	var query string
	var args []interface{}

	if excludePlanID != nil {
		query = "SELECT COUNT(*) FROM plans WHERE slug = $1 AND deleted_at IS NULL AND id != $2"
		args = []interface{}{slug, *excludePlanID}
	} else {
		query = "SELECT COUNT(*) FROM plans WHERE slug = $1 AND deleted_at IS NULL"
		args = []interface{}{slug}
	}

	var count int
	err := s.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Store) ListPlans(ctx context.Context, filters PlanListFilters) (*PaginatedPlans, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 50
	}

	whereClause := "WHERE deleted_at IS NULL"
	var args []interface{}
	argIdx := 1

	if filters.Search != "" {
		whereClause += fmt.Sprintf(" AND search_vector @@ plainto_tsquery('english', $%d)", argIdx)
		args = append(args, filters.Search)
		argIdx++
	}

	if filters.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filters.Status)
		argIdx++
	}

	if filters.Type != "" {
		whereClause += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filters.Type)
		argIdx++
	}

	if filters.Style != "" {
		whereClause += fmt.Sprintf(" AND style = $%d", argIdx)
		args = append(args, filters.Style)
		argIdx++
	}

	if filters.BedsMin != nil {
		whereClause += fmt.Sprintf(" AND beds >= $%d", argIdx)
		args = append(args, *filters.BedsMin)
		argIdx++
	}

	if filters.BedsMax != nil {
		whereClause += fmt.Sprintf(" AND beds <= $%d", argIdx)
		args = append(args, *filters.BedsMax)
		argIdx++
	}

	if filters.BathsMin != nil {
		whereClause += fmt.Sprintf(" AND baths >= $%d", argIdx)
		args = append(args, *filters.BathsMin)
		argIdx++
	}

	if filters.BathsMax != nil {
		whereClause += fmt.Sprintf(" AND baths <= $%d", argIdx)
		args = append(args, *filters.BathsMax)
		argIdx++
	}

	if filters.HeatedSFMin != nil {
		whereClause += fmt.Sprintf(" AND heated_sf >= $%d", argIdx)
		args = append(args, *filters.HeatedSFMin)
		argIdx++
	}

	if filters.HeatedSFMax != nil {
		whereClause += fmt.Sprintf(" AND heated_sf <= $%d", argIdx)
		args = append(args, *filters.HeatedSFMax)
		argIdx++
	}

	if filters.MissingSlot != "" {
		whereClause += fmt.Sprintf(` AND NOT EXISTS (
			SELECT 1 FROM files f 
			WHERE f.plan_id = plans.id 
			AND f.category = 'website' 
			AND f.slot = $%d
		)`, argIdx)
		args = append(args, filters.MissingSlot)
		argIdx++
	}

	orderBy := "name"
	switch filters.Sort {
	case "heated_sf", "total_sf", "created_at", "updated_at":
		orderBy = filters.Sort
	}

	orderDir := "ASC"
	if strings.ToLower(filters.Order) == "desc" {
		orderDir = "DESC"
	}

	countQuery := "SELECT COUNT(*) FROM plans " + whereClause
	var total int
	err := s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, name, slug, type, style, status, beds, baths, half_baths,
		       main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
		       garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
		       notes, deleted_at, created_at, updated_at, created_by, updated_by
		FROM plans
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, orderDir, argIdx, argIdx+1)

	args = append(args, filters.Limit, (filters.Page-1)*filters.Limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []PlanRow
	for rows.Next() {
		var p PlanRow
		err := rows.Scan(p.scanFields()...)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + filters.Limit - 1) / filters.Limit

	return &PaginatedPlans{
		Plans:      plans,
		Total:      total,
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *Store) GetPlanByID(ctx context.Context, planID string) (*PlanRow, error) {
	var p PlanRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, type, style, status, beds, baths, half_baths,
		       main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
		       garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
		       notes, deleted_at, created_at, updated_at, created_by, updated_by
		FROM plans
		WHERE id = $1 AND deleted_at IS NULL
	`, planID).Scan(p.scanFields()...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetPlanBySlug(ctx context.Context, slug string) (*PlanRow, error) {
	var p PlanRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, slug, type, style, status, beds, baths, half_baths,
		       main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
		       garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
		       notes, deleted_at, created_at, updated_at, created_by, updated_by
		FROM plans
		WHERE slug = $1 AND deleted_at IS NULL
	`, slug).Scan(p.scanFields()...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

type CreatePlanInput struct {
	Name              string
	Type              *string
	Style             *string
	Beds              *int
	Baths             *int
	HalfBaths         *int
	MainSF            *int
	UpperSF           *int
	LowerSF           *int
	PorchDeckSF       *int
	GarageSF          *int
	GarageApartmentSF *int
	UnfinishedSF      *int
	HeatedSF          *int
	TotalSF           *int
	Notes             *string
	CreatedBy         string
}

func (s *Store) CreatePlan(ctx context.Context, input CreatePlanInput) (*PlanRow, error) {
	baseSlug := generateSlug(input.Name)
	slug, err := s.makeUniqueSlug(ctx, baseSlug)
	if err != nil {
		return nil, err
	}

	var p PlanRow
	err = s.pool.QueryRow(ctx, `
		INSERT INTO plans (
			name, slug, type, style, beds, baths, half_baths,
			main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
			garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
			notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING 
			id, name, slug, type, style, status, beds, baths, half_baths,
			main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
			garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
			notes, deleted_at, created_at, updated_at, created_by, updated_by
	`,
		input.Name, slug, input.Type, input.Style, input.Beds, input.Baths, input.HalfBaths,
		input.MainSF, input.UpperSF, input.LowerSF, input.PorchDeckSF, input.GarageSF,
		input.GarageApartmentSF, input.UnfinishedSF, input.HeatedSF, input.TotalSF,
		input.Notes, input.CreatedBy,
	).Scan(p.scanFields()...)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type UpdatePlanInput struct {
	Name              string
	Type              *string
	Style             *string
	Beds              *int
	Baths             *int
	HalfBaths         *int
	MainSF            *int
	UpperSF           *int
	LowerSF           *int
	PorchDeckSF       *int
	GarageSF          *int
	GarageApartmentSF *int
	UnfinishedSF      *int
	HeatedSF          *int
	TotalSF           *int
	Notes             *string
	UpdatedBy         string
}

func (s *Store) UpdatePlan(ctx context.Context, planID string, input UpdatePlanInput) (*PlanRow, error) {
	var p PlanRow
	err := s.pool.QueryRow(ctx, `
		UPDATE plans SET
			name = $2,
			type = $3,
			style = $4,
			beds = $5,
			baths = $6,
			half_baths = $7,
			main_sf = $8,
			upper_sf = $9,
			lower_sf = $10,
			porch_deck_sf = $11,
			garage_sf = $12,
			garage_apartment_sf = $13,
			unfinished_sf = $14,
			heated_sf = $15,
			total_sf = $16,
			notes = $17,
			updated_by = $18,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING 
			id, name, slug, type, style, status, beds, baths, half_baths,
			main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
			garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
			notes, deleted_at, created_at, updated_at, created_by, updated_by
	`,
		planID, input.Name, input.Type, input.Style, input.Beds, input.Baths, input.HalfBaths,
		input.MainSF, input.UpperSF, input.LowerSF, input.PorchDeckSF, input.GarageSF,
		input.GarageApartmentSF, input.UnfinishedSF, input.HeatedSF, input.TotalSF,
		input.Notes, input.UpdatedBy,
	).Scan(p.scanFields()...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *Store) SoftDeletePlan(ctx context.Context, planID string) error {
	result, err := s.pool.Exec(ctx,
		"UPDATE plans SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		planID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) RestorePlan(ctx context.Context, planID string) error {
	result, err := s.pool.Exec(ctx,
		"UPDATE plans SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL",
		planID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) DuplicatePlan(ctx context.Context, sourcePlanID string, newName string, createdBy string) (*PlanRow, error) {
	source, err := s.GetPlanByID(ctx, sourcePlanID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, pgx.ErrNoRows
	}

	baseSlug := generateSlug(newName)
	slug, err := s.makeUniqueSlug(ctx, baseSlug)
	if err != nil {
		return nil, err
	}

	var p PlanRow
	err = s.pool.QueryRow(ctx, `
		INSERT INTO plans (
			name, slug, type, style, status, beds, baths, half_baths,
			main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
			garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
			notes, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING 
			id, name, slug, type, style, status, beds, baths, half_baths,
			main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
			garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
			notes, deleted_at, created_at, updated_at, created_by, updated_by
	`,
		newName, slug, source.Type, source.Style, "incomplete", source.Beds, source.Baths, source.HalfBaths,
		source.MainSF, source.UpperSF, source.LowerSF, source.PorchDeckSF, source.GarageSF,
		source.GarageApartmentSF, source.UnfinishedSF, source.HeatedSF, source.TotalSF,
		source.Notes, createdBy,
	).Scan(p.scanFields()...)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) FlagPlan(ctx context.Context, planID string) error {
	result, err := s.pool.Exec(ctx,
		"UPDATE plans SET status = 'flagged' WHERE id = $1 AND deleted_at IS NULL",
		planID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) UnflagPlan(ctx context.Context, planID string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE plans SET status = CASE 
			WHEN EXISTS (
				SELECT 1 FROM files 
				WHERE plan_id = $1 
				AND category = 'website' 
				AND slot IN ('render-front', 'elevation-front', 'elevation-left', 
							 'elevation-rear', 'elevation-right', 'floor-plan-main', 'poster')
				GROUP BY plan_id
				HAVING COUNT(DISTINCT slot) = 7
			) THEN 'complete'
			ELSE 'incomplete'
		END
		WHERE id = $1 AND deleted_at IS NULL AND status = 'flagged'
	`, planID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetFilledSlots(ctx context.Context, planID string) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT slot FROM files 
		WHERE plan_id = $1 AND category = 'website' AND slot IS NOT NULL
	`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []string
	for rows.Next() {
		var slot string
		if err := rows.Scan(&slot); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slots, nil
}

func (s *Store) RecalculatePlanStatus(ctx context.Context, planID string) (string, error) {
	var status string
	err := s.pool.QueryRow(ctx, `
		UPDATE plans SET status = CASE 
			WHEN status = 'flagged' THEN 'flagged'
			WHEN EXISTS (
				SELECT 1 FROM files 
				WHERE plan_id = $1 
				AND category = 'website' 
				AND slot IN ('render-front', 'elevation-front', 'elevation-left', 
							 'elevation-rear', 'elevation-right', 'floor-plan-main', 'poster')
				GROUP BY plan_id
				HAVING COUNT(DISTINCT slot) = 7
			) THEN 'complete'
			ELSE 'incomplete'
		END
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING status
	`, planID).Scan(&status)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return status, nil
}

func (s *Store) LogActivity(ctx context.Context, userID *uuid.UUID, planID *uuid.UUID, action string, detail map[string]interface{}) error {
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO activity_log (user_id, plan_id, action, detail)
		VALUES ($1, $2, $3, $4)
	`, userID, planID, action, detailJSON)

	return err
}

type DashboardStats struct {
	Total      int `json:"total"`
	Complete   int `json:"complete"`
	Incomplete int `json:"incomplete"`
	Flagged    int `json:"flagged"`
}

func (s *Store) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats DashboardStats

	err := s.pool.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'complete') as complete,
			COUNT(*) FILTER (WHERE status = 'incomplete') as incomplete,
			COUNT(*) FILTER (WHERE status = 'flagged') as flagged
		FROM plans
		WHERE deleted_at IS NULL
	`).Scan(&stats.Total, &stats.Complete, &stats.Incomplete, &stats.Flagged)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (s *Store) GetRecentPlans(ctx context.Context, limit int) ([]PlanRow, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id, name, slug, type, style, status, beds, baths, half_baths,
		       main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
		       garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
		       notes, deleted_at, created_at, updated_at, created_by, updated_by
		FROM plans
		WHERE deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []PlanRow
	for rows.Next() {
		var p PlanRow
		err := rows.Scan(p.scanFields()...)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func (s *Store) GetRecentlyImportedPlans(ctx context.Context) ([]PlanRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, slug, type, style, status, beds, baths, half_baths,
		       main_sf, upper_sf, lower_sf, porch_deck_sf, garage_sf,
		       garage_apartment_sf, unfinished_sf, heated_sf, total_sf,
		       notes, deleted_at, created_at, updated_at, created_by, updated_by
		FROM plans
		WHERE deleted_at IS NULL
		  AND created_at >= NOW() - INTERVAL '24 hours'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []PlanRow
	for rows.Next() {
		var p PlanRow
		err := rows.Scan(p.scanFields()...)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}
