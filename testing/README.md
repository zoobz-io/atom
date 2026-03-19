# Testing

This directory contains testing utilities and comprehensive tests for the atom package.

## Structure

```
testing/
├── README.md           # This file
├── helpers.go          # Test utilities for users
├── helpers_test.go     # Tests for helpers
├── integration/        # Integration tests
│   ├── README.md
│   ├── roundtrip_test.go
│   └── concurrent_test.go
└── benchmarks/         # Performance benchmarks
    ├── README.md
    └── core_test.go
```

## Test Categories

### Unit Tests

Located in the main package (`*_test.go` files alongside source).

Run with:
```bash
go test ./...
```

### Integration Tests

Located in `testing/integration/`. Test complete workflows and edge cases.

Run with:
```bash
go test ./testing/integration/...
```

### Benchmarks

Located in `testing/benchmarks/`. Measure performance.

Run with:
```bash
go test -bench=. ./testing/benchmarks/...
```

## Using Test Helpers

Import the testing package:

```go
import atomtest "github.com/zoobz-io/atom/testing"
```

### AtomBuilder

Fluent builder for test atoms:

```go
a := atomtest.NewAtomBuilder().
    String("Name", "Alice").
    Int("Age", 30).
    Build()
```

### RoundTripValidator

Automated round-trip testing:

```go
validator := atomtest.NewRoundTripValidator(atomizer)
validator.Validate(t, &User{Name: "Alice"})
```

### MustUse

Panic-on-error registration for tests:

```go
atomizer := atomtest.MustUse[User](t)
```

## Coverage

Run with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Race Detection

Run with race detector:

```bash
go test -race ./...
```
