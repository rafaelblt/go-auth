package ed25519

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Keyring struct {
	mu      sync.Mutex
	keys    KeyStore
	current atomic.Pointer[key]
}

type KeyringConfig struct {
	KeyStore KeyStore
}

func NewKeyring(ctx context.Context, cfg KeyringConfig) (*Keyring, error) {
	if cfg.KeyStore == nil {
		return nil, errors.New("key store nil")
	}

	keyring := Keyring{
		keys: cfg.KeyStore,
	}

	current, err := keyring.keys.LoadCurrent(ctx)
	if err != nil {
		return nil, fmt.Errorf("key store load current failed: %w", err)
	}
	if current == nil {
		current, err = newKey()
		if err != nil {
			return nil, fmt.Errorf("new key failed: %w", err)
		}
		keyring.keys.UpdateCurrent(ctx, current)
	}
	keyring.current.Store(current)

	return &keyring, nil
}

func (k *Keyring) SigningKey() *key {
	return k.current.Load()
}

func (k *Keyring) PublicKeyByID(id string) (ed25519.PublicKey, error) {
	current := k.current.Load()
	if id != current.id {
		return nil, errKidUnknown
	}
	return current.public, nil
}

func (k *Keyring) PublicKeys() []port.PublicKey {
	current := k.current.Load()
	return []port.PublicKey{current.dto}
}

func (k *Keyring) Rotate(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	next, err := newKey()
	if err != nil {
		return fmt.Errorf("new key failed: %w", err)
	}

	if err = k.keys.UpdateCurrent(ctx, next); err != nil {
		return fmt.Errorf("key store update current failed: %w", err)
	}

	k.current.Store(next)
	return nil
}
