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

func TestClone_Nil(t *testing.T) {
	var a *Atom
	clone := a.Clone()
	if clone != nil {
		t.Error("expected nil clone for nil receiver")
	}
}

func TestClone_Empty(t *testing.T) {
	a := &Atom{}
	clone := a.Clone()
	if clone == nil {
		t.Fatal("expected non-nil clone")
	}
	if clone == a {
		t.Error("clone should be a different pointer")
	}
}

func TestClone_Scalars(t *testing.T) {
	now := time.Now()
	a := &Atom{
		Strings: map[string]string{"name": "test"},
		Ints:    map[string]int64{"count": 42},
		Uints:   map[string]uint64{"id": 123},
		Floats:  map[string]float64{"rate": 3.14},
		Bools:   map[string]bool{"active": true},
		Times:   map[string]time.Time{"created": now},
		Bytes:   map[string][]byte{"data": {0x01, 0x02, 0x03}},
	}

	clone := a.Clone()

	// Verify values copied
	if clone.Strings["name"] != "test" {
		t.Error("string not cloned")
	}
	if clone.Ints["count"] != 42 {
		t.Error("int not cloned")
	}
	if clone.Uints["id"] != 123 {
		t.Error("uint not cloned")
	}
	if clone.Floats["rate"] != 3.14 {
		t.Error("float not cloned")
	}
	if clone.Bools["active"] != true {
		t.Error("bool not cloned")
	}
	if !clone.Times["created"].Equal(now) {
		t.Error("time not cloned")
	}
	if len(clone.Bytes["data"]) != 3 || clone.Bytes["data"][0] != 0x01 {
		t.Error("bytes not cloned")
	}

	// Verify deep copy - modifying original doesn't affect clone
	a.Strings["name"] = "modified"
	a.Ints["count"] = 999
	a.Bytes["data"][0] = 0xFF

	if clone.Strings["name"] != "test" {
		t.Error("clone string affected by original modification")
	}
	if clone.Ints["count"] != 42 {
		t.Error("clone int affected by original modification")
	}
	if clone.Bytes["data"][0] != 0x01 {
		t.Error("clone bytes affected by original modification")
	}
}

func TestClone_Pointers(t *testing.T) {
	strVal := "hello"
	intVal := int64(42)
	uintVal := uint64(123)
	floatVal := 3.14
	boolVal := true
	timeVal := time.Now()
	bytesVal := []byte{0x01, 0x02}

	a := &Atom{
		StringPtrs: map[string]*string{"s": &strVal},
		IntPtrs:    map[string]*int64{"i": &intVal},
		UintPtrs:   map[string]*uint64{"u": &uintVal},
		FloatPtrs:  map[string]*float64{"f": &floatVal},
		BoolPtrs:   map[string]*bool{"b": &boolVal},
		TimePtrs:   map[string]*time.Time{"t": &timeVal},
		BytePtrs:   map[string]*[]byte{"by": &bytesVal},
	}

	clone := a.Clone()

	// Verify values copied
	if *clone.StringPtrs["s"] != "hello" {
		t.Error("string pointer not cloned")
	}
	if *clone.IntPtrs["i"] != 42 {
		t.Error("int pointer not cloned")
	}
	if *clone.UintPtrs["u"] != 123 {
		t.Error("uint pointer not cloned")
	}
	if *clone.FloatPtrs["f"] != 3.14 {
		t.Error("float pointer not cloned")
	}
	if *clone.BoolPtrs["b"] != true {
		t.Error("bool pointer not cloned")
	}
	if !clone.TimePtrs["t"].Equal(timeVal) {
		t.Error("time pointer not cloned")
	}
	if len(*clone.BytePtrs["by"]) != 2 {
		t.Error("byte pointer not cloned")
	}

	// Verify deep copy - pointers are different
	if clone.StringPtrs["s"] == a.StringPtrs["s"] {
		t.Error("string pointer should be different address")
	}
	if clone.IntPtrs["i"] == a.IntPtrs["i"] {
		t.Error("int pointer should be different address")
	}

	// Modify original pointer value
	*a.StringPtrs["s"] = "modified"
	if *clone.StringPtrs["s"] != "hello" {
		t.Error("clone string pointer affected by original modification")
	}
}

func TestClone_Slices(t *testing.T) {
	a := &Atom{
		StringSlices: map[string][]string{"tags": {"a", "b", "c"}},
		IntSlices:    map[string][]int64{"nums": {1, 2, 3}},
		UintSlices:   map[string][]uint64{"ids": {10, 20}},
		FloatSlices:  map[string][]float64{"rates": {1.1, 2.2}},
		BoolSlices:   map[string][]bool{"flags": {true, false}},
		TimeSlices:   map[string][]time.Time{"dates": {time.Now()}},
		ByteSlices:   map[string][][]byte{"chunks": {{0x01}, {0x02, 0x03}}},
	}

	clone := a.Clone()

	// Verify values copied
	if len(clone.StringSlices["tags"]) != 3 || clone.StringSlices["tags"][0] != "a" {
		t.Error("string slice not cloned")
	}
	if len(clone.IntSlices["nums"]) != 3 || clone.IntSlices["nums"][0] != 1 {
		t.Error("int slice not cloned")
	}
	if len(clone.ByteSlices["chunks"]) != 2 || clone.ByteSlices["chunks"][1][0] != 0x02 {
		t.Error("byte slices not cloned")
	}

	// Verify deep copy
	a.StringSlices["tags"][0] = "modified"
	a.IntSlices["nums"][0] = 999
	a.ByteSlices["chunks"][0][0] = 0xFF

	if clone.StringSlices["tags"][0] != "a" {
		t.Error("clone string slice affected by original modification")
	}
	if clone.IntSlices["nums"][0] != 1 {
		t.Error("clone int slice affected by original modification")
	}
	if clone.ByteSlices["chunks"][0][0] != 0x01 {
		t.Error("clone byte slices affected by original modification")
	}
}

func TestClone_Nested(t *testing.T) {
	child := Atom{
		Strings: map[string]string{"name": "child"},
		Ints:    map[string]int64{"value": 100},
	}
	a := &Atom{
		Strings: map[string]string{"name": "parent"},
		Nested:  map[string]Atom{"child": child},
	}

	clone := a.Clone()

	// Verify nested copied
	if clone.Nested["child"].Strings["name"] != "child" {
		t.Error("nested atom not cloned")
	}
	if clone.Nested["child"].Ints["value"] != 100 {
		t.Error("nested atom int not cloned")
	}

	// Verify deep copy
	a.Nested["child"].Strings["name"] = "modified"
	if clone.Nested["child"].Strings["name"] != "child" {
		t.Error("clone nested affected by original modification")
	}
}

func TestClone_NestedSlices(t *testing.T) {
	items := []Atom{
		{Strings: map[string]string{"name": "item1"}},
		{Strings: map[string]string{"name": "item2"}},
	}
	a := &Atom{
		NestedSlices: map[string][]Atom{"items": items},
	}

	clone := a.Clone()

	// Verify nested slices copied
	if len(clone.NestedSlices["items"]) != 2 {
		t.Error("nested slices not cloned")
	}
	if clone.NestedSlices["items"][0].Strings["name"] != "item1" {
		t.Error("nested slice item not cloned")
	}

	// Verify deep copy
	a.NestedSlices["items"][0].Strings["name"] = "modified"
	if clone.NestedSlices["items"][0].Strings["name"] != "item1" {
		t.Error("clone nested slice affected by original modification")
	}
}
