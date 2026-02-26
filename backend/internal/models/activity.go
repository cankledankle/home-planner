package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	ID        uuid.UUID       `json:"id"`
	UserID    *uuid.UUID      `json:"user_id"`
	PlanID    *uuid.UUID      `json:"plan_id"`
	Action    string          `json:"action"`
	Detail    json.RawMessage `json:"detail"`
	CreatedAt time.Time       `json:"created_at"`
	User      *User           `json:"user,omitempty"`
	Plan      *Plan           `json:"plan,omitempty"`
}

type ActivityAction string

const (
	ActionPlanCreated     ActivityAction = "plan.created"
	ActionPlanUpdated     ActivityAction = "plan.updated"
	ActionPlanDeleted     ActivityAction = "plan.deleted"
	ActionPlanRestored    ActivityAction = "plan.restored"
	ActionPlanFlagged     ActivityAction = "plan.flagged"
	ActionPlanUnflagged   ActivityAction = "plan.unflagged"
	ActionPlanDuplicated  ActivityAction = "plan.duplicated"
	ActionFileUploaded    ActivityAction = "file.uploaded"
	ActionFileDeleted     ActivityAction = "file.deleted"
	ActionFileSlotChanged ActivityAction = "file.slot_changed"
	ActionUserCreated     ActivityAction = "user.created"
	ActionUserUpdated     ActivityAction = "user.updated"
	ActionUserDeleted     ActivityAction = "user.deleted"
	ActionAuthLogin       ActivityAction = "auth.login"
	ActionAuthLogout      ActivityAction = "auth.logout"
	ActionExportCSV       ActivityAction = "export.csv"
	ActionExportZIP       ActivityAction = "export.zip"
)
