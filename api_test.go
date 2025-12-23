package atom

import (
	"testing"
	"time"
)

func TestTableTypeConstants(t *testing.T) {
	tests := []struct {
		table    TableType
		expected string
	}{
		{TableStrings, "strings"},
		{TableInts, "ints"},
		{TableFloats, "floats"},
		{TableBools, "bools"},
		{TableTimes, "times"},
	}

	for _, tt := range tests {
		if string(tt.table) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.table)
		}
	}
}

func TestAllTables(t *testing.T) {
	tables := AllTables()

	if len(tables) != 5 {
		t.Errorf("expected 5 tables, got %d", len(tables))
	}

	expected := []TableType{TableStrings, TableInts, TableFloats, TableBools, TableTimes}
	for i, table := range tables {
		if table != expected[i] {
			t.Errorf("expected %s at index %d, got %s", expected[i], i, table)
		}
	}
}

func TestTableTypePrefix(t *testing.T) {
	tests := []struct {
		table    TableType
		expected string
	}{
		{TableStrings, "strings:"},
		{TableInts, "ints:"},
		{TableFloats, "floats:"},
		{TableBools, "bools:"},
		{TableTimes, "times:"},
	}

	for _, tt := range tests {
		if tt.table.Prefix() != tt.expected {
			t.Errorf("expected prefix %s for %s, got %s", tt.expected, tt.table, tt.table.Prefix())
		}
	}
}

func TestNewAtoms(t *testing.T) {
	id := "test-id-123"
	atoms := NewAtoms(id)

	if atoms.ID != id {
		t.Errorf("expected ID %s, got %s", id, atoms.ID)
	}

	if atoms.Strings == nil {
		t.Error("expected Strings map to be initialized")
	}
	if atoms.Ints == nil {
		t.Error("expected Ints map to be initialized")
	}
	if atoms.Floats == nil {
		t.Error("expected Floats map to be initialized")
	}
	if atoms.Bools == nil {
		t.Error("expected Bools map to be initialized")
	}
	if atoms.Times == nil {
		t.Error("expected Times map to be initialized")
	}
}

func TestNewAtomsEmptyID(t *testing.T) {
	atoms := NewAtoms("")

	if atoms.ID != "" {
		t.Errorf("expected empty ID, got %s", atoms.ID)
	}

	// Maps should still be initialized
	if atoms.Strings == nil {
		t.Error("expected Strings map to be initialized even with empty ID")
	}
}

func TestAtomsMapsAreUsable(t *testing.T) {
	atoms := NewAtoms("test")

	// Verify maps can be written to
	atoms.Strings["key"] = "value"
	atoms.Ints["count"] = 42
	atoms.Floats["rate"] = 3.14
	atoms.Bools["active"] = true
	atoms.Times["created"] = time.Now()

	if atoms.Strings["key"] != "value" {
		t.Error("failed to store string value")
	}
	if atoms.Ints["count"] != 42 {
		t.Error("failed to store int value")
	}
	if atoms.Floats["rate"] != 3.14 {
		t.Error("failed to store float value")
	}
	if atoms.Bools["active"] != true {
		t.Error("failed to store bool value")
	}
	if atoms.Times["created"].IsZero() {
		t.Error("failed to store time value")
	}
}

func TestFieldDescriptor(t *testing.T) {
	fd := FieldDescriptor{
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

func TestFieldDescriptorAllTables(t *testing.T) {
	tests := []struct {
		name  string
		table TableType
	}{
		{"StringField", TableStrings},
		{"IntField", TableInts},
		{"FloatField", TableFloats},
		{"BoolField", TableBools},
		{"TimeField", TableTimes},
	}

	for _, tt := range tests {
		fd := FieldDescriptor{Name: tt.name, Table: tt.table}
		if fd.Name != tt.name {
			t.Errorf("expected Name %s, got %s", tt.name, fd.Name)
		}
		if fd.Table != tt.table {
			t.Errorf("expected Table %s, got %s", tt.table, fd.Table)
		}
	}
}
