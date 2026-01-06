// Package atom provides type-segregated atomic value storage.
package atom

import (
	"time"

	"github.com/zoobzio/sentinel"
)

func init() {
	sentinel.Tag("time")
}

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
	TableStringMaps   Table = "string_maps"
	TableIntMaps      Table = "int_maps"
	TableUintMaps     Table = "uint_maps"
	TableFloatMaps    Table = "float_maps"
	TableBoolMaps     Table = "bool_maps"
	TableTimeMaps     Table = "time_maps"
	TableByteMaps     Table = "byte_maps"
	TableNestedMaps   Table = "nested_maps"
)

// AllTables returns all table types in canonical order.
func AllTables() []Table {
	return []Table{
		TableStrings, TableInts, TableUints, TableFloats, TableBools, TableTimes, TableBytes,
		TableBytePtrs, TableStringPtrs, TableIntPtrs, TableUintPtrs, TableFloatPtrs, TableBoolPtrs, TableTimePtrs,
		TableStringSlices, TableIntSlices, TableUintSlices, TableFloatSlices, TableBoolSlices, TableTimeSlices, TableByteSlices,
		TableStringMaps, TableIntMaps, TableUintMaps, TableFloatMaps, TableBoolMaps, TableTimeMaps, TableByteMaps, TableNestedMaps,
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

	// Maps (string-keyed)
	StringMaps map[string]map[string]string
	IntMaps    map[string]map[string]int64
	UintMaps   map[string]map[string]uint64
	FloatMaps  map[string]map[string]float64
	BoolMaps   map[string]map[string]bool
	TimeMaps   map[string]map[string]time.Time
	ByteMaps   map[string]map[string][]byte

	// Nested
	Nested       map[string]Atom
	NestedSlices map[string][]Atom
	NestedMaps   map[string]map[string]Atom

	// Metadata (placed last for optimal alignment)
	Spec Spec
}

// Field maps a field name to its storage table.
type Field struct {
	Name  string
	Table Table
}

// Clone returns a deep copy of the Atom.
func (a *Atom) Clone() *Atom {
	if a == nil {
		return nil
	}

	clone := &Atom{Spec: a.Spec}

	// Clone scalar maps
	if a.Strings != nil {
		clone.Strings = make(map[string]string, len(a.Strings))
		for k, v := range a.Strings {
			clone.Strings[k] = v
		}
	}
	if a.Ints != nil {
		clone.Ints = make(map[string]int64, len(a.Ints))
		for k, v := range a.Ints {
			clone.Ints[k] = v
		}
	}
	if a.Uints != nil {
		clone.Uints = make(map[string]uint64, len(a.Uints))
		for k, v := range a.Uints {
			clone.Uints[k] = v
		}
	}
	if a.Floats != nil {
		clone.Floats = make(map[string]float64, len(a.Floats))
		for k, v := range a.Floats {
			clone.Floats[k] = v
		}
	}
	if a.Bools != nil {
		clone.Bools = make(map[string]bool, len(a.Bools))
		for k, v := range a.Bools {
			clone.Bools[k] = v
		}
	}
	if a.Times != nil {
		clone.Times = make(map[string]time.Time, len(a.Times))
		for k, v := range a.Times {
			clone.Times[k] = v
		}
	}
	if a.Bytes != nil {
		clone.Bytes = make(map[string][]byte, len(a.Bytes))
		for k, v := range a.Bytes {
			cp := make([]byte, len(v))
			copy(cp, v)
			clone.Bytes[k] = cp
		}
	}

	// Clone pointer maps (preserving nil entries)
	if a.StringPtrs != nil {
		clone.StringPtrs = make(map[string]*string, len(a.StringPtrs))
		for k, v := range a.StringPtrs {
			if v != nil {
				cp := *v
				clone.StringPtrs[k] = &cp
			} else {
				clone.StringPtrs[k] = nil
			}
		}
	}
	if a.IntPtrs != nil {
		clone.IntPtrs = make(map[string]*int64, len(a.IntPtrs))
		for k, v := range a.IntPtrs {
			if v != nil {
				cp := *v
				clone.IntPtrs[k] = &cp
			} else {
				clone.IntPtrs[k] = nil
			}
		}
	}
	if a.UintPtrs != nil {
		clone.UintPtrs = make(map[string]*uint64, len(a.UintPtrs))
		for k, v := range a.UintPtrs {
			if v != nil {
				cp := *v
				clone.UintPtrs[k] = &cp
			} else {
				clone.UintPtrs[k] = nil
			}
		}
	}
	if a.FloatPtrs != nil {
		clone.FloatPtrs = make(map[string]*float64, len(a.FloatPtrs))
		for k, v := range a.FloatPtrs {
			if v != nil {
				cp := *v
				clone.FloatPtrs[k] = &cp
			} else {
				clone.FloatPtrs[k] = nil
			}
		}
	}
	if a.BoolPtrs != nil {
		clone.BoolPtrs = make(map[string]*bool, len(a.BoolPtrs))
		for k, v := range a.BoolPtrs {
			if v != nil {
				cp := *v
				clone.BoolPtrs[k] = &cp
			} else {
				clone.BoolPtrs[k] = nil
			}
		}
	}
	if a.TimePtrs != nil {
		clone.TimePtrs = make(map[string]*time.Time, len(a.TimePtrs))
		for k, v := range a.TimePtrs {
			if v != nil {
				cp := *v
				clone.TimePtrs[k] = &cp
			} else {
				clone.TimePtrs[k] = nil
			}
		}
	}
	if a.BytePtrs != nil {
		clone.BytePtrs = make(map[string]*[]byte, len(a.BytePtrs))
		for k, v := range a.BytePtrs {
			if v != nil {
				cp := make([]byte, len(*v))
				copy(cp, *v)
				clone.BytePtrs[k] = &cp
			} else {
				clone.BytePtrs[k] = nil
			}
		}
	}

	// Clone slice maps
	if a.StringSlices != nil {
		clone.StringSlices = make(map[string][]string, len(a.StringSlices))
		for k, v := range a.StringSlices {
			cp := make([]string, len(v))
			copy(cp, v)
			clone.StringSlices[k] = cp
		}
	}
	if a.IntSlices != nil {
		clone.IntSlices = make(map[string][]int64, len(a.IntSlices))
		for k, v := range a.IntSlices {
			cp := make([]int64, len(v))
			copy(cp, v)
			clone.IntSlices[k] = cp
		}
	}
	if a.UintSlices != nil {
		clone.UintSlices = make(map[string][]uint64, len(a.UintSlices))
		for k, v := range a.UintSlices {
			cp := make([]uint64, len(v))
			copy(cp, v)
			clone.UintSlices[k] = cp
		}
	}
	if a.FloatSlices != nil {
		clone.FloatSlices = make(map[string][]float64, len(a.FloatSlices))
		for k, v := range a.FloatSlices {
			cp := make([]float64, len(v))
			copy(cp, v)
			clone.FloatSlices[k] = cp
		}
	}
	if a.BoolSlices != nil {
		clone.BoolSlices = make(map[string][]bool, len(a.BoolSlices))
		for k, v := range a.BoolSlices {
			cp := make([]bool, len(v))
			copy(cp, v)
			clone.BoolSlices[k] = cp
		}
	}
	if a.TimeSlices != nil {
		clone.TimeSlices = make(map[string][]time.Time, len(a.TimeSlices))
		for k, v := range a.TimeSlices {
			cp := make([]time.Time, len(v))
			copy(cp, v)
			clone.TimeSlices[k] = cp
		}
	}
	if a.ByteSlices != nil {
		clone.ByteSlices = make(map[string][][]byte, len(a.ByteSlices))
		for k, v := range a.ByteSlices {
			cp := make([][]byte, len(v))
			for i, b := range v {
				cp[i] = make([]byte, len(b))
				copy(cp[i], b)
			}
			clone.ByteSlices[k] = cp
		}
	}

	// Clone maps
	if a.StringMaps != nil {
		clone.StringMaps = make(map[string]map[string]string, len(a.StringMaps))
		for k, v := range a.StringMaps {
			m := make(map[string]string, len(v))
			for mk, mv := range v {
				m[mk] = mv
			}
			clone.StringMaps[k] = m
		}
	}
	if a.IntMaps != nil {
		clone.IntMaps = make(map[string]map[string]int64, len(a.IntMaps))
		for k, v := range a.IntMaps {
			m := make(map[string]int64, len(v))
			for mk, mv := range v {
				m[mk] = mv
			}
			clone.IntMaps[k] = m
		}
	}
	if a.UintMaps != nil {
		clone.UintMaps = make(map[string]map[string]uint64, len(a.UintMaps))
		for k, v := range a.UintMaps {
			m := make(map[string]uint64, len(v))
			for mk, mv := range v {
				m[mk] = mv
			}
			clone.UintMaps[k] = m
		}
	}
	if a.FloatMaps != nil {
		clone.FloatMaps = make(map[string]map[string]float64, len(a.FloatMaps))
		for k, v := range a.FloatMaps {
			m := make(map[string]float64, len(v))
			for mk, mv := range v {
				m[mk] = mv
			}
			clone.FloatMaps[k] = m
		}
	}
	if a.BoolMaps != nil {
		clone.BoolMaps = make(map[string]map[string]bool, len(a.BoolMaps))
		for k, v := range a.BoolMaps {
			m := make(map[string]bool, len(v))
			for mk, mv := range v {
				m[mk] = mv
			}
			clone.BoolMaps[k] = m
		}
	}
	if a.TimeMaps != nil {
		clone.TimeMaps = make(map[string]map[string]time.Time, len(a.TimeMaps))
		for k, v := range a.TimeMaps {
			m := make(map[string]time.Time, len(v))
			for mk, mv := range v {
				m[mk] = mv
			}
			clone.TimeMaps[k] = m
		}
	}
	if a.ByteMaps != nil {
		clone.ByteMaps = make(map[string]map[string][]byte, len(a.ByteMaps))
		for k, v := range a.ByteMaps {
			m := make(map[string][]byte, len(v))
			for mk, mv := range v {
				cp := make([]byte, len(mv))
				copy(cp, mv)
				m[mk] = cp
			}
			clone.ByteMaps[k] = m
		}
	}

	// Clone nested
	if a.Nested != nil {
		clone.Nested = make(map[string]Atom, len(a.Nested))
		for k := range a.Nested {
			v := a.Nested[k]
			clone.Nested[k] = *v.Clone()
		}
	}
	if a.NestedSlices != nil {
		clone.NestedSlices = make(map[string][]Atom, len(a.NestedSlices))
		for k, v := range a.NestedSlices {
			cp := make([]Atom, len(v))
			for i := range v {
				cp[i] = *v[i].Clone()
			}
			clone.NestedSlices[k] = cp
		}
	}
	if a.NestedMaps != nil {
		clone.NestedMaps = make(map[string]map[string]Atom, len(a.NestedMaps))
		for k, v := range a.NestedMaps {
			m := make(map[string]Atom, len(v))
			for mk := range v {
				mv := v[mk]
				m[mk] = *mv.Clone()
			}
			clone.NestedMaps[k] = m
		}
	}

	return clone
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
