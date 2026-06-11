# Testing Guide

## Testing Strategy

juvia uses a layered testing approach:

1. **Unit tests** — Business logic, utility functions, data validation
2. **Integration tests** — API endpoints, database operations, socket communication
3. **E2E tests** — Critical user flows
4. **Performance tests** — Load testing, benchmarks
5. **Security tests** — Penetration testing, fuzzing
6. **Manual testing** — End-to-end flows on target server

## Unit Tests

### Go (Backend)
```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/db/...

# Run with race detector
go test -race ./...
```

**Conventions:**
- Test files: `*_test.go` alongside source files
- Table-driven tests preferred
- Use `testify/require` for assertions
- Mock external dependencies (database, socket, exec)

### React (Frontend)
```bash
# Run tests
npm test

# Run with coverage
npm test -- --coverage

# Run specific file
npm test -- src/lib/api.test.ts
```

**Conventions:**
- Test files: `*.test.ts` or `*.test.tsx` alongside components
- Use `@testing-library/react` for component tests
- Use `msw` for API mocking
- Aim for 80% coverage on critical paths

## Integration Tests

### API Endpoints
Test REST endpoints with a test database.

### Agent Socket Communication
Test JSON-RPC over Unix socket.

## End-to-End (E2E) Tests

Using Playwright for critical user flows.

Run E2E tests:
```bash
npx playwright test
```

## Performance Testing

### Load Testing with k6
Use k6 for load testing.

### SQLite Query Benchmarks
```bash
go test -bench=. -benchmem ./internal/db/...
```

## Security Testing

### OWASP ZAP Scan
Run ZAP baseline scan against the API.

### Static Analysis
```bash
gosec ./...
```

### Fuzz Testing
```bash
go test -fuzz=FuzzWebsiteCreate -fuzztime=10s ./internal/api/...
```

## CI/CD Pipeline

GitHub Actions workflow in `.github/workflows/test.yml`.

### Remote Deployment Pipeline
Build, copy to remote, restart services.

## Test Data

Use factories for consistent test data.

## See Also

- [AGENTS.md](../AGENTS.md) — Build and run commands
- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) — System architecture
- [SECURITY.md](../docs/SECURITY.md) — Security testing considerations
