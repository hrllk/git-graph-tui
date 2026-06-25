# Decisions

## 2026-06-24: Adopt golangci-lint for analysis, gofumpt/goimports for formatting

- Use `gofumpt` and `goimports` as formatter tools.
- Keep `golangci-lint` focused on analysis linters such as `errcheck`, `govet`, `ineffassign`, and `staticcheck`.
- Use the module path from `go.mod` (`hrllk/graphkeeper`) for `goimports.local-prefixes`.

## 2026-06-24: Offline fallback for lint bootstrap

- `scripts/bootstrap` tries to install `golangci-lint` into `.bin/` when network access is available.
- If installation is blocked, it writes a local shim that runs `gofmt -l` and `go vet ./...` so `scripts/check` still works offline.

## 2026-06-25: Centered shell layout with 10% outer margins and remapped section digits

- Keep the app shell centered with roughly 10% outer margins on both axes.
- Split `Global / Context` at a 3:7 width ratio.
- Map digit shortcuts to `1 Graph`, `2 Local`, `3 Remote`, `4 Tags`.
- Keep `Graph` height equal to the combined `Local + Remote + Tags` rail height.
- Render confirm/reset popups with a clamped overlay width so the surrounding layout stays intact.
