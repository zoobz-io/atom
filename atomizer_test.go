package atom

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// mustUse is a test helper that calls Use and fails if there's an error.
func mustUse[T any](t *testing.T) *Atomizer[T] {
	t.Helper()
	atomizer, err := Use[T]()
	if err != nil {
		t.Fatalf("Use[%T]() failed: %v", *new(T), err)
	}
	return atomizer
}

// TestUser is a test struct with common field types.
type TestUser struct {
	CreatedAt time.Time
	Name      string
	_         string
	Age       int64
	Balance   float64
	Active    bool
}

// TestMinimal is a minimal struct for edge case testing.
type TestMinimal struct {
	Name string
}

// TestExtended tests bytes and pointer types.
type TestExtended struct {
	OptName *string
	OptAge  *int64
	OptRate *float64
	OptFlag *bool
	OptTime *time.Time
	Name    string
	Data    []byte
}

// TestWithSlices tests slice field handling.
type TestWithSlices struct {
	Name   string
	Tags   []string
	Scores []int64
	Rates  []float64
}

// TestAddress is a nested struct for testing.
type TestAddress struct {
	Street string
	City   string
}

// TestWithNested tests nested struct handling.
type TestWithNested struct {
	Name    string
	Address TestAddress
}

// TestWithNestedSlice tests slice of nested structs.
type TestWithNestedSlice struct {
	Name      string
	Addresses []TestAddress
}

// TestWithNestedPtr tests pointer to nested struct handling.
type TestWithNestedPtr struct {
	Address *TestAddress
	Name    string
}

// TestMoreSlices tests additional slice types for coverage.
type TestMoreSlices struct {
	Flags      []bool
	Timestamps []time.Time
	Chunks     [][]byte
}

// TestSelfReferential tests circular type references.
type TestSelfReferential struct {
	Name     string
	Children []TestSelfReferential
}

// TestWithNestedPtrSlice tests slice of pointers to nested structs.
type TestWithNestedPtrSlice struct {
	Name      string
	Addresses []*TestAddress
}

func TestUse(t *testing.T) {
	atomizer, err := Use[TestUser]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomizer == nil {
		t.Fatal("expected non-nil atomizer")
	}
}

func TestUseIdempotent(t *testing.T) {
	a1 := mustUse[TestUser](t)
	a2 := mustUse[TestUser](t)
	if a1.inner != a2.inner {
		t.Error("Use should return same underlying atomizer")
	}
}

func TestAtomizerMetadata(t *testing.T) {
	atomizer := mustUse[TestUser](t)

	meta := atomizer.Spec()
	if meta.TypeName != "TestUser" {
		t.Errorf("expected TypeName TestUser, got %s", meta.TypeName)
	}
}

func TestAtomMetadata(t *testing.T) {
	atomizer := mustUse[TestUser](t)

	user := &TestUser{Name: "Alice", Age: 30}
	atom := atomizer.Atomize(user)

	if atom.Spec.TypeName != "TestUser" {
		t.Errorf("expected atom.Spec.TypeName TestUser, got %s", atom.Spec.TypeName)
	}
}

func TestAtomizerFields(t *testing.T) {
	atomizer := mustUse[TestUser](t)

	fields := atomizer.Fields()
	if len(fields) == 0 {
		t.Error("expected at least one field")
	}

	// Check that we have expected fields
	fieldNames := make(map[string]Table)
	for _, f := range fields {
		fieldNames[f.Name] = f.Table
	}

	// Name should be string
	if table, ok := fieldNames["Name"]; ok {
		if table != TableStrings {
			t.Errorf("expected Name to be TableStrings, got %s", table)
		}
	} else {
		t.Error("expected Name field")
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

	// password should NOT be present (unexported)
	if _, ok := fieldNames["password"]; ok {
		t.Error("unexported field 'password' should not be present")
	}
}

func TestAtomizerFieldsIn(t *testing.T) {
	atomizer := mustUse[TestUser](t)

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
	atomizer := mustUse[TestUser](t)

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
	atomizer := mustUse[TestUser](t)

	user := &TestUser{
		Name:      "Alice",
		Age:       30,
		Balance:   100.50,
		Active:    true,
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	atom := atomizer.Atomize(user)

	if atom.Strings["Name"] != "Alice" {
		t.Errorf("expected Name Alice, got %s", atom.Strings["Name"])
	}
	if atom.Ints["Age"] != 30 {
		t.Errorf("expected Age 30, got %d", atom.Ints["Age"])
	}
	if atom.Floats["Balance"] != 100.50 {
		t.Errorf("expected Balance 100.50, got %f", atom.Floats["Balance"])
	}
	if atom.Bools["Active"] != true {
		t.Errorf("expected Active true, got %v", atom.Bools["Active"])
	}
	if !atom.Times["CreatedAt"].Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected CreatedAt 2024-01-01, got %v", atom.Times["CreatedAt"])
	}
}

func TestAtomizerDeatomize(t *testing.T) {
	atomizer := mustUse[TestUser](t)

	atom := atomizer.NewAtom()
	atom.Strings["Name"] = "Bob"
	atom.Ints["Age"] = 25
	atom.Floats["Balance"] = 50.25
	atom.Bools["Active"] = false
	atom.Times["CreatedAt"] = time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	user, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

func TestAtomizerRoundTrip(t *testing.T) {
	atomizer := mustUse[TestUser](t)

	original := &TestUser{
		Name:      "Charlie",
		Age:       35,
		Balance:   200.75,
		Active:    true,
		CreatedAt: time.Date(2024, 3, 15, 12, 30, 0, 0, time.UTC),
	}

	// Atomize
	atom := atomizer.Atomize(original)

	// Deatomize
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	// Verify round-trip
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

func TestAtomizerMinimalStruct(t *testing.T) {
	atomizer := mustUse[TestMinimal](t)

	obj := &TestMinimal{Name: "minimal"}
	atom := atomizer.Atomize(obj)

	if atom.Strings["Name"] != "minimal" {
		t.Errorf("expected Name minimal, got %s", atom.Strings["Name"])
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restored.Name != "minimal" {
		t.Errorf("expected Name minimal, got %s", restored.Name)
	}
}

func TestAtomizerExtendedFields(t *testing.T) {
	atomizer := mustUse[TestExtended](t)

	fields := atomizer.Fields()
	fieldMap := make(map[string]Table)
	for _, f := range fields {
		fieldMap[f.Name] = f.Table
	}

	// Verify bytes field
	if table, ok := fieldMap["Data"]; ok {
		if table != TableBytes {
			t.Errorf("expected Data to be TableBytes, got %s", table)
		}
	} else {
		t.Error("expected Data field to be present")
	}

	// Verify pointer fields
	if table, ok := fieldMap["OptName"]; ok {
		if table != TableStringPtrs {
			t.Errorf("expected OptName to be TableStringPtrs, got %s", table)
		}
	}
	if table, ok := fieldMap["OptAge"]; ok {
		if table != TableIntPtrs {
			t.Errorf("expected OptAge to be TableIntPtrs, got %s", table)
		}
	}
	if table, ok := fieldMap["OptRate"]; ok {
		if table != TableFloatPtrs {
			t.Errorf("expected OptRate to be TableFloatPtrs, got %s", table)
		}
	}
	if table, ok := fieldMap["OptFlag"]; ok {
		if table != TableBoolPtrs {
			t.Errorf("expected OptFlag to be TableBoolPtrs, got %s", table)
		}
	}
	if table, ok := fieldMap["OptTime"]; ok {
		if table != TableTimePtrs {
			t.Errorf("expected OptTime to be TableTimePtrs, got %s", table)
		}
	}
}

func TestAtomizerExtendedRoundTripWithValues(t *testing.T) {
	atomizer := mustUse[TestExtended](t)

	name := "Alice"
	age := int64(30)
	rate := 3.14
	flag := true
	ts := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	original := &TestExtended{
		Name:    "test",
		Data:    []byte{0x01, 0x02, 0x03},
		OptName: &name,
		OptAge:  &age,
		OptRate: &rate,
		OptFlag: &flag,
		OptTime: &ts,
	}

	atom := atomizer.Atomize(original)

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if !bytes.Equal(restored.Data, original.Data) {
		t.Errorf("expected Data %v, got %v", original.Data, restored.Data)
	}
	if *restored.OptName != *original.OptName {
		t.Errorf("expected OptName %s, got %s", *original.OptName, *restored.OptName)
	}
	if *restored.OptAge != *original.OptAge {
		t.Errorf("expected OptAge %d, got %d", *original.OptAge, *restored.OptAge)
	}
	if *restored.OptRate != *original.OptRate {
		t.Errorf("expected OptRate %f, got %f", *original.OptRate, *restored.OptRate)
	}
	if *restored.OptFlag != *original.OptFlag {
		t.Errorf("expected OptFlag %v, got %v", *original.OptFlag, *restored.OptFlag)
	}
	if !restored.OptTime.Equal(*original.OptTime) {
		t.Errorf("expected OptTime %v, got %v", *original.OptTime, *restored.OptTime)
	}
}

func TestAtomizerExtendedRoundTripWithNils(t *testing.T) {
	atomizer := mustUse[TestExtended](t)

	original := &TestExtended{
		Name:    "test",
		Data:    nil,
		OptName: nil,
		OptAge:  nil,
		OptRate: nil,
		OptFlag: nil,
		OptTime: nil,
	}

	atom := atomizer.Atomize(original)

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if restored.Data != nil {
		t.Errorf("expected Data nil, got %v", restored.Data)
	}
	if restored.OptName != nil {
		t.Errorf("expected OptName nil, got %v", restored.OptName)
	}
	if restored.OptAge != nil {
		t.Errorf("expected OptAge nil, got %v", restored.OptAge)
	}
	if restored.OptRate != nil {
		t.Errorf("expected OptRate nil, got %v", restored.OptRate)
	}
	if restored.OptFlag != nil {
		t.Errorf("expected OptFlag nil, got %v", restored.OptFlag)
	}
	if restored.OptTime != nil {
		t.Errorf("expected OptTime nil, got %v", restored.OptTime)
	}
}

func TestAtomizerSliceFields(t *testing.T) {
	atomizer := mustUse[TestWithSlices](t)

	fields := atomizer.Fields()
	fieldMap := make(map[string]Table)
	for _, f := range fields {
		fieldMap[f.Name] = f.Table
	}

	if table, ok := fieldMap["Tags"]; ok {
		if table != TableStringSlices {
			t.Errorf("expected Tags to be TableStringSlices, got %s", table)
		}
	} else {
		t.Error("expected Tags field to be present")
	}

	if table, ok := fieldMap["Scores"]; ok {
		if table != TableIntSlices {
			t.Errorf("expected Scores to be TableIntSlices, got %s", table)
		}
	}

	if table, ok := fieldMap["Rates"]; ok {
		if table != TableFloatSlices {
			t.Errorf("expected Rates to be TableFloatSlices, got %s", table)
		}
	}
}

func TestAtomizerSliceRoundTrip(t *testing.T) {
	atomizer := mustUse[TestWithSlices](t)

	original := &TestWithSlices{
		Name:   "slices",
		Tags:   []string{"a", "b", "c"},
		Scores: []int64{100, 200, 300},
		Rates:  []float64{1.1, 2.2, 3.3},
	}

	atom := atomizer.Atomize(original)

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if len(restored.Tags) != len(original.Tags) {
		t.Errorf("expected %d tags, got %d", len(original.Tags), len(restored.Tags))
	}
	if len(restored.Scores) != len(original.Scores) {
		t.Errorf("expected %d scores, got %d", len(original.Scores), len(restored.Scores))
	}
	if len(restored.Rates) != len(original.Rates) {
		t.Errorf("expected %d rates, got %d", len(original.Rates), len(restored.Rates))
	}
}

func TestAtomizerNestedRoundTrip(t *testing.T) {
	atomizer := mustUse[TestWithNested](t)

	original := &TestWithNested{
		Name: "Alice",
		Address: TestAddress{
			Street: "123 Main St",
			City:   "Springfield",
		},
	}

	atom := atomizer.Atomize(original)

	// Verify nested Atom structure
	if _, ok := atom.Nested["Address"]; !ok {
		t.Fatal("expected Address in Nested map")
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if restored.Address.Street != original.Address.Street {
		t.Errorf("expected Street %s, got %s", original.Address.Street, restored.Address.Street)
	}
	if restored.Address.City != original.Address.City {
		t.Errorf("expected City %s, got %s", original.Address.City, restored.Address.City)
	}
}

func TestAtomizerNestedSliceRoundTrip(t *testing.T) {
	atomizer := mustUse[TestWithNestedSlice](t)

	original := &TestWithNestedSlice{
		Name: "Alice",
		Addresses: []TestAddress{
			{Street: "123 Main St", City: "Springfield"},
			{Street: "456 Oak Ave", City: "Shelbyville"},
		},
	}

	atom := atomizer.Atomize(original)

	// Verify nested slice structure
	if _, ok := atom.NestedSlices["Addresses"]; !ok {
		t.Fatal("expected Addresses in NestedSlices map")
	}
	if len(atom.NestedSlices["Addresses"]) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(atom.NestedSlices["Addresses"]))
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if len(restored.Addresses) != len(original.Addresses) {
		t.Errorf("expected %d addresses, got %d", len(original.Addresses), len(restored.Addresses))
	}
	if restored.Addresses[0].Street != original.Addresses[0].Street {
		t.Errorf("expected Street %s, got %s", original.Addresses[0].Street, restored.Addresses[0].Street)
	}
	if restored.Addresses[1].City != original.Addresses[1].City {
		t.Errorf("expected City %s, got %s", original.Addresses[1].City, restored.Addresses[1].City)
	}
}

func TestAtomizerNestedPtrRoundTrip(t *testing.T) {
	atomizer := mustUse[TestWithNestedPtr](t)

	// Test with non-nil pointer
	original := &TestWithNestedPtr{
		Name: "Bob",
		Address: &TestAddress{
			Street: "789 Pine Rd",
			City:   "Metropolis",
		},
	}

	atom := atomizer.Atomize(original)

	// Verify nested structure
	if _, ok := atom.Nested["Address"]; !ok {
		t.Fatal("expected Address in Nested map")
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if restored.Address == nil {
		t.Fatal("expected Address to be non-nil")
	}
	if restored.Address.Street != original.Address.Street {
		t.Errorf("expected Street %s, got %s", original.Address.Street, restored.Address.Street)
	}
	if restored.Address.City != original.Address.City {
		t.Errorf("expected City %s, got %s", original.Address.City, restored.Address.City)
	}
}

func TestAtomizerNestedPtrNil(t *testing.T) {
	atomizer := mustUse[TestWithNestedPtr](t)

	// Test with nil pointer
	original := &TestWithNestedPtr{
		Name:    "Charlie",
		Address: nil,
	}

	atom := atomizer.Atomize(original)

	// Nil pointer should not be stored in Nested map
	if _, ok := atom.Nested["Address"]; ok {
		t.Error("expected Address to NOT be in Nested map when nil")
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if restored.Address != nil {
		t.Error("expected Address to be nil")
	}
}

func TestAtomizerMoreSlicesRoundTrip(t *testing.T) {
	atomizer := mustUse[TestMoreSlices](t)

	original := &TestMoreSlices{
		Flags: []bool{true, false, true, false},
		Timestamps: []time.Time{
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC),
		},
		Chunks: [][]byte{
			{0x01, 0x02, 0x03},
			{0x04, 0x05},
			{0x06},
		},
	}

	atom := atomizer.Atomize(original)

	// Verify slice storage
	if len(atom.BoolSlices["Flags"]) != 4 {
		t.Errorf("expected 4 flags, got %d", len(atom.BoolSlices["Flags"]))
	}
	if len(atom.TimeSlices["Timestamps"]) != 2 {
		t.Errorf("expected 2 timestamps, got %d", len(atom.TimeSlices["Timestamps"]))
	}
	if len(atom.ByteSlices["Chunks"]) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(atom.ByteSlices["Chunks"]))
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	// Verify bool slice
	if len(restored.Flags) != len(original.Flags) {
		t.Errorf("expected %d flags, got %d", len(original.Flags), len(restored.Flags))
	}
	for i, v := range original.Flags {
		if restored.Flags[i] != v {
			t.Errorf("Flags[%d]: expected %v, got %v", i, v, restored.Flags[i])
		}
	}

	// Verify time slice
	if len(restored.Timestamps) != len(original.Timestamps) {
		t.Errorf("expected %d timestamps, got %d", len(original.Timestamps), len(restored.Timestamps))
	}
	for i, v := range original.Timestamps {
		if !restored.Timestamps[i].Equal(v) {
			t.Errorf("Timestamps[%d]: expected %v, got %v", i, v, restored.Timestamps[i])
		}
	}

	// Verify byte slice
	if len(restored.Chunks) != len(original.Chunks) {
		t.Errorf("expected %d chunks, got %d", len(original.Chunks), len(restored.Chunks))
	}
	for i, v := range original.Chunks {
		if !bytes.Equal(restored.Chunks[i], v) {
			t.Errorf("Chunks[%d]: expected %v, got %v", i, v, restored.Chunks[i])
		}
	}
}

func TestAtomizerSelfReferential(t *testing.T) {
	atomizer := mustUse[TestSelfReferential](t)

	original := &TestSelfReferential{
		Name: "root",
		Children: []TestSelfReferential{
			{Name: "child1"},
			{
				Name: "child2",
				Children: []TestSelfReferential{
					{Name: "grandchild"},
				},
			},
		},
	}

	atom := atomizer.Atomize(original)

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != "root" {
		t.Errorf("expected Name root, got %s", restored.Name)
	}
	if len(restored.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(restored.Children))
	}
	if restored.Children[0].Name != "child1" {
		t.Errorf("expected child1, got %s", restored.Children[0].Name)
	}
	if restored.Children[1].Name != "child2" {
		t.Errorf("expected child2, got %s", restored.Children[1].Name)
	}
	if len(restored.Children[1].Children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(restored.Children[1].Children))
	}
	if restored.Children[1].Children[0].Name != "grandchild" {
		t.Errorf("expected grandchild, got %s", restored.Children[1].Children[0].Name)
	}
}

func TestAtomizerNewAtom(t *testing.T) {
	atomizer := mustUse[TestUser](t)

	atom := atomizer.NewAtom()

	// TestUser has: string (Name), int64 (Age), float64 (Balance), bool (Active), time.Time (CreatedAt)
	if atom.Strings == nil {
		t.Error("expected Strings to be allocated")
	}
	if atom.Ints == nil {
		t.Error("expected Ints to be allocated")
	}
	if atom.Floats == nil {
		t.Error("expected Floats to be allocated")
	}
	if atom.Bools == nil {
		t.Error("expected Bools to be allocated")
	}
	if atom.Times == nil {
		t.Error("expected Times to be allocated")
	}

	// These should NOT be allocated for TestUser
	if atom.Bytes != nil {
		t.Error("expected Bytes to be nil")
	}
	if atom.StringPtrs != nil {
		t.Error("expected StringPtrs to be nil")
	}
}

func TestAtomizerNewAtomExtended(t *testing.T) {
	atomizer := mustUse[TestExtended](t)

	atom := atomizer.NewAtom()

	// TestExtended has: string (Name), []byte (Data), *string, *int64, *float64, *bool, *time.Time
	if atom.Strings == nil {
		t.Error("expected Strings to be allocated for Name")
	}
	if atom.Bytes == nil {
		t.Error("expected Bytes to be allocated")
	}
	if atom.StringPtrs == nil {
		t.Error("expected StringPtrs to be allocated")
	}
	if atom.IntPtrs == nil {
		t.Error("expected IntPtrs to be allocated")
	}
	if atom.FloatPtrs == nil {
		t.Error("expected FloatPtrs to be allocated")
	}
	if atom.BoolPtrs == nil {
		t.Error("expected BoolPtrs to be allocated")
	}
	if atom.TimePtrs == nil {
		t.Error("expected TimePtrs to be allocated")
	}

	// These should NOT be allocated for TestExtended
	if atom.Ints != nil {
		t.Error("expected Ints to be nil")
	}
	if atom.Floats != nil {
		t.Error("expected Floats to be nil")
	}
}

// TestCustomAtomizable is a struct that implements Atomizable/Deatomizable.
// This simulates code-generated types that bypass reflection.
type TestCustomAtomizable struct {
	ID      string
	Secret  string // Will be handled specially
	Counter int64
}

func (t *TestCustomAtomizable) Atomize(a *Atom) {
	a.Strings["ID"] = t.ID
	a.Strings["Secret"] = "redacted:" + t.Secret // Custom handling
	a.Ints["Counter"] = t.Counter
}

func (t *TestCustomAtomizable) Deatomize(a *Atom) error {
	t.ID = a.Strings["ID"]
	// Strip "redacted:" prefix if present
	secret := a.Strings["Secret"]
	if len(secret) > 9 && secret[:9] == "redacted:" {
		t.Secret = secret[9:]
	} else {
		t.Secret = secret
	}
	t.Counter = a.Ints["Counter"]
	return nil
}

func TestAtomizableInterface(t *testing.T) {
	atomizer := mustUse[TestCustomAtomizable](t)

	original := &TestCustomAtomizable{
		ID:      "custom-1",
		Secret:  "password123",
		Counter: 42,
	}

	atom := atomizer.Atomize(original)

	// Verify custom atomization was used
	if atom.Strings["Secret"] != "redacted:password123" {
		t.Errorf("expected Secret to be redacted, got %s", atom.Strings["Secret"])
	}
	if atom.Strings["ID"] != "custom-1" {
		t.Errorf("expected ID custom-1, got %s", atom.Strings["ID"])
	}
	if atom.Ints["Counter"] != 42 {
		t.Errorf("expected Counter 42, got %d", atom.Ints["Counter"])
	}
}

func TestDeatomizableInterface(t *testing.T) {
	atomizer := mustUse[TestCustomAtomizable](t)

	atom := atomizer.NewAtom()
	atom.Strings["ID"] = "custom-2"
	atom.Strings["Secret"] = "redacted:mysecret"
	atom.Ints["Counter"] = 100

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify custom deatomization was used
	if restored.ID != "custom-2" {
		t.Errorf("expected ID custom-2, got %s", restored.ID)
	}
	if restored.Secret != "mysecret" {
		t.Errorf("expected Secret mysecret (stripped prefix), got %s", restored.Secret)
	}
	if restored.Counter != 100 {
		t.Errorf("expected Counter 100, got %d", restored.Counter)
	}
}

func TestAtomizableRoundTrip(t *testing.T) {
	atomizer := mustUse[TestCustomAtomizable](t)

	original := &TestCustomAtomizable{
		ID:      "roundtrip-1",
		Secret:  "topsecret",
		Counter: 999,
	}

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("expected ID %s, got %s", original.ID, restored.ID)
	}
	if restored.Secret != original.Secret {
		t.Errorf("expected Secret %s, got %s", original.Secret, restored.Secret)
	}
	if restored.Counter != original.Counter {
		t.Errorf("expected Counter %d, got %d", original.Counter, restored.Counter)
	}
}

// TestPartialAtomizable only implements Atomizable, not Deatomizable.
type TestPartialAtomizable struct {
	Name  string
	Value int64
}

func (t *TestPartialAtomizable) Atomize(a *Atom) {
	a.Strings["Name"] = "custom:" + t.Name
	a.Ints["Value"] = t.Value * 2 // Custom: double the value
}

func TestPartialAtomizableInterface(t *testing.T) {
	atomizer := mustUse[TestPartialAtomizable](t)

	original := &TestPartialAtomizable{
		Name:  "test",
		Value: 50,
	}

	atom := atomizer.Atomize(original)

	// Custom atomization should be used
	if atom.Strings["Name"] != "custom:test" {
		t.Errorf("expected Name custom:test, got %s", atom.Strings["Name"])
	}
	if atom.Ints["Value"] != 100 {
		t.Errorf("expected Value 100 (doubled), got %d", atom.Ints["Value"])
	}

	// Deatomize should use reflection (no Deatomizable)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Reflection uses the atom values directly
	if restored.Name != "custom:test" {
		t.Errorf("expected Name custom:test, got %s", restored.Name)
	}
	if restored.Value != 100 {
		t.Errorf("expected Value 100, got %d", restored.Value)
	}
}

// TestConcurrentType is used for concurrent registration tests.
type TestConcurrentType struct {
	Name  string
	Value int64
}

func TestConcurrentRegistration(t *testing.T) {
	// Reset would be needed here if we had a Reset function
	// For now, we just test that concurrent Use calls don't panic

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// All goroutines try to register the same type concurrently
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			atomizer := mustUse[TestConcurrentType](t)
			if atomizer == nil {
				t.Error("expected non-nil atomizer")
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentAtomizeDeatomize(t *testing.T) {
	atomizer := mustUse[TestConcurrentType](t)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent atomize operations
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			obj := &TestConcurrentType{
				Name:  "concurrent",
				Value: int64(id),
			}
			atom := atomizer.Atomize(obj)
			if atom.Strings["Name"] != "concurrent" {
				t.Errorf("expected Name concurrent, got %s", atom.Strings["Name"])
			}
		}(i)
	}

	// Concurrent deatomize operations
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			atom := atomizer.NewAtom()
			atom.Strings["Name"] = "deatomized"
			atom.Ints["Value"] = int64(id)

			restored, err := atomizer.Deatomize(atom)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if restored.Name != "deatomized" {
				t.Errorf("expected Name deatomized, got %s", restored.Name)
			}
		}(i)
	}

	wg.Wait()
}

// Types for testing concurrent registration of different types.
type TestConcurrentA struct{ A string }
type TestConcurrentB struct{ B int64 }
type TestConcurrentC struct{ C float64 }
type TestConcurrentD struct{ D bool }

func TestConcurrentDifferentTypes(t *testing.T) {
	const goroutines = 25
	var wg sync.WaitGroup
	wg.Add(goroutines * 4)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if _, err := Use[TestConcurrentA](); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := Use[TestConcurrentB](); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := Use[TestConcurrentC](); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := Use[TestConcurrentD](); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}

	wg.Wait()

	// Verify all types are properly registered
	a := mustUse[TestConcurrentA](t)
	b := mustUse[TestConcurrentB](t)
	c := mustUse[TestConcurrentC](t)
	d := mustUse[TestConcurrentD](t)

	if a == nil || b == nil || c == nil || d == nil {
		t.Error("expected all atomizers to be non-nil")
	}
}

// TestUnsupportedMap tests that map fields produce an error.
type TestUnsupportedMap struct {
	Labels map[string]string
	Name   string
}

func TestUseUnsupportedMapField(t *testing.T) {
	_, err := Use[TestUnsupportedMap]()
	if err == nil {
		t.Fatal("expected error for unsupported map field")
	}
	if !strings.Contains(err.Error(), "map types are not supported") {
		t.Errorf("expected error about map types, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Labels") {
		t.Errorf("expected error to mention field name 'Labels', got: %v", err)
	}
}

// TestUnsupportedChannel tests that channel fields produce an error.
type TestUnsupportedChannel struct {
	Events chan int
	Name   string
}

func TestUseUnsupportedChannelField(t *testing.T) {
	_, err := Use[TestUnsupportedChannel]()
	if err == nil {
		t.Fatal("expected error for unsupported channel field")
	}
	if !strings.Contains(err.Error(), "unsupported type") {
		t.Errorf("expected error about unsupported type, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Events") {
		t.Errorf("expected error to mention field name 'Events', got: %v", err)
	}
}

// TestUnsupportedFunc tests that function fields produce an error.
type TestUnsupportedFunc struct {
	Callback func()
	Name     string
}

func TestUseUnsupportedFuncField(t *testing.T) {
	_, err := Use[TestUnsupportedFunc]()
	if err == nil {
		t.Fatal("expected error for unsupported func field")
	}
	if !strings.Contains(err.Error(), "unsupported type") {
		t.Errorf("expected error about unsupported type, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Callback") {
		t.Errorf("expected error to mention field name 'Callback', got: %v", err)
	}
}

// --- Overflow detection tests ---

// TestOverflowInt8 tests overflow detection for int8 fields.
type TestOverflowInt8 struct {
	Value int8
}

func TestDeatomizeInt8Overflow(t *testing.T) {
	atomizer := mustUse[TestOverflowInt8](t)

	atom := atomizer.NewAtom()
	atom.Ints["Value"] = 128 // max int8 is 127

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected overflow error")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("expected overflow error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Value") {
		t.Errorf("expected error to mention field name, got: %v", err)
	}
}

func TestDeatomizeInt8Underflow(t *testing.T) {
	atomizer := mustUse[TestOverflowInt8](t)

	atom := atomizer.NewAtom()
	atom.Ints["Value"] = -129 // min int8 is -128

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected underflow error")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("expected overflow error, got: %v", err)
	}
}

// TestOverflowUint8 tests overflow detection for uint8 fields.
type TestOverflowUint8 struct {
	Value uint8
}

func TestDeatomizeUint8Overflow(t *testing.T) {
	atomizer := mustUse[TestOverflowUint8](t)

	atom := atomizer.NewAtom()
	atom.Uints["Value"] = 256 // max uint8 is 255

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected overflow error")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("expected overflow error, got: %v", err)
	}
}

// TestOverflowFloat32 tests overflow detection for float32 fields.
type TestOverflowFloat32 struct {
	Value float32
}

func TestDeatomizeFloat32Overflow(t *testing.T) {
	atomizer := mustUse[TestOverflowFloat32](t)

	atom := atomizer.NewAtom()
	atom.Floats["Value"] = 3.5e38 // max float32 is ~3.4e38

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected overflow error")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("expected overflow error, got: %v", err)
	}
}

// TestOverflowIntSlice tests overflow in slice elements.
type TestOverflowIntSlice struct {
	Values []int8
}

func TestDeatomizeIntSliceOverflow(t *testing.T) {
	atomizer := mustUse[TestOverflowIntSlice](t)

	atom := atomizer.NewAtom()
	atom.IntSlices["Values"] = []int64{10, 20, 300, 40} // 300 overflows int8

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected overflow error")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("expected overflow error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "[2]") {
		t.Errorf("expected error to mention index [2], got: %v", err)
	}
}

// TestOverflowIntPtr tests overflow in pointer fields.
type TestOverflowIntPtr struct {
	Value *int8
}

func TestDeatomizeIntPtrOverflow(t *testing.T) {
	atomizer := mustUse[TestOverflowIntPtr](t)

	overflow := int64(200)
	atom := atomizer.NewAtom()
	atom.IntPtrs["Value"] = &overflow

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected overflow error")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("expected overflow error, got: %v", err)
	}
}

// TestMapPreSizing verifies that maps are pre-sized based on field count.
type TestPreSizedMaps struct {
	A string
	B string
	C string
	X int64
	Y int64
}

func TestMapPreSizing(t *testing.T) {
	atomizer := mustUse[TestPreSizedMaps](t)

	// Verify field counts per table
	stringFields := atomizer.FieldsIn(TableStrings)
	if len(stringFields) != 3 {
		t.Errorf("expected 3 string fields, got %d", len(stringFields))
	}

	intFields := atomizer.FieldsIn(TableInts)
	if len(intFields) != 2 {
		t.Errorf("expected 2 int fields, got %d", len(intFields))
	}

	// Verify NewAtom creates maps (we can't check capacity, but we can verify correctness)
	atom := atomizer.NewAtom()
	if atom.Strings == nil {
		t.Error("expected Strings map to be allocated")
	}
	if atom.Ints == nil {
		t.Error("expected Ints map to be allocated")
	}
	if atom.Floats != nil {
		t.Error("expected Floats map to be nil (no float fields)")
	}
}

// --- Named primitive type tests (enums) ---

// Status is a named int type (enum pattern).
type Status int

const (
	StatusPending Status = iota
	StatusActive
	StatusClosed
)

// Priority is a named int8 type.
type Priority int8

// UserID is a named string type.
type UserID string

// Score is a named float64 type.
type Score float64

// Enabled is a named bool type.
type Enabled bool

// TestNamedPrimitives tests struct with named primitive types.
type TestNamedPrimitives struct {
	UserID   UserID
	Status   Status
	Score    Score
	Priority Priority
	Enabled  Enabled
}

func TestNamedPrimitiveTypes(t *testing.T) {
	atomizer := mustUse[TestNamedPrimitives](t)

	original := &TestNamedPrimitives{
		Status:   StatusActive,
		Priority: 5,
		UserID:   "user-123",
		Score:    98.5,
		Enabled:  true,
	}

	atom := atomizer.Atomize(original)

	// Verify storage in correct tables
	if atom.Ints["Status"] != int64(StatusActive) {
		t.Errorf("expected Status %d, got %d", StatusActive, atom.Ints["Status"])
	}
	if atom.Ints["Priority"] != 5 {
		t.Errorf("expected Priority 5, got %d", atom.Ints["Priority"])
	}
	if atom.Strings["UserID"] != "user-123" {
		t.Errorf("expected UserID user-123, got %s", atom.Strings["UserID"])
	}
	if atom.Floats["Score"] != 98.5 {
		t.Errorf("expected Score 98.5, got %f", atom.Floats["Score"])
	}
	if atom.Bools["Enabled"] != true {
		t.Errorf("expected Enabled true, got %v", atom.Bools["Enabled"])
	}
}

func TestNamedPrimitiveRoundTrip(t *testing.T) {
	atomizer := mustUse[TestNamedPrimitives](t)

	original := &TestNamedPrimitives{
		Status:   StatusClosed,
		Priority: -3,
		UserID:   "admin-456",
		Score:    77.25,
		Enabled:  false,
	}

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.Status != original.Status {
		t.Errorf("expected Status %d, got %d", original.Status, restored.Status)
	}
	if restored.Priority != original.Priority {
		t.Errorf("expected Priority %d, got %d", original.Priority, restored.Priority)
	}
	if restored.UserID != original.UserID {
		t.Errorf("expected UserID %s, got %s", original.UserID, restored.UserID)
	}
	if restored.Score != original.Score {
		t.Errorf("expected Score %f, got %f", original.Score, restored.Score)
	}
	if restored.Enabled != original.Enabled {
		t.Errorf("expected Enabled %v, got %v", original.Enabled, restored.Enabled)
	}
}

// TestNamedPrimitiveSlices tests slices of named primitive types.
type TestNamedPrimitiveSlices struct {
	Statuses   []Status
	Priorities []Priority
	UserIDs    []UserID
	Scores     []Score
}

func TestNamedPrimitiveSliceRoundTrip(t *testing.T) {
	atomizer := mustUse[TestNamedPrimitiveSlices](t)

	original := &TestNamedPrimitiveSlices{
		Statuses:   []Status{StatusPending, StatusActive, StatusClosed},
		Priorities: []Priority{1, 2, 3},
		UserIDs:    []UserID{"a", "b", "c"},
		Scores:     []Score{1.1, 2.2, 3.3},
	}

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(restored.Statuses) != len(original.Statuses) {
		t.Fatalf("expected %d statuses, got %d", len(original.Statuses), len(restored.Statuses))
	}
	for i, v := range original.Statuses {
		if restored.Statuses[i] != v {
			t.Errorf("Statuses[%d]: expected %d, got %d", i, v, restored.Statuses[i])
		}
	}

	if len(restored.Priorities) != len(original.Priorities) {
		t.Fatalf("expected %d priorities, got %d", len(original.Priorities), len(restored.Priorities))
	}
	for i, v := range original.Priorities {
		if restored.Priorities[i] != v {
			t.Errorf("Priorities[%d]: expected %d, got %d", i, v, restored.Priorities[i])
		}
	}

	if len(restored.UserIDs) != len(original.UserIDs) {
		t.Fatalf("expected %d userIDs, got %d", len(original.UserIDs), len(restored.UserIDs))
	}
	for i, v := range original.UserIDs {
		if restored.UserIDs[i] != v {
			t.Errorf("UserIDs[%d]: expected %s, got %s", i, v, restored.UserIDs[i])
		}
	}

	if len(restored.Scores) != len(original.Scores) {
		t.Fatalf("expected %d scores, got %d", len(original.Scores), len(restored.Scores))
	}
	for i, v := range original.Scores {
		if restored.Scores[i] != v {
			t.Errorf("Scores[%d]: expected %f, got %f", i, v, restored.Scores[i])
		}
	}
}

// TestNamedPrimitivePointers tests pointers to named primitive types.
type TestNamedPrimitivePointers struct {
	Status   *Status
	Priority *Priority
	UserID   *UserID
}

func TestNamedPrimitivePointerRoundTrip(t *testing.T) {
	atomizer := mustUse[TestNamedPrimitivePointers](t)

	status := StatusActive
	priority := Priority(7)
	userID := UserID("ptr-user")

	original := &TestNamedPrimitivePointers{
		Status:   &status,
		Priority: &priority,
		UserID:   &userID,
	}

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.Status == nil || *restored.Status != *original.Status {
		t.Errorf("expected Status %d, got %v", *original.Status, restored.Status)
	}
	if restored.Priority == nil || *restored.Priority != *original.Priority {
		t.Errorf("expected Priority %d, got %v", *original.Priority, restored.Priority)
	}
	if restored.UserID == nil || *restored.UserID != *original.UserID {
		t.Errorf("expected UserID %s, got %v", *original.UserID, restored.UserID)
	}
}

func TestNamedPrimitivePointerNil(t *testing.T) {
	atomizer := mustUse[TestNamedPrimitivePointers](t)

	original := &TestNamedPrimitivePointers{
		Status:   nil,
		Priority: nil,
		UserID:   nil,
	}

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.Status != nil {
		t.Errorf("expected Status nil, got %v", restored.Status)
	}
	if restored.Priority != nil {
		t.Errorf("expected Priority nil, got %v", restored.Priority)
	}
	if restored.UserID != nil {
		t.Errorf("expected UserID nil, got %v", restored.UserID)
	}
}

// --- Named byte slice type tests (net.IP, json.RawMessage style) ---

// RawData is a named []byte type (like json.RawMessage).
type RawData []byte

// TestNamedByteSlice tests struct with named []byte type.
type TestNamedByteSlice struct {
	Name string
	Data RawData
}

func TestNamedByteSliceRoundTrip(t *testing.T) {
	atomizer := mustUse[TestNamedByteSlice](t)

	original := &TestNamedByteSlice{
		Name: "test",
		Data: RawData{0x01, 0x02, 0x03, 0x04},
	}

	atom := atomizer.Atomize(original)

	// Verify stored in Bytes table
	if !bytes.Equal(atom.Bytes["Data"], []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Errorf("expected Data in Bytes table, got %v", atom.Bytes["Data"])
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if !bytes.Equal(restored.Data, original.Data) {
		t.Errorf("expected Data %v, got %v", original.Data, restored.Data)
	}
}

// TestNamedByteSlicePtr tests pointer to named []byte type.
type TestNamedByteSlicePtr struct {
	Data *RawData
}

func TestNamedByteSlicePtrRoundTrip(t *testing.T) {
	atomizer := mustUse[TestNamedByteSlicePtr](t)

	data := RawData{0xAA, 0xBB}
	original := &TestNamedByteSlicePtr{
		Data: &data,
	}

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.Data == nil {
		t.Fatal("expected Data to be non-nil")
	}
	if !bytes.Equal(*restored.Data, *original.Data) {
		t.Errorf("expected Data %v, got %v", *original.Data, *restored.Data)
	}
}

// TestNamedByteSliceSlice tests slice of named []byte types.
type TestNamedByteSliceSlice struct {
	Chunks []RawData
}

func TestNamedByteSliceSliceRoundTrip(t *testing.T) {
	atomizer := mustUse[TestNamedByteSliceSlice](t)

	original := &TestNamedByteSliceSlice{
		Chunks: []RawData{
			{0x01, 0x02},
			{0x03, 0x04, 0x05},
			{0x06},
		},
	}

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(restored.Chunks) != len(original.Chunks) {
		t.Fatalf("expected %d chunks, got %d", len(original.Chunks), len(restored.Chunks))
	}
	for i, v := range original.Chunks {
		if !bytes.Equal(restored.Chunks[i], v) {
			t.Errorf("Chunks[%d]: expected %v, got %v", i, v, restored.Chunks[i])
		}
	}
}

// --- Fixed-size byte array tests ---

// UUID is a fixed-size byte array type.
type UUID [16]byte

// Hash256 is a SHA-256 hash type.
type Hash256 [32]byte

// TestFixedByteArray tests struct with fixed-size byte arrays.
type TestFixedByteArray struct {
	ID   UUID
	Hash Hash256
	Code [4]byte // unnamed fixed array
}

func TestFixedByteArrayRoundTrip(t *testing.T) {
	atomizer := mustUse[TestFixedByteArray](t)

	original := &TestFixedByteArray{
		ID:   UUID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		Hash: Hash256{0xAA, 0xBB, 0xCC}, // rest are zeros
		Code: [4]byte{0xDE, 0xAD, 0xBE, 0xEF},
	}

	atom := atomizer.Atomize(original)

	// Verify stored in Bytes table
	if len(atom.Bytes["ID"]) != 16 {
		t.Errorf("expected ID length 16, got %d", len(atom.Bytes["ID"]))
	}
	if len(atom.Bytes["Hash"]) != 32 {
		t.Errorf("expected Hash length 32, got %d", len(atom.Bytes["Hash"]))
	}
	if len(atom.Bytes["Code"]) != 4 {
		t.Errorf("expected Code length 4, got %d", len(atom.Bytes["Code"]))
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("expected ID %v, got %v", original.ID, restored.ID)
	}
	if restored.Hash != original.Hash {
		t.Errorf("expected Hash %v, got %v", original.Hash, restored.Hash)
	}
	if restored.Code != original.Code {
		t.Errorf("expected Code %v, got %v", original.Code, restored.Code)
	}
}

func TestFixedByteArrayZeroValue(t *testing.T) {
	atomizer := mustUse[TestFixedByteArray](t)

	original := &TestFixedByteArray{} // all zeros

	atom := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("expected zero ID, got %v", restored.ID)
	}
	if restored.Hash != original.Hash {
		t.Errorf("expected zero Hash, got %v", restored.Hash)
	}
	if restored.Code != original.Code {
		t.Errorf("expected zero Code, got %v", restored.Code)
	}
}

func TestFixedByteArraySizeMismatch(t *testing.T) {
	atomizer := mustUse[TestFixedByteArray](t)

	atom := atomizer.NewAtom()
	atom.Bytes["ID"] = []byte{0x01, 0x02, 0x03} // Wrong size: 3 instead of 16
	atom.Bytes["Hash"] = make([]byte, 32)
	atom.Bytes["Code"] = make([]byte, 4)

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected error for size mismatch")
	}
	if !strings.Contains(err.Error(), "does not match array size") {
		t.Errorf("expected size mismatch error, got: %v", err)
	}
}

// TestNamedPrimitiveOverflow tests overflow detection for named types.
func TestNamedPrimitiveOverflow(t *testing.T) {
	atomizer := mustUse[TestNamedPrimitives](t)

	atom := atomizer.NewAtom()
	atom.Ints["Status"] = 1
	atom.Ints["Priority"] = 200 // Overflow: Priority is int8, max 127
	atom.Strings["UserID"] = "test"
	atom.Floats["Score"] = 1.0
	atom.Bools["Enabled"] = true

	_, err := atomizer.Deatomize(atom)
	if err == nil {
		t.Fatal("expected overflow error for Priority")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Errorf("expected overflow error, got: %v", err)
	}
}

// TestValidRangeValues ensures valid values don't produce errors.
func TestDeatomizeValidRanges(t *testing.T) {
	t.Run("int8 valid", func(t *testing.T) {
		atomizer := mustUse[TestOverflowInt8](t)
		atom := atomizer.NewAtom()
		atom.Ints["Value"] = 127 // max valid int8
		result, err := atomizer.Deatomize(atom)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value != 127 {
			t.Errorf("expected 127, got %d", result.Value)
		}
	})

	t.Run("uint8 valid", func(t *testing.T) {
		atomizer := mustUse[TestOverflowUint8](t)
		atom := atomizer.NewAtom()
		atom.Uints["Value"] = 255 // max valid uint8
		result, err := atomizer.Deatomize(atom)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value != 255 {
			t.Errorf("expected 255, got %d", result.Value)
		}
	})

	t.Run("float32 valid", func(t *testing.T) {
		atomizer := mustUse[TestOverflowFloat32](t)
		atom := atomizer.NewAtom()
		atom.Floats["Value"] = 3.14159
		result, err := atomizer.Deatomize(atom)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Value < 3.14 || result.Value > 3.15 {
			t.Errorf("expected ~3.14159, got %f", result.Value)
		}
	})
}

// --- Slice of pointer to nested structs tests ---

func TestNestedPtrSliceRoundTrip(t *testing.T) {
	atomizer := mustUse[TestWithNestedPtrSlice](t)

	addr1 := &TestAddress{Street: "123 Main St", City: "Springfield"}
	addr2 := &TestAddress{Street: "456 Oak Ave", City: "Shelbyville"}

	original := &TestWithNestedPtrSlice{
		Name:      "Alice",
		Addresses: []*TestAddress{addr1, addr2},
	}

	atom := atomizer.Atomize(original)

	// Verify nested slices storage
	if len(atom.NestedSlices["Addresses"]) != 2 {
		t.Errorf("expected 2 addresses in NestedSlices, got %d", len(atom.NestedSlices["Addresses"]))
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	if len(restored.Addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(restored.Addresses))
	}
	if restored.Addresses[0].Street != addr1.Street {
		t.Errorf("expected Street %s, got %s", addr1.Street, restored.Addresses[0].Street)
	}
	if restored.Addresses[1].City != addr2.City {
		t.Errorf("expected City %s, got %s", addr2.City, restored.Addresses[1].City)
	}
}

func TestNestedPtrSliceWithNilElements(t *testing.T) {
	atomizer := mustUse[TestWithNestedPtrSlice](t)

	addr1 := &TestAddress{Street: "123 Main St", City: "Springfield"}

	original := &TestWithNestedPtrSlice{
		Name:      "Bob",
		Addresses: []*TestAddress{addr1, nil, addr1}, // nil element in middle
	}

	atom := atomizer.Atomize(original)

	// Nil elements should be skipped
	if len(atom.NestedSlices["Addresses"]) != 2 {
		t.Errorf("expected 2 addresses (nil skipped), got %d", len(atom.NestedSlices["Addresses"]))
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	// Restored slice will have 2 elements (nil was skipped)
	if len(restored.Addresses) != 2 {
		t.Errorf("expected 2 addresses after round-trip, got %d", len(restored.Addresses))
	}
}

func TestNestedPtrSliceEmpty(t *testing.T) {
	atomizer := mustUse[TestWithNestedPtrSlice](t)

	original := &TestWithNestedPtrSlice{
		Name:      "Charlie",
		Addresses: []*TestAddress{},
	}

	atom := atomizer.Atomize(original)

	// Empty slice should not create entry
	if _, ok := atom.NestedSlices["Addresses"]; ok {
		t.Error("expected empty slice to not create NestedSlices entry")
	}

	restored, err := atomizer.Deatomize(atom)
	if err != nil {
		t.Fatalf("unexpected error on deatomize: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("expected Name %s, got %s", original.Name, restored.Name)
	}
	// Empty/nil slice after deatomize is acceptable
	if len(restored.Addresses) != 0 {
		t.Errorf("expected nil or empty addresses, got %v", restored.Addresses)
	}
}

func TestNestedPtrSliceAllNil(t *testing.T) {
	atomizer := mustUse[TestWithNestedPtrSlice](t)

	original := &TestWithNestedPtrSlice{
		Name:      "Dave",
		Addresses: []*TestAddress{nil, nil, nil}, // All nil
	}

	atom := atomizer.Atomize(original)

	// All-nil slice should result in no entry (all elements skipped)
	if _, ok := atom.NestedSlices["Addresses"]; ok {
		t.Error("expected all-nil slice to not create NestedSlices entry")
	}
}

func TestNestedSliceEmptySlice(t *testing.T) {
	atomizer := mustUse[TestWithNestedSlice](t)

	original := &TestWithNestedSlice{
		Name:      "Empty",
		Addresses: []TestAddress{},
	}

	atom := atomizer.Atomize(original)

	// Empty slice should not create entry
	if _, ok := atom.NestedSlices["Addresses"]; ok {
		t.Error("expected empty slice to not create NestedSlices entry")
	}
}
