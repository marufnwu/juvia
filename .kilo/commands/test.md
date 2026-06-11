# Test Command

## Description
Run all tests for juvia.

## Usage

### Run all tests
```bash
go test ./...
```

### Run with coverage
```bash
go test -cover ./...
```

### Run with race detector
```bash
go test -race ./...
```

### Run specific package
```bash
go test ./internal/db/...
go test ./internal/api/...
```

### Run tests with verbose output
```bash
go test -v ./...
```

### Run tests matching pattern
```bash
go test -run "TestWebsite" ./...
```

## Frontend Tests

```bash
cd web
npm test
npm test -- --coverage
```

## Integration Tests

Integration tests require a test database and may start the socket server:

```bash
# Run integration tests (may require sudo for socket creation)
go test -tags=integration ./...
```

## CI/CD

See [TESTING.md](../docs/TESTING.md) for the full CI/CD pipeline configuration.

## See Also
- [TESTING.md](../docs/TESTING.md) — Testing guide
- [AGENTS.md](../AGENTS.md) — Developer context