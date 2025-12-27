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
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name"}},
			{Name: "Age", Type: "int64", Tags: map[string]string{"json": "age"}},
			{Name: "Score", Type: "float64", Tags: map[string]string{"json": "score"}},
			{Name: "Active", Type: "bool", Tags: map[string]string{"json": "active"}},
		},
	}

	atom := &Atom{
		Spec:    spec,
		Strings: map[string]string{"Name": "John"},
		Ints:    map[string]int64{"Age": 30},
		Floats:  map[string]float64{"Score": 95.5},
		Bools:   map[string]bool{"Active": true},
	}

	result := atom.Flatten("json")

	if result["name"] != "John" {
		t.Errorf("expected name=John, got %v", result["name"])
	}
	if result["age"] != int64(30) {
		t.Errorf("expected age=30, got %v", result["age"])
	}
	if result["score"] != 95.5 {
		t.Errorf("expected score=95.5, got %v", result["score"])
	}
	if result["active"] != true {
		t.Errorf("expected active=true, got %v", result["active"])
	}
}

func TestFlatten_FallbackToFieldName(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string", Tags: map[string]string{}},
			{Name: "Age", Type: "int64", Tags: map[string]string{"json": ""}},
		},
	}

	atom := &Atom{
		Spec:    spec,
		Strings: map[string]string{"Name": "John"},
		Ints:    map[string]int64{"Age": 30},
	}

	result := atom.Flatten("json")

	if result["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", result["Name"])
	}
	if result["Age"] != int64(30) {
		t.Errorf("expected Age=30, got %v", result["Age"])
	}
}

func TestFlatten_SkipDashTag(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name"}},
			{Name: "Internal", Type: "string", Tags: map[string]string{"json": "-"}},
		},
	}

	atom := &Atom{
		Spec:    spec,
		Strings: map[string]string{"Name": "John", "Internal": "secret"},
	}

	result := atom.Flatten("json")

	if result["name"] != "John" {
		t.Errorf("expected name=John, got %v", result["name"])
	}
	if _, exists := result["Internal"]; exists {
		t.Error("expected Internal to be skipped")
	}
	if _, exists := result["-"]; exists {
		t.Error("expected - key to not exist")
	}
}

func TestFlatten_Pointers(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "*string", Tags: map[string]string{"json": "name"}},
			{Name: "Age", Type: "*int64", Tags: map[string]string{"json": "age"}},
			{Name: "Nickname", Type: "*string", Tags: map[string]string{"json": "nickname"}},
		},
	}

	name := "John"
	age := int64(30)
	atom := &Atom{
		Spec:       spec,
		StringPtrs: map[string]*string{"Name": &name, "Nickname": nil},
		IntPtrs:    map[string]*int64{"Age": &age},
	}

	result := atom.Flatten("json")

	if result["name"] != "John" {
		t.Errorf("expected name=John, got %v", result["name"])
	}
	if result["age"] != int64(30) {
		t.Errorf("expected age=30, got %v", result["age"])
	}
	if _, exists := result["nickname"]; exists {
		t.Error("expected nil pointer to be omitted")
	}
}

func TestFlatten_Slices(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Tags", Type: "[]string", Tags: map[string]string{"json": "tags"}},
			{Name: "Scores", Type: "[]int64", Tags: map[string]string{"json": "scores"}},
		},
	}

	atom := &Atom{
		Spec:         spec,
		StringSlices: map[string][]string{"Tags": {"go", "rust"}},
		IntSlices:    map[string][]int64{"Scores": {100, 95, 88}},
	}

	result := atom.Flatten("json")

	tags, ok := result["tags"].([]string)
	if !ok {
		t.Fatalf("expected tags to be []string, got %T", result["tags"])
	}
	if len(tags) != 2 || tags[0] != "go" || tags[1] != "rust" {
		t.Errorf("expected tags=[go, rust], got %v", tags)
	}

	scores, ok := result["scores"].([]int64)
	if !ok {
		t.Fatalf("expected scores to be []int64, got %T", result["scores"])
	}
	if len(scores) != 3 || scores[0] != 100 {
		t.Errorf("expected scores=[100, 95, 88], got %v", scores)
	}
}

func TestFlatten_Time(t *testing.T) {
	spec := Spec{
		TypeName:    "Event",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "CreatedAt", Type: "time.Time", Tags: map[string]string{"json": "created_at"}},
		},
	}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	atom := &Atom{
		Spec:  spec,
		Times: map[string]time.Time{"CreatedAt": now},
	}

	result := atom.Flatten("json")

	if result["created_at"] != now {
		t.Errorf("expected created_at=%v, got %v", now, result["created_at"])
	}
}

func TestFlatten_Nested(t *testing.T) {
	addressSpec := Spec{
		TypeName:    "Address",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "City", Type: "string", Tags: map[string]string{"json": "city"}},
			{Name: "Zip", Type: "string", Tags: map[string]string{"json": "zip"}},
		},
	}

	userSpec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name"}},
			{Name: "Address", Type: "Address", Kind: sentinel.KindStruct, Tags: map[string]string{"json": "address"}},
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

	result := atom.Flatten("json")

	if result["name"] != "John" {
		t.Errorf("expected name=John, got %v", result["name"])
	}

	addr, ok := result["address"].(map[string]any)
	if !ok {
		t.Fatalf("expected address to be map[string]any, got %T", result["address"])
	}
	if addr["city"] != "NYC" {
		t.Errorf("expected city=NYC, got %v", addr["city"])
	}
	if addr["zip"] != "10001" {
		t.Errorf("expected zip=10001, got %v", addr["zip"])
	}
}

func TestUnflatten_Scalars(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name"}},
			{Name: "Age", Type: "int64", Tags: map[string]string{"json": "age"}},
			{Name: "Score", Type: "float64", Tags: map[string]string{"json": "score"}},
			{Name: "Active", Type: "bool", Tags: map[string]string{"json": "active"}},
		},
	}

	data := map[string]any{
		"name":   "John",
		"age":    float64(30), // JSON numbers are float64
		"score":  95.5,
		"active": true,
	}

	atom := Unflatten(data, spec, "json")

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
			{Name: "Name", Type: "*string", Tags: map[string]string{"json": "name"}},
			{Name: "Age", Type: "*int64", Tags: map[string]string{"json": "age"}},
		},
	}

	data := map[string]any{
		"name": "John",
		"age":  float64(30),
	}

	atom := Unflatten(data, spec, "json")

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
			{Name: "Tags", Type: "[]string", Tags: map[string]string{"json": "tags"}},
			{Name: "Scores", Type: "[]int64", Tags: map[string]string{"json": "scores"}},
		},
	}

	data := map[string]any{
		"tags":   []any{"go", "rust"},
		"scores": []any{float64(100), float64(95)},
	}

	atom := Unflatten(data, spec, "json")

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
			{Name: "CreatedAt", Type: "time.Time", Tags: map[string]string{"json": "created_at"}},
		},
	}

	data := map[string]any{
		"created_at": "2024-01-15T10:30:00Z",
	}

	atom := Unflatten(data, spec, "json")

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
			{Name: "CreatedAt", Type: "time.Time", Tags: map[string]string{"json": "created_at"}},
		},
	}

	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	data := map[string]any{
		"created_at": expected,
	}

	atom := Unflatten(data, spec, "json")

	if !atom.Times["CreatedAt"].Equal(expected) {
		t.Errorf("expected CreatedAt=%v, got %v", expected, atom.Times["CreatedAt"])
	}
}

func TestRoundtrip(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name"}},
			{Name: "Age", Type: "int64", Tags: map[string]string{"json": "age"}},
			{Name: "Score", Type: "float64", Tags: map[string]string{"json": "score"}},
			{Name: "Active", Type: "bool", Tags: map[string]string{"json": "active"}},
			{Name: "Tags", Type: "[]string", Tags: map[string]string{"json": "tags"}},
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
	flat := original.Flatten("json")
	restored := Unflatten(flat, spec, "json")

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

func TestFlatten_DifferentTagKeys(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name", "bson": "user_name", "db": "user_name_col"}},
		},
	}

	atom := &Atom{
		Spec:    spec,
		Strings: map[string]string{"Name": "John"},
	}

	jsonResult := atom.Flatten("json")
	if jsonResult["name"] != "John" {
		t.Errorf("json: expected name=John, got %v", jsonResult["name"])
	}

	bsonResult := atom.Flatten("bson")
	if bsonResult["user_name"] != "John" {
		t.Errorf("bson: expected user_name=John, got %v", bsonResult["user_name"])
	}

	dbResult := atom.Flatten("db")
	if dbResult["user_name_col"] != "John" {
		t.Errorf("db: expected user_name_col=John, got %v", dbResult["user_name_col"])
	}
}

func TestFlatten_Uints(t *testing.T) {
	spec := Spec{
		TypeName:    "Counter",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Count", Type: "uint64", Tags: map[string]string{"json": "count"}},
		},
	}

	atom := &Atom{
		Spec:  spec,
		Uints: map[string]uint64{"Count": 42},
	}

	result := atom.Flatten("json")

	if result["count"] != uint64(42) {
		t.Errorf("expected count=42, got %v", result["count"])
	}
}

func TestUnflatten_Uints(t *testing.T) {
	spec := Spec{
		TypeName:    "Counter",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Count", Type: "uint64", Tags: map[string]string{"json": "count"}},
		},
	}

	data := map[string]any{
		"count": float64(42), // JSON numbers
	}

	atom := Unflatten(data, spec, "json")

	if atom.Uints["Count"] != 42 {
		t.Errorf("expected Count=42, got %v", atom.Uints["Count"])
	}
}

func TestFlatten_Bytes(t *testing.T) {
	spec := Spec{
		TypeName:    "Document",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Data", Type: "[]uint8", Tags: map[string]string{"json": "data"}},
		},
	}

	atom := &Atom{
		Spec:  spec,
		Bytes: map[string][]byte{"Data": []byte("hello")},
	}

	result := atom.Flatten("json")

	data, ok := result["data"].([]byte)
	if !ok || string(data) != "hello" {
		t.Errorf("expected data=hello, got %v", result["data"])
	}
}

func TestUnflatten_NilValue(t *testing.T) {
	spec := Spec{
		TypeName:    "User",
		PackageName: "test",
		Fields: []sentinel.FieldMetadata{
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name"}},
			{Name: "Age", Type: "int64", Tags: map[string]string{"json": "age"}},
		},
	}

	data := map[string]any{
		"name": "John",
		"age":  nil,
	}

	atom := Unflatten(data, spec, "json")

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
			{Name: "Name", Type: "string", Tags: map[string]string{"json": "name"}},
		},
	}

	data := map[string]any{
		"name":    "John",
		"unknown": "value",
	}

	atom := Unflatten(data, spec, "json")

	if atom.Strings["Name"] != "John" {
		t.Errorf("expected Name=John, got %v", atom.Strings["Name"])
	}
	// Should not panic or error on unknown field
}
