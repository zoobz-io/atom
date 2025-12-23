package atom

import (
	"github.com/zoobzio/sentinel"
)

func init() {
	sentinel.Tag("atom")
}

// Atomizer provides typed bidirectional resolution for any type T.
// T must implement Validator to ensure data integrity before storage.
type Atomizer[T Validator] struct { //nolint:govet // field order matches logical grouping
	metadata  sentinel.Metadata
	fields    []FieldDescriptor
	fieldMap  map[string]TableType
	atomize   Atomize[T]
	deatomize Deatomize[T]
}

// New creates an Atomizer for type T using the provided callbacks.
// Inspects T at construction time to build field metadata.
func New[T Validator](atomize Atomize[T], deatomize Deatomize[T]) *Atomizer[T] {
	metadata := sentinel.Inspect[T]()
	fields, fieldMap := buildFieldDescriptors(metadata)

	return &Atomizer[T]{
		metadata:  metadata,
		fields:    fields,
		fieldMap:  fieldMap,
		atomize:   atomize,
		deatomize: deatomize,
	}
}

// buildFieldDescriptors analyzes sentinel metadata to map fields to tables.
func buildFieldDescriptors(meta sentinel.Metadata) (fields []FieldDescriptor, fieldMap map[string]TableType) {
	fields = make([]FieldDescriptor, 0, len(meta.Fields))
	fieldMap = make(map[string]TableType)

	for _, f := range meta.Fields {
		if atomTag, ok := f.Tags["atom"]; ok && atomTag == "-" {
			continue
		}

		table := tableFromType(f.Type)
		if table == "" {
			continue
		}

		name := f.Name
		if atomTag, ok := f.Tags["atom"]; ok && atomTag != "" && atomTag != "id" {
			name = atomTag
		}

		fd := FieldDescriptor{Name: name, Table: table}
		fields = append(fields, fd)
		fieldMap[name] = table
	}

	return fields, fieldMap
}

// tableFromType maps Go type strings to TableType.
func tableFromType(goType string) TableType {
	switch goType {
	case "string":
		return TableStrings
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return TableInts
	case "float32", "float64":
		return TableFloats
	case "bool":
		return TableBools
	case "time.Time":
		return TableTimes
	default:
		return ""
	}
}

// Metadata returns the sentinel.Metadata for type T.
func (a *Atomizer[T]) Metadata() sentinel.Metadata {
	return a.metadata
}

// Fields returns all field descriptors.
func (a *Atomizer[T]) Fields() []FieldDescriptor {
	return a.fields
}

// FieldsIn returns field names stored in the given table.
func (a *Atomizer[T]) FieldsIn(table TableType) []string {
	var result []string
	for _, f := range a.fields {
		if f.Table == table {
			result = append(result, f.Name)
		}
	}
	return result
}

// TableFor returns the table type for a field name.
func (a *Atomizer[T]) TableFor(field string) (TableType, bool) {
	table, ok := a.fieldMap[field]
	return table, ok
}

// Atomize converts an object to its atomic representation.
func (a *Atomizer[T]) Atomize(obj *T) (Atoms, error) {
	if err := (*obj).Validate(); err != nil {
		return Atoms{}, err
	}
	atoms := a.atomize(obj)
	if atoms.ID == "" {
		return Atoms{}, ErrMissingID
	}
	return atoms, nil
}

// Deatomize reconstructs an object from its atomic representation.
func (a *Atomizer[T]) Deatomize(atoms Atoms) (*T, error) {
	return a.deatomize(atoms)
}
