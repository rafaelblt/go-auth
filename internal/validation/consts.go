package validation

// Issue Codes
const (
	CodeTooLong           = "TOO_LONG"
	CodeTooShort          = "TOO_SHORT"
	CodeInvalidCharacters = "INVALID_CHARACTERS"
	CodeRequired          = "REQUIRED"
	CodeNotPositive       = "NOT_POSITIVE"
	CodeNotAllowed        = "NOT_ALLOWED"
)

// Issue Detail Keys
const (
	KeyMaxLength  = "max"
	KeyMinLength  = "min"
	KeyUnitLength = "unit"
	KeyAllowed    = "allowed"
)

type LengthUnit string

const (
	UnitCodePoint LengthUnit = "code_point"
	UnitByte      LengthUnit = "byte"
)
