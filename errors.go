package atom

import "errors"

var (
	// ErrDecode indicates decoding failed.
	ErrDecode = errors.New("atom: decoding failed")

	// ErrMissingID indicates the Atoms.ID field is empty.
	ErrMissingID = errors.New("atom: missing ID in atoms")
)
