package port

import "context"

type PublicKeyProvider interface {
	PublicKeys(ctx context.Context) []PublicKey
}

type PublicKey struct {
	ID        string
	Type      string
	Algorithm string
	Curve     string
	Key       []byte
}
