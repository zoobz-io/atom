package atom

import (
	"encoding/binary"
	"math"
	"time"
)

// String encoding: stored as raw bytes (UTF-8).
func encodeString(s string) []byte {
	return []byte(s)
}

func decodeString(data []byte) string {
	return string(data)
}

// Int64 encoding: big-endian binary for sortability.
func encodeInt64(v int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v)) //nolint:gosec // intentional bit pattern conversion
	return buf
}

func decodeInt64(data []byte) (int64, error) {
	if len(data) != 8 {
		return 0, ErrDecode
	}
	return int64(binary.BigEndian.Uint64(data)), nil //nolint:gosec // intentional bit pattern conversion
}

// Float64 encoding: IEEE 754 binary representation.
func encodeFloat64(v float64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(v))
	return buf
}

func decodeFloat64(data []byte) (float64, error) {
	if len(data) != 8 {
		return 0, ErrDecode
	}
	return math.Float64frombits(binary.BigEndian.Uint64(data)), nil
}

// Bool encoding: single byte (0 = false, 1 = true).
func encodeBool(v bool) []byte {
	if v {
		return []byte{1}
	}
	return []byte{0}
}

func decodeBool(data []byte) (bool, error) {
	if len(data) != 1 {
		return false, ErrDecode
	}
	return data[0] != 0, nil
}

// Time encoding: RFC3339Nano string for human readability and cross-system compatibility.
func encodeTime(t time.Time) []byte {
	return []byte(t.Format(time.RFC3339Nano))
}

func decodeTime(data []byte) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, string(data))
}
