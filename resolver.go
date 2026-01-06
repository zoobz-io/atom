package atom

import (
	"reflect"
	"time"
)

var (
	atomizableType   = reflect.TypeFor[Atomizable]()
	deatomizableType = reflect.TypeFor[Deatomizable]()
)

// reflectAtomizer performs reflection-based atomization for a single type.
type reflectAtomizer struct {
	typ             reflect.Type
	tableSet        map[Table]int
	spec            Spec
	plan            []fieldPlan
	hasAtomizable   bool
	hasDeatomizable bool
}

// atomize converts a struct value to an Atom.
func (ra *reflectAtomizer) atomize(src any, dst *Atom) {
	v := reflect.ValueOf(src)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := range ra.plan {
		fp := &ra.plan[i]
		fv := v.FieldByIndex(fp.index)
		atomizeField(fp, fv, dst)
	}
}

// deatomize reconstructs a struct value from an Atom.
func (ra *reflectAtomizer) deatomize(src *Atom, dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := range ra.plan {
		fp := &ra.plan[i]
		fv := v.FieldByIndex(fp.index)
		if err := deatomizeField(fp, src, fv); err != nil {
			return err
		}
	}
	return nil
}

// newAtom creates an Atom with only the maps needed for this type.
func (ra *reflectAtomizer) newAtom() *Atom {
	return allocateAtom(ra.spec, ra.tableSet)
}

// tables returns the count of fields per table used by this atomizer.
func (ra *reflectAtomizer) tables() map[Table]int {
	return ra.tableSet
}

// buildReflectAtomizerWithSpec creates a new reflectAtomizer for the given type with spec.
// Returns an error if any field has an unsupported type.
func buildReflectAtomizerWithSpec(typ reflect.Type, spec Spec) (*reflectAtomizer, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	plan, err := buildFieldPlan(typ)
	if err != nil {
		return nil, err
	}
	tableSet := computeTableSet(plan)

	return &reflectAtomizer{
		typ:             typ,
		spec:            spec,
		plan:            plan,
		tableSet:        tableSet,
		hasAtomizable:   implementsAtomizable(typ),
		hasDeatomizable: implementsDeatomizable(typ),
	}, nil
}

// implementsAtomizable checks if a type implements Atomizable.
func implementsAtomizable(t reflect.Type) bool {
	// Check both value and pointer receiver
	if t.Implements(atomizableType) {
		return true
	}
	if t.Kind() != reflect.Ptr {
		return reflect.PointerTo(t).Implements(atomizableType)
	}
	return false
}

// implementsDeatomizable checks if a type implements Deatomizable.
func implementsDeatomizable(t reflect.Type) bool {
	// Check both value and pointer receiver
	if t.Implements(deatomizableType) {
		return true
	}
	if t.Kind() != reflect.Ptr {
		return reflect.PointerTo(t).Implements(deatomizableType)
	}
	return false
}

// allocateAtom creates an Atom with pre-sized maps based on field counts.
func allocateAtom(spec Spec, tableSet map[Table]int) *Atom {
	atom := &Atom{
		Spec:         spec,
		Nested:       make(map[string]Atom),
		NestedSlices: make(map[string][]Atom),
	}

	if n := tableSet[TableStrings]; n > 0 {
		atom.Strings = make(map[string]string, n)
	}
	if n := tableSet[TableInts]; n > 0 {
		atom.Ints = make(map[string]int64, n)
	}
	if n := tableSet[TableUints]; n > 0 {
		atom.Uints = make(map[string]uint64, n)
	}
	if n := tableSet[TableFloats]; n > 0 {
		atom.Floats = make(map[string]float64, n)
	}
	if n := tableSet[TableBools]; n > 0 {
		atom.Bools = make(map[string]bool, n)
	}
	if n := tableSet[TableTimes]; n > 0 {
		atom.Times = make(map[string]time.Time, n)
	}
	if n := tableSet[TableBytes]; n > 0 {
		atom.Bytes = make(map[string][]byte, n)
	}
	if n := tableSet[TableBytePtrs]; n > 0 {
		atom.BytePtrs = make(map[string]*[]byte, n)
	}
	if n := tableSet[TableStringPtrs]; n > 0 {
		atom.StringPtrs = make(map[string]*string, n)
	}
	if n := tableSet[TableIntPtrs]; n > 0 {
		atom.IntPtrs = make(map[string]*int64, n)
	}
	if n := tableSet[TableUintPtrs]; n > 0 {
		atom.UintPtrs = make(map[string]*uint64, n)
	}
	if n := tableSet[TableFloatPtrs]; n > 0 {
		atom.FloatPtrs = make(map[string]*float64, n)
	}
	if n := tableSet[TableBoolPtrs]; n > 0 {
		atom.BoolPtrs = make(map[string]*bool, n)
	}
	if n := tableSet[TableTimePtrs]; n > 0 {
		atom.TimePtrs = make(map[string]*time.Time, n)
	}
	if n := tableSet[TableStringSlices]; n > 0 {
		atom.StringSlices = make(map[string][]string, n)
	}
	if n := tableSet[TableIntSlices]; n > 0 {
		atom.IntSlices = make(map[string][]int64, n)
	}
	if n := tableSet[TableUintSlices]; n > 0 {
		atom.UintSlices = make(map[string][]uint64, n)
	}
	if n := tableSet[TableFloatSlices]; n > 0 {
		atom.FloatSlices = make(map[string][]float64, n)
	}
	if n := tableSet[TableBoolSlices]; n > 0 {
		atom.BoolSlices = make(map[string][]bool, n)
	}
	if n := tableSet[TableTimeSlices]; n > 0 {
		atom.TimeSlices = make(map[string][]time.Time, n)
	}
	if n := tableSet[TableByteSlices]; n > 0 {
		atom.ByteSlices = make(map[string][][]byte, n)
	}
	if n := tableSet[TableStringMaps]; n > 0 {
		atom.StringMaps = make(map[string]map[string]string, n)
	}
	if n := tableSet[TableIntMaps]; n > 0 {
		atom.IntMaps = make(map[string]map[string]int64, n)
	}
	if n := tableSet[TableUintMaps]; n > 0 {
		atom.UintMaps = make(map[string]map[string]uint64, n)
	}
	if n := tableSet[TableFloatMaps]; n > 0 {
		atom.FloatMaps = make(map[string]map[string]float64, n)
	}
	if n := tableSet[TableBoolMaps]; n > 0 {
		atom.BoolMaps = make(map[string]map[string]bool, n)
	}
	if n := tableSet[TableTimeMaps]; n > 0 {
		atom.TimeMaps = make(map[string]map[string]time.Time, n)
	}
	if n := tableSet[TableByteMaps]; n > 0 {
		atom.ByteMaps = make(map[string]map[string][]byte, n)
	}
	if n := tableSet[TableNestedMaps]; n > 0 {
		atom.NestedMaps = make(map[string]map[string]Atom, n)
	}

	return atom
}

// computeTableSet determines which tables are used by a field plan
// and counts how many fields use each table.
func computeTableSet(plans []fieldPlan) map[Table]int {
	set := make(map[Table]int)
	for i := range plans {
		fp := &plans[i]
		if fp.table != "" {
			set[fp.table]++
		}
		// Nested types contribute their tables too (for allocation, not sizing)
		if fp.nested != nil {
			for t, count := range fp.nested.tables() {
				if set[t] == 0 {
					set[t] = count
				}
			}
		}
	}
	return set
}
