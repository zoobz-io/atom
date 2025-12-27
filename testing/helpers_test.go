package testing

import (
	"testing"
	"time"

	"github.com/zoobzio/atom"
)

type testUser struct {
	Name   string
	Age    int64
	Active bool
	Score  float64
}

func TestMustUse(t *testing.T) {
	atomizer := MustUse[testUser](t)
	if atomizer == nil {
		t.Fatal("expected non-nil atomizer")
	}
}

func TestAtomBuilder(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	a := NewAtomBuilder().
		String("Name", "Alice").
		Int("Age", 30).
		Uint("Count", 100).
		Float("Score", 95.5).
		Bool("Active", true).
		Time("Created", now).
		Bytes("Data", []byte("hello")).
		Build()

	if a.Strings["Name"] != "Alice" {
		t.Errorf("Name: got %q, want %q", a.Strings["Name"], "Alice")
	}
	if a.Ints["Age"] != 30 {
		t.Errorf("Age: got %d, want %d", a.Ints["Age"], 30)
	}
	if a.Uints["Count"] != 100 {
		t.Errorf("Count: got %d, want %d", a.Uints["Count"], 100)
	}
	if a.Floats["Score"] != 95.5 {
		t.Errorf("Score: got %g, want %g", a.Floats["Score"], 95.5)
	}
	if a.Bools["Active"] != true {
		t.Errorf("Active: got %v, want %v", a.Bools["Active"], true)
	}
	if !a.Times["Created"].Equal(now) {
		t.Errorf("Created: got %v, want %v", a.Times["Created"], now)
	}
	if string(a.Bytes["Data"]) != "hello" {
		t.Errorf("Data: got %q, want %q", a.Bytes["Data"], "hello")
	}
}

func TestAtomBuilderSlices(t *testing.T) {
	a := NewAtomBuilder().
		StringSlice("Tags", []string{"a", "b"}).
		IntSlice("Scores", []int64{1, 2, 3}).
		Build()

	if len(a.StringSlices["Tags"]) != 2 {
		t.Errorf("Tags length: got %d, want %d", len(a.StringSlices["Tags"]), 2)
	}
	if len(a.IntSlices["Scores"]) != 3 {
		t.Errorf("Scores length: got %d, want %d", len(a.IntSlices["Scores"]), 3)
	}
}

func TestAtomBuilderNested(t *testing.T) {
	inner := NewAtomBuilder().
		String("Street", "123 Main").
		Build()

	a := NewAtomBuilder().
		String("Name", "Alice").
		Nested("Address", inner).
		Build()

	if _, ok := a.Nested["Address"]; !ok {
		t.Fatal("expected Address in Nested")
	}
	if a.Nested["Address"].Strings["Street"] != "123 Main" {
		t.Error("nested Street mismatch")
	}
}

func TestRoundTripValidator(t *testing.T) {
	atomizer := MustUse[testUser](t)
	validator := NewRoundTripValidator(atomizer)

	original := &testUser{
		Name:   "Alice",
		Age:    30,
		Active: true,
		Score:  95.5,
	}

	validator.Validate(t, original)
}

func TestRoundTripValidatorAll(t *testing.T) {
	atomizer := MustUse[testUser](t)
	validator := NewRoundTripValidator(atomizer)

	cases := []*testUser{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25, Active: true},
		{Name: "", Age: 0}, // Edge case
	}

	validator.ValidateAll(t, cases)
}

func TestEqual(t *testing.T) {
	a := NewAtomBuilder().String("Name", "Alice").Int("Age", 30).Build()
	b := NewAtomBuilder().String("Name", "Alice").Int("Age", 30).Build()
	c := NewAtomBuilder().String("Name", "Bob").Int("Age", 30).Build()

	if !Equal(a, b) {
		t.Error("expected a and b to be equal")
	}
	if Equal(a, c) {
		t.Error("expected a and c to be different")
	}
}

func TestEqualFields(t *testing.T) {
	a := NewAtomBuilder().String("Name", "Alice").Int("Age", 30).Build()
	b := NewAtomBuilder().String("Name", "Alice").Int("Age", 25).Build()

	if !EqualFields(a, b, "Name") {
		t.Error("expected Name fields to be equal")
	}
	// Age is different, but we're only comparing Name
}

func TestDiff(t *testing.T) {
	a := NewAtomBuilder().String("Name", "Alice").Int("Age", 30).Build()
	b := NewAtomBuilder().String("Name", "Bob").Int("Age", 30).Build()

	diff := Diff(a, b)
	if diff == "<no differences>" {
		t.Error("expected differences")
	}
}

func TestFieldGetters(t *testing.T) {
	a := NewAtomBuilder().
		String("Name", "Alice").
		Int("Age", 30).
		Uint("Count", 100).
		Float("Score", 95.5).
		Bool("Active", true).
		Build()

	if v, ok := GetString(a, "Name"); !ok || v != "Alice" {
		t.Errorf("GetString: got %q, %v", v, ok)
	}
	if v, ok := GetInt(a, "Age"); !ok || v != 30 {
		t.Errorf("GetInt: got %d, %v", v, ok)
	}
	if v, ok := GetUint(a, "Count"); !ok || v != 100 {
		t.Errorf("GetUint: got %d, %v", v, ok)
	}
	if v, ok := GetFloat(a, "Score"); !ok || v != 95.5 {
		t.Errorf("GetFloat: got %g, %v", v, ok)
	}
	if v, ok := GetBool(a, "Active"); !ok || v != true {
		t.Errorf("GetBool: got %v, %v", v, ok)
	}

	// Missing field
	if _, ok := GetString(a, "Missing"); ok {
		t.Error("expected Missing to not exist")
	}
}

func TestAssertions(t *testing.T) {
	a := NewAtomBuilder().
		String("Name", "Alice").
		Int("Age", 30).
		Float("Score", 95.5).
		Bool("Active", true).
		Build()

	// These should pass
	AssertHasString(t, a, "Name", "Alice")
	AssertHasInt(t, a, "Age", 30)
	AssertHasFloat(t, a, "Score", 95.5)
	AssertHasBool(t, a, "Active", true)
	AssertMissingField(t, a, atom.TableStrings, "NotThere")
}

func TestRandomGenerators(t *testing.T) {
	// RandomString
	s := RandomString(10)
	if len(s) != 10 {
		t.Errorf("RandomString length: got %d, want 10", len(s))
	}

	// RandomInt
	for i := 0; i < 100; i++ {
		n := RandomInt(10, 20)
		if n < 10 || n > 20 {
			t.Errorf("RandomInt out of range: %d", n)
		}
	}

	// RandomFloat
	for i := 0; i < 100; i++ {
		f := RandomFloat(0, 1)
		if f < 0 || f >= 1 {
			t.Errorf("RandomFloat out of range: %g", f)
		}
	}

	// RandomBytes
	b := RandomBytes(16)
	if len(b) != 16 {
		t.Errorf("RandomBytes length: got %d, want 16", len(b))
	}

	// RandomTime
	tm := RandomTime()
	if tm.After(time.Now()) {
		t.Error("RandomTime should be in the past")
	}
}
