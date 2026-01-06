package atom

import (
	"errors"
	"testing"
)

func TestErrOverflow(t *testing.T) {
	type Small struct {
		Value int8
	}

	atomizer, err := Use[Small]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	// Create atom with value that overflows int8
	atom := &Atom{
		Ints: map[string]int64{"Value": 200}, // max int8 is 127
	}

	_, err = atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected overflow error")
	}
	if !errors.Is(err, ErrOverflow) {
		t.Errorf("expected ErrOverflow, got: %v", err)
	}
}

func TestErrUnsupportedType(t *testing.T) {
	type Bad struct {
		Ch chan int
	}

	_, err := Use[Bad]()
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("expected ErrUnsupportedType, got: %v", err)
	}
}

func TestErrSizeMismatch(t *testing.T) {
	type Fixed struct {
		Data [4]byte
	}

	atomizer, err := Use[Fixed]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	// Create atom with wrong size byte slice
	atom := &Atom{
		Bytes: map[string][]byte{"Data": {1, 2, 3}}, // 3 bytes, need 4
	}

	_, err = atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected size mismatch error")
	}
	if !errors.Is(err, ErrSizeMismatch) {
		t.Errorf("expected ErrSizeMismatch, got: %v", err)
	}
}
