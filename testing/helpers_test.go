package testing

import (
	"fmt"
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

func TestAtomBuilderPtrs(t *testing.T) {
	str := "hello"
	i := int64(42)

	a := NewAtomBuilder().
		StringPtr("Name", &str).
		IntPtr("Age", &i).
		Build()

	if a.StringPtrs["Name"] == nil || *a.StringPtrs["Name"] != "hello" {
		t.Errorf("StringPtr: expected 'hello', got %v", a.StringPtrs["Name"])
	}
	if a.IntPtrs["Age"] == nil || *a.IntPtrs["Age"] != 42 {
		t.Errorf("IntPtr: expected 42, got %v", a.IntPtrs["Age"])
	}

	// Test with nil pointers
	aNil := NewAtomBuilder().
		StringPtr("Name", nil).
		IntPtr("Age", nil).
		Build()

	if aNil.StringPtrs["Name"] != nil {
		t.Error("expected nil StringPtr")
	}
	if aNil.IntPtrs["Age"] != nil {
		t.Error("expected nil IntPtr")
	}
}

func TestAtomBuilderNestedSlice(t *testing.T) {
	items := []atom.Atom{
		*NewAtomBuilder().String("Name", "Item1").Build(),
		*NewAtomBuilder().String("Name", "Item2").Build(),
	}

	a := NewAtomBuilder().
		NestedSlice("Items", items).
		Build()

	if len(a.NestedSlices["Items"]) != 2 {
		t.Errorf("expected 2 items, got %d", len(a.NestedSlices["Items"]))
	}
	if a.NestedSlices["Items"][0].Strings["Name"] != "Item1" {
		t.Error("first item name mismatch")
	}
}

func TestAtomBuilderWithSpec(t *testing.T) {
	spec := atom.Spec{
		TypeName:    "TestType",
		PackageName: "testpkg",
	}

	a := NewAtomBuilder().
		WithSpec(spec).
		String("Name", "Test").
		Build()

	if a.Spec.TypeName != "TestType" {
		t.Errorf("expected TypeName='TestType', got %q", a.Spec.TypeName)
	}
	if a.Spec.PackageName != "testpkg" {
		t.Errorf("expected PackageName='testpkg', got %q", a.Spec.PackageName)
	}
}

func TestGetTimeAndBytes(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	data := []byte{0x01, 0x02, 0x03}

	a := NewAtomBuilder().
		Time("Created", now).
		Bytes("Data", data).
		Build()

	if tm, ok := GetTime(a, "Created"); !ok || !tm.Equal(now) {
		t.Errorf("GetTime: got %v, %v", tm, ok)
	}
	if b, ok := GetBytes(a, "Data"); !ok || len(b) != 3 {
		t.Errorf("GetBytes: got %v, %v", b, ok)
	}

	// Missing fields
	if _, ok := GetTime(a, "Missing"); ok {
		t.Error("expected Missing time to not exist")
	}
	if _, ok := GetBytes(a, "Missing"); ok {
		t.Error("expected Missing bytes to not exist")
	}
}

func TestAssertHasNested(t *testing.T) {
	inner := NewAtomBuilder().String("City", "NYC").Build()
	a := NewAtomBuilder().Nested("Address", inner).Build()

	// Should pass
	AssertHasNested(t, a, "Address")
}

func TestAssertMissingFieldAllTypes(t *testing.T) {
	a := NewAtomBuilder().Build() // Empty atom

	// Should all pass - fields are missing
	AssertMissingField(t, a, atom.TableStrings, "Name")
	AssertMissingField(t, a, atom.TableInts, "Age")
	AssertMissingField(t, a, atom.TableFloats, "Score")
	AssertMissingField(t, a, atom.TableBools, "Active")
}

func TestEqualFieldsAllTypes(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	data := []byte{0x01}

	a := NewAtomBuilder().
		String("Name", "Alice").
		Int("Age", 30).
		Float("Score", 95.5).
		Bool("Active", true).
		Time("Created", now).
		Bytes("Data", data).
		Build()

	b := NewAtomBuilder().
		String("Name", "Alice").
		Int("Age", 30).
		Float("Score", 95.5).
		Bool("Active", true).
		Time("Created", now).
		Bytes("Data", data).
		Build()

	c := NewAtomBuilder().
		String("Name", "Bob").
		Int("Age", 25).
		Float("Score", 80.0).
		Bool("Active", false).
		Time("Created", now.Add(time.Hour)).
		Bytes("Data", []byte{0x02}).
		Build()

	// Same fields should be equal
	if !EqualFields(a, b, "Name", "Age", "Score", "Active", "Created", "Data") {
		t.Error("expected a and b to have equal fields")
	}

	// Different fields should not be equal
	if EqualFields(a, c, "Name") {
		t.Error("expected Name fields to differ")
	}
	if EqualFields(a, c, "Age") {
		t.Error("expected Age fields to differ")
	}
	if EqualFields(a, c, "Score") {
		t.Error("expected Score fields to differ")
	}
	if EqualFields(a, c, "Active") {
		t.Error("expected Active fields to differ")
	}
	if EqualFields(a, c, "Created") {
		t.Error("expected Created fields to differ")
	}
	if EqualFields(a, c, "Data") {
		t.Error("expected Data fields to differ")
	}
}

func TestDiffNoDifferences(t *testing.T) {
	a := NewAtomBuilder().String("Name", "Alice").Int("Age", 30).Build()
	b := NewAtomBuilder().String("Name", "Alice").Int("Age", 30).Build()

	diff := Diff(a, b)
	if diff != "<no differences>" {
		t.Errorf("expected no differences, got: %s", diff)
	}
}

func TestDiffMissingFields(t *testing.T) {
	a := NewAtomBuilder().String("Name", "Alice").String("Email", "alice@test.com").Build()
	b := NewAtomBuilder().String("Name", "Alice").Build()

	diff := Diff(a, b)
	if diff == "<no differences>" {
		t.Error("expected differences for missing Email")
	}

	// Test missing in a
	diff2 := Diff(b, a)
	if diff2 == "<no differences>" {
		t.Error("expected differences for extra Email")
	}
}

func TestDiffIntValues(t *testing.T) {
	a := NewAtomBuilder().Int("Age", 30).Build()
	b := NewAtomBuilder().Int("Age", 25).Build()

	diff := Diff(a, b)
	if diff == "<no differences>" {
		t.Error("expected differences for Age")
	}

	// Test missing int
	c := NewAtomBuilder().Int("Age", 30).Int("Count", 10).Build()
	d := NewAtomBuilder().Int("Age", 30).Build()

	diff2 := Diff(c, d)
	if diff2 == "<no differences>" {
		t.Error("expected differences for missing Count")
	}
}

// spyT captures test failures without failing the actual test.
type spyT struct {
	testing.TB
	errors []string
}

func (*spyT) Helper() {}

func (s *spyT) Errorf(format string, args ...interface{}) {
	s.errors = append(s.errors, fmt.Sprintf(format, args...))
}

func (s *spyT) hasError() bool {
	return len(s.errors) > 0
}

func TestAssertHasStringFailures(t *testing.T) {
	a := NewAtomBuilder().String("Name", "Alice").Build()

	// Missing field
	spy := &spyT{}
	AssertHasString(spy, a, "Missing", "value")
	if !spy.hasError() {
		t.Error("expected error for missing field")
	}

	// Wrong value
	spy = &spyT{}
	AssertHasString(spy, a, "Name", "Bob")
	if !spy.hasError() {
		t.Error("expected error for wrong value")
	}
}

func TestAssertHasIntFailures(t *testing.T) {
	a := NewAtomBuilder().Int("Age", 30).Build()

	// Missing field
	spy := &spyT{}
	AssertHasInt(spy, a, "Missing", 0)
	if !spy.hasError() {
		t.Error("expected error for missing field")
	}

	// Wrong value
	spy = &spyT{}
	AssertHasInt(spy, a, "Age", 25)
	if !spy.hasError() {
		t.Error("expected error for wrong value")
	}
}

func TestAssertHasFloatFailures(t *testing.T) {
	a := NewAtomBuilder().Float("Score", 95.5).Build()

	// Missing field
	spy := &spyT{}
	AssertHasFloat(spy, a, "Missing", 0)
	if !spy.hasError() {
		t.Error("expected error for missing field")
	}

	// Wrong value
	spy = &spyT{}
	AssertHasFloat(spy, a, "Score", 80.0)
	if !spy.hasError() {
		t.Error("expected error for wrong value")
	}
}

func TestAssertHasBoolFailures(t *testing.T) {
	a := NewAtomBuilder().Bool("Active", true).Build()

	// Missing field
	spy := &spyT{}
	AssertHasBool(spy, a, "Missing", false)
	if !spy.hasError() {
		t.Error("expected error for missing field")
	}

	// Wrong value
	spy = &spyT{}
	AssertHasBool(spy, a, "Active", false)
	if !spy.hasError() {
		t.Error("expected error for wrong value")
	}
}

func TestAssertHasNestedFailure(t *testing.T) {
	a := NewAtomBuilder().Build() // No nested fields

	spy := &spyT{}
	AssertHasNested(spy, a, "Missing")
	if !spy.hasError() {
		t.Error("expected error for missing nested field")
	}
}

func TestAssertMissingFieldWhenPresent(t *testing.T) {
	a := NewAtomBuilder().
		String("Name", "Alice").
		Int("Age", 30).
		Float("Score", 95.5).
		Bool("Active", true).
		Build()

	// String present
	spy := &spyT{}
	AssertMissingField(spy, a, atom.TableStrings, "Name")
	if !spy.hasError() {
		t.Error("expected error when String field present")
	}

	// Int present
	spy = &spyT{}
	AssertMissingField(spy, a, atom.TableInts, "Age")
	if !spy.hasError() {
		t.Error("expected error when Int field present")
	}

	// Float present
	spy = &spyT{}
	AssertMissingField(spy, a, atom.TableFloats, "Score")
	if !spy.hasError() {
		t.Error("expected error when Float field present")
	}

	// Bool present
	spy = &spyT{}
	AssertMissingField(spy, a, atom.TableBools, "Active")
	if !spy.hasError() {
		t.Error("expected error when Bool field present")
	}
}
