package port

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
