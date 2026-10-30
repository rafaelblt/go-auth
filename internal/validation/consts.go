package validation

// Issue Codes
const (
	CodeTooLong           = "too_long"
	CodeTooShort          = "too_short"
	CodeInvalidCharacters = "invalid_characters"
	CodeSameAsCurrent     = "same_as_current"
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
