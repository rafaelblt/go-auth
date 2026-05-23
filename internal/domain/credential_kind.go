package domain

type CredentialKind struct {
	value string
}

var (
	CredentialKindPassword = CredentialKind{"password"}
)

func (k CredentialKind) String() string { return k.value }
func (k CredentialKind) IsZero() bool   { return k.value == "" }
