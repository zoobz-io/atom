package atom

import (
	"testing"
	"time"
)

type codecTestUser struct {
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	Email     *string   `db:"email"`
	CreatedAt time.Time `db:"created_at" time:"rfc3339"`
	UpdatedAt time.Time `db:"updated_at" time:"unix"`
	Score     float64   `db:"score"`
	Active    bool      `db:"active"`
	NoTag     string    // Should be skipped
}

type codecTestNested struct {
	ID      int              `db:"id"`
	Profile codecTestProfile // No db tag - uses field name as key
}

type codecTestProfile struct {
	Bio     string `db:"bio"`
	Website string `db:"website"`
}

type codecTestSlices struct {
	Tags   []string  `db:"tags"`
	Scores []float64 `db:"scores"`
}

func TestCodecFor(t *testing.T) {
	// Register type
	_, err := Use[codecTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	// Get codec
	spec := mustSpec[codecTestUser](t)
	codec, ok := CodecFor(spec)
	if !ok {
		t.Fatal("CodecFor returned false for registered type")
	}
	if codec == nil {
		t.Fatal("CodecFor returned nil codec")
	}
}

func TestCodecEncodeMap(t *testing.T) {
	atomizer, err := Use[codecTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	email := "test@example.com"
	user := &codecTestUser{
		Name:      "Alice",
		Age:       30,
		Email:     &email,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Score:     95.5,
		Active:    true,
		NoTag:     "should not appear",
	}

	a := atomizer.Atomize(user)
	codec, _ := CodecFor(a.Spec)
	m := codec.EncodeMap(a)

	// Check db-tagged fields are present with correct keys
	if m["name"] != "Alice" {
		t.Errorf("name = %v, want Alice", m["name"])
	}
	if m["age"] != int64(30) {
		t.Errorf("age = %v, want 30", m["age"])
	}
	if m["email"] != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", m["email"])
	}
	if m["score"] != 95.5 {
		t.Errorf("score = %v, want 95.5", m["score"])
	}
	if m["active"] != true {
		t.Errorf("active = %v, want true", m["active"])
	}

	// Check time formats
	if m["created_at"] != "2024-01-15T10:30:00Z" {
		t.Errorf("created_at = %v, want RFC3339 string", m["created_at"])
	}
	if m["updated_at"] != int64(1717243200) {
		t.Errorf("updated_at = %v, want unix timestamp", m["updated_at"])
	}

	// Check NoTag field is NOT present
	if _, ok := m["NoTag"]; ok {
		t.Error("NoTag should not be in encoded map")
	}
}

func TestCodecDecodeMap(t *testing.T) {
	_, err := Use[codecTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	spec := mustSpec[codecTestUser](t)
	codec, _ := CodecFor(spec)

	m := map[string]any{
		"name":       "Bob",
		"age":        float64(25), // JSON numbers are float64
		"email":      "bob@example.com",
		"created_at": "2024-02-20T15:45:00Z",
		"updated_at": int64(1717243200),
		"score":      float64(88.5),
		"active":     false,
		"unknown":    "should be ignored",
	}

	a, err := codec.DecodeMap(m)
	if err != nil {
		t.Fatalf("DecodeMap failed: %v", err)
	}

	// Verify decoded values
	if a.Strings["Name"] != "Bob" {
		t.Errorf("Name = %v, want Bob", a.Strings["Name"])
	}
	if a.Ints["Age"] != 25 {
		t.Errorf("Age = %v, want 25", a.Ints["Age"])
	}
	if a.StringPtrs["Email"] == nil || *a.StringPtrs["Email"] != "bob@example.com" {
		t.Errorf("Email = %v, want bob@example.com", a.StringPtrs["Email"])
	}
	if a.Floats["Score"] != 88.5 {
		t.Errorf("Score = %v, want 88.5", a.Floats["Score"])
	}
	if a.Bools["Active"] != false {
		t.Errorf("Active = %v, want false", a.Bools["Active"])
	}

	// Verify time parsing
	expectedCreated := time.Date(2024, 2, 20, 15, 45, 0, 0, time.UTC)
	if !a.Times["CreatedAt"].Equal(expectedCreated) {
		t.Errorf("CreatedAt = %v, want %v", a.Times["CreatedAt"], expectedCreated)
	}
	expectedUpdated := time.Unix(1717243200, 0)
	if !a.Times["UpdatedAt"].Equal(expectedUpdated) {
		t.Errorf("UpdatedAt = %v, want %v", a.Times["UpdatedAt"], expectedUpdated)
	}
}

func TestCodecNullHandling(t *testing.T) {
	_, err := Use[codecTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	spec := mustSpec[codecTestUser](t)
	codec, _ := CodecFor(spec)

	m := map[string]any{
		"name":  "Charlie",
		"age":   float64(40),
		"email": nil, // null pointer field
		"score": nil, // null scalar field - should be skipped
	}

	a, err := codec.DecodeMap(m)
	if err != nil {
		t.Fatalf("DecodeMap failed: %v", err)
	}

	// Pointer field should be nil
	if ptr, ok := a.StringPtrs["Email"]; !ok {
		t.Error("Email should be present in StringPtrs")
	} else if ptr != nil {
		t.Error("Email should be nil")
	}

	// Scalar field with null should be skipped (not present)
	if _, ok := a.Floats["Score"]; ok {
		t.Error("Score should not be present when null")
	}
}

func TestCodecNestedStruct(t *testing.T) {
	_, err := Use[codecTestNested]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	spec := mustSpec[codecTestNested](t)
	codec, _ := CodecFor(spec)

	m := map[string]any{
		"id": float64(123),
		"Profile": map[string]any{ // Uses field name since no db tag
			"bio":     "Hello world",
			"website": "https://example.com",
		},
	}

	a, err := codec.DecodeMap(m)
	if err != nil {
		t.Fatalf("DecodeMap failed: %v", err)
	}

	if a.Ints["ID"] != 123 {
		t.Errorf("ID = %v, want 123", a.Ints["ID"])
	}

	nested, ok := a.Nested["Profile"]
	if !ok {
		t.Fatal("Profile nested atom not found")
	}
	if nested.Strings["Bio"] != "Hello world" {
		t.Errorf("Bio = %v, want Hello world", nested.Strings["Bio"])
	}
	if nested.Strings["Website"] != "https://example.com" {
		t.Errorf("Website = %v, want https://example.com", nested.Strings["Website"])
	}
}

func TestCodecSlices(t *testing.T) {
	atomizer, err := Use[codecTestSlices]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	original := &codecTestSlices{
		Tags:   []string{"go", "rust", "python"},
		Scores: []float64{1.1, 2.2, 3.3},
	}

	a := atomizer.Atomize(original)
	codec, _ := CodecFor(a.Spec)

	// Encode
	m := codec.EncodeMap(a)
	if tags, ok := m["tags"].([]string); !ok || len(tags) != 3 {
		t.Errorf("tags = %v, want []string with 3 elements", m["tags"])
	}
	if scores, ok := m["scores"].([]float64); !ok || len(scores) != 3 {
		t.Errorf("scores = %v, want []float64 with 3 elements", m["scores"])
	}

	// Decode
	decoded, err := codec.DecodeMap(m)
	if err != nil {
		t.Fatalf("DecodeMap failed: %v", err)
	}

	if len(decoded.StringSlices["Tags"]) != 3 {
		t.Errorf("Tags len = %v, want 3", len(decoded.StringSlices["Tags"]))
	}
	if len(decoded.FloatSlices["Scores"]) != 3 {
		t.Errorf("Scores len = %v, want 3", len(decoded.FloatSlices["Scores"]))
	}
}

func TestCodecRoundTrip(t *testing.T) {
	atomizer, err := Use[codecTestUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	email := "roundtrip@example.com"
	original := &codecTestUser{
		Name:      "RoundTrip",
		Age:       42,
		Email:     &email,
		CreatedAt: time.Date(2024, 3, 1, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 3, 1, 8, 0, 0, 0, time.UTC),
		Score:     100.0,
		Active:    true,
	}

	a := atomizer.Atomize(original)
	codec, _ := CodecFor(a.Spec)

	// Encode then decode
	m := codec.EncodeMap(a)
	decoded, err := codec.DecodeMap(m)
	if err != nil {
		t.Fatalf("DecodeMap failed: %v", err)
	}

	// Verify all fields match
	if decoded.Strings["Name"] != original.Name {
		t.Errorf("Name mismatch: %v != %v", decoded.Strings["Name"], original.Name)
	}
	if decoded.Ints["Age"] != int64(original.Age) {
		t.Errorf("Age mismatch: %v != %v", decoded.Ints["Age"], original.Age)
	}
	if *decoded.StringPtrs["Email"] != *original.Email {
		t.Errorf("Email mismatch: %v != %v", *decoded.StringPtrs["Email"], *original.Email)
	}
	if decoded.Floats["Score"] != original.Score {
		t.Errorf("Score mismatch: %v != %v", decoded.Floats["Score"], original.Score)
	}
	if decoded.Bools["Active"] != original.Active {
		t.Errorf("Active mismatch: %v != %v", decoded.Bools["Active"], original.Active)
	}
}

// Helper to get spec for a type.
func mustSpec[T any](t *testing.T) Spec {
	t.Helper()
	atomizer, err := Use[T]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}
	return atomizer.Spec()
}
