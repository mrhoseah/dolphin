# Contributing to Dolphin

Thanks for your interest in contributing! Please follow these guidelines to keep the project healthy and welcoming.

## Development Setup
- Go 1.21+
- `go mod tidy`
- Run tests: `go test ./...`
- Lint (optional): `golangci-lint run`

## Pull Requests
- Fork and create a feature branch (`feat/xyz`, `fix/bug-123`).
- Keep PRs focused and small.
- Add tests for new behavior.
- Update docs/README when relevant.
- Follow semantic commit messages where possible: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `ci:`

## Coding Style
- Prefer clarity over cleverness.
- Public APIs must have Go doc comments.
- Avoid breaking changes; if needed, propose in an issue first.

## Security
- Do not open public issues for vulnerabilities.
- See `SECURITY.md` for responsible disclosure.

## License
By contributing, you agree your contributions are licensed under the MIT license in `LICENSE`.

