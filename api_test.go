package atom

import (
	"testing"
	"time"
)

func TestTableConstants(t *testing.T) {
	tests := []struct {
		table    Table
		expected string
	}{
		{TableStrings, "strings"},
		{TableInts, "ints"},
		{TableUints, "uints"},
		{TableFloats, "floats"},
		{TableBools, "bools"},
		{TableTimes, "times"},
		{TableBytes, "bytes"},
		{TableBytePtrs, "byte_ptrs"},
		{TableStringPtrs, "string_ptrs"},
		{TableIntPtrs, "int_ptrs"},
		{TableUintPtrs, "uint_ptrs"},
		{TableFloatPtrs, "float_ptrs"},
		{TableBoolPtrs, "bool_ptrs"},
		{TableTimePtrs, "time_ptrs"},
		{TableStringSlices, "string_slices"},
		{TableIntSlices, "int_slices"},
		{TableUintSlices, "uint_slices"},
		{TableFloatSlices, "float_slices"},
		{TableBoolSlices, "bool_slices"},
		{TableTimeSlices, "time_slices"},
		{TableByteSlices, "byte_slices"},
	}

	for _, tt := range tests {
		if string(tt.table) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.table)
		}
	}
}

func TestAllTables(t *testing.T) {
	tables := AllTables()

	if len(tables) != 21 {
		t.Errorf("expected 21 tables, got %d", len(tables))
	}

	expected := []Table{
		TableStrings, TableInts, TableUints, TableFloats, TableBools, TableTimes, TableBytes,
		TableBytePtrs, TableStringPtrs, TableIntPtrs, TableUintPtrs, TableFloatPtrs, TableBoolPtrs, TableTimePtrs,
		TableStringSlices, TableIntSlices, TableUintSlices, TableFloatSlices, TableBoolSlices, TableTimeSlices, TableByteSlices,
	}
	for i, table := range tables {
		if table != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, table)
		}
	}
}

func TestTablePrefix(t *testing.T) {
	tests := []struct {
		table    Table
		expected string
	}{
		{TableStrings, "strings:"},
		{TableInts, "ints:"},
		{TableUints, "uints:"},
		{TableFloats, "floats:"},
		{TableBools, "bools:"},
		{TableTimes, "times:"},
		{TableBytes, "bytes:"},
		{TableBytePtrs, "byte_ptrs:"},
		{TableStringPtrs, "string_ptrs:"},
		{TableIntPtrs, "int_ptrs:"},
		{TableUintPtrs, "uint_ptrs:"},
		{TableFloatPtrs, "float_ptrs:"},
		{TableBoolPtrs, "bool_ptrs:"},
		{TableTimePtrs, "time_ptrs:"},
		{TableStringSlices, "string_slices:"},
		{TableIntSlices, "int_slices:"},
		{TableUintSlices, "uint_slices:"},
		{TableFloatSlices, "float_slices:"},
		{TableBoolSlices, "bool_slices:"},
		{TableTimeSlices, "time_slices:"},
		{TableByteSlices, "byte_slices:"},
	}

	for _, tt := range tests {
		if tt.table.Prefix() != tt.expected {
			t.Errorf("expected prefix %s for %s, got %s", tt.expected, tt.table, tt.table.Prefix())
		}
	}
}

func TestAtomMapsAreUsable(t *testing.T) {
	atom := Atom{
		Strings:      make(map[string]string),
		Ints:         make(map[string]int64),
		Uints:        make(map[string]uint64),
		Floats:       make(map[string]float64),
		Bools:        make(map[string]bool),
		Times:        make(map[string]time.Time),
		Bytes:        make(map[string][]byte),
		BytePtrs:     make(map[string]*[]byte),
		StringPtrs:   make(map[string]*string),
		IntPtrs:      make(map[string]*int64),
		UintPtrs:     make(map[string]*uint64),
		FloatPtrs:    make(map[string]*float64),
		BoolPtrs:     make(map[string]*bool),
		TimePtrs:     make(map[string]*time.Time),
		StringSlices: make(map[string][]string),
		IntSlices:    make(map[string][]int64),
		UintSlices:   make(map[string][]uint64),
		FloatSlices:  make(map[string][]float64),
		BoolSlices:   make(map[string][]bool),
		TimeSlices:   make(map[string][]time.Time),
		ByteSlices:   make(map[string][][]byte),
		Nested:       make(map[string]Atom),
		NestedSlices: make(map[string][]Atom),
	}

	// Verify maps can be written to
	atom.Strings["key"] = "value"
	atom.Ints["count"] = 42
	atom.Floats["rate"] = 3.14
	atom.Bools["active"] = true
	atom.Times["created"] = time.Now()
	atom.Bytes["data"] = []byte{0x01, 0x02, 0x03}

	strVal := "nullable"
	intVal := int64(123)
	floatVal := 2.71
	boolVal := true
	timeVal := time.Now()
	atom.StringPtrs["opt_str"] = &strVal
	atom.IntPtrs["opt_int"] = &intVal
	atom.FloatPtrs["opt_float"] = &floatVal
	atom.BoolPtrs["opt_bool"] = &boolVal
	atom.TimePtrs["opt_time"] = &timeVal

	if atom.Strings["key"] != "value" {
		t.Error("failed to store string value")
	}
	if atom.Ints["count"] != 42 {
		t.Error("failed to store int value")
	}
	if atom.Floats["rate"] != 3.14 {
		t.Error("failed to store float value")
	}
	if atom.Bools["active"] != true {
		t.Error("failed to store bool value")
	}
	if atom.Times["created"].IsZero() {
		t.Error("failed to store time value")
	}
	if len(atom.Bytes["data"]) != 3 {
		t.Error("failed to store bytes value")
	}
	if *atom.StringPtrs["opt_str"] != "nullable" {
		t.Error("failed to store string pointer value")
	}
	if *atom.IntPtrs["opt_int"] != 123 {
		t.Error("failed to store int pointer value")
	}
	if *atom.FloatPtrs["opt_float"] != 2.71 {
		t.Error("failed to store float pointer value")
	}
	if *atom.BoolPtrs["opt_bool"] != true {
		t.Error("failed to store bool pointer value")
	}
	if atom.TimePtrs["opt_time"].IsZero() {
		t.Error("failed to store time pointer value")
	}

	// Test nil pointer storage
	atom.StringPtrs["nil_str"] = nil
	if atom.StringPtrs["nil_str"] != nil {
		t.Error("failed to store nil string pointer")
	}

	// Test slice storage
	atom.StringSlices["tags"] = []string{"a", "b", "c"}
	atom.IntSlices["scores"] = []int64{1, 2, 3}
	atom.FloatSlices["rates"] = []float64{1.1, 2.2}
	atom.BoolSlices["flags"] = []bool{true, false}
	atom.TimeSlices["dates"] = []time.Time{time.Now()}
	atom.ByteSlices["chunks"] = [][]byte{{0x01}, {0x02}}

	if len(atom.StringSlices["tags"]) != 3 {
		t.Error("failed to store string slice")
	}
	if len(atom.IntSlices["scores"]) != 3 {
		t.Error("failed to store int slice")
	}
	if len(atom.FloatSlices["rates"]) != 2 {
		t.Error("failed to store float slice")
	}
	if len(atom.BoolSlices["flags"]) != 2 {
		t.Error("failed to store bool slice")
	}
	if len(atom.TimeSlices["dates"]) != 1 {
		t.Error("failed to store time slice")
	}
	if len(atom.ByteSlices["chunks"]) != 2 {
		t.Error("failed to store byte slice")
	}

	// Test nested Atom storage
	child := Atom{Strings: map[string]string{"name": "child"}}
	atom.Nested["child"] = child

	if atom.Nested["child"].Strings["name"] != "child" {
		t.Error("failed to store nested Atom")
	}

	// Test nested Atom slice storage
	child1 := Atom{Strings: map[string]string{"name": "item-1"}}
	child2 := Atom{Strings: map[string]string{"name": "item-2"}}
	atom.NestedSlices["items"] = []Atom{child1, child2}

	if len(atom.NestedSlices["items"]) != 2 {
		t.Error("failed to store nested Atom slice")
	}
	if atom.NestedSlices["items"][0].Strings["name"] != "item-1" {
		t.Error("failed to store first nested Atom in slice")
	}
}

func TestField(t *testing.T) {
	fd := Field{
		Name:  "CustomerID",
		Table: TableStrings,
	}

	if fd.Name != "CustomerID" {
		t.Errorf("expected Name CustomerID, got %s", fd.Name)
	}
	if fd.Table != TableStrings {
		t.Errorf("expected Table TableStrings, got %s", fd.Table)
	}
}

func TestFieldAllTables(t *testing.T) {
	tests := []struct {
		name  string
		table Table
	}{
		{"StringField", TableStrings},
		{"IntField", TableInts},
		{"UintField", TableUints},
		{"FloatField", TableFloats},
		{"BoolField", TableBools},
		{"TimeField", TableTimes},
		{"BytesField", TableBytes},
		{"BytePtrField", TableBytePtrs},
		{"StringPtrField", TableStringPtrs},
		{"IntPtrField", TableIntPtrs},
		{"UintPtrField", TableUintPtrs},
		{"FloatPtrField", TableFloatPtrs},
		{"BoolPtrField", TableBoolPtrs},
		{"TimePtrField", TableTimePtrs},
		{"StringSliceField", TableStringSlices},
		{"IntSliceField", TableIntSlices},
		{"UintSliceField", TableUintSlices},
		{"FloatSliceField", TableFloatSlices},
		{"BoolSliceField", TableBoolSlices},
		{"TimeSliceField", TableTimeSlices},
		{"ByteSliceField", TableByteSlices},
	}

	for _, tt := range tests {
		fd := Field{Name: tt.name, Table: tt.table}
		if fd.Name != tt.name {
			t.Errorf("expected Name %s, got %s", tt.name, fd.Name)
		}
		if fd.Table != tt.table {
			t.Errorf("expected Table %s, got %s", tt.table, fd.Table)
		}
	}
}
