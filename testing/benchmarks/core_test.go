// Package benchmarks provides performance benchmarks for the atom package.
//
//nolint:errcheck // Benchmarks intentionally ignore errors for performance measurement.
package benchmarks

import (
	"testing"
	"time"

	"github.com/zoobzio/atom"
)

// Simple struct for basic benchmarks.
type SimpleUser struct {
	Name   string
	Age    int64
	Score  float64
	Active bool
}

// Medium complexity struct.
type MediumUser struct {
	ID        int64
	Name      string
	Email     string
	Age       int64
	Score     float64
	Active    bool
	CreatedAt time.Time
	Tags      []string
	Balance   float64
	Verified  bool
}

// Complex struct with nested types.
type ComplexUser struct {
	ID        int64
	Name      string
	Email     string
	Age       int64
	Score     float64
	Active    bool
	CreatedAt time.Time
	Tags      []string
	Address   Address
	Orders    []Order
}

type Address struct {
	Street  string
	City    string
	ZipCode string
	Country string
}

type Order struct {
	ID    int64
	Total float64
	Items []string
}

// Struct implementing Atomizable.
type FastUser struct {
	Name  string
	Age   int64
	Score float64
}

func (u *FastUser) Atomize(a *atom.Atom) {
	a.Strings["Name"] = u.Name
	a.Ints["Age"] = u.Age
	a.Floats["Score"] = u.Score
}

func (u *FastUser) Deatomize(a *atom.Atom) error {
	u.Name = a.Strings["Name"]
	u.Age = a.Ints["Age"]
	u.Score = a.Floats["Score"]
	return nil
}

// Atomize benchmarks.

func BenchmarkAtomize_Simple(b *testing.B) {
	atomizer, _ := atom.Use[SimpleUser]()
	user := &SimpleUser{Name: "Alice", Age: 30, Score: 95.5, Active: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atomizer.Atomize(user)
	}
}

func BenchmarkAtomize_Medium(b *testing.B) {
	atomizer, _ := atom.Use[MediumUser]()
	now := time.Now()
	user := &MediumUser{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       30,
		Score:     95.5,
		Active:    true,
		CreatedAt: now,
		Tags:      []string{"admin", "verified"},
		Balance:   1000.50,
		Verified:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atomizer.Atomize(user)
	}
}

func BenchmarkAtomize_Complex(b *testing.B) {
	atomizer, _ := atom.Use[ComplexUser]()
	now := time.Now()
	user := &ComplexUser{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       30,
		Score:     95.5,
		Active:    true,
		CreatedAt: now,
		Tags:      []string{"admin", "verified"},
		Address: Address{
			Street:  "123 Main St",
			City:    "Springfield",
			ZipCode: "12345",
			Country: "USA",
		},
		Orders: []Order{
			{ID: 1, Total: 99.99, Items: []string{"item1", "item2"}},
			{ID: 2, Total: 149.99, Items: []string{"item3"}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atomizer.Atomize(user)
	}
}

func BenchmarkAtomize_WithInterface(b *testing.B) {
	atomizer, _ := atom.Use[FastUser]()
	user := &FastUser{Name: "Alice", Age: 30, Score: 95.5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atomizer.Atomize(user)
	}
}

// Deatomize benchmarks.

func BenchmarkDeatomize_Simple(b *testing.B) {
	atomizer, _ := atom.Use[SimpleUser]()
	a := atomizer.Atomize(&SimpleUser{Name: "Alice", Age: 30, Score: 95.5, Active: true})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = atomizer.Deatomize(a)
	}
}

func BenchmarkDeatomize_Medium(b *testing.B) {
	atomizer, _ := atom.Use[MediumUser]()
	now := time.Now()
	a := atomizer.Atomize(&MediumUser{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       30,
		Score:     95.5,
		Active:    true,
		CreatedAt: now,
		Tags:      []string{"admin", "verified"},
		Balance:   1000.50,
		Verified:  true,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = atomizer.Deatomize(a)
	}
}

func BenchmarkDeatomize_Complex(b *testing.B) {
	atomizer, _ := atom.Use[ComplexUser]()
	now := time.Now()
	a := atomizer.Atomize(&ComplexUser{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       30,
		Score:     95.5,
		Active:    true,
		CreatedAt: now,
		Tags:      []string{"admin", "verified"},
		Address: Address{
			Street:  "123 Main St",
			City:    "Springfield",
			ZipCode: "12345",
			Country: "USA",
		},
		Orders: []Order{
			{ID: 1, Total: 99.99, Items: []string{"item1", "item2"}},
			{ID: 2, Total: 149.99, Items: []string{"item3"}},
		},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = atomizer.Deatomize(a)
	}
}

func BenchmarkDeatomize_WithInterface(b *testing.B) {
	atomizer, _ := atom.Use[FastUser]()
	a := atomizer.Atomize(&FastUser{Name: "Alice", Age: 30, Score: 95.5})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = atomizer.Deatomize(a)
	}
}

// Round-trip benchmarks.

func BenchmarkRoundTrip_Simple(b *testing.B) {
	atomizer, _ := atom.Use[SimpleUser]()
	user := &SimpleUser{Name: "Alice", Age: 30, Score: 95.5, Active: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := atomizer.Atomize(user)
		_, _ = atomizer.Deatomize(a)
	}
}

func BenchmarkRoundTrip_Complex(b *testing.B) {
	atomizer, _ := atom.Use[ComplexUser]()
	now := time.Now()
	user := &ComplexUser{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       30,
		Score:     95.5,
		Active:    true,
		CreatedAt: now,
		Tags:      []string{"admin", "verified"},
		Address: Address{
			Street:  "123 Main St",
			City:    "Springfield",
			ZipCode: "12345",
			Country: "USA",
		},
		Orders: []Order{
			{ID: 1, Total: 99.99, Items: []string{"item1", "item2"}},
			{ID: 2, Total: 149.99, Items: []string{"item3"}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := atomizer.Atomize(user)
		_, _ = atomizer.Deatomize(a)
	}
}

// Registration benchmarks.

func BenchmarkUse_Cached(b *testing.B) {
	// Pre-register
	_, _ = atom.Use[SimpleUser]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = atom.Use[SimpleUser]()
	}
}

// NewAtom benchmarks.

func BenchmarkNewAtom_Simple(b *testing.B) {
	atomizer, _ := atom.Use[SimpleUser]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atomizer.NewAtom()
	}
}

func BenchmarkNewAtom_Complex(b *testing.B) {
	atomizer, _ := atom.Use[ComplexUser]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = atomizer.NewAtom()
	}
}

// Parallel benchmarks.

func BenchmarkAtomize_Parallel(b *testing.B) {
	atomizer, _ := atom.Use[SimpleUser]()
	user := &SimpleUser{Name: "Alice", Age: 30, Score: 95.5, Active: true}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = atomizer.Atomize(user)
		}
	})
}

func BenchmarkDeatomize_Parallel(b *testing.B) {
	atomizer, _ := atom.Use[SimpleUser]()
	a := atomizer.Atomize(&SimpleUser{Name: "Alice", Age: 30, Score: 95.5, Active: true})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = atomizer.Deatomize(a)
		}
	})
}

func BenchmarkRoundTrip_Parallel(b *testing.B) {
	atomizer, _ := atom.Use[SimpleUser]()
	user := &SimpleUser{Name: "Alice", Age: 30, Score: 95.5, Active: true}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a := atomizer.Atomize(user)
			_, _ = atomizer.Deatomize(a)
		}
	})
}
