package postgres

import "time"

type model struct {
	ID        string    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type userModel struct {
	model
	Username  string    `db:"username"`
	Status    string    `db:"status"`
}

type credentialModel struct {
	model
	UserID   string `db:"user_id"`
	Kind     string `db:"kind"`
	Provider string `db:"provider"`
	Secret   string `db:"secret"`
}
