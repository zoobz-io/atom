package atom

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestEncodeDecodeString(t *testing.T) {
	tests := []string{
		"",
		"hello",
		"hello world",
		"unicode: 日本語",
		"emoji: 🎉",
		"special chars: \n\t\r",
		string(make([]byte, 1000)), // long string
	}

	for _, s := range tests {
		encoded := encodeString(s)
		decoded := decodeString(encoded)
		if decoded != s {
			t.Errorf("string roundtrip failed: expected %q, got %q", s, decoded)
		}
	}
}

func TestEncodeDecodeInt64(t *testing.T) {
	tests := []int64{
		0,
		1,
		-1,
		42,
		-42,
		math.MaxInt64,
		math.MinInt64,
		1234567890,
		-1234567890,
	}

	for _, n := range tests {
		encoded := encodeInt64(n)
		decoded, err := decodeInt64(encoded)
		if err != nil {
			t.Errorf("decodeInt64 error for %d: %v", n, err)
			continue
		}
		if decoded != n {
			t.Errorf("int64 roundtrip failed: expected %d, got %d", n, decoded)
		}
	}
}

func TestDecodeInt64InvalidLength(t *testing.T) {
	tests := [][]byte{
		{},
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 5, 6, 7},
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	for _, data := range tests {
		_, err := decodeInt64(data)
		if !errors.Is(err, ErrDecode) {
			t.Errorf("expected ErrDecode for data length %d, got %v", len(data), err)
		}
	}
}

func TestEncodeDecodeFloat64(t *testing.T) {
	tests := []float64{
		0,
		1.0,
		-1.0,
		3.14159265359,
		-3.14159265359,
		math.MaxFloat64,
		math.SmallestNonzeroFloat64,
		math.Inf(1),
		math.Inf(-1),
	}

	for _, f := range tests {
		encoded := encodeFloat64(f)
		decoded, err := decodeFloat64(encoded)
		if err != nil {
			t.Errorf("decodeFloat64 error for %f: %v", f, err)
			continue
		}
		if decoded != f {
			t.Errorf("float64 roundtrip failed: expected %f, got %f", f, decoded)
		}
	}
}

func TestEncodeDecodeFloat64NaN(t *testing.T) {
	nan := math.NaN()
	encoded := encodeFloat64(nan)
	decoded, err := decodeFloat64(encoded)
	if err != nil {
		t.Errorf("decodeFloat64 error for NaN: %v", err)
	}
	if !math.IsNaN(decoded) {
		t.Errorf("expected NaN, got %f", decoded)
	}
}

func TestDecodeFloat64InvalidLength(t *testing.T) {
	tests := [][]byte{
		{},
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 5, 6, 7},
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	for _, data := range tests {
		_, err := decodeFloat64(data)
		if !errors.Is(err, ErrDecode) {
			t.Errorf("expected ErrDecode for data length %d, got %v", len(data), err)
		}
	}
}

func TestEncodeDecodeBool(t *testing.T) {
	tests := []bool{true, false}

	for _, b := range tests {
		encoded := encodeBool(b)
		decoded, err := decodeBool(encoded)
		if err != nil {
			t.Errorf("decodeBool error for %v: %v", b, err)
			continue
		}
		if decoded != b {
			t.Errorf("bool roundtrip failed: expected %v, got %v", b, decoded)
		}
	}
}

func TestEncodeBoolValues(t *testing.T) {
	trueEncoded := encodeBool(true)
	falseEncoded := encodeBool(false)

	if len(trueEncoded) != 1 {
		t.Errorf("expected 1 byte for true, got %d", len(trueEncoded))
	}
	if len(falseEncoded) != 1 {
		t.Errorf("expected 1 byte for false, got %d", len(falseEncoded))
	}
	if trueEncoded[0] != 1 {
		t.Errorf("expected byte value 1 for true, got %d", trueEncoded[0])
	}
	if falseEncoded[0] != 0 {
		t.Errorf("expected byte value 0 for false, got %d", falseEncoded[0])
	}
}

func TestDecodeBoolNonZero(t *testing.T) {
	// Any non-zero value should decode to true
	tests := []byte{1, 2, 127, 255}

	for _, b := range tests {
		decoded, err := decodeBool([]byte{b})
		if err != nil {
			t.Errorf("decodeBool error for byte %d: %v", b, err)
			continue
		}
		if !decoded {
			t.Errorf("expected true for byte %d, got false", b)
		}
	}
}

func TestDecodeBoolInvalidLength(t *testing.T) {
	tests := [][]byte{
		{},
		{1, 2},
		{1, 2, 3},
	}

	for _, data := range tests {
		_, err := decodeBool(data)
		if !errors.Is(err, ErrDecode) {
			t.Errorf("expected ErrDecode for data length %d, got %v", len(data), err)
		}
	}
}

func TestEncodeDecodeTime(t *testing.T) {
	tests := []time.Time{
		time.Now(),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2099, 12, 31, 23, 59, 59, 999999999, time.UTC),
		time.Time{}, // zero time
	}

	for _, tm := range tests {
		encoded := encodeTime(tm)
		decoded, err := decodeTime(encoded)
		if err != nil {
			t.Errorf("decodeTime error for %v: %v", tm, err)
			continue
		}
		// Compare using Equal to handle timezone differences
		if !decoded.Equal(tm) {
			t.Errorf("time roundtrip failed: expected %v, got %v", tm, decoded)
		}
	}
}

func TestEncodeTimeFormat(t *testing.T) {
	tm := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC)
	encoded := encodeTime(tm)

	// Should be RFC3339Nano format
	expected := "2024-06-15T14:30:45.123456789Z"
	if string(encoded) != expected {
		t.Errorf("expected %s, got %s", expected, string(encoded))
	}
}

func TestDecodeTimeInvalid(t *testing.T) {
	tests := []string{
		"",
		"invalid",
		"2024-13-01T00:00:00Z", // invalid month
		"not a date",
	}

	for _, s := range tests {
		_, err := decodeTime([]byte(s))
		if err == nil {
			t.Errorf("expected error for invalid time string %q", s)
		}
	}
}

func TestEncodeInt64BigEndian(t *testing.T) {
	// Verify big-endian encoding for sortability
	small := encodeInt64(1)
	large := encodeInt64(1000)

	// In big-endian, larger numbers should be lexicographically larger
	// when comparing positive numbers
	if string(small) >= string(large) {
		t.Error("expected big-endian encoding to preserve order for positive numbers")
	}
}
