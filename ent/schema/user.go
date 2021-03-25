package schema

import (
	"time"

	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StructTag(`json:"id,primary_key"`),
		field.String("username").StructTag(`json:"username"`),
		field.String("password_digest").StructTag(`json:"password_digest"`),
		field.String("nickname").StructTag(`json:"nickname"`),
		field.String("status").StructTag(`json:"status"`),
		field.String("avatar").StructTag(`json:"avatar" size:"1000"`),
		field.Time("created_at").StructTag(`json:"created_at"`).Default(time.Now),
		field.Time("updated_at").StructTag(`json:"updated_at"`).Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").StructTag(`json:"deleted_at"`),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
