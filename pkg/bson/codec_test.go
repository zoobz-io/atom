package bson

import (
	"errors"
	"testing"
	"time"

	"github.com/zoobzio/atom"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type bsonTestUser struct {
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	Email     *string   `db:"email"`
	CreatedAt time.Time `db:"created_at" time:"native"`
	Score     float64   `db:"score"`
	Active    bool      `db:"active"`
}

type bsonTestNested struct {
	ID      int             `db:"id"`
	Profile bsonTestProfile // Uses field name
}

type bsonTestProfile struct {
	Bio string `db:"bio"`
}

func TestEncode(t *testing.T) {
	atomizer, err := atom.Use[bsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	email := "test@example.com"
	user := &bsonTestUser{
		Name:      "Alice",
		Age:       30,
		Email:     &email,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Score:     95.5,
		Active:    true,
	}

	a := atomizer.Atomize(user)
	m, err := Encode(a)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if m["name"] != "Alice" {
		t.Errorf("name = %v, want Alice", m["name"])
	}
	if m["age"] != int64(30) {
		t.Errorf("age = %v, want 30", m["age"])
	}
	if m["email"] != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", m["email"])
	}
	if m["active"] != true {
		t.Errorf("active = %v, want true", m["active"])
	}
}

func TestDecode(t *testing.T) {
	atomizer, err := atom.Use[bsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	m := bson.M{
		"name":       "Charlie",
		"age":        int64(35),
		"email":      "charlie@example.com",
		"created_at": time.Date(2024, 2, 20, 15, 45, 0, 0, time.UTC),
		"score":      88.5,
		"active":     true,
	}

	a, err := Decode(m, atomizer.Spec())
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
}

func TestDecodeNested(t *testing.T) {
	atomizer, err := atom.Use[bsonTestNested]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	m := bson.M{
		"id": int64(123),
		"Profile": bson.M{
			"bio": "Hello world",
		},
	}

	a, err := Decode(m, atomizer.Spec())
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

func TestMarshalUnmarshal(t *testing.T) {
	atomizer, err := atom.Use[bsonTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	email := "roundtrip@example.com"
	original := &bsonTestUser{
		Name:      "RoundTrip",
		Age:       42,
		Email:     &email,
		CreatedAt: time.Date(2024, 3, 1, 8, 0, 0, 0, time.UTC),
		Score:     100.0,
		Active:    true,
	}

	a := atomizer.Atomize(original)

	// Marshal
	data, err := Marshal(a)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	decoded, err := Unmarshal(data, a.Spec)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify
	if decoded.Strings["Name"] != original.Name {
		t.Errorf("Name mismatch: %v != %v", decoded.Strings["Name"], original.Name)
	}
	if decoded.Ints["Age"] != int64(original.Age) {
		t.Errorf("Age mismatch: %v != %v", decoded.Ints["Age"], original.Age)
	}
}

func TestEncodeNoCodec(t *testing.T) {
	a := &atom.Atom{
		Spec: atom.Spec{},
	}

	_, err := Encode(a)
	if !errors.Is(err, atom.ErrNoCodec) {
		t.Errorf("Expected ErrNoCodec, got %v", err)
	}
}
