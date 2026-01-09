package entity

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `db:"id" json:"id"`
	Email    string    `db:"email" json:"email"`
	Fullname string    `db:"full_name" json:"full_name"`
	Username string    `db:"username" json:"username"`
	Password string    `db:"password" json:"password"`
	Active   bool      `db:"is_active" json:"is_active"`
}

type AuthClaims struct {
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	Fullname   string      `json:"full_name"`
	Roles      []uuid.UUID `json:"role_ids"`
	RoleString []string    `json:"roles"`
	Permission []string    `json:"permissions"`
	UserID     uuid.UUID   `json:"user_id"`
}
