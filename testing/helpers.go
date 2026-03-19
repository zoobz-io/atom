// Package testing provides utilities for testing atom-based applications.
package testing

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/zoobz-io/atom"
)

// MustUse registers an atomizer and fails the test on error.
func MustUse[T any](t testing.TB) *atom.Atomizer[T] {
	t.Helper()
	atomizer, err := atom.Use[T]()
	if err != nil {
		t.Fatalf("atom.Use[%T]() failed: %v", *new(T), err)
	}
	return atomizer
}

// AtomBuilder provides a fluent interface for constructing test atoms.
type AtomBuilder struct {
	atom *atom.Atom
}

// NewAtomBuilder creates a new AtomBuilder.
func NewAtomBuilder() *AtomBuilder {
	return &AtomBuilder{
		atom: &atom.Atom{
			Strings:      make(map[string]string),
			Ints:         make(map[string]int64),
			Uints:        make(map[string]uint64),
			Floats:       make(map[string]float64),
			Bools:        make(map[string]bool),
			Times:        make(map[string]time.Time),
			Bytes:        make(map[string][]byte),
			StringPtrs:   make(map[string]*string),
			IntPtrs:      make(map[string]*int64),
			UintPtrs:     make(map[string]*uint64),
			FloatPtrs:    make(map[string]*float64),
			BoolPtrs:     make(map[string]*bool),
			TimePtrs:     make(map[string]*time.Time),
			BytePtrs:     make(map[string]*[]byte),
			StringSlices: make(map[string][]string),
			IntSlices:    make(map[string][]int64),
			UintSlices:   make(map[string][]uint64),
			FloatSlices:  make(map[string][]float64),
			BoolSlices:   make(map[string][]bool),
			TimeSlices:   make(map[string][]time.Time),
			ByteSlices:   make(map[string][][]byte),
			Nested:       make(map[string]atom.Atom),
			NestedSlices: make(map[string][]atom.Atom),
		},
	}
}

// String adds a string field.
func (b *AtomBuilder) String(key, value string) *AtomBuilder {
	b.atom.Strings[key] = value
	return b
}

// Int adds an int64 field.
func (b *AtomBuilder) Int(key string, value int64) *AtomBuilder {
	b.atom.Ints[key] = value
	return b
}

// Uint adds a uint64 field.
func (b *AtomBuilder) Uint(key string, value uint64) *AtomBuilder {
	b.atom.Uints[key] = value
	return b
}

// Float adds a float64 field.
func (b *AtomBuilder) Float(key string, value float64) *AtomBuilder {
	b.atom.Floats[key] = value
	return b
}

// Bool adds a bool field.
func (b *AtomBuilder) Bool(key string, value bool) *AtomBuilder {
	b.atom.Bools[key] = value
	return b
}

// Time adds a time.Time field.
func (b *AtomBuilder) Time(key string, value time.Time) *AtomBuilder {
	b.atom.Times[key] = value
	return b
}

// Bytes adds a []byte field.
func (b *AtomBuilder) Bytes(key string, value []byte) *AtomBuilder {
	b.atom.Bytes[key] = value
	return b
}

// StringPtr adds a *string field.
func (b *AtomBuilder) StringPtr(key string, value *string) *AtomBuilder {
	b.atom.StringPtrs[key] = value
	return b
}

// IntPtr adds a *int64 field.
func (b *AtomBuilder) IntPtr(key string, value *int64) *AtomBuilder {
	b.atom.IntPtrs[key] = value
	return b
}

// StringSlice adds a []string field.
func (b *AtomBuilder) StringSlice(key string, value []string) *AtomBuilder {
	b.atom.StringSlices[key] = value
	return b
}

// IntSlice adds a []int64 field.
func (b *AtomBuilder) IntSlice(key string, value []int64) *AtomBuilder {
	b.atom.IntSlices[key] = value
	return b
}

// Nested adds a nested atom.
func (b *AtomBuilder) Nested(key string, value *atom.Atom) *AtomBuilder {
	if value != nil {
		b.atom.Nested[key] = *value
	}
	return b
}

// NestedSlice adds a slice of nested atoms.
func (b *AtomBuilder) NestedSlice(key string, value []atom.Atom) *AtomBuilder {
	b.atom.NestedSlices[key] = value
	return b
}

// WithSpec sets the atom's spec.
func (b *AtomBuilder) WithSpec(spec atom.Spec) *AtomBuilder {
	b.atom.Spec = spec
	return b
}

// Build returns the constructed atom.
func (b *AtomBuilder) Build() *atom.Atom {
	return b.atom
}

// RoundTripValidator validates that atomize/deatomize preserves data.
type RoundTripValidator[T any] struct {
	atomizer *atom.Atomizer[T]
}

// NewRoundTripValidator creates a new validator.
func NewRoundTripValidator[T any](atomizer *atom.Atomizer[T]) *RoundTripValidator[T] {
	return &RoundTripValidator[T]{atomizer: atomizer}
}

// Validate checks that round-trip preserves the value.
func (v *RoundTripValidator[T]) Validate(t testing.TB, original *T) {
	t.Helper()

	a := v.atomizer.Atomize(original)
	restored, err := v.atomizer.Deatomize(a)
	if err != nil {
		t.Fatalf("Deatomize failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("round-trip mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}
}

// ValidateAll checks multiple values.
func (v *RoundTripValidator[T]) ValidateAll(t *testing.T, cases []*T) {
	t.Helper()
	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			v.Validate(t, c)
		})
	}
}

// Equal checks if two atoms are deeply equal.
func Equal(a, b *atom.Atom) bool {
	return reflect.DeepEqual(a, b)
}

// EqualFields checks if specific fields are equal.
func EqualFields(a, b *atom.Atom, fields ...string) bool {
	for _, f := range fields {
		if a.Strings[f] != b.Strings[f] {
			return false
		}
		if a.Ints[f] != b.Ints[f] {
			return false
		}
		if a.Floats[f] != b.Floats[f] {
			return false
		}
		if a.Bools[f] != b.Bools[f] {
			return false
		}
		if !a.Times[f].Equal(b.Times[f]) {
			return false
		}
		if !bytes.Equal(a.Bytes[f], b.Bytes[f]) {
			return false
		}
	}
	return true
}

// Diff returns a human-readable diff between two atoms.
func Diff(a, b *atom.Atom) string {
	var sb strings.Builder

	diffMap := func(name string, m1, m2 map[string]string) {
		for k, v1 := range m1 {
			if v2, ok := m2[k]; !ok {
				sb.WriteString(fmt.Sprintf("%s[%s]: %q vs <missing>\n", name, k, v1))
			} else if v1 != v2 {
				sb.WriteString(fmt.Sprintf("%s[%s]: %q vs %q\n", name, k, v1, v2))
			}
		}
		for k := range m2 {
			if _, ok := m1[k]; !ok {
				sb.WriteString(fmt.Sprintf("%s[%s]: <missing> vs %q\n", name, k, m2[k]))
			}
		}
	}

	diffMapInt := func(name string, m1, m2 map[string]int64) {
		for k, v1 := range m1 {
			if v2, ok := m2[k]; !ok {
				sb.WriteString(fmt.Sprintf("%s[%s]: %d vs <missing>\n", name, k, v1))
			} else if v1 != v2 {
				sb.WriteString(fmt.Sprintf("%s[%s]: %d vs %d\n", name, k, v1, v2))
			}
		}
		for k := range m2 {
			if _, ok := m1[k]; !ok {
				sb.WriteString(fmt.Sprintf("%s[%s]: <missing> vs %d\n", name, k, m2[k]))
			}
		}
	}

	diffMap("Strings", a.Strings, b.Strings)
	diffMapInt("Ints", a.Ints, b.Ints)
	// Add more as needed...

	if sb.Len() == 0 {
		return "<no differences>"
	}
	return sb.String()
}

// Field Getters.

// GetString retrieves a string field.
func GetString(a *atom.Atom, key string) (string, bool) {
	v, ok := a.Strings[key]
	return v, ok
}

// GetInt retrieves an int64 field.
func GetInt(a *atom.Atom, key string) (int64, bool) {
	v, ok := a.Ints[key]
	return v, ok
}

// GetUint retrieves a uint64 field.
func GetUint(a *atom.Atom, key string) (uint64, bool) {
	v, ok := a.Uints[key]
	return v, ok
}

// GetFloat retrieves a float64 field.
func GetFloat(a *atom.Atom, key string) (float64, bool) {
	v, ok := a.Floats[key]
	return v, ok
}

// GetBool retrieves a bool field.
func GetBool(a *atom.Atom, key string) (value bool, ok bool) {
	value, ok = a.Bools[key]
	return value, ok
}

// GetTime retrieves a time.Time field.
func GetTime(a *atom.Atom, key string) (time.Time, bool) {
	v, ok := a.Times[key]
	return v, ok
}

// GetBytes retrieves a []byte field.
func GetBytes(a *atom.Atom, key string) ([]byte, bool) {
	v, ok := a.Bytes[key]
	return v, ok
}

// Assertions.

// AssertHasString asserts a string field value.
func AssertHasString(t testing.TB, a *atom.Atom, key, expected string) {
	t.Helper()
	if got, ok := a.Strings[key]; !ok {
		t.Errorf("expected Strings[%q] to exist", key)
	} else if got != expected {
		t.Errorf("Strings[%q]: got %q, want %q", key, got, expected)
	}
}

// AssertHasInt asserts an int64 field value.
func AssertHasInt(t testing.TB, a *atom.Atom, key string, expected int64) {
	t.Helper()
	if got, ok := a.Ints[key]; !ok {
		t.Errorf("expected Ints[%q] to exist", key)
	} else if got != expected {
		t.Errorf("Ints[%q]: got %d, want %d", key, got, expected)
	}
}

// AssertHasFloat asserts a float64 field value.
func AssertHasFloat(t testing.TB, a *atom.Atom, key string, expected float64) {
	t.Helper()
	if got, ok := a.Floats[key]; !ok {
		t.Errorf("expected Floats[%q] to exist", key)
	} else if got != expected {
		t.Errorf("Floats[%q]: got %g, want %g", key, got, expected)
	}
}

// AssertHasBool asserts a bool field value.
func AssertHasBool(t testing.TB, a *atom.Atom, key string, expected bool) {
	t.Helper()
	if got, ok := a.Bools[key]; !ok {
		t.Errorf("expected Bools[%q] to exist", key)
	} else if got != expected {
		t.Errorf("Bools[%q]: got %v, want %v", key, got, expected)
	}
}

// AssertHasNested asserts a nested field exists.
func AssertHasNested(t testing.TB, a *atom.Atom, key string) {
	t.Helper()
	if _, ok := a.Nested[key]; !ok {
		t.Errorf("expected Nested[%q] to exist", key)
	}
}

// AssertMissingField asserts a field does not exist.
func AssertMissingField(t testing.TB, a *atom.Atom, table atom.Table, key string) {
	t.Helper()
	switch table {
	case atom.TableStrings:
		if _, ok := a.Strings[key]; ok {
			t.Errorf("expected Strings[%q] to be missing", key)
		}
	case atom.TableInts:
		if _, ok := a.Ints[key]; ok {
			t.Errorf("expected Ints[%q] to be missing", key)
		}
	case atom.TableFloats:
		if _, ok := a.Floats[key]; ok {
			t.Errorf("expected Floats[%q] to be missing", key)
		}
	case atom.TableBools:
		if _, ok := a.Bools[key]; ok {
			t.Errorf("expected Bools[%q] to be missing", key)
		}
	}
}

// Random Data Generators.

// RandomString generates a random string of given length.
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// RandomInt generates a random int64 in [lo, hi].
func RandomInt(lo, hi int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(hi-lo+1))
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return lo + n.Int64()
}

// RandomFloat generates a random float64 in [lo, hi).
func RandomFloat(lo, hi float64) float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<53))
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return lo + (float64(n.Int64())/float64(1<<53))*(hi-lo)
}

// RandomBytes generates random bytes of given length.
func RandomBytes(length int) []byte {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return b
}

// RandomTime generates a random time within the last year.
func RandomTime() time.Time {
	now := time.Now()
	n, err := rand.Int(rand.Reader, big.NewInt(int64(365*24*time.Hour)))
	if err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	delta := time.Duration(n.Int64())
	return now.Add(-delta)
}
