# CLAUDE.md

Local proxy that serves PS4/PS5 `.pkg` files from disk; see README.md.

## Rules

- Use only Go standard library packages. No external dependencies.
- Do not commit anything.
- Do not build the project yourself (no `go build`, `build.ps1`, or `build.sh`); the user builds it.
- Do not write comments in code.
- Use pointers for reusable/shared values (e.g. a single shared `*http.Client`) instead of recreating them.
- Use goroutines for independent or concurrent work (e.g. probing parts, background link generation).
- Keep `gofmt` clean and `go vet` passing.
