package validation

// Issue Codes
const (
	CodeTooLong           = "too_long"
	CodeTooShort          = "too_short"
	CodeInvalidCharacters = "invalid_characters"
	CodeRequired          = "required"
	CodeNotPositive       = "not_positive"
	CodeTooSmall          = "too_small"
	CodeNotAllowed        = "not_allowed"
	CodeSameAsCurrent     = "same_as_current"
)

// Issue Detail Keys
const (
	KeyMaxLength  = "max"
	KeyMinLength  = "min"
	KeyUnitLength = "unit"
	KeyMinValue   = "min"
	KeyAllowed    = "allowed"
)

type LengthUnit string

const (
	UnitCodePoint LengthUnit = "code_point"
	UnitByte      LengthUnit = "byte"
)
