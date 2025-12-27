# atom

[![CI Status](https://github.com/zoobzio/atom/workflows/CI/badge.svg)](https://github.com/zoobzio/atom/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobzio/atom/graph/badge.svg?branch=main)](https://codecov.io/gh/zoobzio/atom)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobzio/atom)](https://goreportcard.com/report/github.com/zoobzio/atom)
[![CodeQL](https://github.com/zoobzio/atom/workflows/CodeQL/badge.svg)](https://github.com/zoobzio/atom/security/code-scanning)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobzio/atom.svg)](https://pkg.go.dev/github.com/zoobzio/atom)
[![License](https://img.shields.io/github/license/zoobzio/atom)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobzio/atom)](go.mod)
[![Release](https://img.shields.io/github/v/release/zoobzio/atom)](https://github.com/zoobzio/atom/releases)

Type-segregated atomic value decomposition for Go — break structs into typed atoms, reconstruct them later.

## The Problem

Storing complex structs often means choosing between:
- **Full serialization** — JSON/protobuf blobs that can't be queried field-by-field
- **ORM mapping** — heavy frameworks with reflection overhead on every operation
- **Manual decomposition** — tedious, error-prone code for each type

## The Solution

Atom provides a clean abstraction for decomposing structs into typed atomic values:

```go
type User struct {
    ID        string
    Name      string
    Age       int64
    Balance   float64
    Active    bool
    CreatedAt time.Time
}

// Register the type once
atomizer, _ := atom.Use[User]()

// Decompose to atom
user := &User{Name: "Alice", Age: 30, Balance: 100.50, Active: true}
a := atomizer.Atomize(user)
// a.Strings["Name"] = "Alice"
// a.Ints["Age"] = 30
// a.Floats["Balance"] = 100.50
// a.Bools["Active"] = true

// Reconstruct from atom
restored, _ := atomizer.Deatomize(a)
```

Atom handles:
- **Type segregation** — scalars, pointers, slices, nested objects in separate typed maps
- **Field metadata** — automatic introspection via sentinel
- **Numeric width conversion** — int8/uint32/etc. safely converted with overflow detection

## Features

- **Storage-agnostic** — bring your own persistence layer
- **Type-safe generics** — `Atomizer[T]` catches errors at compile time
- **Sentinel integration** — automatic field discovery and metadata
- **Nullable fields** — pointer types (`*string`, `*int64`, etc.) with explicit nil handling
- **Slice support** — type-safe storage for `[]string`, `[]int64`, etc.
- **Nested composition** — embed Atoms within Atoms for complex object graphs
- **Custom implementations** — implement `Atomizable`/`Deatomizable` interfaces to bypass reflection

## Install

```bash
go get github.com/zoobzio/atom@latest
```

Requires Go 1.23+.

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/zoobzio/atom"
)

type Order struct {
    ID     string
    Total  float64
    Status string
}

func main() {
    // Register the type
    atomizer, err := atom.Use[Order]()
    if err != nil {
        panic(err)
    }

    order := &Order{ID: "order-123", Total: 99.99, Status: "pending"}

    // Decompose
    a := atomizer.Atomize(order)
    fmt.Printf("ID: %s, Total: %.2f\n", a.Strings["ID"], a.Floats["Total"])

    // Reconstruct
    restored, _ := atomizer.Deatomize(a)
    fmt.Printf("Restored: %+v\n", restored)
}
```

## Usage

### Basic Types

All Go primitive types are supported and stored in type-segregated maps:

```go
type Example struct {
    Name    string     // → Atom.Strings
    Age     int        // → Atom.Ints (as int64)
    Count   uint64     // → Atom.Uints
    Rate    float64    // → Atom.Floats
    Active  bool       // → Atom.Bools
    Created time.Time  // → Atom.Times
    Data    []byte     // → Atom.Bytes
}
```

### Nullable Fields (Pointers)

Use pointer types to represent optional/nullable fields:

```go
type Profile struct {
    ID       string
    Nickname *string  // → Atom.StringPtrs (nil preserved)
    Age      *int64   // → Atom.IntPtrs
    Bio      *[]byte  // → Atom.BytePtrs
}

atomizer, _ := atom.Use[Profile]()
profile := &Profile{ID: "user-1", Nickname: nil}

a := atomizer.Atomize(profile)
// a.StringPtrs["Nickname"] == nil (explicitly stored)
```

### Slice Fields

Store collections of primitive values:

```go
type Article struct {
    ID     string
    Tags   []string   // → Atom.StringSlices
    Scores []int64    // → Atom.IntSlices
    Counts []uint32   // → Atom.UintSlices (as []uint64)
}
```

### Nested Structs

Compose complex objects using nested Atoms:

```go
type Address struct {
    Street string
    City   string
}

type Person struct {
    ID      string
    Name    string
    Address Address    // → Atom.Nested["Address"]
    Friends []Person   // → Atom.NestedSlices["Friends"]
}

atomizer, _ := atom.Use[Person]()
```

### Custom Atomization

Implement `Atomizable` and/or `Deatomizable` to bypass reflection:

```go
type Custom struct {
    Value int
}

func (c *Custom) Atomize(a *atom.Atom) {
    a.Ints["Value"] = int64(c.Value * 2) // custom logic
}

func (c *Custom) Deatomize(a *atom.Atom) error {
    c.Value = int(a.Ints["Value"] / 2)
    return nil
}
```

This enables code generation for high-performance scenarios.

## API Reference

### Core Types

| Type | Purpose |
|------|---------|
| `Atomizer[T]` | Generic atomizer for type T |
| `Atom` | Container for decomposed atomic values |
| `Field` | Maps field name to storage table |
| `Atomizable` | Interface for custom atomization |
| `Deatomizable` | Interface for custom deatomization |

### Functions

| Function | Purpose |
|----------|---------|
| `Use[T]()` | Register and return an `Atomizer[T]` for type T |
| `AllTables()` | Return all table types in canonical order |

### Atomizer Methods

| Method | Purpose |
|--------|---------|
| `Atomize(obj)` | Decompose object to atom |
| `Deatomize(atom)` | Reconstruct object from atom |
| `NewAtom()` | Create an Atom with pre-sized maps for type T |
| `Spec()` | Get type specification for type T |
| `Fields()` | Get all field descriptors |
| `FieldsIn(table)` | Get field names for a table type |
| `TableFor(field)` | Get table type for a field name |

### Table Types

**Scalars:**

| TableType | Go Types | Atom Field |
|-----------|----------|------------|
| `TableStrings` | `string` | `Strings` |
| `TableInts` | `int`, `int8`...`int64` | `Ints` |
| `TableUints` | `uint`, `uint8`...`uint64` | `Uints` |
| `TableFloats` | `float32`, `float64` | `Floats` |
| `TableBools` | `bool` | `Bools` |
| `TableTimes` | `time.Time` | `Times` |
| `TableBytes` | `[]byte` | `Bytes` |

**Pointers (nullable):**

| TableType | Go Types | Atom Field |
|-----------|----------|------------|
| `TableStringPtrs` | `*string` | `StringPtrs` |
| `TableIntPtrs` | `*int`, `*int8`...`*int64` | `IntPtrs` |
| `TableUintPtrs` | `*uint`, `*uint8`...`*uint64` | `UintPtrs` |
| `TableFloatPtrs` | `*float32`, `*float64` | `FloatPtrs` |
| `TableBoolPtrs` | `*bool` | `BoolPtrs` |
| `TableTimePtrs` | `*time.Time` | `TimePtrs` |
| `TableBytePtrs` | `*[]byte` | `BytePtrs` |

**Slices:**

| TableType | Go Types | Atom Field |
|-----------|----------|------------|
| `TableStringSlices` | `[]string` | `StringSlices` |
| `TableIntSlices` | `[]int`, `[]int64`, etc. | `IntSlices` |
| `TableUintSlices` | `[]uint`, `[]uint64`, etc. | `UintSlices` |
| `TableFloatSlices` | `[]float32`, `[]float64` | `FloatSlices` |
| `TableBoolSlices` | `[]bool` | `BoolSlices` |
| `TableTimeSlices` | `[]time.Time` | `TimeSlices` |
| `TableByteSlices` | `[][]byte` | `ByteSlices` |

**Nested:**

| Field | Purpose |
|-------|---------|
| `Nested` | `map[string]Atom` for single nested structs |
| `NestedSlices` | `map[string][]Atom` for slices of nested structs |

## Design

Atom is intentionally minimal. It provides:

1. **Decomposition abstraction** — the `Atom` container and transformation types
2. **Field introspection** — via sentinel integration

It does **not** provide:
- Storage backends
- Query interfaces
- Caching layers

This design allows atom to be used within storage libraries without circular dependencies.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE) for details.
