package atom

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestIntConverter(t *testing.T) {
	tests := []struct {
		name    string
		typ     reflect.Type
		input   int64
		wantErr bool
	}{
		{"int8 valid", reflect.TypeFor[int8](), 100, false},
		{"int8 overflow positive", reflect.TypeFor[int8](), 200, true},
		{"int8 overflow negative", reflect.TypeFor[int8](), -200, true},
		{"int16 valid", reflect.TypeFor[int16](), 1000, false},
		{"int16 overflow", reflect.TypeFor[int16](), 40000, true},
		{"int32 valid", reflect.TypeFor[int32](), 100000, false},
		{"int32 overflow", reflect.TypeFor[int32](), 3000000000, true},
		{"int64 valid", reflect.TypeFor[int64](), math.MaxInt64, false},
		{"int valid", reflect.TypeFor[int](), 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := intConverter(tt.typ)
			rv, err := conv.fromInt64(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %d", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if rv.Int() != tt.input {
					t.Errorf("got %d, want %d", rv.Int(), tt.input)
				}
			}
		})
	}
}

func TestIntConverterRoundTrip(t *testing.T) {
	conv := intConverter(reflect.TypeFor[int32]())
	original := int64(42)

	rv := reflect.New(reflect.TypeFor[int32]()).Elem()
	rv.SetInt(original)

	got := conv.toInt64(rv)
	if got != original {
		t.Errorf("toInt64: got %d, want %d", got, original)
	}
}

func TestUintConverter(t *testing.T) {
	tests := []struct {
		name    string
		typ     reflect.Type
		input   uint64
		wantErr bool
	}{
		{"uint8 valid", reflect.TypeFor[uint8](), 200, false},
		{"uint8 overflow", reflect.TypeFor[uint8](), 300, true},
		{"uint16 valid", reflect.TypeFor[uint16](), 60000, false},
		{"uint16 overflow", reflect.TypeFor[uint16](), 70000, true},
		{"uint32 valid", reflect.TypeFor[uint32](), 3000000000, false},
		{"uint32 overflow", reflect.TypeFor[uint32](), 5000000000, true},
		{"uint64 valid", reflect.TypeFor[uint64](), math.MaxUint64, false},
		{"uint valid", reflect.TypeFor[uint](), 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := uintConverter(tt.typ)
			rv, err := conv.fromUint64(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %d", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if rv.Uint() != tt.input {
					t.Errorf("got %d, want %d", rv.Uint(), tt.input)
				}
			}
		})
	}
}

func TestUintConverterRoundTrip(t *testing.T) {
	conv := uintConverter(reflect.TypeFor[uint32]())
	original := uint64(42)

	rv := reflect.New(reflect.TypeFor[uint32]()).Elem()
	rv.SetUint(original)

	got := conv.toUint64(rv)
	if got != original {
		t.Errorf("toUint64: got %d, want %d", got, original)
	}
}

func TestFloatConverter(t *testing.T) {
	tests := []struct {
		name    string
		typ     reflect.Type
		input   float64
		wantErr bool
	}{
		{"float32 valid", reflect.TypeFor[float32](), 100.5, false},
		{"float32 overflow positive", reflect.TypeFor[float32](), math.MaxFloat64, true},
		{"float32 overflow negative", reflect.TypeFor[float32](), -math.MaxFloat64, true},
		{"float32 inf", reflect.TypeFor[float32](), math.Inf(1), false},
		{"float32 nan", reflect.TypeFor[float32](), math.NaN(), false},
		{"float64 valid", reflect.TypeFor[float64](), math.MaxFloat64, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := floatConverter(tt.typ)
			rv, err := conv.fromFloat64(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %g", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !math.IsNaN(tt.input) && rv.Float() != tt.input {
					t.Errorf("got %g, want %g", rv.Float(), tt.input)
				}
			}
		})
	}
}

func TestFloatConverterRoundTrip(t *testing.T) {
	conv := floatConverter(reflect.TypeFor[float64]())
	original := 3.14159

	rv := reflect.New(reflect.TypeFor[float64]()).Elem()
	rv.SetFloat(original)

	got := conv.toFloat64(rv)
	if got != original {
		t.Errorf("toFloat64: got %g, want %g", got, original)
	}
}

func TestAtomizeScalar(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		fp    fieldPlan
		value any
		check func(*Atom) bool
	}{
		{
			name:  "string",
			fp:    fieldPlan{name: "F", table: TableStrings},
			value: "hello",
			check: func(a *Atom) bool { return a.Strings["F"] == "hello" },
		},
		{
			name:  "int64",
			fp:    fieldPlan{name: "F", table: TableInts, converter: intConverter(reflect.TypeFor[int64]())},
			value: int64(42),
			check: func(a *Atom) bool { return a.Ints["F"] == 42 },
		},
		{
			name:  "float64",
			fp:    fieldPlan{name: "F", table: TableFloats, converter: floatConverter(reflect.TypeFor[float64]())},
			value: float64(3.14),
			check: func(a *Atom) bool { return a.Floats["F"] == 3.14 },
		},
		{
			name:  "bool",
			fp:    fieldPlan{name: "F", table: TableBools},
			value: true,
			check: func(a *Atom) bool { return a.Bools["F"] == true },
		},
		{
			name:  "time",
			fp:    fieldPlan{name: "F", table: TableTimes},
			value: now,
			check: func(a *Atom) bool { return a.Times["F"].Equal(now) },
		},
		{
			name:  "bytes",
			fp:    fieldPlan{name: "F", table: TableBytes},
			value: []byte("data"),
			check: func(a *Atom) bool { return string(a.Bytes["F"]) == "data" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := &Atom{
				Strings: make(map[string]string),
				Ints:    make(map[string]int64),
				Floats:  make(map[string]float64),
				Bools:   make(map[string]bool),
				Times:   make(map[string]time.Time),
				Bytes:   make(map[string][]byte),
			}
			fv := reflect.ValueOf(tt.value)
			atomizeScalar(&tt.fp, fv, atom)
			if !tt.check(atom) {
				t.Errorf("check failed for %s", tt.name)
			}
		})
	}
}

func TestAtomizePointerNil(t *testing.T) {
	atom := &Atom{
		StringPtrs: make(map[string]*string),
		IntPtrs:    make(map[string]*int64),
		FloatPtrs:  make(map[string]*float64),
		BoolPtrs:   make(map[string]*bool),
		TimePtrs:   make(map[string]*time.Time),
	}

	tables := []Table{TableStringPtrs, TableIntPtrs, TableFloatPtrs, TableBoolPtrs, TableTimePtrs}
	for _, table := range tables {
		fp := fieldPlan{name: "F", table: table, kind: kindPointer}
		var nilPtr *string
		fv := reflect.ValueOf(nilPtr)
		atomizePointer(&fp, fv, atom)
	}

	// Verify nil was stored
	if _, ok := atom.StringPtrs["F"]; !ok {
		t.Error("StringPtrs should have nil entry")
	}
}

func TestAtomizeSliceNil(t *testing.T) {
	atom := &Atom{
		StringSlices: make(map[string][]string),
	}

	fp := fieldPlan{name: "F", table: TableStringSlices, kind: kindSlice}
	var nilSlice []string
	fv := reflect.ValueOf(nilSlice)
	atomizeSlice(&fp, fv, atom)

	if _, ok := atom.StringSlices["F"]; ok {
		t.Error("nil slice should not create entry")
	}
}

func TestAtomizeSliceTypes(t *testing.T) {
	tests := []struct {
		name  string
		fp    fieldPlan
		value any
		check func(*Atom) bool
	}{
		{
			name:  "string slice",
			fp:    fieldPlan{name: "F", table: TableStringSlices, kind: kindSlice},
			value: []string{"a", "b"},
			check: func(a *Atom) bool { return len(a.StringSlices["F"]) == 2 },
		},
		{
			name:  "int slice",
			fp:    fieldPlan{name: "F", table: TableIntSlices, kind: kindSlice, converter: intConverter(reflect.TypeFor[int64]())},
			value: []int64{1, 2, 3},
			check: func(a *Atom) bool { return len(a.IntSlices["F"]) == 3 },
		},
		{
			name:  "float slice",
			fp:    fieldPlan{name: "F", table: TableFloatSlices, kind: kindSlice, converter: floatConverter(reflect.TypeFor[float64]())},
			value: []float64{1.1, 2.2},
			check: func(a *Atom) bool { return len(a.FloatSlices["F"]) == 2 },
		},
		{
			name:  "bool slice",
			fp:    fieldPlan{name: "F", table: TableBoolSlices, kind: kindSlice},
			value: []bool{true, false},
			check: func(a *Atom) bool { return len(a.BoolSlices["F"]) == 2 },
		},
		{
			name:  "time slice",
			fp:    fieldPlan{name: "F", table: TableTimeSlices, kind: kindSlice},
			value: []time.Time{time.Now(), time.Now()},
			check: func(a *Atom) bool { return len(a.TimeSlices["F"]) == 2 },
		},
		{
			name:  "byte slices",
			fp:    fieldPlan{name: "F", table: TableByteSlices, kind: kindSlice},
			value: [][]byte{[]byte("a"), []byte("b")},
			check: func(a *Atom) bool { return len(a.ByteSlices["F"]) == 2 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := &Atom{
				StringSlices: make(map[string][]string),
				IntSlices:    make(map[string][]int64),
				FloatSlices:  make(map[string][]float64),
				BoolSlices:   make(map[string][]bool),
				TimeSlices:   make(map[string][]time.Time),
				ByteSlices:   make(map[string][][]byte),
			}
			fv := reflect.ValueOf(tt.value)
			atomizeSlice(&tt.fp, fv, atom)
			if !tt.check(atom) {
				t.Errorf("check failed for %s", tt.name)
			}
		})
	}
}

func TestDeatomizeScalar(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		fp    fieldPlan
		atom  *Atom
		check func(reflect.Value) bool
	}{
		{
			name: "string",
			fp:   fieldPlan{name: "F", table: TableStrings},
			atom: &Atom{Strings: map[string]string{"F": "hello"}},
			check: func(v reflect.Value) bool {
				return v.String() == "hello"
			},
		},
		{
			name: "int64",
			fp:   fieldPlan{name: "F", table: TableInts, converter: intConverter(reflect.TypeFor[int64]())},
			atom: &Atom{Ints: map[string]int64{"F": 42}},
			check: func(v reflect.Value) bool {
				return v.Int() == 42
			},
		},
		{
			name: "float64",
			fp:   fieldPlan{name: "F", table: TableFloats, converter: floatConverter(reflect.TypeFor[float64]())},
			atom: &Atom{Floats: map[string]float64{"F": 3.14}},
			check: func(v reflect.Value) bool {
				return v.Float() == 3.14
			},
		},
		{
			name: "bool",
			fp:   fieldPlan{name: "F", table: TableBools},
			atom: &Atom{Bools: map[string]bool{"F": true}},
			check: func(v reflect.Value) bool {
				return v.Bool() == true
			},
		},
		{
			name: "time",
			fp:   fieldPlan{name: "F", table: TableTimes},
			atom: &Atom{Times: map[string]time.Time{"F": now}},
			check: func(v reflect.Value) bool {
				tv, ok := v.Interface().(time.Time)
				return ok && tv.Equal(now)
			},
		},
		{
			name: "bytes",
			fp:   fieldPlan{name: "F", table: TableBytes},
			atom: &Atom{Bytes: map[string][]byte{"F": []byte("data")}},
			check: func(v reflect.Value) bool {
				return string(v.Bytes()) == "data"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create appropriate zero value
			var fv reflect.Value
			switch tt.fp.table {
			case TableStrings:
				fv = reflect.New(reflect.TypeFor[string]()).Elem()
			case TableInts:
				fv = reflect.New(reflect.TypeFor[int64]()).Elem()
			case TableFloats:
				fv = reflect.New(reflect.TypeFor[float64]()).Elem()
			case TableBools:
				fv = reflect.New(reflect.TypeFor[bool]()).Elem()
			case TableTimes:
				fv = reflect.New(reflect.TypeFor[time.Time]()).Elem()
			case TableBytes:
				fv = reflect.New(reflect.TypeFor[[]byte]()).Elem()
			}

			if err := deatomizeScalar(&tt.fp, tt.atom, fv); err != nil {
				t.Fatalf("deatomizeScalar error: %v", err)
			}
			if !tt.check(fv) {
				t.Errorf("check failed for %s", tt.name)
			}
		})
	}
}

func TestDeatomizeScalarMissingKey(t *testing.T) {
	atom := &Atom{Strings: make(map[string]string)}
	fp := fieldPlan{name: "Missing", table: TableStrings}
	fv := reflect.New(reflect.TypeFor[string]()).Elem()
	fv.SetString("original")

	if err := deatomizeScalar(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizeScalar error: %v", err)
	}

	// Value should be unchanged
	if fv.String() != "original" {
		t.Errorf("value should remain unchanged when key missing")
	}
}

func TestDeatomizeScalarOverflow(t *testing.T) {
	atom := &Atom{Ints: map[string]int64{"F": 1000}}
	fp := fieldPlan{name: "F", table: TableInts, converter: intConverter(reflect.TypeFor[int8]())}
	fv := reflect.New(reflect.TypeFor[int8]()).Elem()

	err := deatomizeScalar(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizePointerNil(t *testing.T) {
	atom := &Atom{StringPtrs: map[string]*string{"F": nil}}
	fp := fieldPlan{name: "F", table: TableStringPtrs, kind: kindPointer}
	fv := reflect.New(reflect.TypeFor[*string]()).Elem()

	if err := deatomizePointer(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizePointer error: %v", err)
	}

	if !fv.IsNil() {
		t.Error("expected nil pointer")
	}
}

func TestDeatomizePointerValue(t *testing.T) {
	val := "hello"
	atom := &Atom{StringPtrs: map[string]*string{"F": &val}}
	fp := fieldPlan{name: "F", table: TableStringPtrs, kind: kindPointer}
	fv := reflect.New(reflect.TypeFor[*string]()).Elem()

	if err := deatomizePointer(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizePointer error: %v", err)
	}

	if fv.IsNil() {
		t.Fatal("expected non-nil pointer")
	}
	if fv.Elem().String() != "hello" {
		t.Errorf("got %q, want %q", fv.Elem().String(), "hello")
	}
}

func TestDeatomizePointerOverflow(t *testing.T) {
	val := int64(1000)
	atom := &Atom{IntPtrs: map[string]*int64{"F": &val}}
	fp := fieldPlan{
		name:      "F",
		table:     TableIntPtrs,
		kind:      kindPointer,
		converter: intConverter(reflect.TypeFor[int8]()),
	}
	fv := reflect.New(reflect.TypeFor[*int8]()).Elem()

	err := deatomizePointer(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizeSliceOverflow(t *testing.T) {
	atom := &Atom{IntSlices: map[string][]int64{"F": {1, 2, 1000}}}
	fp := fieldPlan{
		name:      "F",
		table:     TableIntSlices,
		kind:      kindSlice,
		converter: intConverter(reflect.TypeFor[int8]()),
		elemType:  reflect.TypeFor[int8](),
	}
	fv := reflect.New(reflect.TypeFor[[]int8]()).Elem()

	err := deatomizeSlice(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizeSliceTypes(t *testing.T) {
	tests := []struct {
		name  string
		fp    fieldPlan
		atom  *Atom
		fvTyp reflect.Type
		check func(reflect.Value) bool
	}{
		{
			name:  "string slice",
			fp:    fieldPlan{name: "F", table: TableStringSlices, kind: kindSlice},
			atom:  &Atom{StringSlices: map[string][]string{"F": {"a", "b"}}},
			fvTyp: reflect.TypeFor[[]string](),
			check: func(v reflect.Value) bool { return v.Len() == 2 },
		},
		{
			name:  "bool slice",
			fp:    fieldPlan{name: "F", table: TableBoolSlices, kind: kindSlice},
			atom:  &Atom{BoolSlices: map[string][]bool{"F": {true, false}}},
			fvTyp: reflect.TypeFor[[]bool](),
			check: func(v reflect.Value) bool { return v.Len() == 2 },
		},
		{
			name:  "time slice",
			fp:    fieldPlan{name: "F", table: TableTimeSlices, kind: kindSlice},
			atom:  &Atom{TimeSlices: map[string][]time.Time{"F": {time.Now()}}},
			fvTyp: reflect.TypeFor[[]time.Time](),
			check: func(v reflect.Value) bool { return v.Len() == 1 },
		},
		{
			name:  "byte slices",
			fp:    fieldPlan{name: "F", table: TableByteSlices, kind: kindSlice},
			atom:  &Atom{ByteSlices: map[string][][]byte{"F": {[]byte("a")}}},
			fvTyp: reflect.TypeFor[[][]byte](),
			check: func(v reflect.Value) bool { return v.Len() == 1 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := reflect.New(tt.fvTyp).Elem()
			if err := deatomizeSlice(&tt.fp, tt.atom, fv); err != nil {
				t.Fatalf("deatomizeSlice error: %v", err)
			}
			if !tt.check(fv) {
				t.Errorf("check failed for %s", tt.name)
			}
		})
	}
}

func TestAtomizeFieldDispatch(t *testing.T) {
	// Test that atomizeField dispatches to correct handler
	atom := &Atom{Strings: make(map[string]string)}
	fp := fieldPlan{name: "F", table: TableStrings, kind: kindScalar}
	fv := reflect.ValueOf("test")

	atomizeField(&fp, fv, atom)

	if atom.Strings["F"] != "test" {
		t.Errorf("got %q, want %q", atom.Strings["F"], "test")
	}
}

func TestDeatomizeFieldDispatch(t *testing.T) {
	// Test that deatomizeField dispatches to correct handler
	atom := &Atom{Strings: map[string]string{"F": "test"}}
	fp := fieldPlan{name: "F", table: TableStrings, kind: kindScalar}
	fv := reflect.New(reflect.TypeFor[string]()).Elem()

	if err := deatomizeField(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizeField error: %v", err)
	}

	if fv.String() != "test" {
		t.Errorf("got %q, want %q", fv.String(), "test")
	}
}

func TestAtomizeScalarUint(t *testing.T) {
	atom := &Atom{Uints: make(map[string]uint64)}
	fp := fieldPlan{name: "F", table: TableUints, converter: uintConverter(reflect.TypeFor[uint64]())}
	fv := reflect.ValueOf(uint64(42))

	atomizeScalar(&fp, fv, atom)

	if atom.Uints["F"] != 42 {
		t.Errorf("got %d, want %d", atom.Uints["F"], 42)
	}
}

func TestAtomizeScalarNilBytes(t *testing.T) {
	atom := &Atom{Bytes: make(map[string][]byte)}
	fp := fieldPlan{name: "F", table: TableBytes}

	var nilBytes []byte
	fv := reflect.ValueOf(nilBytes)
	atomizeScalar(&fp, fv, atom)

	if _, ok := atom.Bytes["F"]; ok {
		t.Error("nil bytes should not create entry")
	}
}

func TestAtomizeScalarByteArray(t *testing.T) {
	atom := &Atom{Bytes: make(map[string][]byte)}
	fp := fieldPlan{name: "F", table: TableBytes}

	arr := [4]byte{1, 2, 3, 4}
	fv := reflect.ValueOf(arr)
	atomizeScalar(&fp, fv, atom)

	if len(atom.Bytes["F"]) != 4 {
		t.Errorf("got length %d, want 4", len(atom.Bytes["F"]))
	}
	for i, b := range atom.Bytes["F"] {
		if b != byte(i+1) {
			t.Errorf("byte[%d]: got %d, want %d", i, b, i+1)
		}
	}
}

func TestAtomizePointerAllTypes(t *testing.T) {
	now := time.Now()
	str := "hello"
	i64 := int64(42)
	u64 := uint64(100)
	f64 := float64(3.14)
	b := true
	bytes := []byte("data")

	tests := []struct {
		name  string
		fp    fieldPlan
		value any
		check func(*Atom) bool
	}{
		{
			name:  "string ptr",
			fp:    fieldPlan{name: "F", table: TableStringPtrs, kind: kindPointer},
			value: &str,
			check: func(a *Atom) bool { return a.StringPtrs["F"] != nil && *a.StringPtrs["F"] == "hello" },
		},
		{
			name:  "int ptr",
			fp:    fieldPlan{name: "F", table: TableIntPtrs, kind: kindPointer, converter: intConverter(reflect.TypeFor[int64]())},
			value: &i64,
			check: func(a *Atom) bool { return a.IntPtrs["F"] != nil && *a.IntPtrs["F"] == 42 },
		},
		{
			name:  "uint ptr",
			fp:    fieldPlan{name: "F", table: TableUintPtrs, kind: kindPointer, converter: uintConverter(reflect.TypeFor[uint64]())},
			value: &u64,
			check: func(a *Atom) bool { return a.UintPtrs["F"] != nil && *a.UintPtrs["F"] == 100 },
		},
		{
			name:  "float ptr",
			fp:    fieldPlan{name: "F", table: TableFloatPtrs, kind: kindPointer, converter: floatConverter(reflect.TypeFor[float64]())},
			value: &f64,
			check: func(a *Atom) bool { return a.FloatPtrs["F"] != nil && *a.FloatPtrs["F"] == 3.14 },
		},
		{
			name:  "bool ptr",
			fp:    fieldPlan{name: "F", table: TableBoolPtrs, kind: kindPointer},
			value: &b,
			check: func(a *Atom) bool { return a.BoolPtrs["F"] != nil && *a.BoolPtrs["F"] == true },
		},
		{
			name:  "time ptr",
			fp:    fieldPlan{name: "F", table: TableTimePtrs, kind: kindPointer},
			value: &now,
			check: func(a *Atom) bool { return a.TimePtrs["F"] != nil && a.TimePtrs["F"].Equal(now) },
		},
		{
			name:  "bytes ptr",
			fp:    fieldPlan{name: "F", table: TableBytePtrs, kind: kindPointer},
			value: &bytes,
			check: func(a *Atom) bool { return a.BytePtrs["F"] != nil && string(*a.BytePtrs["F"]) == "data" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			atom := &Atom{
				StringPtrs: make(map[string]*string),
				IntPtrs:    make(map[string]*int64),
				UintPtrs:   make(map[string]*uint64),
				FloatPtrs:  make(map[string]*float64),
				BoolPtrs:   make(map[string]*bool),
				TimePtrs:   make(map[string]*time.Time),
				BytePtrs:   make(map[string]*[]byte),
			}
			fv := reflect.ValueOf(tt.value)
			atomizePointer(&tt.fp, fv, atom)
			if !tt.check(atom) {
				t.Errorf("check failed for %s", tt.name)
			}
		})
	}
}

func TestAtomizePointerNilAllTypes(t *testing.T) {
	atom := &Atom{
		StringPtrs: make(map[string]*string),
		IntPtrs:    make(map[string]*int64),
		UintPtrs:   make(map[string]*uint64),
		FloatPtrs:  make(map[string]*float64),
		BoolPtrs:   make(map[string]*bool),
		TimePtrs:   make(map[string]*time.Time),
		BytePtrs:   make(map[string]*[]byte),
	}

	tests := []struct {
		name  string
		table Table
		check func() bool
	}{
		{"uint ptr", TableUintPtrs, func() bool { _, ok := atom.UintPtrs["F"]; return ok }},
		{"byte ptr", TableBytePtrs, func() bool { _, ok := atom.BytePtrs["F"]; return ok }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := fieldPlan{name: "F", table: tt.table, kind: kindPointer}
			fv := reflect.Zero(reflect.PointerTo(reflect.TypeFor[string]()))
			atomizePointer(&fp, fv, atom)

			if !tt.check() {
				t.Errorf("nil pointer should create entry for %s", tt.name)
			}
		})
	}
}

func TestAtomizeSliceUint(t *testing.T) {
	atom := &Atom{UintSlices: make(map[string][]uint64)}
	fp := fieldPlan{
		name:      "F",
		table:     TableUintSlices,
		kind:      kindSlice,
		converter: uintConverter(reflect.TypeFor[uint64]()),
	}
	fv := reflect.ValueOf([]uint64{1, 2, 3})

	atomizeSlice(&fp, fv, atom)

	if len(atom.UintSlices["F"]) != 3 {
		t.Errorf("got length %d, want 3", len(atom.UintSlices["F"]))
	}
}

func TestDeatomizeScalarUint(t *testing.T) {
	atom := &Atom{Uints: map[string]uint64{"F": 42}}
	fp := fieldPlan{name: "F", table: TableUints, converter: uintConverter(reflect.TypeFor[uint64]())}
	fv := reflect.New(reflect.TypeFor[uint64]()).Elem()

	if err := deatomizeScalar(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizeScalar error: %v", err)
	}

	if fv.Uint() != 42 {
		t.Errorf("got %d, want 42", fv.Uint())
	}
}

func TestDeatomizeScalarUintOverflow(t *testing.T) {
	atom := &Atom{Uints: map[string]uint64{"F": 1000}}
	fp := fieldPlan{name: "F", table: TableUints, converter: uintConverter(reflect.TypeFor[uint8]())}
	fv := reflect.New(reflect.TypeFor[uint8]()).Elem()

	err := deatomizeScalar(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizeScalarFloatOverflow(t *testing.T) {
	atom := &Atom{Floats: map[string]float64{"F": math.MaxFloat64}}
	fp := fieldPlan{name: "F", table: TableFloats, converter: floatConverter(reflect.TypeFor[float32]())}
	fv := reflect.New(reflect.TypeFor[float32]()).Elem()

	err := deatomizeScalar(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizeScalarByteArray(t *testing.T) {
	atom := &Atom{Bytes: map[string][]byte{"F": {1, 2, 3, 4}}}
	fp := fieldPlan{name: "F", table: TableBytes}

	// Create a [4]byte array
	arrType := reflect.ArrayOf(4, reflect.TypeFor[byte]())
	fv := reflect.New(arrType).Elem()

	if err := deatomizeScalar(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizeScalar error: %v", err)
	}

	expected := []uint64{1, 2, 3, 4}
	for i, want := range expected {
		if got := fv.Index(i).Uint(); got != want {
			t.Errorf("byte[%d]: got %d, want %d", i, got, want)
		}
	}
}

func TestDeatomizeScalarByteArrayLengthMismatch(t *testing.T) {
	atom := &Atom{Bytes: map[string][]byte{"F": {1, 2, 3}}}
	fp := fieldPlan{name: "F", table: TableBytes}

	// Create a [4]byte array (length mismatch)
	arrType := reflect.ArrayOf(4, reflect.TypeFor[byte]())
	fv := reflect.New(arrType).Elem()

	err := deatomizeScalar(&fp, atom, fv)
	if err == nil {
		t.Error("expected length mismatch error")
	}
}

func TestDeatomizePointerAllTypes(t *testing.T) {
	now := time.Now()
	str := "hello"
	i64 := int64(42)
	u64 := uint64(100)
	f64 := float64(3.14)
	b := true
	bytes := []byte("data")

	tests := []struct {
		name  string
		fp    fieldPlan
		atom  *Atom
		fvTyp reflect.Type
		check func(reflect.Value) bool
	}{
		{
			name:  "uint ptr",
			fp:    fieldPlan{name: "F", table: TableUintPtrs, kind: kindPointer, converter: uintConverter(reflect.TypeFor[uint64]())},
			atom:  &Atom{UintPtrs: map[string]*uint64{"F": &u64}},
			fvTyp: reflect.TypeFor[*uint64](),
			check: func(v reflect.Value) bool { return !v.IsNil() && v.Elem().Uint() == 100 },
		},
		{
			name:  "float ptr",
			fp:    fieldPlan{name: "F", table: TableFloatPtrs, kind: kindPointer, converter: floatConverter(reflect.TypeFor[float64]())},
			atom:  &Atom{FloatPtrs: map[string]*float64{"F": &f64}},
			fvTyp: reflect.TypeFor[*float64](),
			check: func(v reflect.Value) bool { return !v.IsNil() && v.Elem().Float() == 3.14 },
		},
		{
			name:  "bool ptr",
			fp:    fieldPlan{name: "F", table: TableBoolPtrs, kind: kindPointer},
			atom:  &Atom{BoolPtrs: map[string]*bool{"F": &b}},
			fvTyp: reflect.TypeFor[*bool](),
			check: func(v reflect.Value) bool { return !v.IsNil() && v.Elem().Bool() == true },
		},
		{
			name:  "time ptr",
			fp:    fieldPlan{name: "F", table: TableTimePtrs, kind: kindPointer},
			atom:  &Atom{TimePtrs: map[string]*time.Time{"F": &now}},
			fvTyp: reflect.TypeFor[*time.Time](),
			check: func(v reflect.Value) bool {
				if v.IsNil() {
					return false
				}
				t, ok := v.Elem().Interface().(time.Time)
				return ok && t.Equal(now)
			},
		},
		{
			name:  "bytes ptr",
			fp:    fieldPlan{name: "F", table: TableBytePtrs, kind: kindPointer},
			atom:  &Atom{BytePtrs: map[string]*[]byte{"F": &bytes}},
			fvTyp: reflect.TypeFor[*[]byte](),
			check: func(v reflect.Value) bool { return !v.IsNil() && string(v.Elem().Bytes()) == "data" },
		},
		{
			name:  "string ptr nil",
			fp:    fieldPlan{name: "F", table: TableStringPtrs, kind: kindPointer},
			atom:  &Atom{StringPtrs: map[string]*string{"F": nil}},
			fvTyp: reflect.TypeFor[*string](),
			check: func(v reflect.Value) bool { return v.IsNil() },
		},
		{
			name:  "uint ptr nil",
			fp:    fieldPlan{name: "F", table: TableUintPtrs, kind: kindPointer, converter: uintConverter(reflect.TypeFor[uint64]())},
			atom:  &Atom{UintPtrs: map[string]*uint64{"F": nil}},
			fvTyp: reflect.TypeFor[*uint64](),
			check: func(v reflect.Value) bool { return v.IsNil() },
		},
		{
			name:  "float ptr nil",
			fp:    fieldPlan{name: "F", table: TableFloatPtrs, kind: kindPointer, converter: floatConverter(reflect.TypeFor[float64]())},
			atom:  &Atom{FloatPtrs: map[string]*float64{"F": nil}},
			fvTyp: reflect.TypeFor[*float64](),
			check: func(v reflect.Value) bool { return v.IsNil() },
		},
		{
			name:  "bool ptr nil",
			fp:    fieldPlan{name: "F", table: TableBoolPtrs, kind: kindPointer},
			atom:  &Atom{BoolPtrs: map[string]*bool{"F": nil}},
			fvTyp: reflect.TypeFor[*bool](),
			check: func(v reflect.Value) bool { return v.IsNil() },
		},
		{
			name:  "time ptr nil",
			fp:    fieldPlan{name: "F", table: TableTimePtrs, kind: kindPointer},
			atom:  &Atom{TimePtrs: map[string]*time.Time{"F": nil}},
			fvTyp: reflect.TypeFor[*time.Time](),
			check: func(v reflect.Value) bool { return v.IsNil() },
		},
		{
			name:  "bytes ptr nil",
			fp:    fieldPlan{name: "F", table: TableBytePtrs, kind: kindPointer},
			atom:  &Atom{BytePtrs: map[string]*[]byte{"F": nil}},
			fvTyp: reflect.TypeFor[*[]byte](),
			check: func(v reflect.Value) bool { return v.IsNil() },
		},
	}

	// Add non-nil cases with existing values
	tests = append(tests, struct {
		name  string
		fp    fieldPlan
		atom  *Atom
		fvTyp reflect.Type
		check func(reflect.Value) bool
	}{
		name:  "string ptr",
		fp:    fieldPlan{name: "F", table: TableStringPtrs, kind: kindPointer},
		atom:  &Atom{StringPtrs: map[string]*string{"F": &str}},
		fvTyp: reflect.TypeFor[*string](),
		check: func(v reflect.Value) bool { return !v.IsNil() && v.Elem().String() == "hello" },
	})

	tests = append(tests, struct {
		name  string
		fp    fieldPlan
		atom  *Atom
		fvTyp reflect.Type
		check func(reflect.Value) bool
	}{
		name:  "int ptr",
		fp:    fieldPlan{name: "F", table: TableIntPtrs, kind: kindPointer, converter: intConverter(reflect.TypeFor[int64]())},
		atom:  &Atom{IntPtrs: map[string]*int64{"F": &i64}},
		fvTyp: reflect.TypeFor[*int64](),
		check: func(v reflect.Value) bool { return !v.IsNil() && v.Elem().Int() == 42 },
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := reflect.New(tt.fvTyp).Elem()
			if err := deatomizePointer(&tt.fp, tt.atom, fv); err != nil {
				t.Fatalf("deatomizePointer error: %v", err)
			}
			if !tt.check(fv) {
				t.Errorf("check failed for %s", tt.name)
			}
		})
	}
}

func TestDeatomizePointerUintOverflow(t *testing.T) {
	val := uint64(1000)
	atom := &Atom{UintPtrs: map[string]*uint64{"F": &val}}
	fp := fieldPlan{
		name:      "F",
		table:     TableUintPtrs,
		kind:      kindPointer,
		converter: uintConverter(reflect.TypeFor[uint8]()),
	}
	fv := reflect.New(reflect.TypeFor[*uint8]()).Elem()

	err := deatomizePointer(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizePointerFloatOverflow(t *testing.T) {
	val := math.MaxFloat64
	atom := &Atom{FloatPtrs: map[string]*float64{"F": &val}}
	fp := fieldPlan{
		name:      "F",
		table:     TableFloatPtrs,
		kind:      kindPointer,
		converter: floatConverter(reflect.TypeFor[float32]()),
	}
	fv := reflect.New(reflect.TypeFor[*float32]()).Elem()

	err := deatomizePointer(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizeSliceUint(t *testing.T) {
	atom := &Atom{UintSlices: map[string][]uint64{"F": {1, 2, 3}}}
	fp := fieldPlan{
		name:      "F",
		table:     TableUintSlices,
		kind:      kindSlice,
		converter: uintConverter(reflect.TypeFor[uint64]()),
	}
	fv := reflect.New(reflect.TypeFor[[]uint64]()).Elem()

	if err := deatomizeSlice(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizeSlice error: %v", err)
	}

	if fv.Len() != 3 {
		t.Errorf("got length %d, want 3", fv.Len())
	}
}

func TestDeatomizeSliceUintOverflow(t *testing.T) {
	atom := &Atom{UintSlices: map[string][]uint64{"F": {1, 2, 1000}}}
	fp := fieldPlan{
		name:      "F",
		table:     TableUintSlices,
		kind:      kindSlice,
		converter: uintConverter(reflect.TypeFor[uint8]()),
	}
	fv := reflect.New(reflect.TypeFor[[]uint8]()).Elem()

	err := deatomizeSlice(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestDeatomizeSliceFloat(t *testing.T) {
	atom := &Atom{FloatSlices: map[string][]float64{"F": {1.1, 2.2, 3.3}}}
	fp := fieldPlan{
		name:      "F",
		table:     TableFloatSlices,
		kind:      kindSlice,
		converter: floatConverter(reflect.TypeFor[float64]()),
	}
	fv := reflect.New(reflect.TypeFor[[]float64]()).Elem()

	if err := deatomizeSlice(&fp, atom, fv); err != nil {
		t.Fatalf("deatomizeSlice error: %v", err)
	}

	if fv.Len() != 3 {
		t.Errorf("got length %d, want 3", fv.Len())
	}
}

func TestDeatomizeSliceFloatOverflow(t *testing.T) {
	atom := &Atom{FloatSlices: map[string][]float64{"F": {1.1, math.MaxFloat64}}}
	fp := fieldPlan{
		name:      "F",
		table:     TableFloatSlices,
		kind:      kindSlice,
		converter: floatConverter(reflect.TypeFor[float32]()),
	}
	fv := reflect.New(reflect.TypeFor[[]float32]()).Elem()

	err := deatomizeSlice(&fp, atom, fv)
	if err == nil {
		t.Error("expected overflow error")
	}
}
