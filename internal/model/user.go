package model

import (
	"time"

	"github.com/uptrace/bun"
)

// User represents the users table in the database.
type User struct {
	bun.BaseModel `bun:"table:t_user,alias:tu"`

	ID        int64      `bun:",pk,autoincrement" json:"id"`
	Name      string     `bun:",notnull" json:"name"`
	Email     string     `bun:",notnull,unique" json:"email"`
	CreatedAt *time.Time `bun:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at"`
	RoleID    *int64     `bun:"role_id" json:"role_id,omitempty"`
	DeletedAt *time.Time `bun:"deleted_at"`

	// RELATION
	Role *Role `bun:"rel:belongs-to,join:role_id=id"`
}
