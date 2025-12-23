// Package atom provides type-segregated atomic value storage.
package atom

import "time"

// TableType identifies the segregated storage table for atomic values.
type TableType string

// Table type constants for type-segregated storage.
const (
	TableStrings TableType = "strings"
	TableInts    TableType = "ints"
	TableFloats  TableType = "floats"
	TableBools   TableType = "bools"
	TableTimes   TableType = "times"
)

// AllTables returns all table types in canonical order.
func AllTables() []TableType {
	return []TableType{TableStrings, TableInts, TableFloats, TableBools, TableTimes}
}

// Prefix returns the storage key prefix for this table type.
func (t TableType) Prefix() string {
	return string(t) + ":"
}

// Atoms holds decomposed atomic values by type.
type Atoms struct { //nolint:govet // field order matches logical grouping
	ID      string
	Strings map[string]string
	Ints    map[string]int64
	Floats  map[string]float64
	Bools   map[string]bool
	Times   map[string]time.Time
}

// NewAtoms creates an Atoms with initialized maps.
func NewAtoms(id string) *Atoms {
	return &Atoms{
		ID:      id,
		Strings: make(map[string]string),
		Ints:    make(map[string]int64),
		Floats:  make(map[string]float64),
		Bools:   make(map[string]bool),
		Times:   make(map[string]time.Time),
	}
}

// Atomize converts a value to its atomic representation.
type Atomize[T any] func(*T) Atoms

// Deatomize reconstructs a value from its atomic representation.
type Deatomize[T any] func(Atoms) (*T, error)

// FieldDescriptor maps a field name to its storage table.
type FieldDescriptor struct {
	Name  string
	Table TableType
}

// Validator provides validation before atomization.
type Validator interface {
	Validate() error
}
