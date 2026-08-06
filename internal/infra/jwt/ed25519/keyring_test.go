package ed25519

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

type FakeKeyStore struct {
	currentKey     *key
	loadContexts   []context.Context
	updateContexts []context.Context
	updateKeys     []*key
}

func NewFakeKeyStore() *FakeKeyStore {
	fake := FakeKeyStore{
		currentKey:     nil,
		loadContexts:   make([]context.Context, 0),
		updateContexts: make([]context.Context, 0),
		updateKeys:     make([]*key, 0),
	}
	return &fake
}

func (fks *FakeKeyStore) LoadCurrent(ctx context.Context) (*key, error) {
	fks.loadContexts = append(fks.loadContexts, ctx)

	return fks.currentKey, nil
}

func (fks *FakeKeyStore) UpdateCurrent(ctx context.Context, key *key) error {
	fks.updateContexts = append(fks.updateContexts, ctx)
	fks.updateKeys = append(fks.updateKeys, key)

	return nil
}

func (fks *FakeKeyStore) ContextsUsedInLoad() []context.Context {
	return fks.loadContexts
}

func (fks *FakeKeyStore) ContextsUsedInUpdate() []context.Context {
	return fks.updateContexts
}

func (fks *FakeKeyStore) UpdatedKeys() []*key {
	return fks.updateKeys
}

func (fks *FakeKeyStore) SetCurrent(k *key) {
	fks.currentKey = k
}

func keyringForTest(t *testing.T) *Keyring {
	t.Helper()
	keyring, _ := keyringAndCurrentForTest(t)
	return keyring
}

func keyringAndCurrentForTest(t *testing.T) (*Keyring, *key) {
	t.Helper()

	curr, err := newKey()
	require.NoError(t, err)

	keystore := NewFakeKeyStore()
	keystore.SetCurrent(curr)

	keyring, err := NewKeyring(t.Context(), KeyringConfig{keystore})
	require.NoError(t, err)

	return keyring, curr
}

func keyringAndFakeKeyStoreForTest(t *testing.T) (*Keyring, *FakeKeyStore) {
	t.Helper()

	curr, err := newKey()
	require.NoError(t, err)

	keystore := NewFakeKeyStore()
	keystore.SetCurrent(curr)

	keyring, err := NewKeyring(t.Context(), KeyringConfig{keystore})
	require.NoError(t, err)

	return keyring, keystore
}

// TESTS

func TestNewKeyring_Return(t *testing.T) {
	fakeKeys := NewFakeKeyStore()

	keyring, err := NewKeyring(t.Context(), KeyringConfig{
		KeyStore: fakeKeys,
	})

	require.NoError(t, err)
	require.NotNil(t, keyring)
}

func TestNewKeyring_UseProvidedKeyStore(t *testing.T) {
	expected := NewFakeKeyStore()

	keyring, err := NewKeyring(t.Context(), KeyringConfig{
		KeyStore: expected,
	})

	require.NotNil(t, keyring)
	require.NoError(t, err)

	require.IsType(t, &FakeKeyStore{}, keyring.keys)
	assert.Equal(t, expected, keyring.keys)
}

func TestNewKeyring_LoadCurrentFromKeyStore_WithContext(t *testing.T) {
	current, err := newKey()
	require.NoError(t, err)
	fakeKeys := NewFakeKeyStore()
	fakeKeys.SetCurrent(current)

	keyring, err := NewKeyring(t.Context(), KeyringConfig{
		KeyStore: fakeKeys,
	})

	require.NotNil(t, keyring)
	require.NoError(t, err)

	ctx := testutil.Only(t, fakeKeys.ContextsUsedInLoad())
	assert.Equal(t, t.Context(), ctx)
}

func TestNewKeyring_LoadCurrentFromKeyStore_AndSetTheSameCurrentInKeyring(t *testing.T) {
	expected, err := newKey()
	require.NoError(t, err)
	fakeKeys := NewFakeKeyStore()
	fakeKeys.SetCurrent(expected)

	keyring, err := NewKeyring(t.Context(), KeyringConfig{
		KeyStore: fakeKeys,
	})

	require.NotNil(t, keyring)
	require.NoError(t, err)
	assert.Equal(t, expected, keyring.current.Load())
}

func TestNewKeyring_UpdateCurrentInKeyStore_WithContext_WhenKeyStoreDoesntHaveCurrentKey(t *testing.T) {
	fakeKeys := NewFakeKeyStore()

	keyring, err := NewKeyring(t.Context(), KeyringConfig{
		KeyStore: fakeKeys,
	})

	require.NotNil(t, keyring)
	require.NoError(t, err)

	ctx := testutil.Only(t, fakeKeys.ContextsUsedInUpdate())
	assert.Equal(t, t.Context(), ctx)
}

func TestNewKeyring_UpdateCurrentInKeyStore_WithTheSameCurrentInKeyring_WhenKeyStoreDoesntHaveCurrentKey(t *testing.T) {
	fakeKeys := NewFakeKeyStore()

	keyring, err := NewKeyring(t.Context(), KeyringConfig{
		KeyStore: fakeKeys,
	})

	require.NotNil(t, keyring)
	require.NoError(t, err)

	current := testutil.Only(t, fakeKeys.UpdatedKeys())
	assert.Equal(t, current, keyring.current.Load())
}

func TestKeyring_SigningKey_ReturnsCurrentKey(t *testing.T) {
	keyring, current := keyringAndCurrentForTest(t)

	key := keyring.SigningKey()

	require.NotNil(t, key)
	assert.Equal(t, current, key)
}

func TestKeyring_PublicKeyByID_ReturnsCurrentPublicKey(t *testing.T) {
	keyring, current := keyringAndCurrentForTest(t)

	key, err := keyring.PublicKeyByID(current.id)

	require.NoError(t, err)
	require.NotNil(t, key)
	assert.Equal(t, current.public, key)
}

func TestKeyring_PublicKeyByID_ReturnsErrKidUnknown(t *testing.T) {
	keyring := keyringForTest(t)

	key, err := keyring.PublicKeyByID("unknown")

	assert.Nil(t, key)
	assert.ErrorIs(t, err, errKidUnknown)
}

func TestKeyring_Rotate_ChangesTheCurrentKey(t *testing.T) {
	keyring, previous := keyringAndCurrentForTest(t)

	err := keyring.Rotate(t.Context())

	require.NoError(t, err)
	current := keyring.current.Load()
	require.NotNil(t, current)
	assert.NotEqual(t, previous, current)
}

func TestKeyring_Rotate_UpdateCurrentInKeyStore_WithContext(t *testing.T) {
	keyring, fakeKeys := keyringAndFakeKeyStoreForTest(t)

	err := keyring.Rotate(t.Context())

	require.NoError(t, err)
	ctx := testutil.Only(t, fakeKeys.ContextsUsedInUpdate())
	assert.Equal(t, t.Context(), ctx)
}

func TestKeyring_Rotate_UpdateCurrentInKeyStore_WithTheSameCurrentInKeyring(t *testing.T) {
	keyring, fakeKeys := keyringAndFakeKeyStoreForTest(t)

	err := keyring.Rotate(t.Context())

	require.NoError(t, err)
	expected := testutil.Only(t, fakeKeys.UpdatedKeys())
	assert.Equal(t, expected, keyring.current.Load())
}

func TestKeyring_PublicKeys_ReturnsOnlyCurrent(t *testing.T) {
	keyring, current := keyringAndCurrentForTest(t)

	keys := keyring.PublicKeys()

	actual := testutil.Only(t, keys)
	assert.Equal(t, current.dto, actual)
}
