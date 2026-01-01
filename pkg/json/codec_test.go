package json //nolint:revive // intentional name conflict for clear API

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/zoobzio/atom"
)

type jsonTestUser struct {
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	Email     *string   `db:"email"`
	CreatedAt time.Time `db:"created_at" time:"rfc3339"`
	Score     float64   `db:"score"`
	Active    bool      `db:"active"`
}

type jsonTestNested struct {
	ID      int             `db:"id"`
	Profile jsonTestProfile // Uses field name
}

type jsonTestProfile struct {
	Bio string `db:"bio"`
}

func TestEncode(t *testing.T) {
	atomizer, err := atom.Use[jsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	email := "test@example.com"
	user := &jsonTestUser{
		Name:      "Alice",
		Age:       30,
		Email:     &email,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Score:     95.5,
		Active:    true,
	}

	a := atomizer.Atomize(user)
	data, err := Encode(a)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Verify it's valid JSON
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify fields
	if m["name"] != "Alice" {
		t.Errorf("name = %v, want Alice", m["name"])
	}
	if m["age"] != float64(30) { // JSON numbers are float64
		t.Errorf("age = %v, want 30", m["age"])
	}
	if m["email"] != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", m["email"])
	}
	if m["active"] != true {
		t.Errorf("active = %v, want true", m["active"])
	}
}

func TestEncodeIndent(t *testing.T) {
	atomizer, err := atom.Use[jsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	user := &jsonTestUser{Name: "Bob", Age: 25, Active: false}
	a := atomizer.Atomize(user)

	data, err := EncodeIndent(a, "", "  ")
	if err != nil {
		t.Fatalf("EncodeIndent failed: %v", err)
	}

	// Should contain newlines and indentation
	if len(data) == 0 {
		t.Fatal("EncodeIndent returned empty data")
	}

	// Verify it's valid JSON
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
}

func TestDecode(t *testing.T) {
	_, err := atom.Use[jsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	jsonData := []byte(`{
		"name": "Charlie",
		"age": 35,
		"email": "charlie@example.com",
		"created_at": "2024-02-20T15:45:00Z",
		"score": 88.5,
		"active": true
	}`)

	atomizer, err := atom.Use[jsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}
	a, err := Decode(jsonData, atomizer.Spec())
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if a.Strings["Name"] != "Charlie" {
		t.Errorf("Name = %v, want Charlie", a.Strings["Name"])
	}
	if a.Ints["Age"] != 35 {
		t.Errorf("Age = %v, want 35", a.Ints["Age"])
	}
	if a.StringPtrs["Email"] == nil || *a.StringPtrs["Email"] != "charlie@example.com" {
		t.Errorf("Email = %v, want charlie@example.com", a.StringPtrs["Email"])
	}
	if a.Floats["Score"] != 88.5 {
		t.Errorf("Score = %v, want 88.5", a.Floats["Score"])
	}
	if a.Bools["Active"] != true {
		t.Errorf("Active = %v, want true", a.Bools["Active"])
	}

	expectedTime := time.Date(2024, 2, 20, 15, 45, 0, 0, time.UTC)
	if !a.Times["CreatedAt"].Equal(expectedTime) {
		t.Errorf("CreatedAt = %v, want %v", a.Times["CreatedAt"], expectedTime)
	}
}

func TestDecodeNested(t *testing.T) {
	_, err := atom.Use[jsonTestNested]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	jsonData := []byte(`{
		"id": 123,
		"Profile": {
			"bio": "Hello world"
		}
	}`)

	atomizer, err := atom.Use[jsonTestNested]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}
	a, err := Decode(jsonData, atomizer.Spec())
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if a.Ints["ID"] != 123 {
		t.Errorf("ID = %v, want 123", a.Ints["ID"])
	}

	nested, ok := a.Nested["Profile"]
	if !ok {
		t.Fatal("Profile nested not found")
	}
	if nested.Strings["Bio"] != "Hello world" {
		t.Errorf("Bio = %v, want Hello world", nested.Strings["Bio"])
	}
}

func TestRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[jsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	email := "roundtrip@example.com"
	original := &jsonTestUser{
		Name:      "RoundTrip",
		Age:       42,
		Email:     &email,
		CreatedAt: time.Date(2024, 3, 1, 8, 0, 0, 0, time.UTC),
		Score:     100.0,
		Active:    true,
	}

	a := atomizer.Atomize(original)

	// Encode
	data, err := Encode(a)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	decoded, err := Decode(data, a.Spec)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify
	if decoded.Strings["Name"] != original.Name {
		t.Errorf("Name mismatch: %v != %v", decoded.Strings["Name"], original.Name)
	}
	if decoded.Ints["Age"] != int64(original.Age) {
		t.Errorf("Age mismatch: %v != %v", decoded.Ints["Age"], original.Age)
	}
	if *decoded.StringPtrs["Email"] != *original.Email {
		t.Errorf("Email mismatch")
	}
	if decoded.Floats["Score"] != original.Score {
		t.Errorf("Score mismatch")
	}
	if decoded.Bools["Active"] != original.Active {
		t.Errorf("Active mismatch")
	}
}

func TestEncodeNoCodec(t *testing.T) {
	// Create an atom with a spec that has no registered codec
	a := &atom.Atom{
		Spec: atom.Spec{}, // Empty spec - no codec registered
	}

	_, err := Encode(a)
	if !errors.Is(err, atom.ErrNoCodec) {
		t.Errorf("Expected ErrNoCodec, got %v", err)
	}
}

func TestDecodeInvalidJSON(t *testing.T) {
	atomizer, err := atom.Use[jsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	_, err = Decode([]byte("not valid json"), atomizer.Spec())
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
