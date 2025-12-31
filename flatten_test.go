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

func TestUnflatten_PointerTypes(t *testing.T) {
	spec := Spec{
		TypeName:    "AllPtrs",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "StrPtr", Type: "*string"},
			{Name: "IntPtr", Type: "*int64"},
			{Name: "UintPtr", Type: "*uint64"},
			{Name: "FloatPtr", Type: "*float64"},
			{Name: "BoolPtr", Type: "*bool"},
			{Name: "TimePtr", Type: "*time.Time"},
			{Name: "BytesPtr", Type: "*[]byte"},
		},
	}

	data := map[string]any{
		"StrPtr":   "hello",
		"IntPtr":   float64(42),
		"UintPtr":  float64(100),
		"FloatPtr": 3.14,
		"BoolPtr":  true,
		"TimePtr":  "2024-01-15T10:30:00Z",
		"BytesPtr": "aGVsbG8=", // base64 for "hello"
	}

	atom := Unflatten(data, spec)

	if atom.StringPtrs["StrPtr"] == nil || *atom.StringPtrs["StrPtr"] != "hello" {
		t.Errorf("expected StrPtr=hello, got %v", atom.StringPtrs["StrPtr"])
	}
	if atom.IntPtrs["IntPtr"] == nil || *atom.IntPtrs["IntPtr"] != 42 {
		t.Errorf("expected IntPtr=42, got %v", atom.IntPtrs["IntPtr"])
	}
	if atom.UintPtrs["UintPtr"] == nil || *atom.UintPtrs["UintPtr"] != 100 {
		t.Errorf("expected UintPtr=100, got %v", atom.UintPtrs["UintPtr"])
	}
	if atom.FloatPtrs["FloatPtr"] == nil || *atom.FloatPtrs["FloatPtr"] != 3.14 {
		t.Errorf("expected FloatPtr=3.14, got %v", atom.FloatPtrs["FloatPtr"])
	}
	if atom.BoolPtrs["BoolPtr"] == nil || *atom.BoolPtrs["BoolPtr"] != true {
		t.Errorf("expected BoolPtr=true, got %v", atom.BoolPtrs["BoolPtr"])
	}
	if atom.TimePtrs["TimePtr"] == nil {
		t.Error("expected TimePtr to be set")
	}
	if atom.BytePtrs["BytesPtr"] == nil || string(*atom.BytePtrs["BytesPtr"]) != "hello" {
		t.Errorf("expected BytesPtr=hello, got %v", atom.BytePtrs["BytesPtr"])
	}
}

func TestUnflatten_SliceTypes(t *testing.T) {
	spec := Spec{
		TypeName:    "AllSlices",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Strings", Type: "[]string"},
			{Name: "Ints", Type: "[]int64"},
			{Name: "Uints", Type: "[]uint64"},
			{Name: "Floats", Type: "[]float64"},
			{Name: "Bools", Type: "[]bool"},
			{Name: "Times", Type: "[]time.Time"},
			{Name: "ByteSlices", Type: "[][]byte"},
		},
	}

	data := map[string]any{
		"Strings":    []any{"a", "b", "c"},
		"Ints":       []any{float64(1), float64(2), float64(3)},
		"Uints":      []any{float64(10), float64(20)},
		"Floats":     []any{1.1, 2.2, 3.3},
		"Bools":      []any{true, false, true},
		"Times":      []any{"2024-01-01T00:00:00Z", "2024-06-15T12:00:00Z"},
		"ByteSlices": []any{"aGVsbG8=", "d29ybGQ="}, // base64 for "hello", "world"
	}

	atom := Unflatten(data, spec)

	if len(atom.StringSlices["Strings"]) != 3 || atom.StringSlices["Strings"][0] != "a" {
		t.Errorf("expected Strings=[a,b,c], got %v", atom.StringSlices["Strings"])
	}
	if len(atom.IntSlices["Ints"]) != 3 || atom.IntSlices["Ints"][0] != 1 {
		t.Errorf("expected Ints=[1,2,3], got %v", atom.IntSlices["Ints"])
	}
	if len(atom.UintSlices["Uints"]) != 2 || atom.UintSlices["Uints"][0] != 10 {
		t.Errorf("expected Uints=[10,20], got %v", atom.UintSlices["Uints"])
	}
	if len(atom.FloatSlices["Floats"]) != 3 || atom.FloatSlices["Floats"][0] != 1.1 {
		t.Errorf("expected Floats=[1.1,2.2,3.3], got %v", atom.FloatSlices["Floats"])
	}
	if len(atom.BoolSlices["Bools"]) != 3 || atom.BoolSlices["Bools"][0] != true {
		t.Errorf("expected Bools=[true,false,true], got %v", atom.BoolSlices["Bools"])
	}
	if len(atom.TimeSlices["Times"]) != 2 {
		t.Errorf("expected 2 times, got %v", len(atom.TimeSlices["Times"]))
	}
	if len(atom.ByteSlices["ByteSlices"]) != 2 || string(atom.ByteSlices["ByteSlices"][0]) != "hello" {
		t.Errorf("expected ByteSlices with hello/world, got %v", atom.ByteSlices["ByteSlices"])
	}
}

func TestUnflatten_Bytes(t *testing.T) {
	spec := Spec{
		TypeName:    "Doc",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Data", Type: "[]byte"},
			{Name: "Raw", Type: "[]uint8"},
		},
	}

	t.Run("base64 string", func(t *testing.T) {
		data := map[string]any{
			"Data": "aGVsbG8=", // base64 for "hello"
		}
		atom := Unflatten(data, spec)
		if string(atom.Bytes["Data"]) != "hello" {
			t.Errorf("expected Data=hello, got %v", string(atom.Bytes["Data"]))
		}
	})

	t.Run("raw bytes", func(t *testing.T) {
		data := map[string]any{
			"Data": []byte("world"),
		}
		atom := Unflatten(data, spec)
		if string(atom.Bytes["Data"]) != "world" {
			t.Errorf("expected Data=world, got %v", string(atom.Bytes["Data"]))
		}
	})

	t.Run("plain string fallback", func(t *testing.T) {
		data := map[string]any{
			"Data": "not-base64!@#",
		}
		atom := Unflatten(data, spec)
		if string(atom.Bytes["Data"]) != "not-base64!@#" {
			t.Errorf("expected Data=not-base64!@#, got %v", string(atom.Bytes["Data"]))
		}
	})
}

func TestUnflatten_DirectSliceTypes(t *testing.T) {
	// Test when slices are passed as typed slices, not []any
	spec := Spec{
		TypeName:    "TypedSlices",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Strings", Type: "[]string"},
			{Name: "Ints", Type: "[]int64"},
			{Name: "Uints", Type: "[]uint64"},
			{Name: "Floats", Type: "[]float64"},
			{Name: "Bools", Type: "[]bool"},
			{Name: "Times", Type: "[]time.Time"},
			{Name: "ByteSlices", Type: "[][]byte"},
		},
	}

	now := time.Now().UTC().Truncate(time.Second)
	data := map[string]any{
		"Strings":    []string{"x", "y"},
		"Ints":       []int64{100, 200},
		"Uints":      []uint64{1, 2, 3},
		"Floats":     []float64{0.1, 0.2},
		"Bools":      []bool{false, true},
		"Times":      []time.Time{now},
		"ByteSlices": [][]byte{[]byte("a"), []byte("b")},
	}

	atom := Unflatten(data, spec)

	if len(atom.StringSlices["Strings"]) != 2 {
		t.Errorf("expected 2 strings, got %d", len(atom.StringSlices["Strings"]))
	}
	if len(atom.IntSlices["Ints"]) != 2 {
		t.Errorf("expected 2 ints, got %d", len(atom.IntSlices["Ints"]))
	}
	if len(atom.UintSlices["Uints"]) != 3 {
		t.Errorf("expected 3 uints, got %d", len(atom.UintSlices["Uints"]))
	}
	if len(atom.FloatSlices["Floats"]) != 2 {
		t.Errorf("expected 2 floats, got %d", len(atom.FloatSlices["Floats"]))
	}
	if len(atom.BoolSlices["Bools"]) != 2 {
		t.Errorf("expected 2 bools, got %d", len(atom.BoolSlices["Bools"]))
	}
	if len(atom.TimeSlices["Times"]) != 1 {
		t.Errorf("expected 1 time, got %d", len(atom.TimeSlices["Times"]))
	}
	if len(atom.ByteSlices["ByteSlices"]) != 2 {
		t.Errorf("expected 2 byte slices, got %d", len(atom.ByteSlices["ByteSlices"]))
	}
}

func TestFlatten_AllPointerTypes(t *testing.T) {
	str := "hello"
	i := int64(42)
	u := uint64(100)
	f := 3.14
	b := true
	tm := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	bs := []byte("data")

	spec := Spec{
		TypeName:    "AllPtrs",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "StrPtr", Type: "*string"},
			{Name: "IntPtr", Type: "*int64"},
			{Name: "UintPtr", Type: "*uint64"},
			{Name: "FloatPtr", Type: "*float64"},
			{Name: "BoolPtr", Type: "*bool"},
			{Name: "TimePtr", Type: "*time.Time"},
			{Name: "BytesPtr", Type: "*[]byte"},
		},
	}

	atom := &Atom{
		Spec:       spec,
		StringPtrs: map[string]*string{"StrPtr": &str},
		IntPtrs:    map[string]*int64{"IntPtr": &i},
		UintPtrs:   map[string]*uint64{"UintPtr": &u},
		FloatPtrs:  map[string]*float64{"FloatPtr": &f},
		BoolPtrs:   map[string]*bool{"BoolPtr": &b},
		TimePtrs:   map[string]*time.Time{"TimePtr": &tm},
		BytePtrs:   map[string]*[]byte{"BytesPtr": &bs},
	}

	result := atom.Flatten()

	if result["StrPtr"] != "hello" {
		t.Errorf("expected StrPtr=hello, got %v", result["StrPtr"])
	}
	if result["IntPtr"] != int64(42) {
		t.Errorf("expected IntPtr=42, got %v", result["IntPtr"])
	}
	if result["UintPtr"] != uint64(100) {
		t.Errorf("expected UintPtr=100, got %v", result["UintPtr"])
	}
	if result["FloatPtr"] != 3.14 {
		t.Errorf("expected FloatPtr=3.14, got %v", result["FloatPtr"])
	}
	if result["BoolPtr"] != true {
		t.Errorf("expected BoolPtr=true, got %v", result["BoolPtr"])
	}
	if result["TimePtr"] != tm {
		t.Errorf("expected TimePtr=%v, got %v", tm, result["TimePtr"])
	}
	if b, ok := result["BytesPtr"].([]byte); !ok || string(b) != "data" {
		t.Errorf("expected BytesPtr=data, got %v", result["BytesPtr"])
	}
}

func TestFlatten_AllSliceTypes(t *testing.T) {
	spec := Spec{
		TypeName:    "AllSlices",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Strings", Type: "[]string"},
			{Name: "Ints", Type: "[]int64"},
			{Name: "Uints", Type: "[]uint64"},
			{Name: "Floats", Type: "[]float64"},
			{Name: "Bools", Type: "[]bool"},
			{Name: "Times", Type: "[]time.Time"},
			{Name: "ByteSlices", Type: "[][]byte"},
		},
	}

	now := time.Now().UTC().Truncate(time.Second)
	atom := &Atom{
		Spec:         spec,
		StringSlices: map[string][]string{"Strings": {"a", "b"}},
		IntSlices:    map[string][]int64{"Ints": {1, 2}},
		UintSlices:   map[string][]uint64{"Uints": {10, 20}},
		FloatSlices:  map[string][]float64{"Floats": {1.1, 2.2}},
		BoolSlices:   map[string][]bool{"Bools": {true, false}},
		TimeSlices:   map[string][]time.Time{"Times": {now}},
		ByteSlices:   map[string][][]byte{"ByteSlices": {[]byte("x"), []byte("y")}},
	}

	result := atom.Flatten()

	if s, ok := result["Strings"].([]string); !ok || len(s) != 2 {
		t.Errorf("expected Strings with 2 elements, got %v", result["Strings"])
	}
	if s, ok := result["Ints"].([]int64); !ok || len(s) != 2 {
		t.Errorf("expected Ints with 2 elements, got %v", result["Ints"])
	}
	if s, ok := result["Uints"].([]uint64); !ok || len(s) != 2 {
		t.Errorf("expected Uints with 2 elements, got %v", result["Uints"])
	}
	if s, ok := result["Floats"].([]float64); !ok || len(s) != 2 {
		t.Errorf("expected Floats with 2 elements, got %v", result["Floats"])
	}
	if s, ok := result["Bools"].([]bool); !ok || len(s) != 2 {
		t.Errorf("expected Bools with 2 elements, got %v", result["Bools"])
	}
	if s, ok := result["Times"].([]time.Time); !ok || len(s) != 1 {
		t.Errorf("expected Times with 1 element, got %v", result["Times"])
	}
	if s, ok := result["ByteSlices"].([][]byte); !ok || len(s) != 2 {
		t.Errorf("expected ByteSlices with 2 elements, got %v", result["ByteSlices"])
	}
}

func TestToInt64_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int64
	}{
		{"float64", float64(42), 42},
		{"int", int(42), 42},
		{"int64", int64(42), 42},
		{"unknown", "not a number", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toInt64(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestToUint64_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected uint64
	}{
		{"float64", float64(42), 42},
		{"uint", uint(42), 42},
		{"uint64", uint64(42), 42},
		{"int", int(42), 42},
		{"unknown", "not a number", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toUint64(tt.input)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestToFloat64_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected float64
	}{
		{"float64", float64(3.14), 3.14},
		{"int", int(42), 42.0},
		{"unknown", "not a number", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toFloat64(tt.input)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestToTime_AllTypes(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	nowRFC3339 := now.Format(time.RFC3339)

	tests := []struct {
		name  string
		input any
		check func(time.Time) bool
	}{
		{"time.Time", now, func(t time.Time) bool { return t.Equal(now) }},
		{"RFC3339 string", nowRFC3339, func(t time.Time) bool { return t.Equal(now) }},
		{"RFC3339Nano string", now.Format(time.RFC3339Nano), func(t time.Time) bool { return t.Equal(now) }},
		{"unknown", 12345, func(t time.Time) bool { return t.IsZero() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toTime(tt.input)
			if !tt.check(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

func TestUnflatten_NestedSlice(t *testing.T) {
	spec := Spec{
		TypeName:    "Container",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Items", Type: "[]Item", Kind: sentinel.KindSlice},
		},
	}

	data := map[string]any{
		"Items": []any{
			map[string]any{"Name": "item1"},
			map[string]any{"Name": "item2"},
		},
	}

	atom := Unflatten(data, spec)

	if len(atom.NestedSlices["Items"]) != 2 {
		t.Errorf("expected 2 nested items, got %d", len(atom.NestedSlices["Items"]))
	}
}
