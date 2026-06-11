# Contributing to juvia

Thank you for your interest in contributing to juvia!

## Getting Started

1. **Fork the repository**
2. **Clone your fork:**
   ```bash
   git clone https://github.com/your-username/juvia.git
   cd juvia
   ```
3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/juvia/juvia.git
   ```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming conventions:
- `feature/` — New features
- `fix/` — Bug fixes
- `docs/` — Documentation only
- `refactor/` — Code refactoring
- `test/` — Adding or updating tests

### 2. Make Your Changes

- Write clean, commented code
- Follow the code style guides in [AGENTS.md](../AGENTS.md)
- Add tests for new functionality
- Update documentation if needed

### 3. Commit Your Changes

```bash
git add .
git commit -m "feat: add website suspension feature"
```

**Commit message format:**
```
<type>: <short description>

<longer description if needed>

Fixes #<issue number>
```

**Types:**
- `feat` — New feature
- `fix` — Bug fix
- `docs` — Documentation changes
- `style` — Formatting (no code change)
- `refactor` — Code refactoring
- `test` — Adding tests
- `chore` — Maintenance tasks

### 4. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then open a Pull Request on GitHub.

## Code Review Checklist

Before submitting a PR, ensure:

- [ ] Code follows Go formatting (`gofmt`)
- [ ] No `fmt.Errorf("...")` without `%w` for wrapping
- [ ] Context is propagated (`ctx` parameter)
- [ ] Errors are handled, not ignored
- [ ] New endpoints have corresponding tests
- [ ] Security-sensitive code reviewed (file ops, auth, etc.)
- [ ] No hardcoded credentials or secrets
- [ ] UI changes follow the design system (see [FRONTEND.md](../docs/FRONTEND.md))
- [ ] Plain-English UI rule followed (see [AGENTS.md](../AGENTS.md))

## Issue Reporting

### Bug Reports
Include:
- Description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Server environment (OS, version)
- Relevant log entries

### Feature Requests
Include:
- Clear description of the feature
- Use case (why is it needed?)
- Suggested implementation approach
- Any relevant mockups or examples

## Security Disclosures

If you discover a security vulnerability, please do NOT open a public issue. Instead, email security@juvia.dev with:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes

We aim to respond within 48 hours and will work with you on a disclosure timeline.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

- Open a Discussion on GitHub
- Join our community chat (link coming soon)

## See Also

- [AGENTS.md](../AGENTS.md) — Development context and conventions
- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) — System architecture
- [ROADMAP.md](../docs/ROADMAP.md) — Implementation phases and priorities