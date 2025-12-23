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

atomizer := atom.New[User](atomizeUser, deatomizeUser)

// Decompose to atoms
atoms, _ := atomizer.Atomize(&user)
// atoms.Strings["Name"] = "Alice"
// atoms.Ints["Age"] = 30
// atoms.Floats["Balance"] = 100.50

// Reconstruct from atoms
user, _ := atomizer.Deatomize(atoms)
```

You provide the transformation functions. Atom handles:
- **Type segregation** — strings, ints, floats, bools, times in separate maps
- **Field metadata** — automatic introspection via sentinel
- **Validation** — runs before atomization
- **Binary encoding** — utilities for storage-ready byte conversion

## Features

- **Storage-agnostic** — bring your own persistence layer
- **Type-safe generics** — `Atomizer[T]` catches errors at compile time
- **Sentinel integration** — automatic field discovery and metadata
- **Binary encoding** — big-endian ints (sortable), RFC3339Nano times
- **Validation hooks** — implement `Validator` interface for pre-atomization checks

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
    "time"

    "github.com/zoobzio/atom"
)

type Order struct {
    ID     string  `atom:"id"`
    Total  float64
    Status string
}

func (o Order) Validate() error {
    if o.ID == "" {
        return fmt.Errorf("ID required")
    }
    return nil
}

func atomizeOrder(o *Order) atom.Atoms {
    atoms := atom.NewAtoms(o.ID)
    atoms.Floats["Total"] = o.Total
    atoms.Strings["Status"] = o.Status
    return *atoms
}

func deatomizeOrder(atoms atom.Atoms) (*Order, error) {
    return &Order{
        ID:     atoms.ID,
        Total:  atoms.Floats["Total"],
        Status: atoms.Strings["Status"],
    }, nil
}

func main() {
    atomizer := atom.New[Order](atomizeOrder, deatomizeOrder)

    order := &Order{ID: "order-123", Total: 99.99, Status: "pending"}

    // Decompose
    atoms, _ := atomizer.Atomize(order)
    fmt.Printf("ID: %s, Total: %.2f\n", atoms.ID, atoms.Floats["Total"])

    // Reconstruct
    restored, _ := atomizer.Deatomize(atoms)
    fmt.Printf("Restored: %+v\n", restored)
}
```

## API Reference

### Core Types

| Type | Purpose |
|------|---------|
| `Atomizer[T]` | Generic atomizer for type T |
| `Atoms` | Container for decomposed atomic values |
| `Atomize[T]` | Function type: `*T` → `Atoms` |
| `Deatomize[T]` | Function type: `Atoms` → `(*T, error)` |
| `Validator` | Interface with `Validate() error` |

### Atomizer Methods

| Method | Purpose |
|--------|---------|
| `New[T](atomize, deatomize)` | Create an atomizer for type T |
| `Atomize(obj)` | Decompose object to atoms (validates first) |
| `Deatomize(atoms)` | Reconstruct object from atoms |
| `Metadata()` | Get sentinel metadata for type T |
| `Fields()` | Get all field descriptors |
| `FieldsIn(table)` | Get field names for a table type |
| `TableFor(field)` | Get table type for a field name |

### Table Types

| TableType | Go Types |
|-----------|----------|
| `TableStrings` | `string` |
| `TableInts` | `int`, `int8`...`int64`, `uint`...`uint64` |
| `TableFloats` | `float32`, `float64` |
| `TableBools` | `bool` |
| `TableTimes` | `time.Time` |

### Encoding Utilities

| Function | Format |
|----------|--------|
| `encodeString` / `decodeString` | UTF-8 bytes |
| `encodeInt64` / `decodeInt64` | Big-endian (sortable) |
| `encodeFloat64` / `decodeFloat64` | IEEE 754 binary |
| `encodeBool` / `decodeBool` | Single byte (0/1) |
| `encodeTime` / `decodeTime` | RFC3339Nano string |

## Design

Atom is intentionally minimal. It provides:

1. **Decomposition abstraction** — the `Atoms` container and transformation types
2. **Field introspection** — via sentinel integration
3. **Encoding utilities** — for storage-ready byte conversion

It does **not** provide:
- Storage backends
- Query interfaces
- Caching layers

This design allows atom to be used within storage libraries (like grub) without circular dependencies.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE) for details.
