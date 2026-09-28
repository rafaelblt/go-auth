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
	ID       string `json:"id"`
	Username string `json:"username"`
}

type AccessToken struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"`
}

type RefreshToken struct {
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"`
}
