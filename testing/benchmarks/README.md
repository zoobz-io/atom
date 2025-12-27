# Benchmarks

Performance benchmarks for the atom package.

## Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./testing/benchmarks/...

# Run with memory allocation stats
go test -bench=. -benchmem ./testing/benchmarks/...

# Run specific benchmark
go test -bench=BenchmarkAtomize ./testing/benchmarks/...

# Run with custom count
go test -bench=. -count=5 ./testing/benchmarks/...
```

## Benchmark Categories

### Core Operations

- `BenchmarkAtomize_*` - Struct to Atom conversion
- `BenchmarkDeatomize_*` - Atom to Struct conversion
- `BenchmarkRoundTrip_*` - Full atomize/deatomize cycle

### By Struct Complexity

- `*_Simple` - Small struct (4 fields)
- `*_Medium` - Medium struct (10 fields)
- `*_Complex` - Large struct with nested types

### Registration

- `BenchmarkUse_Cached` - Cached atomizer retrieval
- `BenchmarkUse_New` - New type registration (one-time cost)

## Performance Targets

| Operation | Target |
|-----------|--------|
| Atomize (simple) | < 500 ns/op |
| Deatomize (simple) | < 500 ns/op |
| Use (cached) | < 100 ns/op |
| Atomize (with interface) | < 100 ns/op |

## Profiling

```bash
# CPU profile
go test -bench=. -cpuprofile=cpu.prof ./testing/benchmarks/...
go tool pprof cpu.prof

# Memory profile
go test -bench=. -memprofile=mem.prof ./testing/benchmarks/...
go tool pprof mem.prof
```
