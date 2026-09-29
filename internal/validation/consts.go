package validation

// Issue Codes
const (
	CodeTooLong           = "too_long"
	CodeTooShort          = "too_short"
	CodeInvalidCharacters = "invalid_characters"
	CodeRequired          = "required"
	CodeNotPositive       = "not_positive"
	CodeNotAllowed        = "not_allowed"
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
