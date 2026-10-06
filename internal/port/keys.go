package port

import (
	"context"
	"time"
)

type PublicKeyProvider interface {
	PublicKeys() []PublicKey
}

type PublicKey struct {
	ID        string
	Type      string
	Algorithm string
	Curve     string
	Key       []byte
}

type SigningKeyStore interface {
	// List returns every stored key, by generation, lowest first.
	List(context.Context) ([]StoredSigningKey, error)
	// Add stores key. It applies only while no key holds its generation:
	// when another instance stored that generation first, Add changes
	// nothing and returns nil, and the caller lists the keys again.
	Add(context.Context, StoredSigningKey) error
	// DeleteBefore deletes every key whose generation is lower than
	// generation.
	DeleteBefore(ctx context.Context, generation int64) error
}

// StoredSigningKey is a signing key as stored. Seed is its Ed25519 private
// key, in the 32-byte form of RFC 8032, or that seed as the keyring sealed it
// when an encryption key is set. Either way, it is never logged.
type StoredSigningKey struct {
	Generation int64
	Seed       []byte
	ActiveAt   time.Time
	CreatedAt  time.Time
}
