package atom

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// typeConverter handles width conversion between struct fields and Atom storage.
type typeConverter struct {
	toInt64     func(reflect.Value) int64
	fromInt64   func(int64) (reflect.Value, error)
	toUint64    func(reflect.Value) uint64
	fromUint64  func(uint64) (reflect.Value, error)
	toFloat64   func(reflect.Value) float64
	fromFloat64 func(float64) (reflect.Value, error)
	origType    reflect.Type
}

// intConverter creates a converter for signed integer types.
func intConverter(t reflect.Type) typeConverter {
	var minVal, maxVal int64
	switch t.Kind() {
	case reflect.Int8:
		minVal, maxVal = math.MinInt8, math.MaxInt8
	case reflect.Int16:
		minVal, maxVal = math.MinInt16, math.MaxInt16
	case reflect.Int32:
		minVal, maxVal = math.MinInt32, math.MaxInt32
	default: // int, int64
		minVal, maxVal = math.MinInt64, math.MaxInt64
	}

	return typeConverter{
		origType: t,
		toInt64: func(v reflect.Value) int64 {
			return v.Int()
		},
		fromInt64: func(i int64) (reflect.Value, error) {
			if i < minVal || i > maxVal {
				return reflect.Value{}, fmt.Errorf("value %d overflows %s (range %d to %d)", i, t.Kind(), minVal, maxVal)
			}
			rv := reflect.New(t).Elem()
			rv.SetInt(i)
			return rv, nil
		},
	}
}

// uintConverter creates a converter for unsigned integer types.
func uintConverter(t reflect.Type) typeConverter {
	var maxVal uint64
	switch t.Kind() {
	case reflect.Uint8:
		maxVal = math.MaxUint8
	case reflect.Uint16:
		maxVal = math.MaxUint16
	case reflect.Uint32:
		maxVal = math.MaxUint32
	default: // uint, uint64
		maxVal = math.MaxUint64
	}

	return typeConverter{
		origType: t,
		toUint64: func(v reflect.Value) uint64 {
			return v.Uint()
		},
		fromUint64: func(u uint64) (reflect.Value, error) {
			if u > maxVal {
				return reflect.Value{}, fmt.Errorf("value %d overflows %s (max %d)", u, t.Kind(), maxVal)
			}
			rv := reflect.New(t).Elem()
			rv.SetUint(u)
			return rv, nil
		},
	}
}

// floatConverter creates a converter for floating point types.
func floatConverter(t reflect.Type) typeConverter {
	return typeConverter{
		origType: t,
		toFloat64: func(v reflect.Value) float64 {
			return v.Float()
		},
		fromFloat64: func(f float64) (reflect.Value, error) {
			if t.Kind() == reflect.Float32 {
				if !math.IsInf(f, 0) && !math.IsNaN(f) {
					if f > math.MaxFloat32 || f < -math.MaxFloat32 {
						return reflect.Value{}, fmt.Errorf("value %g overflows float32 (max %g)", f, math.MaxFloat32)
					}
				}
			}
			rv := reflect.New(t).Elem()
			rv.SetFloat(f)
			return rv, nil
		},
	}
}

// atomizeField converts a single field value to its Atom representation.
func atomizeField(fp *fieldPlan, fv reflect.Value, dst *Atom) {
	switch fp.kind {
	case kindScalar:
		atomizeScalar(fp, fv, dst)
	case kindPointer:
		atomizePointer(fp, fv, dst)
	case kindSlice:
		atomizeSlice(fp, fv, dst)
	case kindNested:
		atomizeNested(fp, fv, dst)
	case kindNestedSlice:
		atomizeNestedSlice(fp, fv, dst)
	case kindNestedPtr:
		atomizeNestedPtr(fp, fv, dst)
	}
}

func atomizeScalar(fp *fieldPlan, fv reflect.Value, dst *Atom) {
	switch fp.table {
	case TableStrings:
		dst.Strings[fp.name] = fv.String()
	case TableInts:
		dst.Ints[fp.name] = fp.converter.toInt64(fv)
	case TableUints:
		dst.Uints[fp.name] = fp.converter.toUint64(fv)
	case TableFloats:
		dst.Floats[fp.name] = fp.converter.toFloat64(fv)
	case TableBools:
		dst.Bools[fp.name] = fv.Bool()
	case TableTimes:
		dst.Times[fp.name] = fv.Interface().(time.Time) //nolint:errcheck // type assertion is safe
	case TableBytes:
		// Handle both slices and fixed-size arrays
		if fv.Kind() == reflect.Array {
			// Convert array to slice
			b := make([]byte, fv.Len())
			for i := 0; i < fv.Len(); i++ {
				b[i] = byte(fv.Index(i).Uint())
			}
			dst.Bytes[fp.name] = b
		} else if !fv.IsNil() {
			dst.Bytes[fp.name] = fv.Bytes()
		}
	}
}

func atomizePointer(fp *fieldPlan, fv reflect.Value, dst *Atom) {
	if fv.IsNil() {
		// Store nil explicitly
		switch fp.table {
		case TableStringPtrs:
			dst.StringPtrs[fp.name] = nil
		case TableIntPtrs:
			dst.IntPtrs[fp.name] = nil
		case TableUintPtrs:
			dst.UintPtrs[fp.name] = nil
		case TableFloatPtrs:
			dst.FloatPtrs[fp.name] = nil
		case TableBoolPtrs:
			dst.BoolPtrs[fp.name] = nil
		case TableTimePtrs:
			dst.TimePtrs[fp.name] = nil
		case TableBytePtrs:
			dst.BytePtrs[fp.name] = nil
		}
		return
	}

	elem := fv.Elem()
	switch fp.table {
	case TableStringPtrs:
		v := elem.String()
		dst.StringPtrs[fp.name] = &v
	case TableIntPtrs:
		v := fp.converter.toInt64(elem)
		dst.IntPtrs[fp.name] = &v
	case TableUintPtrs:
		v := fp.converter.toUint64(elem)
		dst.UintPtrs[fp.name] = &v
	case TableFloatPtrs:
		v := fp.converter.toFloat64(elem)
		dst.FloatPtrs[fp.name] = &v
	case TableBoolPtrs:
		v := elem.Bool()
		dst.BoolPtrs[fp.name] = &v
	case TableTimePtrs:
		v := elem.Interface().(time.Time) //nolint:errcheck // type assertion is safe
		dst.TimePtrs[fp.name] = &v
	case TableBytePtrs:
		v := elem.Bytes()
		dst.BytePtrs[fp.name] = &v
	}
}

func atomizeSlice(fp *fieldPlan, fv reflect.Value, dst *Atom) {
	if fv.IsNil() {
		return
	}

	switch fp.table {
	case TableStringSlices:
		s := make([]string, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			s[i] = fv.Index(i).String()
		}
		dst.StringSlices[fp.name] = s
	case TableIntSlices:
		s := make([]int64, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			s[i] = fp.converter.toInt64(fv.Index(i))
		}
		dst.IntSlices[fp.name] = s
	case TableUintSlices:
		s := make([]uint64, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			s[i] = fp.converter.toUint64(fv.Index(i))
		}
		dst.UintSlices[fp.name] = s
	case TableFloatSlices:
		s := make([]float64, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			s[i] = fp.converter.toFloat64(fv.Index(i))
		}
		dst.FloatSlices[fp.name] = s
	case TableBoolSlices:
		s := make([]bool, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			s[i] = fv.Index(i).Bool()
		}
		dst.BoolSlices[fp.name] = s
	case TableTimeSlices:
		s := make([]time.Time, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			s[i] = fv.Index(i).Interface().(time.Time) //nolint:errcheck // type assertion is safe
		}
		dst.TimeSlices[fp.name] = s
	case TableByteSlices:
		s := make([][]byte, fv.Len())
		for i := 0; i < fv.Len(); i++ {
			s[i] = fv.Index(i).Bytes()
		}
		dst.ByteSlices[fp.name] = s
	}
}

func atomizeNested(fp *fieldPlan, fv reflect.Value, dst *Atom) {
	nestedAtom := fp.nested.newAtom()
	fp.nested.atomize(fv.Addr().Interface(), nestedAtom)
	dst.Nested[fp.name] = *nestedAtom
}

func atomizeNestedPtr(fp *fieldPlan, fv reflect.Value, dst *Atom) {
	if fv.IsNil() {
		// Don't store anything for nil pointers to nested structs
		return
	}

	nestedAtom := fp.nested.newAtom()
	fp.nested.atomize(fv.Interface(), nestedAtom)
	dst.Nested[fp.name] = *nestedAtom
}

func atomizeNestedSlice(fp *fieldPlan, fv reflect.Value, dst *Atom) {
	if fv.IsNil() || fv.Len() == 0 {
		return
	}

	atoms := make([]Atom, 0, fv.Len())
	for i := 0; i < fv.Len(); i++ {
		elem := fv.Index(i)

		// Handle slice of pointers to structs - skip nil elements
		if elem.Kind() == reflect.Ptr {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		nestedAtom := fp.nested.newAtom()
		fp.nested.atomize(elem.Addr().Interface(), nestedAtom)
		atoms = append(atoms, *nestedAtom)
	}
	if len(atoms) > 0 {
		dst.NestedSlices[fp.name] = atoms
	}
}

// deatomizeField reconstructs a single field value from its Atom representation.
func deatomizeField(fp *fieldPlan, src *Atom, fv reflect.Value) error {
	switch fp.kind {
	case kindScalar:
		return deatomizeScalar(fp, src, fv)
	case kindPointer:
		return deatomizePointer(fp, src, fv)
	case kindSlice:
		return deatomizeSlice(fp, src, fv)
	case kindNested:
		return deatomizeNested(fp, src, fv)
	case kindNestedSlice:
		return deatomizeNestedSlice(fp, src, fv)
	case kindNestedPtr:
		return deatomizeNestedPtr(fp, src, fv)
	}
	return nil
}

func deatomizeScalar(fp *fieldPlan, src *Atom, fv reflect.Value) error {
	switch fp.table {
	case TableStrings:
		if v, ok := src.Strings[fp.name]; ok {
			fv.SetString(v)
		}
	case TableInts:
		if v, ok := src.Ints[fp.name]; ok {
			rv, err := fp.converter.fromInt64(v)
			if err != nil {
				return fmt.Errorf("field %q: %w", fp.name, err)
			}
			fv.Set(rv)
		}
	case TableUints:
		if v, ok := src.Uints[fp.name]; ok {
			rv, err := fp.converter.fromUint64(v)
			if err != nil {
				return fmt.Errorf("field %q: %w", fp.name, err)
			}
			fv.Set(rv)
		}
	case TableFloats:
		if v, ok := src.Floats[fp.name]; ok {
			rv, err := fp.converter.fromFloat64(v)
			if err != nil {
				return fmt.Errorf("field %q: %w", fp.name, err)
			}
			fv.Set(rv)
		}
	case TableBools:
		if v, ok := src.Bools[fp.name]; ok {
			fv.SetBool(v)
		}
	case TableTimes:
		if v, ok := src.Times[fp.name]; ok {
			fv.Set(reflect.ValueOf(v))
		}
	case TableBytes:
		if v, ok := src.Bytes[fp.name]; ok {
			// Handle both slices and fixed-size arrays
			if fv.Kind() == reflect.Array {
				// Validate length matches array size
				if len(v) != fv.Len() {
					return fmt.Errorf("field %q: byte slice length %d does not match array size %d", fp.name, len(v), fv.Len())
				}
				// Copy bytes into array
				for i := 0; i < len(v); i++ {
					fv.Index(i).SetUint(uint64(v[i]))
				}
			} else {
				fv.SetBytes(v)
			}
		}
	}
	return nil
}

func deatomizePointer(fp *fieldPlan, src *Atom, fv reflect.Value) error {
	switch fp.table {
	case TableStringPtrs:
		if v, ok := src.StringPtrs[fp.name]; ok {
			if v == nil {
				fv.Set(reflect.Zero(fv.Type()))
			} else {
				ptr := reflect.New(fv.Type().Elem())
				ptr.Elem().SetString(*v)
				fv.Set(ptr)
			}
		}
	case TableIntPtrs:
		if v, ok := src.IntPtrs[fp.name]; ok {
			if v == nil {
				fv.Set(reflect.Zero(fv.Type()))
			} else {
				rv, err := fp.converter.fromInt64(*v)
				if err != nil {
					return fmt.Errorf("field %q: %w", fp.name, err)
				}
				ptr := reflect.New(rv.Type())
				ptr.Elem().Set(rv)
				fv.Set(ptr)
			}
		}
	case TableUintPtrs:
		if v, ok := src.UintPtrs[fp.name]; ok {
			if v == nil {
				fv.Set(reflect.Zero(fv.Type()))
			} else {
				rv, err := fp.converter.fromUint64(*v)
				if err != nil {
					return fmt.Errorf("field %q: %w", fp.name, err)
				}
				ptr := reflect.New(rv.Type())
				ptr.Elem().Set(rv)
				fv.Set(ptr)
			}
		}
	case TableFloatPtrs:
		if v, ok := src.FloatPtrs[fp.name]; ok {
			if v == nil {
				fv.Set(reflect.Zero(fv.Type()))
			} else {
				rv, err := fp.converter.fromFloat64(*v)
				if err != nil {
					return fmt.Errorf("field %q: %w", fp.name, err)
				}
				ptr := reflect.New(rv.Type())
				ptr.Elem().Set(rv)
				fv.Set(ptr)
			}
		}
	case TableBoolPtrs:
		if v, ok := src.BoolPtrs[fp.name]; ok {
			if v == nil {
				fv.Set(reflect.Zero(fv.Type()))
			} else {
				ptr := reflect.New(fv.Type().Elem())
				ptr.Elem().SetBool(*v)
				fv.Set(ptr)
			}
		}
	case TableTimePtrs:
		if v, ok := src.TimePtrs[fp.name]; ok {
			if v == nil {
				fv.Set(reflect.Zero(fv.Type()))
			} else {
				ptr := reflect.New(fv.Type().Elem())
				ptr.Elem().Set(reflect.ValueOf(*v))
				fv.Set(ptr)
			}
		}
	case TableBytePtrs:
		if v, ok := src.BytePtrs[fp.name]; ok {
			if v == nil {
				fv.Set(reflect.Zero(fv.Type()))
			} else {
				ptr := reflect.New(fv.Type().Elem())
				ptr.Elem().SetBytes(*v)
				fv.Set(ptr)
			}
		}
	}
	return nil
}

func deatomizeSlice(fp *fieldPlan, src *Atom, fv reflect.Value) error {
	switch fp.table {
	case TableStringSlices:
		if v, ok := src.StringSlices[fp.name]; ok {
			slice := reflect.MakeSlice(fv.Type(), len(v), len(v))
			for i, val := range v {
				slice.Index(i).SetString(val)
			}
			fv.Set(slice)
		}
	case TableIntSlices:
		if v, ok := src.IntSlices[fp.name]; ok {
			slice := reflect.MakeSlice(fv.Type(), len(v), len(v))
			for i, val := range v {
				rv, err := fp.converter.fromInt64(val)
				if err != nil {
					return fmt.Errorf("field %q[%d]: %w", fp.name, i, err)
				}
				slice.Index(i).Set(rv)
			}
			fv.Set(slice)
		}
	case TableUintSlices:
		if v, ok := src.UintSlices[fp.name]; ok {
			slice := reflect.MakeSlice(fv.Type(), len(v), len(v))
			for i, val := range v {
				rv, err := fp.converter.fromUint64(val)
				if err != nil {
					return fmt.Errorf("field %q[%d]: %w", fp.name, i, err)
				}
				slice.Index(i).Set(rv)
			}
			fv.Set(slice)
		}
	case TableFloatSlices:
		if v, ok := src.FloatSlices[fp.name]; ok {
			slice := reflect.MakeSlice(fv.Type(), len(v), len(v))
			for i, val := range v {
				rv, err := fp.converter.fromFloat64(val)
				if err != nil {
					return fmt.Errorf("field %q[%d]: %w", fp.name, i, err)
				}
				slice.Index(i).Set(rv)
			}
			fv.Set(slice)
		}
	case TableBoolSlices:
		if v, ok := src.BoolSlices[fp.name]; ok {
			slice := reflect.MakeSlice(fv.Type(), len(v), len(v))
			for i, val := range v {
				slice.Index(i).SetBool(val)
			}
			fv.Set(slice)
		}
	case TableTimeSlices:
		if v, ok := src.TimeSlices[fp.name]; ok {
			slice := reflect.MakeSlice(fv.Type(), len(v), len(v))
			for i, val := range v {
				slice.Index(i).Set(reflect.ValueOf(val))
			}
			fv.Set(slice)
		}
	case TableByteSlices:
		if v, ok := src.ByteSlices[fp.name]; ok {
			slice := reflect.MakeSlice(fv.Type(), len(v), len(v))
			for i, val := range v {
				slice.Index(i).SetBytes(val)
			}
			fv.Set(slice)
		}
	}
	return nil
}

func deatomizeNested(fp *fieldPlan, src *Atom, fv reflect.Value) error {
	nestedAtom, ok := src.Nested[fp.name]
	if !ok {
		return nil
	}

	return fp.nested.deatomize(&nestedAtom, fv.Addr().Interface())
}

func deatomizeNestedPtr(fp *fieldPlan, src *Atom, fv reflect.Value) error {
	nestedAtom, ok := src.Nested[fp.name]
	if !ok {
		// Leave as nil
		return nil
	}

	// Create new instance
	ptr := reflect.New(fp.elemType)
	if err := fp.nested.deatomize(&nestedAtom, ptr.Interface()); err != nil {
		return err
	}
	fv.Set(ptr)
	return nil
}

func deatomizeNestedSlice(fp *fieldPlan, src *Atom, fv reflect.Value) error {
	atoms, ok := src.NestedSlices[fp.name]
	if !ok || len(atoms) == 0 {
		return nil
	}

	// Determine if this is a slice of pointers or values
	elemType := fv.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr

	slice := reflect.MakeSlice(fv.Type(), len(atoms), len(atoms))
	for i := range atoms {
		nestedAtom := &atoms[i]

		if isPtr {
			// Slice of pointers to structs
			ptr := reflect.New(elemType.Elem())
			if err := fp.nested.deatomize(nestedAtom, ptr.Interface()); err != nil {
				return err
			}
			slice.Index(i).Set(ptr)
		} else {
			// Slice of struct values
			elem := slice.Index(i)
			if err := fp.nested.deatomize(nestedAtom, elem.Addr().Interface()); err != nil {
				return err
			}
		}
	}
	fv.Set(slice)
	return nil
}
