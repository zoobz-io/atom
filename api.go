// Package atom provides type-segregated atomic value storage.
package atom

import (
	"time"

	"github.com/zoobzio/sentinel"
)

// Spec is metadata describing a struct type.
// Aliased from sentinel.Metadata to decouple downstream users from sentinel.
type Spec = sentinel.Metadata

// Table identifies the segregated storage table for atomic values.
type Table string

// Table constants for type-segregated storage.
const (
	TableStrings      Table = "strings"
	TableInts         Table = "ints"
	TableUints        Table = "uints"
	TableFloats       Table = "floats"
	TableBools        Table = "bools"
	TableTimes        Table = "times"
	TableBytes        Table = "bytes"
	TableBytePtrs     Table = "byte_ptrs"
	TableStringPtrs   Table = "string_ptrs"
	TableIntPtrs      Table = "int_ptrs"
	TableUintPtrs     Table = "uint_ptrs"
	TableFloatPtrs    Table = "float_ptrs"
	TableBoolPtrs     Table = "bool_ptrs"
	TableTimePtrs     Table = "time_ptrs"
	TableStringSlices Table = "string_slices"
	TableIntSlices    Table = "int_slices"
	TableUintSlices   Table = "uint_slices"
	TableFloatSlices  Table = "float_slices"
	TableBoolSlices   Table = "bool_slices"
	TableTimeSlices   Table = "time_slices"
	TableByteSlices   Table = "byte_slices"
)

// AllTables returns all table types in canonical order.
func AllTables() []Table {
	return []Table{
		TableStrings, TableInts, TableUints, TableFloats, TableBools, TableTimes, TableBytes,
		TableBytePtrs, TableStringPtrs, TableIntPtrs, TableUintPtrs, TableFloatPtrs, TableBoolPtrs, TableTimePtrs,
		TableStringSlices, TableIntSlices, TableUintSlices, TableFloatSlices, TableBoolSlices, TableTimeSlices, TableByteSlices,
	}
}

// Prefix returns the storage key prefix for this table.
func (t Table) Prefix() string {
	return string(t) + ":"
}

// Atom holds decomposed atomic values by type.
type Atom struct {
	// Scalars
	Strings map[string]string
	Ints    map[string]int64
	Uints   map[string]uint64
	Floats  map[string]float64
	Bools   map[string]bool
	Times   map[string]time.Time
	Bytes   map[string][]byte

	// Pointers (nullable)
	StringPtrs map[string]*string
	IntPtrs    map[string]*int64
	UintPtrs   map[string]*uint64
	FloatPtrs  map[string]*float64
	BoolPtrs   map[string]*bool
	TimePtrs   map[string]*time.Time
	BytePtrs   map[string]*[]byte

	// Slices
	StringSlices map[string][]string
	IntSlices    map[string][]int64
	UintSlices   map[string][]uint64
	FloatSlices  map[string][]float64
	BoolSlices   map[string][]bool
	TimeSlices   map[string][]time.Time
	ByteSlices   map[string][][]byte

	// Nested
	Nested       map[string]Atom
	NestedSlices map[string][]Atom

	// Metadata (placed last for optimal alignment)
	Spec Spec
}

// Field maps a field name to its storage table.
type Field struct {
	Name  string
	Table Table
}

// Atomizable allows types to provide custom atomization logic.
// If a type implements this interface, it will be used instead of reflection.
// This enables code generation to avoid reflection overhead.
type Atomizable interface {
	Atomize(*Atom)
}

// Deatomizable allows types to provide custom deatomization logic.
// If a type implements this interface, it will be used instead of reflection.
// This enables code generation to avoid reflection overhead.
type Deatomizable interface {
	Deatomize(*Atom) error
}
