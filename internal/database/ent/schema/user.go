package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// User holds the schema definition for the users entity.
type User struct {
	ent.Schema
}

// Mixin returns User mixins.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Fields returns User fields.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("outlet_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("name").
			MaxLen(128).
			NotEmpty(),
		field.String("email").
			MaxLen(128).
			NotEmpty(),
		field.String("password_hash").
			MaxLen(255).
			NotEmpty(),
		field.Enum("status").
			Values(
				"active",
				"inactive",
				"suspended",
				"banned",
			).
			Default("active"),
		field.Enum("role").
			Values(
				"admin",
				"member",
			).
			Default("member"),
	}
}

// Edges returns User edges.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("users").
			Field("organization_id").
			Unique().
			Required(),
	}
}
