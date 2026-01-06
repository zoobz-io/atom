package atom

import (
	"reflect"
	"testing"
	"time"
)

func TestScalarToTable(t *testing.T) {
	tests := []struct {
		name      string
		typ       reflect.Type
		wantTable Table
		wantOK    bool
	}{
		{"string", reflect.TypeFor[string](), TableStrings, true},
		{"int", reflect.TypeFor[int](), TableInts, true},
		{"int8", reflect.TypeFor[int8](), TableInts, true},
		{"int16", reflect.TypeFor[int16](), TableInts, true},
		{"int32", reflect.TypeFor[int32](), TableInts, true},
		{"int64", reflect.TypeFor[int64](), TableInts, true},
		{"uint", reflect.TypeFor[uint](), TableUints, true},
		{"uint8", reflect.TypeFor[uint8](), TableUints, true},
		{"uint16", reflect.TypeFor[uint16](), TableUints, true},
		{"uint32", reflect.TypeFor[uint32](), TableUints, true},
		{"uint64", reflect.TypeFor[uint64](), TableUints, true},
		{"float32", reflect.TypeFor[float32](), TableFloats, true},
		{"float64", reflect.TypeFor[float64](), TableFloats, true},
		{"bool", reflect.TypeFor[bool](), TableBools, true},
		{"time.Time", reflect.TypeFor[time.Time](), TableTimes, true},
		{"unsupported", reflect.TypeFor[complex64](), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, _, ok := scalarToTable(tt.typ)
			if ok != tt.wantOK {
				t.Errorf("got ok=%v, want ok=%v", ok, tt.wantOK)
			}
			if table != tt.wantTable {
				t.Errorf("got table=%v, want table=%v", table, tt.wantTable)
			}
		})
	}
}

func TestPointerTable(t *testing.T) {
	tests := []struct {
		base Table
		want Table
	}{
		{TableStrings, TableStringPtrs},
		{TableInts, TableIntPtrs},
		{TableUints, TableUintPtrs},
		{TableFloats, TableFloatPtrs},
		{TableBools, TableBoolPtrs},
		{TableTimes, TableTimePtrs},
		{TableBytes, TableBytePtrs},
	}

	for _, tt := range tests {
		t.Run(string(tt.base), func(t *testing.T) {
			got := pointerTable(tt.base)
			if got != tt.want {
				t.Errorf("pointerTable(%v) = %v, want %v", tt.base, got, tt.want)
			}
		})
	}
}

func TestSliceTable(t *testing.T) {
	tests := []struct {
		base Table
		want Table
	}{
		{TableStrings, TableStringSlices},
		{TableInts, TableIntSlices},
		{TableUints, TableUintSlices},
		{TableFloats, TableFloatSlices},
		{TableBools, TableBoolSlices},
		{TableTimes, TableTimeSlices},
		{TableBytes, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.base), func(t *testing.T) {
			got := sliceTable(tt.base)
			if got != tt.want {
				t.Errorf("sliceTable(%v) = %v, want %v", tt.base, got, tt.want)
			}
		})
	}
}

func TestBuildFieldPlanScalars(t *testing.T) {
	type Scalars struct {
		T  time.Time
		S  string
		Bs []byte
		I  int64
		F  float64
		B  bool
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Scalars]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 6 {
		t.Errorf("got %d plans, want 6", len(plans))
	}

	expected := map[string]Table{
		"S":  TableStrings,
		"I":  TableInts,
		"F":  TableFloats,
		"B":  TableBools,
		"T":  TableTimes,
		"Bs": TableBytes,
	}

	for _, fp := range plans {
		if want, ok := expected[fp.name]; ok {
			if fp.table != want {
				t.Errorf("field %s: got table %v, want %v", fp.name, fp.table, want)
			}
		}
	}
}

func TestBuildFieldPlanPointers(t *testing.T) {
	type Pointers struct {
		S *string
		I *int64
		F *float64
		B *bool
		T *time.Time
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Pointers]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 5 {
		t.Errorf("got %d plans, want 5", len(plans))
	}

	for _, fp := range plans {
		if fp.kind != kindPointer {
			t.Errorf("field %s: got kind %v, want kindPointer", fp.name, fp.kind)
		}
	}
}

func TestBuildFieldPlanSlices(t *testing.T) {
	type Slices struct {
		S  []string
		I  []int64
		F  []float64
		B  []bool
		T  []time.Time
		Bs [][]byte
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Slices]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 6 {
		t.Errorf("got %d plans, want 6", len(plans))
	}

	expected := map[string]Table{
		"S":  TableStringSlices,
		"I":  TableIntSlices,
		"F":  TableFloatSlices,
		"B":  TableBoolSlices,
		"T":  TableTimeSlices,
		"Bs": TableByteSlices,
	}

	for _, fp := range plans {
		if want, ok := expected[fp.name]; ok {
			if fp.table != want {
				t.Errorf("field %s: got table %v, want %v", fp.name, fp.table, want)
			}
			if fp.kind != kindSlice {
				t.Errorf("field %s: got kind %v, want kindSlice", fp.name, fp.kind)
			}
		}
	}
}

func TestBuildFieldPlanNested(t *testing.T) {
	type Inner struct {
		X int
	}
	type Outer struct {
		N Inner
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Outer]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 1 {
		t.Fatalf("got %d plans, want 1", len(plans))
	}

	fp := plans[0]
	if fp.kind != kindNested {
		t.Errorf("got kind %v, want kindNested", fp.kind)
	}
	if fp.nested == nil {
		t.Error("nested atomizer should not be nil")
	}
}

func TestBuildFieldPlanNestedPtr(t *testing.T) {
	type Inner struct {
		X int
	}
	type Outer struct {
		N *Inner
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Outer]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 1 {
		t.Fatalf("got %d plans, want 1", len(plans))
	}

	fp := plans[0]
	if fp.kind != kindNestedPtr {
		t.Errorf("got kind %v, want kindNestedPtr", fp.kind)
	}
}

func TestBuildFieldPlanNestedSlice(t *testing.T) {
	type Inner struct {
		X int
	}
	type Outer struct {
		Items []Inner
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Outer]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 1 {
		t.Fatalf("got %d plans, want 1", len(plans))
	}

	fp := plans[0]
	if fp.kind != kindNestedSlice {
		t.Errorf("got kind %v, want kindNestedSlice", fp.kind)
	}
}

func TestBuildFieldPlanNestedPtrSlice(t *testing.T) {
	type Inner struct {
		X int
	}
	type Outer struct {
		Items []*Inner
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Outer]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 1 {
		t.Fatalf("got %d plans, want 1", len(plans))
	}

	fp := plans[0]
	if fp.kind != kindNestedSlice {
		t.Errorf("got kind %v, want kindNestedSlice", fp.kind)
	}
}

func TestBuildFieldPlanUnexported(t *testing.T) {
	type Mixed struct {
		Public  string
		private string //nolint:unused
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Mixed]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 1 {
		t.Errorf("got %d plans, want 1 (unexported should be skipped)", len(plans))
	}

	if plans[0].name != "Public" {
		t.Errorf("got name %q, want %q", plans[0].name, "Public")
	}
}

func TestBuildFieldPlanPointerType(t *testing.T) {
	type Simple struct {
		X int
	}

	// Test that pointer to struct works
	plans, err := buildFieldPlan(reflect.TypeFor[*Simple]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	if len(plans) != 1 {
		t.Errorf("got %d plans, want 1", len(plans))
	}
}

func TestBuildFieldPlanSupportedMap(t *testing.T) {
	type WithMap struct {
		M map[string]string
	}

	plans, err := buildFieldPlan(reflect.TypeFor[WithMap]())
	if err != nil {
		t.Errorf("unexpected error for supported map field: %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("got %d plans, want 1", len(plans))
	}
	if plans[0].kind != kindMap {
		t.Errorf("got kind %d, want kindMap (%d)", plans[0].kind, kindMap)
	}
	if plans[0].table != TableStringMaps {
		t.Errorf("got table %s, want TableStringMaps", plans[0].table)
	}
}

func TestBuildFieldPlanUnsupportedMapKey(t *testing.T) {
	type WithIntKeyMap struct {
		M map[int]string
	}

	_, err := buildFieldPlan(reflect.TypeFor[WithIntKeyMap]())
	if err == nil {
		t.Error("expected error for non-string map key")
	}
}

func TestBuildFieldPlanUnsupportedChannel(t *testing.T) {
	type WithChan struct {
		C chan int
	}

	_, err := buildFieldPlan(reflect.TypeFor[WithChan]())
	if err == nil {
		t.Error("expected error for channel field")
	}
}

func TestBuildFieldPlanUnsupportedFunc(t *testing.T) {
	type WithFunc struct {
		F func()
	}

	_, err := buildFieldPlan(reflect.TypeFor[WithFunc]())
	if err == nil {
		t.Error("expected error for func field")
	}
}

func TestBuildFieldPlanUnsupportedSliceElement(t *testing.T) {
	type WithBadSlice struct {
		S []chan int
	}

	_, err := buildFieldPlan(reflect.TypeFor[WithBadSlice]())
	if err == nil {
		t.Error("expected error for slice of channels")
	}
}

func TestBuildFieldPlanUnsupportedPointerElement(t *testing.T) {
	type WithBadPtr struct {
		P *chan int
	}

	_, err := buildFieldPlan(reflect.TypeFor[WithBadPtr]())
	if err == nil {
		t.Error("expected error for pointer to channel")
	}
}

func TestPlanFieldIndex(t *testing.T) {
	type Simple struct {
		A string
		B int
		C bool
	}

	plans, err := buildFieldPlan(reflect.TypeFor[Simple]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	for i, fp := range plans {
		if len(fp.index) != 1 || fp.index[0] != i {
			t.Errorf("field %s: got index %v, want [%d]", fp.name, fp.index, i)
		}
	}
}

func TestPlanFieldConverterAssignment(t *testing.T) {
	type WithInts struct {
		I8  int8
		I16 int16
		I32 int32
		I64 int64
	}

	plans, err := buildFieldPlan(reflect.TypeFor[WithInts]())
	if err != nil {
		t.Fatalf("buildFieldPlan error: %v", err)
	}

	for _, fp := range plans {
		if fp.converter.fromInt64 == nil {
			t.Errorf("field %s: converter.fromInt64 should not be nil", fp.name)
		}
		if fp.converter.toInt64 == nil {
			t.Errorf("field %s: converter.toInt64 should not be nil", fp.name)
		}
	}
}
