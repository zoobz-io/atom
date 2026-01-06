package atom

import "errors"

// Sentinel errors for programmatic error handling.
// Use errors.Is() to check for these error types.
var (
	// ErrOverflow indicates a numeric value exceeds the target type's range.
	// Returned during Deatomize when an int64/uint64/float64 value cannot fit
	// in a narrower type (e.g., int8, uint16, float32).
	ErrOverflow = errors.New("numeric overflow")

	// ErrUnsupportedType indicates a struct field has a type that cannot be atomized.
	// Returned during Use[T]() registration for types like channels, functions,
	// interfaces, or maps with non-string keys.
	ErrUnsupportedType = errors.New("unsupported type")

	// ErrSizeMismatch indicates a byte slice length doesn't match a fixed-size array.
	// Returned during Deatomize when a []byte value is assigned to a [N]byte field
	// but the lengths differ.
	ErrSizeMismatch = errors.New("size mismatch")
)
