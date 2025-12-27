package atom

import (
	"reflect"
	"sync"
	"testing"
)

func TestUseBasic(t *testing.T) {
	type Simple struct {
		Name string
		Age  int
	}

	az, err := Use[Simple]()
	if err != nil {
		t.Fatalf("Use error: %v", err)
	}
	if az == nil {
		t.Fatal("atomizer should not be nil")
	}
}

func TestUseCaching(t *testing.T) {
	type Cached struct {
		Value string
	}

	az1, err := Use[Cached]()
	if err != nil {
		t.Fatalf("first Use error: %v", err)
	}

	az2, err := Use[Cached]()
	if err != nil {
		t.Fatalf("second Use error: %v", err)
	}

	// Should return same inner atomizer
	if az1.inner != az2.inner {
		t.Error("expected same inner atomizer for repeated Use calls")
	}
}

func TestUseConcurrent(t *testing.T) {
	type Concurrent struct {
		Y string
		X int
	}

	var wg sync.WaitGroup
	results := make(chan *Atomizer[Concurrent], 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			az, err := Use[Concurrent]()
			if err != nil {
				errors <- err
				return
			}
			results <- az
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	for err := range errors {
		t.Errorf("concurrent Use error: %v", err)
	}

	// All should have same inner
	var first *reflectAtomizer
	for az := range results {
		if first == nil {
			first = az.inner
		} else if az.inner != first {
			t.Error("concurrent calls returned different inner atomizers")
		}
	}
}

func TestUseUnsupportedType(t *testing.T) {
	type WithMap struct {
		M map[string]int
	}

	_, err := Use[WithMap]()
	if err == nil {
		t.Error("expected error for unsupported map field")
	}
}

func TestUseUnsupportedTypeCached(t *testing.T) {
	type BadType struct {
		C chan int
	}

	_, err1 := Use[BadType]()
	if err1 == nil {
		t.Fatal("expected error for channel field")
	}

	// Second call should return cached error
	_, err2 := Use[BadType]()
	if err2 == nil {
		t.Error("expected cached error on second call")
	}
}

func TestUseNestedType(t *testing.T) {
	type Inner struct {
		Value int
	}
	type Outer struct {
		Nested Inner
	}

	az, err := Use[Outer]()
	if err != nil {
		t.Fatalf("Use error: %v", err)
	}
	if az == nil {
		t.Fatal("atomizer should not be nil")
	}

	// Verify nested atomizer is set up
	if len(az.inner.plan) != 1 {
		t.Fatalf("got %d plans, want 1", len(az.inner.plan))
	}
	if az.inner.plan[0].nested == nil {
		t.Error("nested atomizer should not be nil")
	}
}

func TestUseNestedWithBadField(t *testing.T) {
	type BadInner struct {
		M map[string]int
	}
	type OuterBad struct {
		Nested BadInner
	}

	_, err := Use[OuterBad]()
	if err == nil {
		t.Error("expected error for nested type with map field")
	}
}

func TestUsesNestedType(t *testing.T) {
	type Inner struct {
		X int
	}

	innerType := reflect.TypeFor[Inner]()

	// Create a plan that references the nested type
	plans := []fieldPlan{
		{
			name:   "Nested",
			kind:   kindNested,
			nested: &reflectAtomizer{typ: innerType},
		},
	}

	if !usesNestedType(plans, innerType) {
		t.Error("expected usesNestedType to return true")
	}

	// Test with different type
	otherType := reflect.TypeFor[string]()
	if usesNestedType(plans, otherType) {
		t.Error("expected usesNestedType to return false for non-matching type")
	}
}

func TestUsesNestedTypeEmpty(t *testing.T) {
	var plans []fieldPlan
	if usesNestedType(plans, reflect.TypeFor[int]()) {
		t.Error("expected false for empty plans")
	}
}

func TestUsesNestedTypeNoNested(t *testing.T) {
	plans := []fieldPlan{
		{name: "Scalar", kind: kindScalar, table: TableStrings},
	}
	if usesNestedType(plans, reflect.TypeFor[int]()) {
		t.Error("expected false when no nested fields")
	}
}

func TestEnsureRegisteredPointerType(t *testing.T) {
	type PtrTarget struct {
		Val int
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	// Clear any existing registration
	ptrType := reflect.TypeFor[*PtrTarget]()
	elemType := ptrType.Elem()
	delete(registry, elemType)
	delete(registrationErrors, elemType)

	ra := ensureRegistered(ptrType)
	if ra == nil {
		t.Fatal("ensureRegistered returned nil")
	}

	// Should have registered the element type, not the pointer type
	if ra.typ != elemType {
		t.Errorf("got type %v, want %v", ra.typ, elemType)
	}
}

func TestEnsureRegisteredCached(t *testing.T) {
	type AlreadyRegistered struct {
		X int
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	typ := reflect.TypeFor[AlreadyRegistered]()

	// First registration
	ra1 := ensureRegistered(typ)

	// Second registration should return cached
	ra2 := ensureRegistered(typ)

	if ra1 != ra2 {
		t.Error("expected same atomizer for cached type")
	}
}

func TestEnsureRegisteredWithError(t *testing.T) {
	type HasBadField struct {
		F func()
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	typ := reflect.TypeFor[HasBadField]()
	delete(registry, typ)
	delete(registrationErrors, typ)

	ra := ensureRegistered(typ)

	// Should have recorded an error
	if _, hasErr := registrationErrors[typ]; !hasErr {
		t.Error("expected error to be recorded")
	}

	// Atomizer should still be returned (shell)
	if ra == nil {
		t.Error("atomizer should not be nil even with error")
	}
}

func TestRegistryThreadSafety(_ *testing.T) {
	// Test that concurrent registration doesn't cause races
	type RaceTest1 struct{ A int }
	type RaceTest2 struct{ B string }
	type RaceTest3 struct{ C bool }

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			//nolint:errcheck // Intentionally ignoring errors in race test
			Use[RaceTest1]()
		}()
		go func() {
			defer wg.Done()
			//nolint:errcheck // Intentionally ignoring errors in race test
			Use[RaceTest2]()
		}()
		go func() {
			defer wg.Done()
			//nolint:errcheck // Intentionally ignoring errors in race test
			Use[RaceTest3]()
		}()
	}
	wg.Wait()
}
