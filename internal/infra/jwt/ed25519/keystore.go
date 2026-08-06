package ed25519

import (
	"context"
	"sync"
	"sync/atomic"
)

type KeyStore interface {
	LoadCurrent(context.Context) (*key, error)
	UpdateCurrent(context.Context, *key) error
}

type KeyStoreInMemory struct {
	mu      sync.Mutex
	current atomic.Pointer[key]
}

func NewKeyStoreInMemory() *KeyStoreInMemory {
	return &KeyStoreInMemory{}
}

func (ks *KeyStoreInMemory) LoadCurrent(ctx context.Context) (*key, error) {
	return ks.current.Load(), nil
}

func (ks *KeyStoreInMemory) UpdateCurrent(ctx context.Context, key *key) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.current.Store(key)
	return nil
}
