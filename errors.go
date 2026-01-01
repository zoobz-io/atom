package atom

import "errors"

// Codec errors.
var (
	// ErrNoCodec is returned when no codec is registered for a spec.
	ErrNoCodec = errors.New("no codec registered for spec")
)
