package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Player holds the schema definition for the players entity.
type Player struct {
	ent.Schema
}

// Fields of the Player.
func (Player) Fields() []ent.Field {
	return []ent.Field{
		field.String("platform").
			NotEmpty(),
		field.String("external_id").
			NotEmpty(),
		field.String("nickname").
			NotEmpty(),
		field.Int32("level").
			Default(1),
		field.Int64("exp").
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("last_login_at").
			Optional().
			Nillable(),
		field.Int64("gold").
			Default(0),
		field.Int64("gems").
			Default(0),
		field.Int32("adventure_id").
			Default(1),
		field.Int32("stage_index").
			Default(1),
		field.Int32("highest_stage_cleared").
			Default(0),
		field.Time("last_claim_at").
			Optional().
			Nillable(),
		field.JSON("equipment_json", map[string]any{}).
			Default(map[string]any{}),
		field.JSON("cleared_milestones", []int32{}).
			Default([]int32{}),
	}
}

// Indexes of the Player.
func (Player) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "external_id").
			Unique(),
	}
}
