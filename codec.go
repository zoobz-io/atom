package atom

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// timeFormat specifies how time.Time fields are encoded/decoded.
type timeFormat int

const (
	// timeFormatRFC3339 encodes time as RFC3339 string (default).
	timeFormatRFC3339 timeFormat = iota
	// timeFormatUnix encodes time as Unix timestamp (seconds).
	timeFormatUnix
	// timeFormatUnixMilli encodes time as Unix timestamp (milliseconds).
	timeFormatUnixMilli
	// timeFormatUnixNano encodes time as Unix timestamp (nanoseconds).
	timeFormatUnixNano
	// timeFormatNative preserves time.Time as-is (for formats with native time support).
	timeFormatNative
)

// parseTimeFormat converts a time tag value to a timeFormat.
func parseTimeFormat(tag string) timeFormat {
	switch tag {
	case "unix":
		return timeFormatUnix
	case "unixmilli":
		return timeFormatUnixMilli
	case "unixnano":
		return timeFormatUnixNano
	case "native":
		return timeFormatNative
	default:
		return timeFormatRFC3339
	}
}

// Codec encodes Atoms to maps and decodes maps to Atoms using db tags.
type Codec struct {
	byTag    map[string]*codecPlan // db tag → plan (for decode)
	byField  map[string]*codecPlan // field name → plan (for encode)
	tableSet map[Table]int
	spec     Spec
}

// codecPlan describes how to encode/decode a single field.
type codecPlan struct {
	nested     *Codec
	fieldName  string
	tag        string
	table      Table
	kind       fieldKind
	timeFormat timeFormat
}

var (
	codecMu sync.RWMutex
	codecs  = make(map[reflect.Type]*Codec)
)

// CodecFor retrieves the Codec for a registered type.
// Returns nil, false if the type has not been registered via Use[T]().
func CodecFor(spec Spec) (*Codec, bool) {
	codecMu.RLock()
	c, ok := codecs[spec.ReflectType]
	codecMu.RUnlock()
	return c, ok
}

// registerCodec adds a codec to the registry.
func registerCodec(typ reflect.Type, codec *Codec) {
	codecMu.Lock()
	codecs[typ] = codec
	codecMu.Unlock()
}

// buildCodec creates a Codec from field plans and spec.
func buildCodec(plans []fieldPlan, spec Spec) (*Codec, error) {
	// Use a building set to detect circular references
	building := make(map[reflect.Type]*Codec)
	return buildCodecWithCycleDetection(plans, spec, building)
}

// buildCodecWithCycleDetection builds a codec while tracking types being built.
func buildCodecWithCycleDetection(plans []fieldPlan, spec Spec, building map[reflect.Type]*Codec) (*Codec, error) {
	c := &Codec{
		spec:     spec,
		byTag:    make(map[string]*codecPlan),
		byField:  make(map[string]*codecPlan),
		tableSet: make(map[Table]int),
	}

	// Register early to handle circular references
	if spec.ReflectType != nil {
		building[spec.ReflectType] = c
	}

	// Build a map of field name to tags for this type
	fieldTags := make(map[string]map[string]string, len(spec.Fields))
	for _, f := range spec.Fields {
		fieldTags[f.Name] = f.Tags
	}

	if err := c.buildPlans(plans, fieldTags, building); err != nil {
		return nil, err
	}

	return c, nil
}

// buildPlans builds codec plans from field plans.
func (c *Codec) buildPlans(plans []fieldPlan, fieldTags map[string]map[string]string, building map[reflect.Type]*Codec) error {
	for i := range plans {
		fp := &plans[i]
		tags := fieldTags[fp.name]
		dbTag, hasDB := tags["db"]
		timeTag := tags["time"]

		switch fp.kind {
		case kindScalar, kindPointer, kindSlice:
			if !hasDB {
				continue // No db tag, skip
			}
			plan := &codecPlan{
				fieldName:  fp.name,
				tag:        dbTag,
				table:      fp.table,
				kind:       fp.kind,
				timeFormat: parseTimeFormat(timeTag),
			}
			c.byTag[dbTag] = plan
			c.byField[fp.name] = plan
			c.tableSet[fp.table]++

		case kindNested, kindNestedPtr:
			if fp.nested == nil {
				continue
			}
			// Use db tag if present, otherwise use field name
			tagKey := dbTag
			if !hasDB {
				tagKey = fp.name
			}
			// Get or build nested codec (with cycle detection)
			nestedCodec, err := getOrBuildNestedCodec(fp.nested, building)
			if err != nil {
				return err
			}
			plan := &codecPlan{
				fieldName: fp.name,
				tag:       tagKey,
				kind:      fp.kind,
				nested:    nestedCodec,
			}
			c.byTag[tagKey] = plan
			c.byField[fp.name] = plan

		case kindNestedSlice:
			if fp.nested == nil {
				continue
			}
			// Use db tag if present, otherwise use field name
			tagKey := dbTag
			if !hasDB {
				tagKey = fp.name
			}
			// Get or build nested codec (with cycle detection)
			nestedCodec, err := getOrBuildNestedCodec(fp.nested, building)
			if err != nil {
				return err
			}
			plan := &codecPlan{
				fieldName: fp.name,
				tag:       tagKey,
				kind:      kindNestedSlice,
				nested:    nestedCodec,
			}
			c.byTag[tagKey] = plan
			c.byField[fp.name] = plan
		}
	}

	return nil
}

// getOrBuildNestedCodec retrieves or builds a codec for a nested type.
func getOrBuildNestedCodec(ra *reflectAtomizer, building map[reflect.Type]*Codec) (*Codec, error) {
	// Check if already being built (circular reference)
	if existing, ok := building[ra.typ]; ok {
		return existing, nil
	}

	// Check global registry
	codecMu.RLock()
	if existing, ok := codecs[ra.typ]; ok {
		codecMu.RUnlock()
		return existing, nil
	}
	codecMu.RUnlock()

	// Build new codec
	return buildCodecWithCycleDetection(ra.plan, ra.spec, building)
}

// Spec returns the spec for this codec.
func (c *Codec) Spec() Spec {
	return c.spec
}

// EncodeMap converts an Atom to a map using db tags as keys.
func (c *Codec) EncodeMap(a *Atom) map[string]any {
	if a == nil {
		return nil
	}

	result := make(map[string]any, len(c.byField))

	for _, plan := range c.byField {
		val := c.encodeField(a, plan)
		if val != nil {
			result[plan.tag] = val
		}
	}

	return result
}

// encodeField extracts a field value from an Atom.
func (c *Codec) encodeField(a *Atom, plan *codecPlan) any {
	switch plan.kind {
	case kindScalar:
		return c.encodeScalar(a, plan)
	case kindPointer:
		return c.encodePointer(a, plan)
	case kindSlice:
		return c.encodeSlice(a, plan)
	case kindNested, kindNestedPtr:
		return c.encodeNested(a, plan)
	case kindNestedSlice:
		return c.encodeNestedSlice(a, plan)
	}
	return nil
}

func (*Codec) encodeScalar(a *Atom, plan *codecPlan) any {
	switch plan.table {
	case TableStrings:
		if v, ok := a.Strings[plan.fieldName]; ok {
			return v
		}
	case TableInts:
		if v, ok := a.Ints[plan.fieldName]; ok {
			return v
		}
	case TableUints:
		if v, ok := a.Uints[plan.fieldName]; ok {
			return v
		}
	case TableFloats:
		if v, ok := a.Floats[plan.fieldName]; ok {
			return v
		}
	case TableBools:
		if v, ok := a.Bools[plan.fieldName]; ok {
			return v
		}
	case TableTimes:
		if v, ok := a.Times[plan.fieldName]; ok {
			return encodeTime(v, plan.timeFormat)
		}
	case TableBytes:
		if v, ok := a.Bytes[plan.fieldName]; ok {
			return v
		}
	}
	return nil
}

func (*Codec) encodePointer(a *Atom, plan *codecPlan) any {
	switch plan.table {
	case TableStringPtrs:
		if v, ok := a.StringPtrs[plan.fieldName]; ok && v != nil {
			return *v
		}
	case TableIntPtrs:
		if v, ok := a.IntPtrs[plan.fieldName]; ok && v != nil {
			return *v
		}
	case TableUintPtrs:
		if v, ok := a.UintPtrs[plan.fieldName]; ok && v != nil {
			return *v
		}
	case TableFloatPtrs:
		if v, ok := a.FloatPtrs[plan.fieldName]; ok && v != nil {
			return *v
		}
	case TableBoolPtrs:
		if v, ok := a.BoolPtrs[plan.fieldName]; ok && v != nil {
			return *v
		}
	case TableTimePtrs:
		if v, ok := a.TimePtrs[plan.fieldName]; ok && v != nil {
			return encodeTime(*v, plan.timeFormat)
		}
	case TableBytePtrs:
		if v, ok := a.BytePtrs[plan.fieldName]; ok && v != nil {
			return *v
		}
	}
	return nil
}

func (*Codec) encodeSlice(a *Atom, plan *codecPlan) any {
	switch plan.table {
	case TableStringSlices:
		if v, ok := a.StringSlices[plan.fieldName]; ok {
			return v
		}
	case TableIntSlices:
		if v, ok := a.IntSlices[plan.fieldName]; ok {
			return v
		}
	case TableUintSlices:
		if v, ok := a.UintSlices[plan.fieldName]; ok {
			return v
		}
	case TableFloatSlices:
		if v, ok := a.FloatSlices[plan.fieldName]; ok {
			return v
		}
	case TableBoolSlices:
		if v, ok := a.BoolSlices[plan.fieldName]; ok {
			return v
		}
	case TableTimeSlices:
		if v, ok := a.TimeSlices[plan.fieldName]; ok {
			result := make([]any, len(v))
			for i, t := range v {
				result[i] = encodeTime(t, plan.timeFormat)
			}
			return result
		}
	case TableByteSlices:
		if v, ok := a.ByteSlices[plan.fieldName]; ok {
			return v
		}
	}
	return nil
}

func (*Codec) encodeNested(a *Atom, plan *codecPlan) any {
	if nested, ok := a.Nested[plan.fieldName]; ok {
		return plan.nested.EncodeMap(&nested)
	}
	return nil
}

func (*Codec) encodeNestedSlice(a *Atom, plan *codecPlan) any {
	if nestedSlice, ok := a.NestedSlices[plan.fieldName]; ok {
		result := make([]map[string]any, len(nestedSlice))
		for i := range nestedSlice {
			result[i] = plan.nested.EncodeMap(&nestedSlice[i])
		}
		return result
	}
	return nil
}

// encodeTime converts a time.Time to the appropriate format.
func encodeTime(t time.Time, format timeFormat) any {
	switch format {
	case timeFormatUnix:
		return t.Unix()
	case timeFormatUnixMilli:
		return t.UnixMilli()
	case timeFormatUnixNano:
		return t.UnixNano()
	case timeFormatNative:
		return t
	default: // timeFormatRFC3339
		return t.Format(time.RFC3339Nano)
	}
}

// DecodeMap converts a map to an Atom using db tags as keys.
func (c *Codec) DecodeMap(data map[string]any) (*Atom, error) {
	if data == nil {
		return nil, nil
	}

	atom := allocateAtom(c.spec, c.tableSet)

	for key, val := range data {
		plan, ok := c.byTag[key]
		if !ok {
			continue // Unknown field, skip
		}

		if err := c.decodeField(atom, plan, val); err != nil {
			return nil, fmt.Errorf("field %q: %w", plan.fieldName, err)
		}
	}

	return atom, nil
}

// decodeField places a value into the appropriate Atom field.
func (c *Codec) decodeField(atom *Atom, plan *codecPlan, val any) error {
	if val == nil {
		// Null handling: pointer fields get nil, scalar fields skip
		if plan.kind == kindPointer {
			return c.decodeNullPointer(atom, plan)
		}
		return nil
	}

	switch plan.kind {
	case kindScalar:
		return c.decodeScalar(atom, plan, val)
	case kindPointer:
		return c.decodePointer(atom, plan, val)
	case kindSlice:
		return c.decodeSlice(atom, plan, val)
	case kindNested, kindNestedPtr:
		return c.decodeNested(atom, plan, val)
	case kindNestedSlice:
		return c.decodeNestedSlice(atom, plan, val)
	}
	return nil
}

func (*Codec) decodeNullPointer(atom *Atom, plan *codecPlan) error {
	switch plan.table {
	case TableStringPtrs:
		if atom.StringPtrs == nil {
			atom.StringPtrs = make(map[string]*string)
		}
		atom.StringPtrs[plan.fieldName] = nil
	case TableIntPtrs:
		if atom.IntPtrs == nil {
			atom.IntPtrs = make(map[string]*int64)
		}
		atom.IntPtrs[plan.fieldName] = nil
	case TableUintPtrs:
		if atom.UintPtrs == nil {
			atom.UintPtrs = make(map[string]*uint64)
		}
		atom.UintPtrs[plan.fieldName] = nil
	case TableFloatPtrs:
		if atom.FloatPtrs == nil {
			atom.FloatPtrs = make(map[string]*float64)
		}
		atom.FloatPtrs[plan.fieldName] = nil
	case TableBoolPtrs:
		if atom.BoolPtrs == nil {
			atom.BoolPtrs = make(map[string]*bool)
		}
		atom.BoolPtrs[plan.fieldName] = nil
	case TableTimePtrs:
		if atom.TimePtrs == nil {
			atom.TimePtrs = make(map[string]*time.Time)
		}
		atom.TimePtrs[plan.fieldName] = nil
	case TableBytePtrs:
		if atom.BytePtrs == nil {
			atom.BytePtrs = make(map[string]*[]byte)
		}
		atom.BytePtrs[plan.fieldName] = nil
	}
	return nil
}

func (*Codec) decodeScalar(atom *Atom, plan *codecPlan, val any) error {
	switch plan.table {
	case TableStrings:
		if v, ok := val.(string); ok {
			if atom.Strings == nil {
				atom.Strings = make(map[string]string)
			}
			atom.Strings[plan.fieldName] = v
		}
	case TableInts:
		if atom.Ints == nil {
			atom.Ints = make(map[string]int64)
		}
		atom.Ints[plan.fieldName] = toInt64(val)
	case TableUints:
		if atom.Uints == nil {
			atom.Uints = make(map[string]uint64)
		}
		atom.Uints[plan.fieldName] = toUint64(val)
	case TableFloats:
		if atom.Floats == nil {
			atom.Floats = make(map[string]float64)
		}
		atom.Floats[plan.fieldName] = toFloat64(val)
	case TableBools:
		if v, ok := val.(bool); ok {
			if atom.Bools == nil {
				atom.Bools = make(map[string]bool)
			}
			atom.Bools[plan.fieldName] = v
		}
	case TableTimes:
		if atom.Times == nil {
			atom.Times = make(map[string]time.Time)
		}
		atom.Times[plan.fieldName] = decodeTime(val, plan.timeFormat)
	case TableBytes:
		if atom.Bytes == nil {
			atom.Bytes = make(map[string][]byte)
		}
		atom.Bytes[plan.fieldName] = toBytes(val)
	}
	return nil
}

func (*Codec) decodePointer(atom *Atom, plan *codecPlan, val any) error {
	switch plan.table {
	case TableStringPtrs:
		if v, ok := val.(string); ok {
			if atom.StringPtrs == nil {
				atom.StringPtrs = make(map[string]*string)
			}
			atom.StringPtrs[plan.fieldName] = &v
		}
	case TableIntPtrs:
		if atom.IntPtrs == nil {
			atom.IntPtrs = make(map[string]*int64)
		}
		v := toInt64(val)
		atom.IntPtrs[plan.fieldName] = &v
	case TableUintPtrs:
		if atom.UintPtrs == nil {
			atom.UintPtrs = make(map[string]*uint64)
		}
		v := toUint64(val)
		atom.UintPtrs[plan.fieldName] = &v
	case TableFloatPtrs:
		if atom.FloatPtrs == nil {
			atom.FloatPtrs = make(map[string]*float64)
		}
		v := toFloat64(val)
		atom.FloatPtrs[plan.fieldName] = &v
	case TableBoolPtrs:
		if v, ok := val.(bool); ok {
			if atom.BoolPtrs == nil {
				atom.BoolPtrs = make(map[string]*bool)
			}
			atom.BoolPtrs[plan.fieldName] = &v
		}
	case TableTimePtrs:
		if atom.TimePtrs == nil {
			atom.TimePtrs = make(map[string]*time.Time)
		}
		v := decodeTime(val, plan.timeFormat)
		atom.TimePtrs[plan.fieldName] = &v
	case TableBytePtrs:
		if atom.BytePtrs == nil {
			atom.BytePtrs = make(map[string]*[]byte)
		}
		v := toBytes(val)
		atom.BytePtrs[plan.fieldName] = &v
	}
	return nil
}

func (*Codec) decodeSlice(atom *Atom, plan *codecPlan, val any) error {
	switch plan.table {
	case TableStringSlices:
		if atom.StringSlices == nil {
			atom.StringSlices = make(map[string][]string)
		}
		atom.StringSlices[plan.fieldName] = toStringSlice(val)
	case TableIntSlices:
		if atom.IntSlices == nil {
			atom.IntSlices = make(map[string][]int64)
		}
		atom.IntSlices[plan.fieldName] = toInt64Slice(val)
	case TableUintSlices:
		if atom.UintSlices == nil {
			atom.UintSlices = make(map[string][]uint64)
		}
		atom.UintSlices[plan.fieldName] = toUint64Slice(val)
	case TableFloatSlices:
		if atom.FloatSlices == nil {
			atom.FloatSlices = make(map[string][]float64)
		}
		atom.FloatSlices[plan.fieldName] = toFloat64Slice(val)
	case TableBoolSlices:
		if atom.BoolSlices == nil {
			atom.BoolSlices = make(map[string][]bool)
		}
		atom.BoolSlices[plan.fieldName] = toBoolSlice(val)
	case TableTimeSlices:
		if atom.TimeSlices == nil {
			atom.TimeSlices = make(map[string][]time.Time)
		}
		atom.TimeSlices[plan.fieldName] = toTimeSlice(val, plan.timeFormat)
	case TableByteSlices:
		if atom.ByteSlices == nil {
			atom.ByteSlices = make(map[string][][]byte)
		}
		atom.ByteSlices[plan.fieldName] = toBytesSlice(val)
	}
	return nil
}

func (*Codec) decodeNested(atom *Atom, plan *codecPlan, val any) error {
	m, ok := val.(map[string]any)
	if !ok {
		return nil
	}
	nested, err := plan.nested.DecodeMap(m)
	if err != nil {
		return err
	}
	if atom.Nested == nil {
		atom.Nested = make(map[string]Atom)
	}
	atom.Nested[plan.fieldName] = *nested
	return nil
}

func (*Codec) decodeNestedSlice(atom *Atom, plan *codecPlan, val any) error {
	arr, ok := val.([]any)
	if !ok {
		// Try typed slice
		if mapSlice, ok := val.([]map[string]any); ok {
			atoms := make([]Atom, 0, len(mapSlice))
			for _, m := range mapSlice {
				nested, err := plan.nested.DecodeMap(m)
				if err != nil {
					return err
				}
				atoms = append(atoms, *nested)
			}
			if atom.NestedSlices == nil {
				atom.NestedSlices = make(map[string][]Atom)
			}
			atom.NestedSlices[plan.fieldName] = atoms
		}
		return nil
	}

	atoms := make([]Atom, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			nested, err := plan.nested.DecodeMap(m)
			if err != nil {
				return err
			}
			atoms = append(atoms, *nested)
		}
	}
	if atom.NestedSlices == nil {
		atom.NestedSlices = make(map[string][]Atom)
	}
	atom.NestedSlices[plan.fieldName] = atoms
	return nil
}

// decodeTime converts a value to time.Time based on format.
func decodeTime(val any, format timeFormat) time.Time {
	switch format {
	case timeFormatUnix:
		if v := toInt64(val); v != 0 {
			return time.Unix(v, 0)
		}
	case timeFormatUnixMilli:
		if v := toInt64(val); v != 0 {
			return time.UnixMilli(v)
		}
	case timeFormatUnixNano:
		if v := toInt64(val); v != 0 {
			return time.Unix(0, v)
		}
	case timeFormatNative:
		if t, ok := val.(time.Time); ok {
			return t
		}
	default: // timeFormatRFC3339
		return toTime(val)
	}
	return time.Time{}
}

// Type coercion helpers

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

func toTimeSlice(v any, format timeFormat) []time.Time {
	if arr, ok := v.([]any); ok {
		result := make([]time.Time, 0, len(arr))
		for _, item := range arr {
			result = append(result, decodeTime(item, format))
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
