package atom

import (
	"github.com/zoobzio/sentinel"
)

func init() {
	sentinel.Tag("atom")
}

// Atomizer provides typed bidirectional resolution for type T.
type Atomizer[T any] struct {
	inner *reflectAtomizer
}

// Atomize converts an object to its atomic representation.
// If T implements Atomizable, that method is used instead of reflection.
func (a *Atomizer[T]) Atomize(obj *T) *Atom {
	atom := a.inner.newAtom()

	// Check for Atomizable interface (enables code generation)
	if a.inner.hasAtomizable {
		if az, ok := any(obj).(Atomizable); ok {
			az.Atomize(atom)
			return atom
		}
	}

	a.inner.atomize(obj, atom)
	return atom
}

// Deatomize reconstructs an object from an Atom.
// If T implements Deatomizable, that method is used instead of reflection.
func (a *Atomizer[T]) Deatomize(atom *Atom) (*T, error) {
	obj := new(T)

	// Check for Deatomizable interface (enables code generation)
	if a.inner.hasDeatomizable {
		if dz, ok := any(obj).(Deatomizable); ok {
			if err := dz.Deatomize(atom); err != nil {
				return nil, err
			}
			return obj, nil
		}
	}

	if err := a.inner.deatomize(atom, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// NewAtom creates an Atom with only the maps needed for this type.
func (a *Atomizer[T]) NewAtom() *Atom {
	return a.inner.newAtom()
}

// Spec returns the type specification for type T.
func (a *Atomizer[T]) Spec() Spec {
	return a.inner.spec
}

// Fields returns all fields with their table mappings.
func (a *Atomizer[T]) Fields() []Field {
	var fields []Field
	for i := range a.inner.plan {
		fp := &a.inner.plan[i]
		if fp.table != "" {
			fields = append(fields, Field{
				Name:  fp.name,
				Table: fp.table,
			})
		}
	}
	return fields
}

// FieldsIn returns field names stored in the given table.
func (a *Atomizer[T]) FieldsIn(table Table) []string {
	var result []string
	for i := range a.inner.plan {
		fp := &a.inner.plan[i]
		if fp.table == table {
			result = append(result, fp.name)
		}
	}
	return result
}

// TableFor returns the table for a field name.
func (a *Atomizer[T]) TableFor(field string) (Table, bool) {
	for i := range a.inner.plan {
		fp := &a.inner.plan[i]
		if fp.name == field {
			return fp.table, fp.table != ""
		}
	}
	return "", false
}
