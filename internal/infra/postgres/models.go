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
	ID        string     `db:"id"`
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

func mapSessionToModel(sess *session.Session) (sessionModel, error) {
	if sess == nil {
		return sessionModel{}, errors.New("session nil")
	}
	if sess.IsZero() {
		return sessionModel{}, errors.New("session zero")
	}

	id := sess.ID().String()
	userID := sess.UserID().String()
	issuedAt := sess.IssuedAt()
	var revokedAtPtr *time.Time

	revokedAt, ok := sess.RevokedAt()
	if ok {
		revokedAtPtr = &revokedAt
	}

	model := sessionModel{
		ID:        id,
		UserID:    userID,
		IssuedAt:  issuedAt,
		RevokedAt: revokedAtPtr,
	}

	return model, nil
}

func mapSessionToEntity(model sessionModel) (*session.Session, error) {
	id, err := session.ParseSessionID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("parse session id failed: %w", err)
	}
	userID, err := user.ParseID(model.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id failed: %w", err)
	}

	sess, err := session.RestoreSession(session.SessionRestoreParams{
		ID:        id,
		UserID:    userID,
		IssuedAt:  model.IssuedAt,
		RevokedAt: model.RevokedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore session failed: %w", err)
	}

	return sess, nil
}

