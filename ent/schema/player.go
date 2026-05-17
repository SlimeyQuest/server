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
	}
}

// Indexes of the Player.
func (Player) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "external_id").
			Unique(),
	}
}
