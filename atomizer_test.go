package atom

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestUser is a test struct.
type TestUser struct {
	ID        string
	Name      string
	Age       int64
	Balance   float64
	Active    bool
	CreatedAt time.Time
	Password  string `atom:"-"` // excluded
}

func (u TestUser) Validate() error {
	if strings.Contains(u.ID, ":") {
		return errors.New("ID contains reserved character")
	}
	return nil
}

// Atomize/Deatomize callbacks for TestUser.
func atomizeUser(u *TestUser) Atoms {
	atoms := NewAtoms(u.ID)
	atoms.Strings["Name"] = u.Name
	atoms.Ints["Age"] = u.Age
	atoms.Floats["Balance"] = u.Balance
	atoms.Bools["Active"] = u.Active
	atoms.Times["CreatedAt"] = u.CreatedAt
	return *atoms
}

func deatomizeUser(atoms Atoms) (*TestUser, error) {
	return &TestUser{
		ID:        atoms.ID,
		Name:      atoms.Strings["Name"],
		Age:       atoms.Ints["Age"],
		Balance:   atoms.Floats["Balance"],
		Active:    atoms.Bools["Active"],
		CreatedAt: atoms.Times["CreatedAt"],
	}, nil
}

// TestMinimal is a minimal struct for edge case testing.
type TestMinimal struct {
	ID string
}

func (TestMinimal) Validate() error {
	return nil
}

func atomizeMinimal(m *TestMinimal) Atoms {
	return *NewAtoms(m.ID)
}

func deatomizeMinimal(atoms Atoms) (*TestMinimal, error) {
	return &TestMinimal{ID: atoms.ID}, nil
}

// TestDeatomizeError callbacks that return an error.
type TestDeatomizeError struct {
	ID string
}

func (TestDeatomizeError) Validate() error {
	return nil
}

func atomizeDeatomizeError(e *TestDeatomizeError) Atoms {
	return *NewAtoms(e.ID)
}

func deatomizeDeatomizeError(_ Atoms) (*TestDeatomizeError, error) {
	return nil, errors.New("deatomize error")
}

func newTestUserAtomizer() *Atomizer[TestUser] {
	return New[TestUser](atomizeUser, deatomizeUser)
}

func TestNew(t *testing.T) {
	atomizer := newTestUserAtomizer()

	if atomizer == nil {
		t.Fatal("expected non-nil atomizer")
	}
}

func TestAtomizerMetadata(t *testing.T) {
	atomizer := newTestUserAtomizer()

	meta := atomizer.Metadata()
	if meta.TypeName != "TestUser" {
		t.Errorf("expected TypeName TestUser, got %s", meta.TypeName)
	}
}

func TestAtomizerFields(t *testing.T) {
	atomizer := newTestUserAtomizer()

	fields := atomizer.Fields()
	if len(fields) == 0 {
		t.Error("expected at least one field")
	}

	// Check that we have expected fields
	fieldNames := make(map[string]TableType)
	for _, f := range fields {
		fieldNames[f.Name] = f.Table
	}

	// ID should be string
	if table, ok := fieldNames["ID"]; ok {
		if table != TableStrings {
			t.Errorf("expected ID to be TableStrings, got %s", table)
		}
	}

	// Name should be string
	if table, ok := fieldNames["Name"]; ok {
		if table != TableStrings {
			t.Errorf("expected Name to be TableStrings, got %s", table)
		}
	}

	// Age should be int
	if table, ok := fieldNames["Age"]; ok {
		if table != TableInts {
			t.Errorf("expected Age to be TableInts, got %s", table)
		}
	}

	// Balance should be float
	if table, ok := fieldNames["Balance"]; ok {
		if table != TableFloats {
			t.Errorf("expected Balance to be TableFloats, got %s", table)
		}
	}

	// Active should be bool
	if table, ok := fieldNames["Active"]; ok {
		if table != TableBools {
			t.Errorf("expected Active to be TableBools, got %s", table)
		}
	}

	// CreatedAt should be time
	if table, ok := fieldNames["CreatedAt"]; ok {
		if table != TableTimes {
			t.Errorf("expected CreatedAt to be TableTimes, got %s", table)
		}
	}
}

func TestAtomizerFieldsExcludesTaggedFields(t *testing.T) {
	atomizer := newTestUserAtomizer()

	fields := atomizer.Fields()
	for _, f := range fields {
		if f.Name == "Password" {
			t.Error("Password field should be excluded via atom:\"-\" tag")
		}
	}
}

func TestAtomizerFieldsIn(t *testing.T) {
	atomizer := newTestUserAtomizer()

	stringFields := atomizer.FieldsIn(TableStrings)
	if len(stringFields) == 0 {
		t.Error("expected at least one string field")
	}

	intFields := atomizer.FieldsIn(TableInts)
	if len(intFields) == 0 {
		t.Error("expected at least one int field")
	}
}

func TestAtomizerTableFor(t *testing.T) {
	atomizer := newTestUserAtomizer()

	// Known field
	table, ok := atomizer.TableFor("Name")
	if !ok {
		t.Error("expected Name to be a known field")
	}
	if table != TableStrings {
		t.Errorf("expected Name to be TableStrings, got %s", table)
	}

	// Unknown field
	_, ok = atomizer.TableFor("Unknown")
	if ok {
		t.Error("expected Unknown to not be a known field")
	}
}

func TestAtomizerAtomize(t *testing.T) {
	atomizer := newTestUserAtomizer()

	user := &TestUser{
		ID:        "user-123",
		Name:      "Alice",
		Age:       30,
		Balance:   100.50,
		Active:    true,
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	atoms, err := atomizer.Atomize(user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atoms.ID != "user-123" {
		t.Errorf("expected ID user-123, got %s", atoms.ID)
	}
	if atoms.Strings["Name"] != "Alice" {
		t.Errorf("expected Name Alice, got %s", atoms.Strings["Name"])
	}
	if atoms.Ints["Age"] != 30 {
		t.Errorf("expected Age 30, got %d", atoms.Ints["Age"])
	}
	if atoms.Floats["Balance"] != 100.50 {
		t.Errorf("expected Balance 100.50, got %f", atoms.Floats["Balance"])
	}
	if atoms.Bools["Active"] != true {
		t.Errorf("expected Active true, got %v", atoms.Bools["Active"])
	}
	if !atoms.Times["CreatedAt"].Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected CreatedAt 2024-01-01, got %v", atoms.Times["CreatedAt"])
	}
}

func TestAtomizerAtomizeMissingID(t *testing.T) {
	atomizer := newTestUserAtomizer()

	user := &TestUser{
		ID:   "", // empty ID
		Name: "Alice",
	}

	_, err := atomizer.Atomize(user)
	if !errors.Is(err, ErrMissingID) {
		t.Errorf("expected ErrMissingID, got %v", err)
	}
}

func TestAtomizerAtomizeValidationError(t *testing.T) {
	atomizer := newTestUserAtomizer()

	// ID with colon should fail validation
	user := &TestUser{
		ID:   "user:invalid",
		Name: "Bad ID",
	}

	_, err := atomizer.Atomize(user)
	if err == nil {
		t.Error("expected validation error for ID containing colon")
	}
	if !strings.Contains(err.Error(), "reserved character") {
		t.Errorf("expected reserved character error, got %v", err)
	}
}

func TestAtomizerDeatomize(t *testing.T) {
	atomizer := newTestUserAtomizer()

	atoms := NewAtoms("user-456")
	atoms.Strings["Name"] = "Bob"
	atoms.Ints["Age"] = 25
	atoms.Floats["Balance"] = 50.25
	atoms.Bools["Active"] = false
	atoms.Times["CreatedAt"] = time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	user, err := atomizer.Deatomize(*atoms)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-456" {
		t.Errorf("expected ID user-456, got %s", user.ID)
	}
	if user.Name != "Bob" {
		t.Errorf("expected Name Bob, got %s", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("expected Age 25, got %d", user.Age)
	}
	if user.Balance != 50.25 {
		t.Errorf("expected Balance 50.25, got %f", user.Balance)
	}
	if user.Active != false {
		t.Errorf("expected Active false, got %v", user.Active)
	}
	if !user.CreatedAt.Equal(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected CreatedAt 2024-06-15, got %v", user.CreatedAt)
	}
}

func TestAtomizerDeatomizeError(t *testing.T) {
	atomizer := New[TestDeatomizeError](atomizeDeatomizeError, deatomizeDeatomizeError)

	atoms := NewAtoms("error-obj")
	_, err := atomizer.Deatomize(*atoms)
	if err == nil {
		t.Error("expected error from Deatomize")
	}
}

func TestAtomizerRoundTrip(t *testing.T) {
	atomizer := newTestUserAtomizer()

	original := &TestUser{
		ID:        "user-roundtrip",
		Name:      "Charlie",
		Age:       35,
		Balance:   200.75,
		Active:    true,
		CreatedAt: time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC),
	}

	// Atomize
	atoms, err := atomizer.Atomize(original)
	if err != nil {
		t.Fatalf("unexpected error on atomize: %v", err)
	}

	// Deatomize
	restored, err := atomizer.Deatomize(atoms)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	// Verify round-trip
	if restored.ID != original.ID {
		t.Errorf("expected ID %s, got %s", original.ID, restored.ID)
	}
	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if restored.Age != original.Age {
		t.Errorf("expected Age %d, got %d", original.Age, restored.Age)
	}
	if restored.Balance != original.Balance {
		t.Errorf("expected Balance %f, got %f", original.Balance, restored.Balance)
	}
	if restored.Active != original.Active {
		t.Errorf("expected Active %v, got %v", original.Active, restored.Active)
	}
	if !restored.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", original.CreatedAt, restored.CreatedAt)
	}
}

func TestTableFromType(t *testing.T) {
	tests := []struct {
		goType   string
		expected TableType
	}{
		{"string", TableStrings},
		{"int", TableInts},
		{"int8", TableInts},
		{"int16", TableInts},
		{"int32", TableInts},
		{"int64", TableInts},
		{"uint", TableInts},
		{"uint8", TableInts},
		{"uint16", TableInts},
		{"uint32", TableInts},
		{"uint64", TableInts},
		{"float32", TableFloats},
		{"float64", TableFloats},
		{"bool", TableBools},
		{"time.Time", TableTimes},
		{"unknown", ""},
		{"[]string", ""},
		{"map[string]int", ""},
	}

	for _, tt := range tests {
		result := tableFromType(tt.goType)
		if result != tt.expected {
			t.Errorf("tableFromType(%q) = %q, expected %q", tt.goType, result, tt.expected)
		}
	}
}

func TestAtomizerMinimalStruct(t *testing.T) {
	atomizer := New[TestMinimal](atomizeMinimal, deatomizeMinimal)

	obj := &TestMinimal{ID: "minimal-1"}
	atoms, err := atomizer.Atomize(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atoms.ID != "minimal-1" {
		t.Errorf("expected ID minimal-1, got %s", atoms.ID)
	}

	restored, err := atomizer.Deatomize(atoms)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restored.ID != "minimal-1" {
		t.Errorf("expected ID minimal-1, got %s", restored.ID)
	}
}
