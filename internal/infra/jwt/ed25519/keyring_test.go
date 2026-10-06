package ed25519

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

// t0 is when the keys of a test start signing, unless the test says otherwise.
var t0 = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

// keyringConfigForTest uses the windows bootstrap uses.
func keyringConfigForTest(store port.SigningKeyStore, clock port.Clock) KeyringConfig {
	return KeyringConfig{
		KeyStore:         store,
		Clock:            clock,
		RotationInterval: 7 * 24 * time.Hour,
		PublishBefore:    24 * time.Hour,
		PublishAfter:     25 * time.Hour,
	}
}

func storedKeyForTest(t *testing.T, generation int64, activeAt time.Time) port.StoredSigningKey {
	t.Helper()

	seed, err := newSeed()
	require.NoError(t, err)

	return port.StoredSigningKey{
		Generation: generation,
		Seed:       seed,
		ActiveAt:   activeAt,
		CreatedAt:  activeAt,
	}
}

func restoredKeyForTest(t *testing.T, stored port.StoredSigningKey) *key {
	t.Helper()

	key, err := restoreKey(stored)
	require.NoError(t, err)

	return key
}

func clockAt(now time.Time) *porttest.FakeClock {
	clock := porttest.NewFakeClock()
	clock.SetNow(now)
	return clock
}

func newKeyringForTest(t *testing.T, store port.SigningKeyStore, clock port.Clock) *Keyring {
	t.Helper()

	keyring, err := NewKeyring(t.Context(), keyringConfigForTest(store, clock))
	require.NoError(t, err)

	return keyring
}

// keyringForTest returns a keyring that started over an empty store.
func keyringForTest(t *testing.T) *Keyring {
	t.Helper()
	return newKeyringForTest(t, porttest.NewFakeSigningKeyStore(), porttest.NewFakeClock())
}

// keyringAndCurrentForTest returns a keyring that loaded one stored key,
// signing now, and that key.
func keyringAndCurrentForTest(t *testing.T) (*Keyring, *key) {
	t.Helper()

	clock := porttest.NewFakeClock()
	stored := storedKeyForTest(t, 1, clock.Now())
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(stored)

	return newKeyringForTest(t, store, clock), restoredKeyForTest(t, stored)
}

func dtosOf(keys ...*key) []port.PublicKey {
	dtos := make([]port.PublicKey, 0, len(keys))
	for _, key := range keys {
		dtos = append(dtos, key.dto)
	}
	return dtos
}

func generationsOf(stored []port.StoredSigningKey) []int64 {
	generations := make([]int64, 0, len(stored))
	for _, key := range stored {
		generations = append(generations, key.Generation)
	}
	return generations
}

// spySigningKeyStore wraps a fake store. It records the method and the
// context of every call, and lets a test act before an Add or make
// DeleteBefore fail.
type spySigningKeyStore struct {
	*porttest.FakeSigningKeyStore
	calls     []spyCall
	beforeAdd func()
	deleteErr error
}

type spyCall struct {
	method string
	ctx    context.Context
}

func newSpySigningKeyStore() *spySigningKeyStore {
	return &spySigningKeyStore{FakeSigningKeyStore: porttest.NewFakeSigningKeyStore()}
}

func (s *spySigningKeyStore) List(ctx context.Context) ([]port.StoredSigningKey, error) {
	s.calls = append(s.calls, spyCall{"List", ctx})
	return s.FakeSigningKeyStore.List(ctx)
}

func (s *spySigningKeyStore) Add(ctx context.Context, key port.StoredSigningKey) error {
	s.calls = append(s.calls, spyCall{"Add", ctx})
	if s.beforeAdd != nil {
		s.beforeAdd()
	}
	return s.FakeSigningKeyStore.Add(ctx, key)
}

func (s *spySigningKeyStore) DeleteBefore(ctx context.Context, generation int64) error {
	s.calls = append(s.calls, spyCall{"DeleteBefore", ctx})
	if s.deleteErr != nil {
		return s.deleteErr
	}
	return s.FakeSigningKeyStore.DeleteBefore(ctx, generation)
}

func (s *spySigningKeyStore) assertCalledWith(t *testing.T, ctx context.Context, methods ...string) {
	t.Helper()

	called := make([]string, 0, len(s.calls))
	for _, call := range s.calls {
		called = append(called, call.method)
		assert.Equal(t, ctx, call.ctx, "context of %s", call.method)
	}
	assert.Equal(t, methods, called)
}

type contextKey struct{}

// TESTS

func TestNewKeyring_ReturnsError(t *testing.T) {
	failing := porttest.NewFakeSigningKeyStore()
	failing.SetError(errors.New("store failed"))

	corrupt := porttest.NewFakeSigningKeyStore()
	corruptKey := storedKeyForTest(t, 1, t0)
	corruptKey.Seed = corruptKey.Seed[:31]
	corrupt.Insert(corruptKey)

	with := func(override func(cfg *KeyringConfig)) KeyringConfig {
		cfg := keyringConfigForTest(porttest.NewFakeSigningKeyStore(), clockAt(t0))
		override(&cfg)
		return cfg
	}

	testCases := []struct {
		desc string
		cfg  KeyringConfig
	}{
		{"store nil", with(func(cfg *KeyringConfig) { cfg.KeyStore = nil })},
		{"clock nil", with(func(cfg *KeyringConfig) { cfg.Clock = nil })},
		{"rotation interval zero", with(func(cfg *KeyringConfig) { cfg.RotationInterval = 0 })},
		{"publish before zero", with(func(cfg *KeyringConfig) { cfg.PublishBefore = 0 })},
		{"publish before equal to rotation interval", with(func(cfg *KeyringConfig) { cfg.PublishBefore = cfg.RotationInterval })},
		{"publish after zero", with(func(cfg *KeyringConfig) { cfg.PublishAfter = 0 })},
		{"store fails", with(func(cfg *KeyringConfig) { cfg.KeyStore = failing })},
		{"store holds a key with a 31-byte seed", with(func(cfg *KeyringConfig) { cfg.KeyStore = corrupt })},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			keyring, err := NewKeyring(t.Context(), tC.cfg)

			assert.Nil(t, keyring)
			assert.Error(t, err)
		})
	}
}

func TestNewKeyring_AddsTheFirstKey_SigningAtOnce_WhenTheStoreIsEmpty(t *testing.T) {
	store := porttest.NewFakeSigningKeyStore()
	clock := clockAt(t0)

	keyring := newKeyringForTest(t, store, clock)

	require.Len(t, store.Keys(), 1)
	stored := store.Keys()[0]
	assert.Equal(t, int64(1), stored.Generation)
	assert.Equal(t, clock.Now(), stored.ActiveAt)
	assert.Equal(t, clock.Now(), stored.CreatedAt)
	first := restoredKeyForTest(t, stored)
	assert.Equal(t, first.id, keyring.SigningKey().id)
	assert.Equal(t, dtosOf(first), keyring.PublicKeys())
}

func TestNewKeyring_LoadsTheStoredKeys_WithoutAddingOne(t *testing.T) {
	clock := clockAt(t0)
	first := storedKeyForTest(t, 1, t0.Add(-time.Hour))
	second := storedKeyForTest(t, 2, t0.Add(time.Hour))
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(first)
	store.Insert(second)

	keyring := newKeyringForTest(t, store, clock)

	assert.Equal(t, []port.StoredSigningKey{first, second}, store.Keys(), "store changed")
	assert.Equal(t, dtosOf(restoredKeyForTest(t, first), restoredKeyForTest(t, second)), keyring.PublicKeys())
	assert.Equal(t, restoredKeyForTest(t, first).id, keyring.SigningKey().id)
}

func TestNewKeyring_SignsWithTheRivalKey_WhenAnotherInstanceAddedTheGenerationFirst(t *testing.T) {
	store := newSpySigningKeyStore()
	rival := storedKeyForTest(t, 1, t0)
	store.beforeAdd = func() { store.Insert(rival) }

	keyring := newKeyringForTest(t, store, clockAt(t0))

	assert.Equal(t, []port.StoredSigningKey{rival}, store.Keys())
	assert.Equal(t, restoredKeyForTest(t, rival).id, keyring.SigningKey().id)
}

func TestNewKeyring_PassesItsContextToTheStore(t *testing.T) {
	store := newSpySigningKeyStore()
	ctx := context.WithValue(t.Context(), contextKey{}, "new keyring")

	_, err := NewKeyring(ctx, keyringConfigForTest(store, clockAt(t0)))

	require.NoError(t, err)
	store.assertCalledWith(t, ctx, "List", "Add", "List")
}

func TestNewKeyring_LeavesTheRetiredKeysForSyncToDelete(t *testing.T) {
	cfg := keyringConfigForTest(nil, nil)
	first := storedKeyForTest(t, 1, t0)
	second := storedKeyForTest(t, 2, t0.Add(cfg.RotationInterval))
	store := newSpySigningKeyStore()
	store.Insert(first)
	store.Insert(second)
	store.deleteErr = errors.New("permission denied")

	keyring := newKeyringForTest(t, store, clockAt(second.ActiveAt.Add(cfg.PublishAfter)))

	assert.Equal(t, dtosOf(restoredKeyForTest(t, second)), keyring.PublicKeys())
	assert.Equal(t, []int64{1, 2}, generationsOf(store.Keys()))
	store.assertCalledWith(t, t.Context(), "List")
}

func TestKeyring_Sync_AddsTheNextKey_PublishedBeforeItSigns_WhenRotationIsDue(t *testing.T) {
	clock := clockAt(t0)
	first := storedKeyForTest(t, 1, t0)
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(first)
	keyring := newKeyringForTest(t, store, clock)
	cfg := keyringConfigForTest(store, clock)
	clock.SetNow(t0.Add(cfg.RotationInterval - cfg.PublishBefore))

	err := keyring.Sync(t.Context())

	require.NoError(t, err)
	require.Len(t, store.Keys(), 2)
	second := store.Keys()[1]
	assert.Equal(t, int64(2), second.Generation)
	assert.Equal(t, clock.Now().Add(cfg.PublishBefore), second.ActiveAt)
	assert.Equal(t, clock.Now(), second.CreatedAt)
	assert.Equal(t, dtosOf(restoredKeyForTest(t, first), restoredKeyForTest(t, second)), keyring.PublicKeys())
	assert.Equal(t, restoredKeyForTest(t, first).id, keyring.SigningKey().id)
}

func TestKeyring_Sync_AddsNoKey_BeforeRotationIsDue(t *testing.T) {
	clock := clockAt(t0)
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(storedKeyForTest(t, 1, t0))
	keyring := newKeyringForTest(t, store, clock)
	cfg := keyringConfigForTest(store, clock)
	clock.SetNow(t0.Add(cfg.RotationInterval - cfg.PublishBefore - time.Nanosecond))

	err := keyring.Sync(t.Context())

	require.NoError(t, err)
	assert.Equal(t, []int64{1}, generationsOf(store.Keys()))
}

func TestKeyring_Sync_AddsNoKey_WhileTheNextKeyIsPending(t *testing.T) {
	clock := clockAt(t0)
	cfg := keyringConfigForTest(nil, nil)
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(storedKeyForTest(t, 1, t0))
	store.Insert(storedKeyForTest(t, 2, t0.Add(cfg.RotationInterval)))
	keyring := newKeyringForTest(t, store, clock)
	clock.SetNow(t0.Add(cfg.RotationInterval - time.Nanosecond))

	err := keyring.Sync(t.Context())

	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, generationsOf(store.Keys()))
}

func TestKeyring_SigningKey_SwitchesToTheNextKey_WhenItsTimeComes(t *testing.T) {
	t1 := t0.Add(7 * 24 * time.Hour)
	first := storedKeyForTest(t, 1, t0)
	second := storedKeyForTest(t, 2, t1)
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(first)
	store.Insert(second)
	clock := clockAt(t1.Add(-time.Nanosecond))
	keyring := newKeyringForTest(t, store, clock)

	assert.Equal(t, restoredKeyForTest(t, first).id, keyring.SigningKey().id)

	clock.SetNow(t1)

	assert.Equal(t, restoredKeyForTest(t, second).id, keyring.SigningKey().id)
}

func TestKeyring_SigningKey_ReturnsTheFirstToActivate_WhenNoneIsActiveYet(t *testing.T) {
	first := storedKeyForTest(t, 1, t0.Add(5*time.Second))
	second := storedKeyForTest(t, 2, t0.Add(7*24*time.Hour))
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(second)
	store.Insert(first)

	keyring := newKeyringForTest(t, store, clockAt(t0))

	assert.Equal(t, restoredKeyForTest(t, first).id, keyring.SigningKey().id)
}

func TestKeyring_Sync_KeepsThePreviousKey_UntilPublishAfterHasPassed(t *testing.T) {
	cfg := keyringConfigForTest(nil, nil)
	t1 := t0.Add(cfg.RotationInterval)
	first := storedKeyForTest(t, 1, t0)
	second := storedKeyForTest(t, 2, t1)
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(first)
	store.Insert(second)
	clock := clockAt(t1)
	keyring := newKeyringForTest(t, store, clock)
	clock.SetNow(t1.Add(cfg.PublishAfter - time.Nanosecond))

	err := keyring.Sync(t.Context())

	require.NoError(t, err)
	assert.Equal(t, dtosOf(restoredKeyForTest(t, first), restoredKeyForTest(t, second)), keyring.PublicKeys())
	assert.Equal(t, []int64{1, 2}, generationsOf(store.Keys()))
}

func TestKeyring_Sync_DeletesThePreviousKey_OncePublishAfterHasPassed(t *testing.T) {
	cfg := keyringConfigForTest(nil, nil)
	t1 := t0.Add(cfg.RotationInterval)
	first := storedKeyForTest(t, 1, t0)
	second := storedKeyForTest(t, 2, t1)
	store := porttest.NewFakeSigningKeyStore()
	store.Insert(first)
	store.Insert(second)
	clock := clockAt(t1)
	keyring := newKeyringForTest(t, store, clock)
	clock.SetNow(t1.Add(cfg.PublishAfter))

	err := keyring.Sync(t.Context())

	require.NoError(t, err)
	assert.Equal(t, []int64{2}, generationsOf(store.Keys()))
	assert.Equal(t, dtosOf(restoredKeyForTest(t, second)), keyring.PublicKeys())
	public, err := keyring.PublicKeyByID(restoredKeyForTest(t, first).id)
	assert.Nil(t, public)
	assert.ErrorIs(t, err, errKidUnknown)
}

func TestKeyring_Sync_LoadsKeysAddedByAnotherInstance(t *testing.T) {
	clock := clockAt(t0)
	store := porttest.NewFakeSigningKeyStore()
	keyring := newKeyringForTest(t, store, clock)
	pending := storedKeyForTest(t, 2, t0.Add(time.Hour))
	store.Insert(pending)

	err := keyring.Sync(t.Context())

	require.NoError(t, err)
	assert.Contains(t, keyring.PublicKeys(), restoredKeyForTest(t, pending).dto)
}

func TestKeyring_Sync_KeepsTheLoadedKeys_WhenTheStoreFails(t *testing.T) {
	clock := clockAt(t0)
	store := porttest.NewFakeSigningKeyStore()
	keyring := newKeyringForTest(t, store, clock)
	published := keyring.PublicKeys()
	signing := keyring.SigningKey()
	store.Insert(storedKeyForTest(t, 2, t0.Add(time.Hour)))
	storeErr := errors.New("store failed")
	store.SetError(storeErr)

	err := keyring.Sync(t.Context())

	assert.ErrorIs(t, err, storeErr)
	assert.Equal(t, published, keyring.PublicKeys())
	assert.Equal(t, signing, keyring.SigningKey())
}

func TestKeyring_Sync_PublishesTheLoadedKeys_WhenDeletingTheRetiredOnesFails(t *testing.T) {
	cfg := keyringConfigForTest(nil, nil)
	t1 := t0.Add(cfg.RotationInterval)
	first := storedKeyForTest(t, 1, t0)
	second := storedKeyForTest(t, 2, t1)
	store := newSpySigningKeyStore()
	store.Insert(first)
	store.Insert(second)
	deleteErr := errors.New("permission denied")
	store.deleteErr = deleteErr
	clock := clockAt(t1)
	keyring := newKeyringForTest(t, store, clock)
	clock.SetNow(t1.Add(cfg.PublishAfter))

	err := keyring.Sync(t.Context())

	assert.ErrorIs(t, err, deleteErr)
	assert.Equal(t, dtosOf(restoredKeyForTest(t, second)), keyring.PublicKeys())
	assert.Equal(t, []int64{1, 2}, generationsOf(store.Keys()))
}

func TestKeyring_Sync_PassesItsContextToTheStore(t *testing.T) {
	cfg := keyringConfigForTest(nil, nil)
	t1 := t0.Add(cfg.RotationInterval)
	store := newSpySigningKeyStore()
	store.Insert(storedKeyForTest(t, 1, t0))
	store.Insert(storedKeyForTest(t, 2, t1))
	clock := clockAt(t1)
	keyring := newKeyringForTest(t, store, clock)
	store.calls = nil
	// Late enough to retire the first key and to add the third.
	clock.SetNow(t1.Add(cfg.RotationInterval - cfg.PublishBefore))
	ctx := context.WithValue(t.Context(), contextKey{}, "sync")

	err := keyring.Sync(ctx)

	require.NoError(t, err)
	store.assertCalledWith(t, ctx, "List", "Add", "List", "DeleteBefore")
	assert.Equal(t, []int64{2, 3}, generationsOf(store.Keys()))
}

func TestKeyring_ServesTheKeys_WhileSyncRuns(t *testing.T) {
	store := porttest.NewFakeSigningKeyStore()
	keyring := newKeyringForTest(t, store, clockAt(t0))
	done := make(chan struct{})

	var readers sync.WaitGroup
	for range 4 {
		readers.Go(func() {
			for {
				select {
				case <-done:
					return
				default:
				}
				signing := keyring.SigningKey()
				assert.NotEmpty(t, keyring.PublicKeys())
				_, err := keyring.PublicKeyByID(signing.id)
				assert.NoError(t, err)
			}
		})
	}

	for generation := int64(2); generation <= 50; generation++ {
		store.Insert(storedKeyForTest(t, generation, t0.Add(time.Duration(generation)*time.Hour)))
		require.NoError(t, keyring.Sync(t.Context()))
	}
	close(done)
	readers.Wait()

	assert.Len(t, keyring.PublicKeys(), 50)
}

func TestKeyring_PublicKeyByID_ReturnsAnyPublishedKey(t *testing.T) {
	keyring, current := keyringAndCurrentForTest(t)
	stored := storedKeyForTest(t, 2, current.activeAt.Add(time.Hour))
	keyring.store.(*porttest.FakeSigningKeyStore).Insert(stored)
	require.NoError(t, keyring.Sync(t.Context()))
	pending := restoredKeyForTest(t, stored)

	for _, expected := range []*key{current, pending} {
		public, err := keyring.PublicKeyByID(expected.id)

		require.NoError(t, err)
		assert.Equal(t, expected.public, public)
	}
}

func TestKeyring_PublicKeyByID_ReturnsErrKidUnknown(t *testing.T) {
	keyring := keyringForTest(t)

	key, err := keyring.PublicKeyByID("unknown")

	assert.Nil(t, key)
	assert.ErrorIs(t, err, errKidUnknown)
}
