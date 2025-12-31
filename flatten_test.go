package atom

import (
	"testing"
	"time"

	"github.com/zoobzio/sentinel"
)

func TestFlatten_Scalars(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string"},
			{Name: "Age", Type: "int64"},
			{Name: "Score", Type: "float64"},
			{Name: "Active", Type: "bool"},
		},
	}

	atom := &Atom{
		Spec:    spec,
		Strings: map[string]string{"Name": "John"},
		Ints:    map[string]int64{"Age": 30},
		Floats:  map[string]float64{"Score": 95.5},
		Bools:   map[string]bool{"Active": true},
	}

	result := atom.Flatten()

	if result["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", result["Name"])
	}
	if result["Age"] != int64(30) {
		t.Errorf("expected Age=30, got %v", result["Age"])
	}
	if result["Score"] != 95.5 {
		t.Errorf("expected Score=95.5, got %v", result["Score"])
	}
	if result["Active"] != true {
		t.Errorf("expected Active=true, got %v", result["Active"])
	}
}

func TestFlatten_Pointers(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "*string"},
			{Name: "Age", Type: "*int64"},
			{Name: "Nickname", Type: "*string"},
		},
	}

	name := "John"
	age := int64(30)
	atom := &Atom{
		Spec:       spec,
		StringPtrs: map[string]*string{"Name": &name, "Nickname": nil},
		IntPtrs:    map[string]*int64{"Age": &age},
	}

	result := atom.Flatten()

	if result["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", result["Name"])
	}
	if result["Age"] != int64(30) {
		t.Errorf("expected Age=30, got %v", result["Age"])
	}
	if _, exists := result["Nickname"]; exists {
		t.Error("expected nil pointer to be omitted")
	}
}

func TestFlatten_Slices(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Tags", Type: "[]string"},
			{Name: "Scores", Type: "[]int64"},
		},
	}

	atom := &Atom{
		Spec:         spec,
		StringSlices: map[string][]string{"Tags": {"go", "rust"}},
		IntSlices:    map[string][]int64{"Scores": {100, 95, 88}},
	}

	result := atom.Flatten()

	tags, ok := result["Tags"].([]string)
	if !ok {
		t.Fatalf("expected Tags to be []string, got %T", result["Tags"])
	}
	if len(tags) != 2 || tags[0] != "go" || tags[1] != "rust" {
		t.Errorf("expected Tags=[go, rust], got %v", tags)
	}

	scores, ok := result["Scores"].([]int64)
	if !ok {
		t.Fatalf("expected Scores to be []int64, got %T", result["Scores"])
	}
	if len(scores) != 3 || scores[0] != 100 {
		t.Errorf("expected Scores=[100, 95, 88], got %v", scores)
	}
}

func TestFlatten_Time(t *testing.T) {
	spec := Spec{
		TypeName:    "Event",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "CreatedAt", Type: "time.Time"},
		},
	}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	atom := &Atom{
		Spec:  spec,
		Times: map[string]time.Time{"CreatedAt": now},
	}

	result := atom.Flatten()

	if result["CreatedAt"] != now {
		t.Errorf("expected CreatedAt=%v, got %v", now, result["CreatedAt"])
	}
}

func TestFlatten_Nested(t *testing.T) {
	addressSpec := Spec{
		TypeName:    "Address",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "City", Type: "string"},
			{Name: "Zip", Type: "string"},
		},
	}

	userSpec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string"},
			{Name: "Address", Type: "Address", Kind: sentinel.KindStruct},
		},
	}

	atom := &Atom{
		Spec:    userSpec,
		Strings: map[string]string{"Name": "John"},
		Nested: map[string]Atom{
			"Address": {
				Spec:    addressSpec,
				Strings: map[string]string{"City": "NYC", "Zip": "10001"},
			},
		},
	}

	result := atom.Flatten()

	if result["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", result["Name"])
	}

	addr, ok := result["Address"].(map[string]any)
	if !ok {
		t.Fatalf("expected Address to be map[string]any, got %T", result["Address"])
	}
	if addr["City"] != "NYC" {
		t.Errorf("expected City=NYC, got %v", addr["City"])
	}
	if addr["Zip"] != "10001" {
		t.Errorf("expected Zip=10001, got %v", addr["Zip"])
	}
}

func TestUnflatten_Scalars(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string"},
			{Name: "Age", Type: "int64"},
			{Name: "Score", Type: "float64"},
			{Name: "Active", Type: "bool"},
		},
	}

	data := map[string]any{
		"Name":   "John",
		"Age":    float64(30), // JSON numbers are float64
		"Score":  95.5,
		"Active": true,
	}

	atom := Unflatten(data, spec)

	if atom.Strings["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", atom.Strings["Name"])
	}
	if atom.Ints["Age"] != 30 {
		t.Errorf("expected Age=30, got %v", atom.Ints["Age"])
	}
	if atom.Floats["Score"] != 95.5 {
		t.Errorf("expected Score=95.5, got %v", atom.Floats["Score"])
	}
	if atom.Bools["Active"] != true {
		t.Errorf("expected Active=true, got %v", atom.Bools["Active"])
	}
}

func TestUnflatten_Pointers(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "*string"},
			{Name: "Age", Type: "*int64"},
		},
	}

	data := map[string]any{
		"Name": "John",
		"Age":  float64(30),
	}

	atom := Unflatten(data, spec)

	if atom.StringPtrs["Name"] == nil || *atom.StringPtrs["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", atom.StringPtrs["Name"])
	}
	if atom.IntPtrs["Age"] == nil || *atom.IntPtrs["Age"] != 30 {
		t.Errorf("expected Age=30, got %v", atom.IntPtrs["Age"])
	}
}

func TestUnflatten_Slices(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Tags", Type: "[]string"},
			{Name: "Scores", Type: "[]int64"},
		},
	}

	data := map[string]any{
		"Tags":   []any{"go", "rust"},
		"Scores": []any{float64(100), float64(95)},
	}

	atom := Unflatten(data, spec)

	if len(atom.StringSlices["Tags"]) != 2 || atom.StringSlices["Tags"][0] != "go" {
		t.Errorf("expected Tags=[go, rust], got %v", atom.StringSlices["Tags"])
	}
	if len(atom.IntSlices["Scores"]) != 2 || atom.IntSlices["Scores"][0] != 100 {
		t.Errorf("expected Scores=[100, 95], got %v", atom.IntSlices["Scores"])
	}
}

func TestUnflatten_Time(t *testing.T) {
	spec := Spec{
		TypeName:    "Event",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "CreatedAt", Type: "time.Time"},
		},
	}

	data := map[string]any{
		"CreatedAt": "2024-01-15T10:30:00Z",
	}

	atom := Unflatten(data, spec)

	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !atom.Times["CreatedAt"].Equal(expected) {
		t.Errorf("expected CreatedAt=%v, got %v", expected, atom.Times["CreatedAt"])
	}
}

func TestUnflatten_TimeObject(t *testing.T) {
	spec := Spec{
		TypeName:    "Event",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "CreatedAt", Type: "time.Time"},
		},
	}

	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	data := map[string]any{
		"CreatedAt": expected,
	}

	atom := Unflatten(data, spec)

	if !atom.Times["CreatedAt"].Equal(expected) {
		t.Errorf("expected CreatedAt=%v, got %v", expected, atom.Times["CreatedAt"])
	}
}

func TestRoundtrip(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string"},
			{Name: "Age", Type: "int64"},
			{Name: "Score", Type: "float64"},
			{Name: "Active", Type: "bool"},
			{Name: "Tags", Type: "[]string"},
		},
	}

	original := &Atom{
		Spec:         spec,
		Strings:      map[string]string{"Name": "John"},
		Ints:         map[string]int64{"Age": 30},
		Floats:       map[string]float64{"Score": 95.5},
		Bools:        map[string]bool{"Active": true},
		StringSlices: map[string][]string{"Tags": {"go", "rust"}},
	}

	// Flatten then unflatten
	flat := original.Flatten()
	restored := Unflatten(flat, spec)

	if restored.Strings["Name"] != original.Strings["Name"] {
		t.Errorf("Name mismatch: %v != %v", restored.Strings["Name"], original.Strings["Name"])
	}
	if restored.Ints["Age"] != original.Ints["Age"] {
		t.Errorf("Age mismatch: %v != %v", restored.Ints["Age"], original.Ints["Age"])
	}
	if restored.Floats["Score"] != original.Floats["Score"] {
		t.Errorf("Score mismatch: %v != %v", restored.Floats["Score"], original.Floats["Score"])
	}
	if restored.Bools["Active"] != original.Bools["Active"] {
		t.Errorf("Active mismatch: %v != %v", restored.Bools["Active"], original.Bools["Active"])
	}
	if len(restored.StringSlices["Tags"]) != len(original.StringSlices["Tags"]) {
		t.Errorf("Tags length mismatch: %v != %v", len(restored.StringSlices["Tags"]), len(original.StringSlices["Tags"]))
	}
}

func TestFlatten_Uints(t *testing.T) {
	spec := Spec{
		TypeName:    "Counter",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Count", Type: "uint64"},
		},
	}

	atom := &Atom{
		Spec:  spec,
		Uints: map[string]uint64{"Count": 42},
	}

	result := atom.Flatten()

	if result["Count"] != uint64(42) {
		t.Errorf("expected Count=42, got %v", result["Count"])
	}
}

func TestUnflatten_Uints(t *testing.T) {
	spec := Spec{
		TypeName:    "Counter",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Count", Type: "uint64"},
		},
	}

	data := map[string]any{
		"Count": float64(42), // JSON numbers
	}

	atom := Unflatten(data, spec)

	if atom.Uints["Count"] != 42 {
		t.Errorf("expected Count=42, got %v", atom.Uints["Count"])
	}
}

func TestFlatten_Bytes(t *testing.T) {
	spec := Spec{
		TypeName:    "Document",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Data", Type: "[]uint8"},
		},
	}

	atom := &Atom{
		Spec:  spec,
		Bytes: map[string][]byte{"Data": []byte("hello")},
	}

	result := atom.Flatten()

	data, ok := result["Data"].([]byte)
	if !ok || string(data) != "hello" {
		t.Errorf("expected Data=hello, got %v", result["Data"])
	}
}

func TestUnflatten_NilValue(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string"},
			{Name: "Age", Type: "int64"},
		},
	}

	data := map[string]any{
		"Name": "John",
		"Age":  nil,
	}

	atom := Unflatten(data, spec)

	if atom.Strings["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", atom.Strings["Name"])
	}
	// Age should not be set since value is nil
	if _, exists := atom.Ints["Age"]; exists {
		t.Error("expected Age to not be set for nil value")
	}
}

func TestUnflatten_UnknownField(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string"},
		},
	}

	data := map[string]any{
		"Name":    "John",
		"unknown": "value",
	}

	atom := Unflatten(data, spec)

	if atom.Strings["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", atom.Strings["Name"])
	}
	// Should not panic or error on unknown field
}

// Test types for nested unflatten.
type nestedAddress struct {
	Street string
	City   string
}

type nestedPerson struct {
	Name    string
	Address nestedAddress
}

func TestUnflatten_Nested(t *testing.T) {
	// Register the types so sentinel.Lookup can find them
	_, err := Use[nestedAddress]()
	if err != nil {
		t.Fatalf("failed to register nestedAddress: %v", err)
	}
	personAtomizer, err := Use[nestedPerson]()
	if err != nil {
		t.Fatalf("failed to register nestedPerson: %v", err)
	}

	// Create an atom with nested data
	original := &Atom{
		Spec:    personAtomizer.Spec(),
		Strings: map[string]string{"Name": "John"},
		Nested: map[string]Atom{
			"Address": {
				Strings: map[string]string{"Street": "123 Main St", "City": "NYC"},
			},
		},
	}

	// Flatten then unflatten
	flat := original.Flatten()
	restored := Unflatten(flat, personAtomizer.Spec())

	// Verify root level
	if restored.Strings["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", restored.Strings["Name"])
	}

	// Verify nested data was preserved
	nestedAtom, ok := restored.Nested["Address"]
	if !ok {
		t.Fatal("expected Address nested atom to exist")
	}
	if nestedAtom.Strings["Street"] != "123 Main St" {
		t.Errorf("expected Street=123 Main St, got %v", nestedAtom.Strings["Street"])
	}
	if nestedAtom.Strings["City"] != "NYC" {
		t.Errorf("expected City=NYC, got %v", nestedAtom.Strings["City"])
	}
}

func TestUnflatten_NestedRoundtrip(t *testing.T) {
	// Register and use atomizer
	personAtomizer, err := Use[nestedPerson]()
	if err != nil {
		t.Fatalf("failed to register nestedPerson: %v", err)
	}

	// Create original struct
	person := &nestedPerson{
		Name: "Alice",
		Address: nestedAddress{
			Street: "456 Oak Ave",
			City:   "Boston",
		},
	}

	// Atomize -> Flatten -> Unflatten -> Deatomize
	atom := personAtomizer.Atomize(person)
	flat := atom.Flatten()
	restored := Unflatten(flat, personAtomizer.Spec())
	result, err := personAtomizer.Deatomize(restored)
	if err != nil {
		t.Fatalf("deatomize failed: %v", err)
	}

	// Verify full roundtrip
	if result.Name != person.Name {
		t.Errorf("Name mismatch: %v != %v", result.Name, person.Name)
	}
	if result.Address.Street != person.Address.Street {
		t.Errorf("Street mismatch: %v != %v", result.Address.Street, person.Address.Street)
	}
	if result.Address.City != person.Address.City {
		t.Errorf("City mismatch: %v != %v", result.Address.City, person.Address.City)
	}
}
