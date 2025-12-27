package atom

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestImplementsAtomizable(t *testing.T) {
	// Test type that doesn't implement
	if implementsAtomizable(reflect.TypeFor[string]()) {
		t.Error("string should not implement Atomizable")
	}

	// Test pointer to type that doesn't implement
	if implementsAtomizable(reflect.TypeFor[*string]()) {
		t.Error("*string should not implement Atomizable")
	}
}

func TestImplementsDeatomizable(t *testing.T) {
	// Test type that doesn't implement
	if implementsDeatomizable(reflect.TypeFor[int]()) {
		t.Error("int should not implement Deatomizable")
	}

	// Test pointer to type that doesn't implement
	if implementsDeatomizable(reflect.TypeFor[*int]()) {
		t.Error("*int should not implement Deatomizable")
	}
}

func TestAllocateAtomEmpty(t *testing.T) {
	tableSet := make(map[Table]int)
	atom := allocateAtom(Spec{}, tableSet)

	if atom == nil {
		t.Fatal("atom should not be nil")
	}

	// Nested maps should always be allocated
	if atom.Nested == nil {
		t.Error("Nested map should be allocated")
	}
	if atom.NestedSlices == nil {
		t.Error("NestedSlices map should be allocated")
	}

	// Other maps should be nil
	if atom.Strings != nil {
		t.Error("Strings should be nil for empty tableSet")
	}
}

func TestAllocateAtomWithMetadata(t *testing.T) {
	meta := Spec{
		TypeName:    "TestType",
		PackageName: "github.com/example/pkg",
	}
	tableSet := map[Table]int{TableStrings: 1}

	atom := allocateAtom(meta, tableSet)

	if atom.Spec.TypeName != "TestType" {
		t.Errorf("Meta.TypeName: got %q, want %q", atom.Spec.TypeName, "TestType")
	}
	if atom.Spec.PackageName != "github.com/example/pkg" {
		t.Errorf("Meta.PackageName: got %q, want %q", atom.Spec.PackageName, "github.com/example/pkg")
	}
}

func TestAllocateAtomWithTables(t *testing.T) {
	tableSet := map[Table]int{
		TableStrings:      2,
		TableInts:         3,
		TableFloats:       1,
		TableBools:        1,
		TableTimes:        1,
		TableBytes:        1,
		TableStringPtrs:   1,
		TableIntPtrs:      1,
		TableFloatPtrs:    1,
		TableBoolPtrs:     1,
		TableTimePtrs:     1,
		TableStringSlices: 2,
		TableIntSlices:    1,
		TableFloatSlices:  1,
		TableBoolSlices:   1,
		TableTimeSlices:   1,
		TableByteSlices:   1,
	}

	atom := allocateAtom(Spec{}, tableSet)

	// All maps should be allocated
	if atom.Strings == nil {
		t.Error("Strings should be allocated")
	}
	if atom.Ints == nil {
		t.Error("Ints should be allocated")
	}
	if atom.Floats == nil {
		t.Error("Floats should be allocated")
	}
	if atom.Bools == nil {
		t.Error("Bools should be allocated")
	}
	if atom.Times == nil {
		t.Error("Times should be allocated")
	}
	if atom.Bytes == nil {
		t.Error("Bytes should be allocated")
	}
	if atom.StringPtrs == nil {
		t.Error("StringPtrs should be allocated")
	}
	if atom.IntPtrs == nil {
		t.Error("IntPtrs should be allocated")
	}
	if atom.FloatPtrs == nil {
		t.Error("FloatPtrs should be allocated")
	}
	if atom.BoolPtrs == nil {
		t.Error("BoolPtrs should be allocated")
	}
	if atom.TimePtrs == nil {
		t.Error("TimePtrs should be allocated")
	}
	if atom.StringSlices == nil {
		t.Error("StringSlices should be allocated")
	}
	if atom.IntSlices == nil {
		t.Error("IntSlices should be allocated")
	}
	if atom.FloatSlices == nil {
		t.Error("FloatSlices should be allocated")
	}
	if atom.BoolSlices == nil {
		t.Error("BoolSlices should be allocated")
	}
	if atom.TimeSlices == nil {
		t.Error("TimeSlices should be allocated")
	}
	if atom.ByteSlices == nil {
		t.Error("ByteSlices should be allocated")
	}
}

func TestComputeTableSet(t *testing.T) {
	plans := []fieldPlan{
		{name: "A", table: TableStrings, kind: kindScalar},
		{name: "B", table: TableStrings, kind: kindScalar},
		{name: "C", table: TableInts, kind: kindScalar},
	}

	set := computeTableSet(plans)

	if set[TableStrings] != 2 {
		t.Errorf("TableStrings: got %d, want 2", set[TableStrings])
	}
	if set[TableInts] != 1 {
		t.Errorf("TableInts: got %d, want 1", set[TableInts])
	}
}

func TestComputeTableSetNested(t *testing.T) {
	// Create a nested atomizer with its own table set
	innerSet := map[Table]int{TableFloats: 1}
	innerAtomizer := &reflectAtomizer{tableSet: innerSet}

	plans := []fieldPlan{
		{name: "A", table: TableStrings, kind: kindScalar},
		{name: "N", kind: kindNested, nested: innerAtomizer},
	}

	set := computeTableSet(plans)

	if set[TableStrings] != 1 {
		t.Errorf("TableStrings: got %d, want 1", set[TableStrings])
	}
	// Nested tables should be included
	if set[TableFloats] != 1 {
		t.Errorf("TableFloats: got %d, want 1", set[TableFloats])
	}
}

func TestComputeTableSetEmptyTable(t *testing.T) {
	// Nested structs have empty table field
	plans := []fieldPlan{
		{name: "N", table: "", kind: kindNested},
	}

	set := computeTableSet(plans)

	// Empty table should not be counted
	if _, ok := set[""]; ok {
		t.Error("empty table should not be in set")
	}
}

func TestReflectAtomizerAtomize(t *testing.T) {
	type Simple struct {
		Name string
		Age  int64
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Simple]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	ra := &reflectAtomizer{
		typ:      reflect.TypeFor[Simple](),
		plan:     plans,
		tableSet: computeTableSet(plans),
	}

	obj := &Simple{Name: "Alice", Age: 30}
	atom := ra.newAtom()

	ra.atomize(obj, atom)

	if atom.Strings["Name"] != "Alice" {
		t.Errorf("Name: got %q, want %q", atom.Strings["Name"], "Alice")
	}
	if atom.Ints["Age"] != 30 {
		t.Errorf("Age: got %d, want %d", atom.Ints["Age"], 30)
	}
}

func TestReflectAtomizerAtomizeNonPointer(t *testing.T) {
	type Simple struct {
		X int64
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Simple]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}
	ra := &reflectAtomizer{
		typ:      reflect.TypeFor[Simple](),
		plan:     plans,
		tableSet: computeTableSet(plans),
	}

	// Pass non-pointer
	obj := Simple{X: 42}
	atom := ra.newAtom()

	ra.atomize(obj, atom)

	if atom.Ints["X"] != 42 {
		t.Errorf("X: got %d, want %d", atom.Ints["X"], 42)
	}
}

func TestReflectAtomizerDeatomize(t *testing.T) {
	type Simple struct {
		Name string
		Age  int64
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Simple]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}
	ra := &reflectAtomizer{
		typ:      reflect.TypeFor[Simple](),
		plan:     plans,
		tableSet: computeTableSet(plans),
	}

	atom := &Atom{
		Strings: map[string]string{"Name": "Bob"},
		Ints:    map[string]int64{"Age": 25},
	}

	obj := new(Simple)
	if err := ra.deatomize(atom, obj); err != nil {
		t.Fatalf("deatomize error: %v", err)
	}

	if obj.Name != "Bob" {
		t.Errorf("Name: got %q, want %q", obj.Name, "Bob")
	}
	if obj.Age != 25 {
		t.Errorf("Age: got %d, want %d", obj.Age, 25)
	}
}

func TestReflectAtomizerNewAtom(t *testing.T) {
	tableSet := map[Table]int{
		TableStrings: 2,
		TableInts:    1,
	}

	ra := &reflectAtomizer{tableSet: tableSet}
	atom := ra.newAtom()

	if atom.Strings == nil {
		t.Error("Strings should be allocated")
	}
	if atom.Ints == nil {
		t.Error("Ints should be allocated")
	}
	if atom.Floats != nil {
		t.Error("Floats should be nil")
	}
}

func TestReflectAtomizerTables(t *testing.T) {
	tableSet := map[Table]int{
		TableStrings: 3,
		TableBools:   1,
	}

	ra := &reflectAtomizer{tableSet: tableSet}
	got := ra.tables()

	if got[TableStrings] != 3 {
		t.Errorf("TableStrings: got %d, want 3", got[TableStrings])
	}
	if got[TableBools] != 1 {
		t.Errorf("TableBools: got %d, want 1", got[TableBools])
	}
}

func TestBuildReflectAtomizerWithMeta(t *testing.T) {
	type Sample struct {
		Value string
	}

	typ := reflect.TypeFor[Sample]()
	ra, err := buildReflectAtomizerWithSpec(typ, Spec{})
	if err != nil {
		t.Fatalf("buildReflectAtomizerWithSpec error: %v", err)
	}

	if ra.typ != typ {
		t.Errorf("typ: got %v, want %v", ra.typ, typ)
	}
	if len(ra.plan) != 1 {
		t.Errorf("plan length: got %d, want 1", len(ra.plan))
	}
	if ra.tableSet[TableStrings] != 1 {
		t.Errorf("tableSet[TableStrings]: got %d, want 1", ra.tableSet[TableStrings])
	}
}

func TestBuildReflectAtomizerWithMetaPointer(t *testing.T) {
	type Sample struct {
		X int
	}

	ptrType := reflect.TypeFor[*Sample]()
	elemType := ptrType.Elem()

	ra, err := buildReflectAtomizerWithSpec(ptrType, Spec{})
	if err != nil {
		t.Fatalf("buildReflectAtomizerWithSpec error: %v", err)
	}

	// Should dereference to element type
	if ra.typ != elemType {
		t.Errorf("typ: got %v, want %v", ra.typ, elemType)
	}
}

func TestBuildReflectAtomizerWithMetaError(t *testing.T) {
	type BadType struct {
		M map[int]int
	}

	_, err := buildReflectAtomizerWithSpec(reflect.TypeFor[BadType](), Spec{})
	if err == nil {
		t.Error("expected error for unsupported field")
	}
}

func TestReflectAtomizerRoundTrip(t *testing.T) {
	type Complex struct {
		Created time.Time
		Name    string
		Data    []byte
		Count   int64
		Score   float64
		Active  bool
	}

	now := time.Now().Truncate(time.Second)
	original := &Complex{
		Name:    "test",
		Count:   42,
		Score:   3.14,
		Active:  true,
		Created: now,
		Data:    []byte("hello"),
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Complex]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}
	ra := &reflectAtomizer{
		typ:      reflect.TypeFor[Complex](),
		plan:     plans,
		tableSet: computeTableSet(plans),
	}

	atom := ra.newAtom()
	ra.atomize(original, atom)

	result := new(Complex)
	if err := ra.deatomize(atom, result); err != nil {
		t.Fatalf("deatomize error: %v", err)
	}

	if result.Name != original.Name {
		t.Errorf("Name: got %q, want %q", result.Name, original.Name)
	}
	if result.Count != original.Count {
		t.Errorf("Count: got %d, want %d", result.Count, original.Count)
	}
	if result.Score != original.Score {
		t.Errorf("Score: got %g, want %g", result.Score, original.Score)
	}
	if result.Active != original.Active {
		t.Errorf("Active: got %v, want %v", result.Active, original.Active)
	}
	if !result.Created.Equal(original.Created) {
		t.Errorf("Created: got %v, want %v", result.Created, original.Created)
	}
	if !bytes.Equal(result.Data, original.Data) {
		t.Errorf("Data: got %q, want %q", result.Data, original.Data)
	}
}
