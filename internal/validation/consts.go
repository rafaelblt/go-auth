package validation

// Issue Codes
const (
	CodeTooLong           = "TOO_LONG"
	CodeTooShort          = "TOO_SHORT"
	CodeInvalidCharacters = "INVALID_CHARACTERS"
)

// Issue Detail Keys
const (
	KeyMaxLength  = "max"
	KeyMinLength  = "min"
	KeyUnitLength = "unit"
)

type LengthUnit string

const (
	UnitCodePoint LengthUnit = "code_point"
	UnitByte      LengthUnit = "byte"
)
