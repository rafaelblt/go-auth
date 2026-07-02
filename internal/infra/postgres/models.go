package postgres

import "time"

type model struct {
	ID string `db:"id"`
}

type userModel struct {
	model
	Username  string    `db:"username"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type credentialModel struct {
	model
	UserID    string    `db:"user_id"`
	Kind      string    `db:"kind"`
	Provider  string    `db:"provider"`
	Secret    string    `db:"secret"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type sessionModel struct {
	model
	UserID    string     `db:"user_id"`
	IssuedAt  time.Time  `db:"issued_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

type refreshTokenModel struct {
	model
	SessionID string     `db:"session_id"`
	ParentID  *string    `db:"parent_id"`
	Hash      []byte     `db:"hash"`
	IssuedAt  time.Time  `db:"issued_at"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
}
