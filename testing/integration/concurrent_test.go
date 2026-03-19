package integration

import (
	"sync"
	"testing"

	"github.com/zoobz-io/atom"
)

// Types for concurrent testing.
type ConcurrentUser struct {
	Name  string
	Age   int64
	Score float64
}

type ConcurrentOrder struct {
	ID    int64
	Total float64
}

type ConcurrentProduct struct {
	SKU   string
	Price float64
}

func TestConcurrentRegistration(t *testing.T) {
	var wg sync.WaitGroup

	// Register multiple types concurrently
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, err := atom.Use[ConcurrentUser]()
			if err != nil {
				t.Errorf("Use[ConcurrentUser] failed: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			_, err := atom.Use[ConcurrentOrder]()
			if err != nil {
				t.Errorf("Use[ConcurrentOrder] failed: %v", err)
			}
		}()
		go func() {
			defer wg.Done()
			_, err := atom.Use[ConcurrentProduct]()
			if err != nil {
				t.Errorf("Use[ConcurrentProduct] failed: %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentAtomize(t *testing.T) {
	atomizer, err := atom.Use[ConcurrentUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			user := &ConcurrentUser{
				Name:  "User" + string(rune('A'+id%26)),
				Age:   int64(20 + id%50),
				Score: float64(id) * 1.5,
			}
			a := atomizer.Atomize(user)
			if a == nil {
				errors <- nil // Signal error
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent atomize error: %v", err)
		}
	}
}

func TestConcurrentDeatomize(t *testing.T) {
	atomizer, err := atom.Use[ConcurrentUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	// Create a template atom
	template := atomizer.Atomize(&ConcurrentUser{Name: "Template", Age: 25, Score: 100})

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := atomizer.Deatomize(template)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent deatomize error: %v", err)
	}
}

func TestConcurrentRoundTrip(t *testing.T) {
	atomizer, err := atom.Use[ConcurrentUser]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			original := &ConcurrentUser{
				Name:  "User" + string(rune('A'+id%26)),
				Age:   int64(20 + id%50),
				Score: float64(id) * 1.5,
			}

			a := atomizer.Atomize(original)
			restored, err := atomizer.Deatomize(a)
			if err != nil {
				t.Errorf("round-trip %d failed: %v", id, err)
				return
			}

			if restored.Name != original.Name ||
				restored.Age != original.Age ||
				restored.Score != original.Score {
				t.Errorf("round-trip %d mismatch", id)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentDifferentTypes(_ *testing.T) {
	userAtomizer, _ := atom.Use[ConcurrentUser]()    //nolint:errcheck // Test setup
	orderAtomizer, _ := atom.Use[ConcurrentOrder]()   //nolint:errcheck // Test setup
	productAtomizer, _ := atom.Use[ConcurrentProduct]() //nolint:errcheck // Test setup

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(3)

		go func(id int) {
			defer wg.Done()
			user := &ConcurrentUser{Name: "User", Age: int64(id)}
			a := userAtomizer.Atomize(user)
			_, _ = userAtomizer.Deatomize(a) //nolint:errcheck // Concurrent stress test
		}(i)

		go func(id int) {
			defer wg.Done()
			order := &ConcurrentOrder{ID: int64(id), Total: float64(id) * 10}
			a := orderAtomizer.Atomize(order)
			_, _ = orderAtomizer.Deatomize(a) //nolint:errcheck // Concurrent stress test
		}(i)

		go func(id int) {
			defer wg.Done()
			product := &ConcurrentProduct{SKU: "SKU", Price: float64(id)}
			a := productAtomizer.Atomize(product)
			_, _ = productAtomizer.Deatomize(a) //nolint:errcheck // Concurrent stress test
		}(i)
	}

	wg.Wait()
}

// Nested type for concurrent testing.
type ConcurrentNested struct {
	Name  string
	Inner ConcurrentUser
}

func TestConcurrentNestedTypes(t *testing.T) {
	atomizer, err := atom.Use[ConcurrentNested]()
	if err != nil {
		t.Fatalf("Use failed: %v", err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			original := &ConcurrentNested{
				Name: "Outer",
				Inner: ConcurrentUser{
					Name:  "Inner",
					Age:   int64(id),
					Score: float64(id) * 2,
				},
			}

			a := atomizer.Atomize(original)
			restored, err := atomizer.Deatomize(a)
			if err != nil {
				t.Errorf("nested round-trip %d failed: %v", id, err)
				return
			}

			if restored.Inner.Age != original.Inner.Age {
				t.Errorf("nested round-trip %d mismatch", id)
			}
		}(i)
	}

	wg.Wait()
}
