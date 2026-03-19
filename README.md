# atom

[![CI Status](https://github.com/zoobz-io/atom/workflows/CI/badge.svg)](https://github.com/zoobz-io/atom/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobz-io/atom/graph/badge.svg?branch=main)](https://codecov.io/gh/zoobz-io/atom)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobz-io/atom)](https://goreportcard.com/report/github.com/zoobz-io/atom)
[![CodeQL](https://github.com/zoobz-io/atom/workflows/CodeQL/badge.svg)](https://github.com/zoobz-io/atom/security/code-scanning)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobz-io/atom.svg)](https://pkg.go.dev/github.com/zoobz-io/atom)
[![License](https://img.shields.io/github/license/zoobz-io/atom)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobz-io/atom)](go.mod)
[![Release](https://img.shields.io/github/v/release/zoobz-io/atom)](https://github.com/zoobz-io/atom/releases)

Type-segregated struct decomposition for Go.

Break structs into typed maps, work with fields programmatically, reconstruct later — without knowing T.

## Normalized Type-Safe Data Without the Types

User code knows its types. Infrastructure code doesn't need to.

```go
// User side: knows T
type User struct {
    ID      string
    Name    string
    Age     int64
    Balance float64
    Active  bool
}

atomizer, _ := atom.Use[User]()
user := &User{ID: "usr-1", Name: "Alice", Age: 30, Balance: 100.50, Active: true}
a := atomizer.Atomize(user)

// Pass atom to any library...
result := storage.Save(a)            // storage never imports User
validated := validator.Check(a)      // validator never imports User
transformed := migrator.Upgrade(a)   // migrator never imports User

// ...get it back
restored, _ := atomizer.Deatomize(result)
```

The receiving library sees typed maps and metadata — not T:

```go
// Library side: doesn't know T, doesn't need T
func Save(a *atom.Atom) *atom.Atom {
    // Spec describes the struct
    fmt.Println(a.Spec.TypeName) // "User"

    // Typed maps hold the values
    for field, value := range a.Strings {
        db.SetString(field, value)
    }
    for field, value := range a.Ints {
        db.SetInt(field, value)
    }
    for field, value := range a.Floats {
        db.SetFloat(field, value)
    }

    return a
}
```

Type-safe field access. Zero knowledge of the original struct.

## Install

```bash
go get github.com/zoobz-io/atom@latest
```

Requires Go 1.24+.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zoobz-io/atom"
)

type Order struct {
    ID     string
    Total  float64
    Status string
}

func main() {
    // Register the type once
    atomizer, err := atom.Use[Order]()
    if err != nil {
        panic(err)
    }

    order := &Order{ID: "order-123", Total: 99.99, Status: "pending"}

    // Decompose to atom
    a := atomizer.Atomize(order)

    // Work with typed maps
    fmt.Printf("ID: %s\n", a.Strings["ID"])
    fmt.Printf("Total: %.2f\n", a.Floats["Total"])

    // Modify fields
    a.Strings["Status"] = "confirmed"

    // Reconstruct
    restored, _ := atomizer.Deatomize(a)
    fmt.Printf("Status: %s\n", restored.Status) // "confirmed"
}
```

## Capabilities

| Feature                | Description                                                       | Docs                                                |
| ---------------------- | ----------------------------------------------------------------- | --------------------------------------------------- |
| Type Segregation       | Strings, ints, floats, bools, times, bytes in separate typed maps | [Concepts](docs/2.learn/2.concepts.md)              |
| Nullable Fields        | Pointer types (`*string`, `*int64`) with explicit nil handling    | [Basic Usage](docs/3.guides/1.basic-usage.md)       |
| Slices                 | `[]string`, `[]int64`, etc. preserved as typed slices             | [Basic Usage](docs/3.guides/1.basic-usage.md)       |
| Nested Composition     | Embed atoms within atoms for complex object graphs                | [Nested Structs](docs/3.guides/3.nested-structs.md) |
| Field Introspection    | Query fields, tables, and type metadata via Spec                  | [API Reference](docs/5.reference/1.api.md)          |
| Custom Implementations | `Atomizable`/`Deatomizable` interfaces bypass reflection          | [Interfaces](docs/3.guides/4.interfaces.md)         |
| Code Generation        | Generate implementations for zero-reflection paths                | [Code Generation](docs/4.cookbook/1.codegen.md)     |

## Why atom?

- **Type-safe without T** — Libraries work with typed maps, not `any` or reflection
- **Field-level control** — Read, write, transform individual fields programmatically
- **Decoupled** — Infrastructure code never imports user types
- **Zero reflection path** — Implement interfaces or use codegen for production performance
- **Sentinel integration** — Automatic field discovery and metadata extraction

## The Typed Bridge

Atom enables a pattern: **user code owns types, infrastructure owns behaviour**.

Your application defines structs. Libraries accept atoms. Each side works with what it knows — concrete types on one end, typed maps on the other. No shared type imports. No reflection at runtime (with codegen).

```go
// Your domain package defines types
type User struct { ... }
type Order struct { ... }

// Storage library accepts atoms — never sees User or Order
func (s *Store) Put(a *atom.Atom) error { ... }
func (s *Store) Get(spec atom.Spec, id string) (*atom.Atom, error) { ... }

// Your code bridges the two
a := userAtomizer.Atomize(user)
store.Put(a)
```

The contract is the Atom structure. The types stay where they belong.

## Documentation

### Learn

- [Quickstart](docs/2.learn/1.quickstart.md) — Get started in 5 minutes
- [Core Concepts](docs/2.learn/2.concepts.md) — Atoms, tables, specs
- [Architecture](docs/2.learn/3.architecture.md) — Internal design

### Guides

- [Basic Usage](docs/3.guides/1.basic-usage.md) — Common patterns
- [Custom Types](docs/3.guides/2.custom-types.md) — Named types and enums
- [Nested Structs](docs/3.guides/3.nested-structs.md) — Composition
- [Interfaces](docs/3.guides/4.interfaces.md) — Custom serialization
- [Testing](docs/3.guides/5.testing.md) — Test strategies

### Cookbook

- [Code Generation](docs/4.cookbook/1.codegen.md) — Eliminating reflection
- [Serialization](docs/4.cookbook/2.serialization.md) — Encoding atoms
- [Migrations](docs/4.cookbook/3.migrations.md) — Schema evolution

### Reference

- [API Reference](docs/5.reference/1.api.md) — Complete API
- [Tables Reference](docs/5.reference/2.tables.md) — All table types
- [Testing Reference](docs/5.reference/3.testing.md) — Test utilities

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Run `make help` for available commands.

## License

MIT License — see [LICENSE](LICENSE) for details.
