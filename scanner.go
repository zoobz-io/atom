package atom

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// ColScanner is the interface for database row scanning.
// Satisfied by sqlx.Rows and other database libraries.
type ColScanner interface {
	Columns() ([]string, error)
	Scan(dest ...any) error
	Err() error
}

// Scanner efficiently scans database rows directly into Atoms.
type Scanner struct {
	byColumn map[string]*scanFieldPlan
	tableSet map[Table]int
	spec     Spec
}

// scanFieldPlan describes how to scan a single column into an Atom.
type scanFieldPlan struct {
	fieldName string
	column    string
	table     Table
	path      []string
	nullable  bool
}

var (
	scannerMu sync.RWMutex
	scanners  = make(map[reflect.Type]*Scanner)
)

// ScannerFor retrieves the Scanner for a registered type.
// Returns nil, false if the type has not been registered via Use[T]().
func ScannerFor(spec Spec) (*Scanner, bool) {
	scannerMu.RLock()
	s, ok := scanners[spec.ReflectType]
	scannerMu.RUnlock()
	return s, ok
}

// registerScanner adds a scanner to the registry.
// Must be called with appropriate locking by the caller.
func registerScanner(typ reflect.Type, scanner *Scanner) {
	scannerMu.Lock()
	scanners[typ] = scanner
	scannerMu.Unlock()
}

// buildScanner creates a Scanner from field plans and spec.
func buildScanner(plans []fieldPlan, spec Spec) (*Scanner, error) {
	s := &Scanner{
		spec:     spec,
		byColumn: make(map[string]*scanFieldPlan),
		tableSet: make(map[Table]int),
	}

	// Build a map of field name to tags for this type
	fieldTags := make(map[string]map[string]string, len(spec.Fields))
	for _, f := range spec.Fields {
		fieldTags[f.Name] = f.Tags
	}

	if err := s.buildPlans(plans, fieldTags, nil); err != nil {
		return nil, err
	}

	return s, nil
}

// buildPlans recursively builds scan plans from field plans.
func (s *Scanner) buildPlans(plans []fieldPlan, fieldTags map[string]map[string]string, path []string) error {
	for i := range plans {
		fp := &plans[i]
		tags := fieldTags[fp.name]
		dbTag, hasDB := tags["db"]

		switch fp.kind {
		case kindScalar, kindPointer:
			if !hasDB {
				continue // No db tag, skip
			}
			// Check for column name collision
			if existing, ok := s.byColumn[dbTag]; ok {
				return fmt.Errorf("column %q maps to multiple fields: %s and %s",
					dbTag, fieldPath(existing.path, existing.fieldName), fieldPath(path, fp.name))
			}
			plan := &scanFieldPlan{
				fieldName: fp.name,
				column:    dbTag,
				table:     fp.table,
				nullable:  fp.kind == kindPointer,
				path:      path,
			}
			s.byColumn[dbTag] = plan
			s.tableSet[fp.table]++

		case kindNested, kindNestedPtr:
			if hasDB {
				return fmt.Errorf("field %q: struct fields cannot have db tags", fp.name)
			}
			if fp.nested == nil {
				continue
			}
			// Recurse into nested struct using its spec for tag lookup
			nestedPath := append(append([]string(nil), path...), fp.name)
			nestedTags := make(map[string]map[string]string)
			for _, f := range fp.nested.spec.Fields {
				nestedTags[f.Name] = f.Tags
			}
			if err := s.buildPlans(fp.nested.plan, nestedTags, nestedPath); err != nil {
				return err
			}

		case kindSlice, kindNestedSlice:
			if hasDB {
				return fmt.Errorf("field %q: slice fields cannot have db tags", fp.name)
			}
			// Silently skip slice fields - not scannable from database rows
		}
	}

	return nil
}

// fieldPath formats a field path for error messages.
func fieldPath(path []string, fieldName string) string {
	if len(path) == 0 {
		return fieldName
	}
	return strings.Join(path, ".") + "." + fieldName
}

// Scan reads a single row into an Atom.
func (s *Scanner) Scan(cs ColScanner) (*Atom, error) {
	cols, err := cs.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting columns: %w", err)
	}

	plans, dests := s.prepareScan(cols)

	if err := cs.Scan(dests...); err != nil {
		return nil, fmt.Errorf("scanning row: %w", err)
	}

	return s.buildAtom(plans, dests), nil
}

// ScanAll reads all rows into Atoms.
// The next function should return true while there are more rows (typically rows.Next).
func (s *Scanner) ScanAll(cs ColScanner, next func() bool) ([]*Atom, error) {
	cols, err := cs.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting columns: %w", err)
	}

	plans, dests := s.prepareScan(cols)

	var atoms []*Atom
	for next() {
		if err := cs.Scan(dests...); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		atoms = append(atoms, s.buildAtom(plans, dests))
		// Reset destinations for next row
		resetDests(dests)
	}

	if err := cs.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return atoms, nil
}

// prepareScan creates the scan plan and destinations for a column set.
func (s *Scanner) prepareScan(cols []string) (plans []*scanFieldPlan, dests []any) {
	plans = make([]*scanFieldPlan, len(cols))
	dests = make([]any, len(cols))

	for i, col := range cols {
		plan := s.byColumn[col]
		plans[i] = plan

		if plan == nil {
			// Unknown column, use a discard destination
			dests[i] = new(any)
			continue
		}

		dests[i] = makeScanDest(plan.table, plan.nullable)
	}

	return plans, dests
}

// makeScanDest creates a typed scan destination for a table.
func makeScanDest(table Table, nullable bool) any {
	if nullable {
		switch table {
		case TableStringPtrs:
			return new(sql.NullString)
		case TableIntPtrs:
			return new(sql.NullInt64)
		case TableUintPtrs:
			// sql package lacks NullUint64. Values > MaxInt64 will overflow.
			// See docs/3.guides/6.database-scanning.md for details.
			return new(sql.NullInt64)
		case TableFloatPtrs:
			return new(sql.NullFloat64)
		case TableBoolPtrs:
			return new(sql.NullBool)
		case TableTimePtrs:
			return new(sql.NullTime)
		case TableBytePtrs:
			return new([]byte)
		}
	}

	switch table {
	case TableStrings:
		return new(string)
	case TableInts:
		return new(int64)
	case TableUints:
		return new(uint64)
	case TableFloats:
		return new(float64)
	case TableBools:
		return new(bool)
	case TableTimes:
		return new(time.Time)
	case TableBytes:
		return new([]byte)
	}

	return new(any)
}

// resetDests resets scan destinations for reuse.
func resetDests(dests []any) {
	for i, d := range dests {
		switch v := d.(type) {
		case *string:
			*v = ""
		case *int64:
			*v = 0
		case *uint64:
			*v = 0
		case *float64:
			*v = 0
		case *bool:
			*v = false
		case *time.Time:
			*v = time.Time{}
		case *[]byte:
			*v = nil
		case *sql.NullString:
			*v = sql.NullString{}
		case *sql.NullInt64:
			*v = sql.NullInt64{}
		case *sql.NullFloat64:
			*v = sql.NullFloat64{}
		case *sql.NullBool:
			*v = sql.NullBool{}
		case *sql.NullTime:
			*v = sql.NullTime{}
		case *any:
			dests[i] = new(any)
		}
	}
}

// buildAtom constructs an Atom from scanned destinations.
func (s *Scanner) buildAtom(plans []*scanFieldPlan, dests []any) *Atom {
	atom := allocateAtom(s.spec, s.tableSet)

	for i, plan := range plans {
		if plan == nil {
			continue
		}

		dest := dests[i]
		assignToNested(atom, plan.path, func(target *Atom) {
			assignValue(target, plan, dest)
		})
	}

	return atom
}

// assignToNested assigns a value through a path of nested atoms.
func assignToNested(root *Atom, path []string, assign func(*Atom)) {
	if len(path) == 0 {
		assign(root)
		return
	}

	// For nested paths, we need to get/create each level, then assign back up
	nestedAtoms := make([]*Atom, len(path))
	current := root

	// Traverse down, creating atoms as needed
	for i, name := range path {
		if current.Nested == nil {
			current.Nested = make(map[string]Atom)
		}
		nested, ok := current.Nested[name]
		if !ok {
			nested = Atom{
				Nested:       make(map[string]Atom),
				NestedSlices: make(map[string][]Atom),
			}
		}
		nestedAtoms[i] = &nested
		current = &nested
	}

	// Assign to the leaf
	assign(current)

	// Assign back up the chain
	for i := len(path) - 1; i >= 0; i-- {
		parent := root
		if i > 0 {
			parent = nestedAtoms[i-1]
		}
		parent.Nested[path[i]] = *nestedAtoms[i]
	}
}

// assignValue assigns a scanned value to the appropriate Atom table.
func assignValue(atom *Atom, plan *scanFieldPlan, dest any) {
	switch plan.table {
	case TableStrings:
		if v, ok := dest.(*string); ok {
			if atom.Strings == nil {
				atom.Strings = make(map[string]string)
			}
			atom.Strings[plan.fieldName] = *v
		}
	case TableInts:
		if v, ok := dest.(*int64); ok {
			if atom.Ints == nil {
				atom.Ints = make(map[string]int64)
			}
			atom.Ints[plan.fieldName] = *v
		}
	case TableUints:
		if v, ok := dest.(*uint64); ok {
			if atom.Uints == nil {
				atom.Uints = make(map[string]uint64)
			}
			atom.Uints[plan.fieldName] = *v
		}
	case TableFloats:
		if v, ok := dest.(*float64); ok {
			if atom.Floats == nil {
				atom.Floats = make(map[string]float64)
			}
			atom.Floats[plan.fieldName] = *v
		}
	case TableBools:
		if v, ok := dest.(*bool); ok {
			if atom.Bools == nil {
				atom.Bools = make(map[string]bool)
			}
			atom.Bools[plan.fieldName] = *v
		}
	case TableTimes:
		if v, ok := dest.(*time.Time); ok {
			if atom.Times == nil {
				atom.Times = make(map[string]time.Time)
			}
			atom.Times[plan.fieldName] = *v
		}
	case TableBytes:
		if v, ok := dest.(*[]byte); ok {
			if atom.Bytes == nil {
				atom.Bytes = make(map[string][]byte)
			}
			atom.Bytes[plan.fieldName] = *v
		}

	// Nullable types
	case TableStringPtrs:
		if v, ok := dest.(*sql.NullString); ok {
			if atom.StringPtrs == nil {
				atom.StringPtrs = make(map[string]*string)
			}
			if v.Valid {
				val := v.String
				atom.StringPtrs[plan.fieldName] = &val
			} else {
				atom.StringPtrs[plan.fieldName] = nil
			}
		}
	case TableIntPtrs:
		if v, ok := dest.(*sql.NullInt64); ok {
			if atom.IntPtrs == nil {
				atom.IntPtrs = make(map[string]*int64)
			}
			if v.Valid {
				val := v.Int64
				atom.IntPtrs[plan.fieldName] = &val
			} else {
				atom.IntPtrs[plan.fieldName] = nil
			}
		}
	case TableUintPtrs:
		if v, ok := dest.(*sql.NullInt64); ok {
			if atom.UintPtrs == nil {
				atom.UintPtrs = make(map[string]*uint64)
			}
			if v.Valid {
				u := uint64(v.Int64) //nolint:gosec // database values assumed valid
				atom.UintPtrs[plan.fieldName] = &u
			} else {
				atom.UintPtrs[plan.fieldName] = nil
			}
		}
	case TableFloatPtrs:
		if v, ok := dest.(*sql.NullFloat64); ok {
			if atom.FloatPtrs == nil {
				atom.FloatPtrs = make(map[string]*float64)
			}
			if v.Valid {
				val := v.Float64
				atom.FloatPtrs[plan.fieldName] = &val
			} else {
				atom.FloatPtrs[plan.fieldName] = nil
			}
		}
	case TableBoolPtrs:
		if v, ok := dest.(*sql.NullBool); ok {
			if atom.BoolPtrs == nil {
				atom.BoolPtrs = make(map[string]*bool)
			}
			if v.Valid {
				val := v.Bool
				atom.BoolPtrs[plan.fieldName] = &val
			} else {
				atom.BoolPtrs[plan.fieldName] = nil
			}
		}
	case TableTimePtrs:
		if v, ok := dest.(*sql.NullTime); ok {
			if atom.TimePtrs == nil {
				atom.TimePtrs = make(map[string]*time.Time)
			}
			if v.Valid {
				val := v.Time
				atom.TimePtrs[plan.fieldName] = &val
			} else {
				atom.TimePtrs[plan.fieldName] = nil
			}
		}
	case TableBytePtrs:
		if v, ok := dest.(*[]byte); ok {
			if atom.BytePtrs == nil {
				atom.BytePtrs = make(map[string]*[]byte)
			}
			if v != nil && *v != nil {
				atom.BytePtrs[plan.fieldName] = v
			} else {
				atom.BytePtrs[plan.fieldName] = nil
			}
		}
	}
}
