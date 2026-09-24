package e2e

import "time"

// Errors

type ErrorResponseBody struct {
	Error ErrorData `json:"error"`
}

type ErrorData struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Fields  []FieldErrorData `json:"fields"`
}

type FieldErrorData struct {
	Field   string         `json:"field"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details"`
}

// Entities

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccessToken struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RefreshToken struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}
