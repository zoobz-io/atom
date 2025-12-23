package atom

import (
	"errors"
	"testing"
)

func TestErrorsAreDefined(t *testing.T) {
	// Verify all errors are non-nil
	errs := []error{
		ErrDecode,
		ErrMissingID,
	}

	for _, err := range errs {
		if err == nil {
			t.Error("expected error to be non-nil")
		}
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		err      error
		contains string
	}{
		{ErrDecode, "decoding"},
		{ErrMissingID, "missing ID"},
	}

	for _, tt := range tests {
		if tt.err.Error() == "" {
			t.Errorf("expected non-empty error message for %v", tt.err)
		}
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	errs := []error{
		ErrDecode,
		ErrMissingID,
	}

	for i, err1 := range errs {
		for j, err2 := range errs {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("expected %v and %v to be distinct errors", err1, err2)
			}
		}
	}
}

func TestErrorsCanBeWrapped(t *testing.T) {
	wrapped := errors.Join(ErrDecode, errors.New("underlying cause"))

	if !errors.Is(wrapped, ErrDecode) {
		t.Error("expected wrapped error to match ErrDecode")
	}
}

func TestErrorPrefix(t *testing.T) {
	// All errors should have "atom:" prefix
	errs := []error{
		ErrDecode,
		ErrMissingID,
	}

	for _, err := range errs {
		msg := err.Error()
		if len(msg) < 5 || msg[:5] != "atom:" {
			t.Errorf("expected error %q to have 'atom:' prefix", msg)
		}
	}
}
