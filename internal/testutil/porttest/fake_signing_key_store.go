package porttest

import (
	"cmp"
	"context"
	"slices"

	"github.com/rafaelblt/go-auth/internal/port"
)

type FakeSigningKeyStore struct {
	keys map[int64]port.StoredSigningKey
	err  error
}

func NewFakeSigningKeyStore() *FakeSigningKeyStore {
	return &FakeSigningKeyStore{
		keys: make(map[int64]port.StoredSigningKey),
	}
}

func (s *FakeSigningKeyStore) List(ctx context.Context) ([]port.StoredSigningKey, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.Keys(), nil
}

func (s *FakeSigningKeyStore) Add(ctx context.Context, key port.StoredSigningKey) error {
	if s.err != nil {
		return s.err
	}
	if _, taken := s.keys[key.Generation]; taken {
		return nil
	}
	s.Insert(key)
	return nil
}

func (s *FakeSigningKeyStore) DeleteBefore(ctx context.Context, generation int64) error {
	if s.err != nil {
		return s.err
	}
	for stored := range s.keys {
		if stored < generation {
			delete(s.keys, stored)
		}
	}
	return nil
}

// Insert stores key as another instance would, replacing any key of the same
// generation, whatever SetError holds.
func (s *FakeSigningKeyStore) Insert(key port.StoredSigningKey) {
	key.Seed = slices.Clone(key.Seed)
	s.keys[key.Generation] = key
}

// Keys returns the stored keys, by generation, lowest first.
func (s *FakeSigningKeyStore) Keys() []port.StoredSigningKey {
	keys := make([]port.StoredSigningKey, 0, len(s.keys))
	for _, key := range s.keys {
		key.Seed = slices.Clone(key.Seed)
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b port.StoredSigningKey) int {
		return cmp.Compare(a.Generation, b.Generation)
	})
	return keys
}

func (s *FakeSigningKeyStore) SetError(err error) {
	s.err = err
}
