package domain

type CredentialKind string

const (
	CredentialKindPassword CredentialKind = "password"
)

func (k CredentialKind) String() string { return string(k) }
func (k CredentialKind) IsValid() bool {
	switch k {
	case CredentialKindPassword:
		return true
	default:
		return false
	}
}
