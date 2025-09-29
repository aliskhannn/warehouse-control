package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ItemAction string

const (
	ActionInsert ItemAction = "INSERT"
	ActionUpdate ItemAction = "UPDATE"
	ActionDelete ItemAction = "DELETE"
)

type ItemHistory struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	ItemID    uuid.UUID       `db:"item_id" json:"item_id"`
	Action    ItemAction      `db:"action" json:"action"`
	ChangedBy uuid.UUID       `db:"changed_by" json:"changed_by"`
	ChangedAt time.Time       `db:"changed_at" json:"changed_at"`
	OldData   json.RawMessage `db:"old_data,omitempty" json:"old_data,omitempty"`
	NewData   json.RawMessage `db:"new_data,omitempty" json:"new_data,omitempty"`
}
