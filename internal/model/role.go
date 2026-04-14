package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:t_role,alias:tr"`

	ID        int64     `bun:"id,pk,autoincrement"`
	RoleName  string    `bun:"role_name,notnull"`
	RoleDesc  string    `bun:"role_desc,notnull"`
	RoleCode  string    `bun:"role_code,notnull,unique"`
	Status    int16     `bun:"status,notnull,default:1"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}
