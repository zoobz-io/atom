package integration

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/zoobz-io/atom"
)

// Simple struct with basic types.
type SimpleUser struct {
	Name   string
	Age    int64
	Score  float64
	Active bool
}

func TestSimpleRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[SimpleUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	cases := []SimpleUser{
		{Name: "Alice", Age: 30, Score: 95.5, Active: true},
		{Name: "Bob", Age: 0, Score: 0, Active: false},
		{Name: "", Age: -1, Score: -0.5, Active: true},
		{Name: "Unicode: 日本語", Age: math.MaxInt64, Score: math.MaxFloat64, Active: false},
	}

	for _, original := range cases {
		a := atomizer.Atomize(&original)
		restored, err := atomizer.Deatomize(a)
		if err != nil {
			t.Fatalf("Deatomize failed: %v", err)
		}
		if !reflect.DeepEqual(&original, restored) {
			t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, *restored)
		}
	}
}

// Struct with all scalar types.
type AllScalars struct {
	String  string
	Int     int64
	Uint    uint64
	Float   float64
	Bool    bool
	Time    time.Time
	Bytes   []byte
	Int8    int8
	Int16   int16
	Int32   int32
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Float32 float32
}

func TestAllScalarsRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[AllScalars]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	original := &AllScalars{
		String:  "test",
		Int:     -12345,
		Uint:    12345,
		Float:   3.14159,
		Bool:    true,
		Time:    now,
		Bytes:   []byte("hello"),
		Int8:    -100,
		Int16:   -1000,
		Int32:   -100000,
		Uint8:   200,
		Uint16:  60000,
		Uint32:  3000000000,
		Float32: 2.5,
	}

	a := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(a)
	if err != nil {
		t.Fatalf("Deatomize failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}
}

// Struct with pointer types.
type WithPointers struct {
	Name     string
	OptInt   *int64
	OptFloat *float64
	OptBool  *bool
	OptTime  *time.Time
}

func TestPointersRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[WithPointers]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	intVal := int64(42)
	floatVal := 3.14
	boolVal := true
	timeVal := time.Now().Truncate(time.Second)

	cases := []*WithPointers{
		{Name: "all nil"},
		{Name: "all set", OptInt: &intVal, OptFloat: &floatVal, OptBool: &boolVal, OptTime: &timeVal},
		{Name: "partial", OptInt: &intVal, OptBool: &boolVal},
	}

	for _, original := range cases {
		t.Run(original.Name, func(t *testing.T) {
			a := atomizer.Atomize(original)
			restored, err := atomizer.Deatomize(a)
			if err != nil {
				t.Fatalf("Deatomize failed: %v", err)
			}
			if !reflect.DeepEqual(original, restored) {
				t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
			}
		})
	}
}

// Struct with slice types.
type WithSlices struct {
	Strings []string
	Ints    []int64
	Floats  []float64
	Bools   []bool
	Times   []time.Time
	Bytes   [][]byte
}

func TestSlicesRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[WithSlices]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	cases := []*WithSlices{
		{}, // All nil
		{Strings: []string{}}, // Empty slice
		{
			Strings: []string{"a", "b", "c"},
			Ints:    []int64{1, 2, 3},
			Floats:  []float64{1.1, 2.2, 3.3},
			Bools:   []bool{true, false, true},
			Times:   []time.Time{now, now.Add(time.Hour)},
			Bytes:   [][]byte{{1, 2}, {3, 4, 5}},
		},
	}

	for i, original := range cases {
		t.Run(string(rune('0'+i)), func(t *testing.T) {
			a := atomizer.Atomize(original)
			restored, err := atomizer.Deatomize(a)
			if err != nil {
				t.Fatalf("Deatomize failed: %v", err)
			}
			if !reflect.DeepEqual(original, restored) {
				t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
			}
		})
	}
}

// Nested struct.
type Address struct {
	Street  string
	City    string
	ZipCode string
}

type Person struct {
	Name    string
	Address Address
}

func TestNestedRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[Person]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	original := &Person{
		Name: "Alice",
		Address: Address{
			Street:  "123 Main St",
			City:    "Springfield",
			ZipCode: "12345",
		},
	}

	a := atomizer.Atomize(original)

	// Verify nested structure
	if _, ok := a.Nested["Address"]; !ok {
		t.Fatal("expected Address in Nested")
	}

	restored, err := atomizer.Deatomize(a)
	if err != nil {
		t.Fatalf("Deatomize failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}
}

// Pointer to nested struct.
type PersonWithOptionalAddress struct {
	Name    string
	Address *Address
}

func TestNestedPointerRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[PersonWithOptionalAddress]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	cases := []*PersonWithOptionalAddress{
		{Name: "NoAddress", Address: nil},
		{Name: "HasAddress", Address: &Address{Street: "123 Main", City: "Boston"}},
	}

	for _, original := range cases {
		t.Run(original.Name, func(t *testing.T) {
			a := atomizer.Atomize(original)
			restored, err := atomizer.Deatomize(a)
			if err != nil {
				t.Fatalf("Deatomize failed: %v", err)
			}
			if !reflect.DeepEqual(original, restored) {
				t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
			}
		})
	}
}

// Slice of nested structs.
type Team struct {
	Name    string
	Members []Person
}

func TestNestedSliceRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[Team]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	original := &Team{
		Name: "Engineering",
		Members: []Person{
			{Name: "Alice", Address: Address{City: "Boston"}},
			{Name: "Bob", Address: Address{City: "Seattle"}},
		},
	}

	a := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(a)
	if err != nil {
		t.Fatalf("Deatomize failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}
}

// Self-referential struct.
type Node struct {
	Value    int64
	Children []Node
}

func TestSelfReferentialRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[Node]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	original := &Node{
		Value: 1,
		Children: []Node{
			{Value: 2, Children: nil},
			{
				Value: 3,
				Children: []Node{
					{Value: 4, Children: nil},
				},
			},
		},
	}

	a := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(a)
	if err != nil {
		t.Fatalf("Deatomize failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}
}

// Named types.
type UserID string
type Status int

const (
	StatusActive Status = iota
	StatusInactive
)

type UserWithNamedTypes struct {
	ID     UserID
	Status Status
}

func TestNamedTypesRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[UserWithNamedTypes]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	original := &UserWithNamedTypes{
		ID:     "usr_123",
		Status: StatusActive,
	}

	a := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(a)
	if err != nil {
		t.Fatalf("Deatomize failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}

	// Verify type preservation
	if reflect.TypeOf(restored.ID) != reflect.TypeOf(original.ID) {
		t.Error("UserID type not preserved")
	}
	if reflect.TypeOf(restored.Status) != reflect.TypeOf(original.Status) {
		t.Error("Status type not preserved")
	}
}

// Fixed byte array.
type WithFixedArray struct {
	Hash [32]byte
	UUID [16]byte
}

func TestFixedArrayRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[WithFixedArray]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	var hash [32]byte
	var uuid [16]byte
	copy(hash[:], "12345678901234567890123456789012")
	copy(uuid[:], "1234567890123456")

	original := &WithFixedArray{Hash: hash, UUID: uuid}

	a := atomizer.Atomize(original)
	restored, err := atomizer.Deatomize(a)
	if err != nil {
		t.Fatalf("Deatomize failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("mismatch:\noriginal: %+v\nrestored: %+v", original, restored)
	}
}
