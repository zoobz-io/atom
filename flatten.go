package atom

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/zoobzio/sentinel"
)

// Type name constants for unflatten type switching.
const (
	typeString   = "string"
	typeInt      = "int"
	typeInt8     = "int8"
	typeInt16    = "int16"
	typeInt32    = "int32"
	typeInt64    = "int64"
	typeUint     = "uint"
	typeUint8    = "uint8"
	typeUint16   = "uint16"
	typeUint32   = "uint32"
	typeUint64   = "uint64"
	typeFloat32  = "float32"
	typeFloat64  = "float64"
	typeBool     = "bool"
	typeTime     = "time.Time"
	typeBytes    = "[]byte"
	typeBytesAlt = "[]uint8"
)

// Flatten converts an Atom to a struct-shaped map using the specified tag key.
// Field names are resolved from struct tags (e.g., "json", "bson", "db").
// Falls back to field name if tag is missing or "-".
func (a *Atom) Flatten(tagKey string) map[string]any {
	result := make(map[string]any)

	// Build field name -> storage key mapping
	keyMap := buildKeyMap(a.Spec, tagKey)

	// Flatten scalar types
	for name, val := range a.Strings {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.Ints {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.Uints {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.Floats {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.Bools {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.Times {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.Bytes {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}

	// Flatten pointer types (nil pointers are omitted)
	for name, val := range a.StringPtrs {
		if key := keyMap[name]; key != "" && val != nil {
			result[key] = *val
		}
	}
	for name, val := range a.IntPtrs {
		if key := keyMap[name]; key != "" && val != nil {
			result[key] = *val
		}
	}
	for name, val := range a.UintPtrs {
		if key := keyMap[name]; key != "" && val != nil {
			result[key] = *val
		}
	}
	for name, val := range a.FloatPtrs {
		if key := keyMap[name]; key != "" && val != nil {
			result[key] = *val
		}
	}
	for name, val := range a.BoolPtrs {
		if key := keyMap[name]; key != "" && val != nil {
			result[key] = *val
		}
	}
	for name, val := range a.TimePtrs {
		if key := keyMap[name]; key != "" && val != nil {
			result[key] = *val
		}
	}
	for name, val := range a.BytePtrs {
		if key := keyMap[name]; key != "" && val != nil {
			result[key] = *val
		}
	}

	// Flatten slice types
	for name, val := range a.StringSlices {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.IntSlices {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.UintSlices {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.FloatSlices {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.BoolSlices {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.TimeSlices {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}
	for name, val := range a.ByteSlices {
		if key := keyMap[name]; key != "" {
			result[key] = val
		}
	}

	// Flatten nested structs recursively
	for name := range a.Nested {
		if key := keyMap[name]; key != "" {
			nested := a.Nested[name]
			result[key] = nested.Flatten(tagKey)
		}
	}
	for name, nestedSlice := range a.NestedSlices {
		if key := keyMap[name]; key != "" {
			flat := make([]map[string]any, len(nestedSlice))
			for i := range nestedSlice {
				flat[i] = nestedSlice[i].Flatten(tagKey)
			}
			result[key] = flat
		}
	}

	return result
}

// Unflatten reconstructs an Atom from a struct-shaped map using spec and tag key.
func Unflatten(data map[string]any, spec Spec, tagKey string) *Atom {
	atom := &Atom{
		Spec:         spec,
		Nested:       make(map[string]Atom),
		NestedSlices: make(map[string][]Atom),
	}

	// Build storage key -> field metadata mapping
	fieldMap := buildFieldMap(spec, tagKey)

	for storageKey, val := range data {
		field, ok := fieldMap[storageKey]
		if !ok {
			continue
		}

		unflattenField(atom, field, val, tagKey)
	}

	return atom
}

// buildKeyMap creates a mapping from field name to storage key.
func buildKeyMap(spec Spec, tagKey string) map[string]string {
	keyMap := make(map[string]string, len(spec.Fields))
	for _, f := range spec.Fields {
		key := resolveStorageKey(f, tagKey)
		if key != "" {
			keyMap[f.Name] = key
		}
	}
	return keyMap
}

// buildFieldMap creates a mapping from storage key to field metadata.
func buildFieldMap(spec Spec, tagKey string) map[string]sentinel.FieldMetadata {
	fieldMap := make(map[string]sentinel.FieldMetadata, len(spec.Fields))
	for _, f := range spec.Fields {
		key := resolveStorageKey(f, tagKey)
		if key != "" {
			fieldMap[key] = f
		}
	}
	return fieldMap
}

// resolveStorageKey extracts the storage key from a field's tags.
// Returns empty string if field should be skipped.
func resolveStorageKey(f sentinel.FieldMetadata, tagKey string) string {
	if tag, ok := f.Tags[tagKey]; ok && tag != "" {
		// Parse tag value (e.g., "name,omitempty" -> "name")
		key := strings.Split(tag, ",")[0]
		if key == "-" {
			return ""
		}
		if key != "" {
			return key
		}
	}
	// Fall back to field name
	return f.Name
}

// unflattenField places a value into the appropriate typed map based on field metadata.
func unflattenField(atom *Atom, field sentinel.FieldMetadata, val any, tagKey string) {
	if val == nil {
		return
	}

	typeName := field.Type

	// Handle pointer types
	if strings.HasPrefix(typeName, "*") {
		unflattenPointer(atom, field.Name, strings.TrimPrefix(typeName, "*"), val)
		return
	}

	// Handle slice types
	if strings.HasPrefix(typeName, "[]") {
		elemType := strings.TrimPrefix(typeName, "[]")
		unflattenSlice(atom, field.Name, elemType, val, tagKey)
		return
	}

	// Handle scalar and struct types
	unflattenScalar(atom, field, val, tagKey)
}

// unflattenScalar handles scalar and struct types.
func unflattenScalar(atom *Atom, field sentinel.FieldMetadata, val any, tagKey string) {
	name := field.Name
	typeName := field.Type

	switch typeName {
	case typeString:
		if v, ok := val.(string); ok {
			if atom.Strings == nil {
				atom.Strings = make(map[string]string)
			}
			atom.Strings[name] = v
		}
	case typeInt, typeInt8, typeInt16, typeInt32, typeInt64:
		if atom.Ints == nil {
			atom.Ints = make(map[string]int64)
		}
		atom.Ints[name] = toInt64(val)
	case typeUint, typeUint8, typeUint16, typeUint32, typeUint64:
		if atom.Uints == nil {
			atom.Uints = make(map[string]uint64)
		}
		atom.Uints[name] = toUint64(val)
	case typeFloat32, typeFloat64:
		if atom.Floats == nil {
			atom.Floats = make(map[string]float64)
		}
		atom.Floats[name] = toFloat64(val)
	case typeBool:
		if v, ok := val.(bool); ok {
			if atom.Bools == nil {
				atom.Bools = make(map[string]bool)
			}
			atom.Bools[name] = v
		}
	case typeTime:
		if atom.Times == nil {
			atom.Times = make(map[string]time.Time)
		}
		atom.Times[name] = toTime(val)
	case typeBytesAlt, typeBytes:
		if atom.Bytes == nil {
			atom.Bytes = make(map[string][]byte)
		}
		atom.Bytes[name] = toBytes(val)
	default:
		// Assume struct type - try to unflatten as nested
		if field.Kind == sentinel.KindStruct {
			if nested, ok := val.(map[string]any); ok {
				// Find nested spec from relationships or use empty spec
				nestedSpec := findNestedSpec(atom.Spec, field)
				atom.Nested[name] = *Unflatten(nested, nestedSpec, tagKey)
			}
		}
	}
}

// unflattenPointer handles pointer types.
func unflattenPointer(atom *Atom, name, elemType string, val any) {
	switch elemType {
	case typeString:
		if v, ok := val.(string); ok {
			if atom.StringPtrs == nil {
				atom.StringPtrs = make(map[string]*string)
			}
			atom.StringPtrs[name] = &v
		}
	case typeInt, typeInt8, typeInt16, typeInt32, typeInt64:
		if atom.IntPtrs == nil {
			atom.IntPtrs = make(map[string]*int64)
		}
		v := toInt64(val)
		atom.IntPtrs[name] = &v
	case typeUint, typeUint8, typeUint16, typeUint32, typeUint64:
		if atom.UintPtrs == nil {
			atom.UintPtrs = make(map[string]*uint64)
		}
		v := toUint64(val)
		atom.UintPtrs[name] = &v
	case typeFloat32, typeFloat64:
		if atom.FloatPtrs == nil {
			atom.FloatPtrs = make(map[string]*float64)
		}
		v := toFloat64(val)
		atom.FloatPtrs[name] = &v
	case typeBool:
		if v, ok := val.(bool); ok {
			if atom.BoolPtrs == nil {
				atom.BoolPtrs = make(map[string]*bool)
			}
			atom.BoolPtrs[name] = &v
		}
	case typeTime:
		if atom.TimePtrs == nil {
			atom.TimePtrs = make(map[string]*time.Time)
		}
		v := toTime(val)
		atom.TimePtrs[name] = &v
	case typeBytesAlt, typeBytes:
		if atom.BytePtrs == nil {
			atom.BytePtrs = make(map[string]*[]byte)
		}
		v := toBytes(val)
		atom.BytePtrs[name] = &v
	}
}

// unflattenSlice handles slice types.
func unflattenSlice(atom *Atom, name, elemType string, val any, tagKey string) {
	switch elemType {
	case typeString:
		if atom.StringSlices == nil {
			atom.StringSlices = make(map[string][]string)
		}
		atom.StringSlices[name] = toStringSlice(val)
	case typeInt, typeInt8, typeInt16, typeInt32, typeInt64:
		if atom.IntSlices == nil {
			atom.IntSlices = make(map[string][]int64)
		}
		atom.IntSlices[name] = toInt64Slice(val)
	case typeUint, typeUint8, typeUint16, typeUint32, typeUint64:
		if atom.UintSlices == nil {
			atom.UintSlices = make(map[string][]uint64)
		}
		atom.UintSlices[name] = toUint64Slice(val)
	case typeFloat32, typeFloat64:
		if atom.FloatSlices == nil {
			atom.FloatSlices = make(map[string][]float64)
		}
		atom.FloatSlices[name] = toFloat64Slice(val)
	case typeBool:
		if atom.BoolSlices == nil {
			atom.BoolSlices = make(map[string][]bool)
		}
		atom.BoolSlices[name] = toBoolSlice(val)
	case typeTime:
		if atom.TimeSlices == nil {
			atom.TimeSlices = make(map[string][]time.Time)
		}
		atom.TimeSlices[name] = toTimeSlice(val)
	case typeBytesAlt, typeBytes:
		if atom.ByteSlices == nil {
			atom.ByteSlices = make(map[string][][]byte)
		}
		atom.ByteSlices[name] = toBytesSlice(val)
	default:
		// Assume slice of structs
		if arr, ok := val.([]any); ok {
			nested := make([]Atom, 0, len(arr))
			for _, item := range arr {
				if m, ok := item.(map[string]any); ok {
					// TODO: resolve nested spec properly
					nested = append(nested, *Unflatten(m, Spec{}, tagKey))
				}
			}
			atom.NestedSlices[name] = nested
		}
	}
}

// findNestedSpec attempts to find the spec for a nested struct field.
func findNestedSpec(_ Spec, _ sentinel.FieldMetadata) Spec {
	// For now, return empty spec. Nested specs would need to be
	// passed in or resolved from a registry.
	return Spec{}
}

// Type conversion helpers

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	case float32:
		return int64(n)
	default:
		return 0
	}
}

func toUint64(v any) uint64 {
	switch n := v.(type) {
	case uint64:
		return n
	case uint:
		return uint64(n)
	case uint32:
		return uint64(n)
	case float64:
		return uint64(n) //nolint:gosec // Intentional conversion for JSON interop
	case float32:
		return uint64(n) //nolint:gosec // Intentional conversion for JSON interop
	case int64:
		return uint64(n) //nolint:gosec // Intentional conversion, caller validates
	case int:
		return uint64(n) //nolint:gosec // Intentional conversion, caller validates
	default:
		return 0
	}
}

func toFloat64(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}

func toTime(v any) time.Time {
	switch t := v.(type) {
	case time.Time:
		return t
	case string:
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			return parsed
		}
		if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func toBytes(v any) []byte {
	switch b := v.(type) {
	case []byte:
		return b
	case string:
		// Try base64 decode
		if decoded, err := base64.StdEncoding.DecodeString(b); err == nil {
			return decoded
		}
		return []byte(b)
	}
	return nil
}

func toStringSlice(v any) []string {
	if arr, ok := v.([]any); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	if arr, ok := v.([]string); ok {
		return arr
	}
	return nil
}

func toInt64Slice(v any) []int64 {
	if arr, ok := v.([]any); ok {
		result := make([]int64, 0, len(arr))
		for _, item := range arr {
			result = append(result, toInt64(item))
		}
		return result
	}
	if arr, ok := v.([]int64); ok {
		return arr
	}
	return nil
}

func toUint64Slice(v any) []uint64 {
	if arr, ok := v.([]any); ok {
		result := make([]uint64, 0, len(arr))
		for _, item := range arr {
			result = append(result, toUint64(item))
		}
		return result
	}
	if arr, ok := v.([]uint64); ok {
		return arr
	}
	return nil
}

func toFloat64Slice(v any) []float64 {
	if arr, ok := v.([]any); ok {
		result := make([]float64, 0, len(arr))
		for _, item := range arr {
			result = append(result, toFloat64(item))
		}
		return result
	}
	if arr, ok := v.([]float64); ok {
		return arr
	}
	return nil
}

func toBoolSlice(v any) []bool {
	if arr, ok := v.([]any); ok {
		result := make([]bool, 0, len(arr))
		for _, item := range arr {
			if b, ok := item.(bool); ok {
				result = append(result, b)
			}
		}
		return result
	}
	if arr, ok := v.([]bool); ok {
		return arr
	}
	return nil
}

func toTimeSlice(v any) []time.Time {
	if arr, ok := v.([]any); ok {
		result := make([]time.Time, 0, len(arr))
		for _, item := range arr {
			result = append(result, toTime(item))
		}
		return result
	}
	if arr, ok := v.([]time.Time); ok {
		return arr
	}
	return nil
}

func toBytesSlice(v any) [][]byte {
	if arr, ok := v.([]any); ok {
		result := make([][]byte, 0, len(arr))
		for _, item := range arr {
			result = append(result, toBytes(item))
		}
		return result
	}
	if arr, ok := v.([][]byte); ok {
		return arr
	}
	return nil
}
