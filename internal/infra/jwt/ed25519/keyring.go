// Package ed25519 holds the signing keys: the key type, the keyring that loads
// them from a port.SigningKeyStore, rotates them in windows and serves them,
// and the JWT signer. A key's ID is its RFC 7638 thumbprint, derived from the
// key itself, so it needs no storage of its own.
//
// See docs/architecture/tokens.md#signing-keys.
package ed25519

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Keyring struct {
	store            port.SigningKeyStore
	clock            port.Clock
	rotationInterval time.Duration
	publishBefore    time.Duration
	publishAfter     time.Duration

	mu sync.Mutex
	// keys holds the published keys, by generation, and is never empty. Sync
	// replaces the slice, and never changes one already stored.
	keys atomic.Pointer[[]*key]
}

type KeyringConfig struct {
	KeyStore         port.SigningKeyStore
	Clock            port.Clock
	RotationInterval time.Duration
	// PublishBefore is how long a key is published before it starts signing.
	PublishBefore time.Duration
	// PublishAfter is how long a key stays published after the next one
	// starts signing.
	PublishAfter time.Duration
}

func NewKeyring(ctx context.Context, cfg KeyringConfig) (*Keyring, error) {
	if cfg.KeyStore == nil {
		return nil, errors.New("key store nil")
	}
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}
	if cfg.RotationInterval <= 0 {
		return nil, errors.New("rotation interval zero or negative")
	}
	if cfg.PublishBefore <= 0 {
		return nil, errors.New("publish before zero or negative")
	}
	if cfg.PublishAfter <= 0 {
		return nil, errors.New("publish after zero or negative")
	}
	if cfg.PublishBefore >= cfg.RotationInterval {
		return nil, errors.New("publish before not shorter than rotation interval")
	}

	keyring := Keyring{
		store:            cfg.KeyStore,
		clock:            cfg.Clock,
		rotationInterval: cfg.RotationInterval,
		publishBefore:    cfg.PublishBefore,
		publishAfter:     cfg.PublishAfter,
	}

	// The retired keys are left for Sync to delete, so that a delete that
	// fails cannot stop the service from starting.
	if _, err := keyring.load(ctx); err != nil {
		return nil, fmt.Errorf("signing keys load failed: %w", err)
	}

	return &keyring, nil
}

// Sync reads the stored keys, adds the next one when rotation is due, and
// publishes them without the retired ones. Then it deletes the retired keys
// from the store: when that fails, the keys are published all the same.
//
// See docs/development/decisions/0053-signing-keys-are-shared-through-the-database.md.
func (k *Keyring) Sync(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	retiredBelow, err := k.load(ctx)
	if err != nil {
		return err
	}
	if retiredBelow == 0 {
		return nil
	}

	if err := k.store.DeleteBefore(ctx, retiredBelow); err != nil {
		return fmt.Errorf("signing key store delete failed: %w", err)
	}
	return nil
}

// load publishes the stored keys, after adding the next one when it is due,
// and returns the generation below which the stored keys are retired, or 0
// when none is. When it fails, the published keys stay as they were.
func (k *Keyring) load(ctx context.Context) (retiredBelow int64, err error) {
	now := k.clock.Now()

	keys, err := k.list(ctx)
	if err != nil {
		return 0, err
	}

	if next, due := k.next(keys, now); due {
		if err := k.add(ctx, next); err != nil {
			return 0, err
		}
		// Another instance may have stored this generation first, so what is
		// stored decides, not the key generated here.
		if keys, err = k.list(ctx); err != nil {
			return 0, err
		}
	}

	if len(keys) == 0 {
		return 0, errors.New("no signing key stored")
	}

	published, retiredBelow := k.unretired(keys, now)
	k.keys.Store(&published)
	return retiredBelow, nil
}

func (k *Keyring) list(ctx context.Context) ([]*key, error) {
	stored, err := k.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("signing key store list failed: %w", err)
	}

	keys := make([]*key, 0, len(stored))
	for _, s := range stored {
		key, err := restoreKey(s)
		if err != nil {
			return nil, fmt.Errorf("signing key %d restore failed: %w", s.Generation, err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// next says which key to add, if any: the first one, signing at once, when
// none is stored, or else the next one, once the last has PublishBefore left
// to sign, signing when that has passed.
func (k *Keyring) next(keys []*key, now time.Time) (port.StoredSigningKey, bool) {
	if len(keys) == 0 {
		return port.StoredSigningKey{Generation: 1, ActiveAt: now, CreatedAt: now}, true
	}

	last := keys[len(keys)-1]
	if now.Before(last.activeAt.Add(k.rotationInterval - k.publishBefore)) {
		return port.StoredSigningKey{}, false
	}

	next := port.StoredSigningKey{
		Generation: last.generation + 1,
		ActiveAt:   now.Add(k.publishBefore),
		CreatedAt:  now,
	}
	return next, true
}

func (k *Keyring) add(ctx context.Context, next port.StoredSigningKey) error {
	seed, err := newSeed()
	if err != nil {
		return err
	}
	next.Seed = seed

	if err := k.store.Add(ctx, next); err != nil {
		return fmt.Errorf("signing key store add failed: %w", err)
	}
	return nil
}

// unretired drops the keys below the last one that has been signing for
// PublishAfter, and returns the generation they are below, or 0 when it drops
// none. keys are by generation.
func (k *Keyring) unretired(keys []*key, now time.Time) ([]*key, int64) {
	cutoff := now.Add(-k.publishAfter)

	last := 0
	for i, key := range keys {
		if !key.activeAt.After(cutoff) {
			last = i
		}
	}
	if last == 0 {
		return keys, 0
	}
	return keys[last:], keys[last].generation
}

// SigningKey returns the key whose activeAt is the latest not after now. While
// none is active yet, which only a peer's clock running ahead can cause, it
// returns the one that activates first.
func (k *Keyring) SigningKey() *key {
	now := k.clock.Now()

	var signing, first *key
	for _, key := range *k.keys.Load() {
		if !key.activeAt.After(now) && (signing == nil || key.activeAt.After(signing.activeAt)) {
			signing = key
		}
		if first == nil || key.activeAt.Before(first.activeAt) {
			first = key
		}
	}

	if signing == nil {
		return first
	}
	return signing
}

func (k *Keyring) PublicKeyByID(id string) (ed25519.PublicKey, error) {
	for _, key := range *k.keys.Load() {
		if key.id == id {
			return key.public, nil
		}
	}
	return nil, errKidUnknown
}

func (k *Keyring) PublicKeys() []port.PublicKey {
	keys := *k.keys.Load()

	dtos := make([]port.PublicKey, 0, len(keys))
	for _, key := range keys {
		dtos = append(dtos, key.dto)
	}
	return dtos
}
