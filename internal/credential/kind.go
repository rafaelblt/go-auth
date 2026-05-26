package credential

type Kind struct {
	value string
}

var (
	KindPassword = Kind{"password"}
)

func (k Kind) String() string { return k.value }
func (k Kind) IsZero() bool   { return k.value == "" }
