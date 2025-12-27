package atom

import (
	"fmt"
	"reflect"
	"time"
)

// fieldKind represents the category of a field's type.
type fieldKind int

const (
	kindScalar fieldKind = iota
	kindPointer
	kindSlice
	kindNested
	kindNestedSlice
	kindNestedPtr
)

// fieldPlan describes how to atomize/deatomize a single field.
type fieldPlan struct {
	converter typeConverter
	elemType  reflect.Type
	nested    *reflectAtomizer
	name      string
	table     Table
	index     []int
	kind      fieldKind
}

var timeType = reflect.TypeFor[time.Time]()

// buildFieldPlan creates a field plan for the given struct type.
// Returns an error if any field has an unsupported type.
func buildFieldPlan(typ reflect.Type) ([]fieldPlan, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	var plans []fieldPlan
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)

		// Skip unexported fields
		if !sf.IsExported() {
			continue
		}

		fp, err := planField(sf, []int{i})
		if err != nil {
			return nil, fmt.Errorf("type %s: %w", typ.Name(), err)
		}
		plans = append(plans, fp)
	}
	return plans, nil
}

// planField creates a fieldPlan for a single struct field.
// Returns an error if the field type is not supported.
func planField(sf reflect.StructField, index []int) (fieldPlan, error) {
	ft := sf.Type

	fp := fieldPlan{
		name:  sf.Name,
		index: index,
	}

	// Handle pointer types
	if ft.Kind() == reflect.Ptr {
		elemType := ft.Elem()

		// Pointer to []byte or named []byte types (e.g., *net.IP)
		if elemType.Kind() == reflect.Slice && elemType.Elem().Kind() == reflect.Uint8 {
			fp.table = TableBytePtrs
			fp.kind = kindPointer
			fp.elemType = elemType
			return fp, nil
		}

		// Pointer to scalar
		if table, conv, ok := scalarToTable(elemType); ok {
			fp.table = pointerTable(table)
			fp.kind = kindPointer
			fp.converter = conv
			fp.elemType = elemType
			return fp, nil
		}

		// Pointer to struct (nested)
		if elemType.Kind() == reflect.Struct && elemType != timeType {
			fp.kind = kindNestedPtr
			fp.nested = ensureRegistered(elemType)
			fp.elemType = elemType
			return fp, nil
		}

		return fp, fmt.Errorf("field %q: unsupported pointer element type %s", sf.Name, elemType)
	}

	// Handle fixed-size byte arrays ([N]byte) - stored as []byte
	if ft.Kind() == reflect.Array && ft.Elem().Kind() == reflect.Uint8 {
		fp.table = TableBytes
		fp.kind = kindScalar
		fp.elemType = ft // Store array type for size validation during deatomization
		return fp, nil
	}

	// Handle slice types
	if ft.Kind() == reflect.Slice {
		elemType := ft.Elem()

		// []byte and named []byte types (e.g., net.IP, json.RawMessage) - treated as scalar
		if elemType.Kind() == reflect.Uint8 {
			fp.table = TableBytes
			fp.kind = kindScalar
			return fp, nil
		}

		// [][]byte and slices of named []byte types
		if elemType.Kind() == reflect.Slice && elemType.Elem().Kind() == reflect.Uint8 {
			fp.table = TableByteSlices
			fp.kind = kindSlice
			fp.elemType = elemType
			return fp, nil
		}

		// Slice of structs = nested slice
		if elemType.Kind() == reflect.Struct && elemType != timeType {
			fp.kind = kindNestedSlice
			fp.nested = ensureRegistered(elemType)
			fp.elemType = elemType
			return fp, nil
		}

		// Slice of pointers to structs
		if elemType.Kind() == reflect.Ptr {
			ptrElem := elemType.Elem()
			if ptrElem.Kind() == reflect.Struct && ptrElem != timeType {
				fp.kind = kindNestedSlice
				fp.nested = ensureRegistered(ptrElem)
				fp.elemType = elemType
				return fp, nil
			}
		}

		// Slice of scalars
		if table, conv, ok := scalarToTable(elemType); ok {
			fp.table = sliceTable(table)
			fp.kind = kindSlice
			fp.converter = conv
			fp.elemType = elemType
			return fp, nil
		}

		return fp, fmt.Errorf("field %q: unsupported slice element type %s", sf.Name, elemType)
	}

	// Handle map types - explicitly unsupported
	if ft.Kind() == reflect.Map {
		return fp, fmt.Errorf("field %q: map types are not supported", sf.Name)
	}

	// Handle struct (nested)
	if ft.Kind() == reflect.Struct && ft != timeType {
		fp.kind = kindNested
		fp.nested = ensureRegistered(ft)
		return fp, nil
	}

	// Handle scalar types
	if table, conv, ok := scalarToTable(ft); ok {
		fp.table = table
		fp.kind = kindScalar
		fp.converter = conv
		return fp, nil
	}

	return fp, fmt.Errorf("field %q: unsupported type %s", sf.Name, ft)
}

// scalarToTable maps a reflect.Type to its Table.
func scalarToTable(t reflect.Type) (Table, typeConverter, bool) {
	switch t.Kind() {
	case reflect.String:
		return TableStrings, typeConverter{}, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return TableInts, intConverter(t), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return TableUints, uintConverter(t), true
	case reflect.Float32, reflect.Float64:
		return TableFloats, floatConverter(t), true
	case reflect.Bool:
		return TableBools, typeConverter{}, true
	}

	if t == timeType {
		return TableTimes, typeConverter{}, true
	}

	return "", typeConverter{}, false
}

// pointerTable returns the pointer variant of a base table.
func pointerTable(base Table) Table {
	switch base {
	case TableStrings:
		return TableStringPtrs
	case TableInts:
		return TableIntPtrs
	case TableUints:
		return TableUintPtrs
	case TableFloats:
		return TableFloatPtrs
	case TableBools:
		return TableBoolPtrs
	case TableTimes:
		return TableTimePtrs
	case TableBytes:
		return TableBytePtrs
	default:
		return ""
	}
}

// sliceTable returns the slice variant of a base table.
func sliceTable(base Table) Table {
	switch base {
	case TableStrings:
		return TableStringSlices
	case TableInts:
		return TableIntSlices
	case TableUints:
		return TableUintSlices
	case TableFloats:
		return TableFloatSlices
	case TableBools:
		return TableBoolSlices
	case TableTimes:
		return TableTimeSlices
	default:
		return ""
	}
}
