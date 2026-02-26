package db

import (
	"context"
	"fmt"
	"time"
)

type ActivityRow struct {
	ID        string
	UserID    *string
	PlanID    *string
	Action    string
	Detail    []byte
	CreatedAt time.Time
}

type ActivityWithRelations struct {
	ID        string
	UserID    *string
	PlanID    *string
	Action    string
	Detail    []byte
	CreatedAt time.Time
	UserName  *string
	PlanName  *string
}

type ActivityListFilters struct {
	UserID *string
	PlanID *string
	Action string
	Page   int
	Limit  int
}

type PaginatedActivities struct {
	Activities []ActivityWithRelations
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

func ListActivities(ctx context.Context, filters ActivityListFilters) (*PaginatedActivities, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 50
	}

	whereClause := "WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if filters.UserID != nil {
		whereClause += fmt.Sprintf(" AND al.user_id = $%d", argIdx)
		args = append(args, *filters.UserID)
		argIdx++
	}

	if filters.PlanID != nil {
		whereClause += fmt.Sprintf(" AND al.plan_id = $%d", argIdx)
		args = append(args, *filters.PlanID)
		argIdx++
	}

	if filters.Action != "" {
		whereClause += fmt.Sprintf(" AND al.action = $%d", argIdx)
		args = append(args, filters.Action)
		argIdx++
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM activity_log al
		%s
	`, whereClause)

	var total int
	err := Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT 
			al.id, al.user_id, al.plan_id, al.action, al.detail, al.created_at,
			u.name as user_name,
			p.name as plan_name
		FROM activity_log al
		LEFT JOIN users u ON al.user_id = u.id
		LEFT JOIN plans p ON al.plan_id = p.id
		%s
		ORDER BY al.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filters.Limit, (filters.Page-1)*filters.Limit)

	rows, err := Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []ActivityWithRelations
	for rows.Next() {
		var a ActivityWithRelations
		err := rows.Scan(
			&a.ID, &a.UserID, &a.PlanID, &a.Action, &a.Detail, &a.CreatedAt,
			&a.UserName, &a.PlanName,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + filters.Limit - 1) / filters.Limit

	return &PaginatedActivities{
		Activities: activities,
		Total:      total,
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: totalPages,
	}, nil
}

func ListActivitiesForPlan(ctx context.Context, planID string, page, limit int) (*PaginatedActivities, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	countQuery := `SELECT COUNT(*) FROM activity_log WHERE plan_id = $1`

	var total int
	err := Pool.QueryRow(ctx, countQuery, planID).Scan(&total)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			al.id, al.user_id, al.plan_id, al.action, al.detail, al.created_at,
			u.name as user_name,
			p.name as plan_name
		FROM activity_log al
		LEFT JOIN users u ON al.user_id = u.id
		LEFT JOIN plans p ON al.plan_id = p.id
		WHERE al.plan_id = $1
		ORDER BY al.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := Pool.Query(ctx, query, planID, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []ActivityWithRelations
	for rows.Next() {
		var a ActivityWithRelations
		err := rows.Scan(
			&a.ID, &a.UserID, &a.PlanID, &a.Action, &a.Detail, &a.CreatedAt,
			&a.UserName, &a.PlanName,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit

	return &PaginatedActivities{
		Activities: activities,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
