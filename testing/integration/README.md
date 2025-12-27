# Integration Tests

Comprehensive integration tests for the atom package.

## Tests

### roundtrip_test.go

Tests complete atomization/deatomization cycles:
- Simple structs
- Complex nested structures
- All field types
- Edge cases (empty, nil, max values)

### concurrent_test.go

Tests thread safety:
- Concurrent type registration
- Concurrent atomization
- Concurrent deatomization
- Race condition detection

## Running

```bash
# Run all integration tests
go test ./testing/integration/...

# Run with verbose output
go test -v ./testing/integration/...

# Run with race detector
go test -race ./testing/integration/...
```
