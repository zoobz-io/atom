package atom

import (
	"reflect"
	"sync"

	"github.com/zoobzio/sentinel"
)

var (
	registry   = make(map[reflect.Type]*reflectAtomizer)
	registryMu sync.RWMutex
)

// Use registers and returns an Atomizer for type T.
// First call builds the atomizer; subsequent calls return the cached instance.
// Returns an error if the type contains unsupported field types.
func Use[T any]() (*Atomizer[T], error) {
	typ := reflect.TypeFor[T]()

	registryMu.RLock()
	if ra, ok := registry[typ]; ok {
		if err, hasErr := registrationErrors[typ]; hasErr {
			registryMu.RUnlock()
			return nil, err
		}
		registryMu.RUnlock()
		return &Atomizer[T]{inner: ra}, nil
	}
	registryMu.RUnlock()

	// Get sentinel spec while we have the type parameter
	spec := sentinel.Scan[T]()

	registryMu.Lock()
	defer registryMu.Unlock()

	// Double-check after acquiring write lock
	if ra, ok := registry[typ]; ok {
		if err, hasErr := registrationErrors[typ]; hasErr {
			return nil, err
		}
		return &Atomizer[T]{inner: ra}, nil
	}

	ra, err := buildReflectAtomizerWithSpec(typ, spec)
	if err != nil {
		registrationErrors[typ] = err
		return nil, err
	}
	registry[typ] = ra

	// Check for errors in nested types that were registered during field plan building
	for nestedType, nestedErr := range registrationErrors {
		if nestedErr != nil && usesNestedType(ra.plan, nestedType) {
			return nil, nestedErr
		}
	}

	return &Atomizer[T]{inner: ra}, nil
}

// usesNestedType checks if any field plan references the given nested type.
func usesNestedType(plans []fieldPlan, typ reflect.Type) bool {
	for i := range plans {
		if plans[i].nested != nil && plans[i].nested.typ == typ {
			return true
		}
	}
	return false
}

// registrationError holds an error that occurred during type registration.
// This is stored in the registry to propagate errors for nested types.
var registrationErrors = make(map[reflect.Type]error)

// ensureRegistered recursively ensures a type is in the registry.
// Must be called with registryMu held.
// Returns the atomizer and any error that occurred during registration.
func ensureRegistered(typ reflect.Type) *reflectAtomizer {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if ra, ok := registry[typ]; ok {
		return ra
	}

	// Register shell first to handle circular references
	ra := &reflectAtomizer{typ: typ}
	registry[typ] = ra

	// Try to get spec from sentinel (may have been scanned already)
	typeName := typ.Name()
	if typ.PkgPath() != "" {
		typeName = typ.PkgPath() + "." + typ.Name()
	}
	if spec, ok := sentinel.Lookup(typeName); ok {
		ra.spec = spec
	}

	// Build field plan (may recursively call ensureRegistered)
	plan, err := buildFieldPlan(typ)
	if err != nil {
		registrationErrors[typ] = err
		return ra
	}
	ra.plan = plan
	ra.tableSet = computeTableSet(ra.plan)
	ra.hasAtomizable = implementsAtomizable(typ)
	ra.hasDeatomizable = implementsDeatomizable(typ)

	return ra
}
