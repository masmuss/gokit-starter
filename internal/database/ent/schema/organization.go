// Package schema defines the Ent schema definitions for core domain entities.
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// Organization holds the schema definition for the organizations entity.
type Organization struct {
	ent.Schema
}

// Mixin returns Organization mixins.
func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields returns Organization fields.
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("name").
			MaxLen(128).
			NotEmpty(),
		field.String("code").
			MaxLen(16).
			NotEmpty(),
		field.String("type").
			NotEmpty(),
		field.String("status").
			NotEmpty(),
	}
}

// Edges returns Organization edges.
func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Organization.Type).
			From("parent").
			Unique().
			Field("parent_id"),

		edge.To("users", User.Type),
	}
}
