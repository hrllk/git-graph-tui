# graphkeeper Roadmap

## Goal

This document shows the order for cleaning up `graphkeeper`.
It is about structure first, not new features first.

## Step 1: Freeze the current state

Goal:
- lock the current behavior with tests
- set a baseline before changes

Work:
- run `go test ./...`
- run `go build ./cmd/graphkeeper`
- check the main user flows
- list temporary test files for cleanup

Done when:
- the current behavior still works
- we have a clear test baseline

## Step 2: Clean up the entrypoint

Goal:
- make the main path small and clear

Work:
- add `cmd/graphkeeper/main.go`
- add `internal/bootstrap/app.go`
- keep main as wiring only
- update the README build path

Done when:
- the entrypoint is thin
- app setup lives in one place

## Step 3: Split Git and graph logic

Goal:
- separate raw Git work from graph rules

Work:
- keep raw Git calls in `internal/git`
- move lane/order/selection logic to `internal/graph`
- add pure function tests

Done when:
- graph logic is no longer tied to UI code
- Git calls and Git interpretation are separate

## Step 4: Split the Bubble Tea app files

Goal:
- make `model.go` smaller and easier to read

Work:
- move update logic to `update.go`
- move command code to `commands.go`
- move cursor work to `navigation.go`
- move action flow to `actions.go`
- keep `view.go` focused on rendering

Done when:
- state changes, command creation, and rendering are split

## Step 5: Clean up the UI layer

Goal:
- make styles and widgets easy to reuse

Work:
- add `internal/ui/theme.go`
- add `internal/ui/widgets.go`
- move shared render helpers out of view code

Done when:
- styles are not spread across many files
- render code is easier to follow

## Step 6: Keep docs in sync

Goal:
- make the structure easy to understand for new readers

Work:
- keep `docs/architecture.md`
- keep `docs/roadmap.md`
- keep `docs/cli-structure-plan.md`
- update README when the code layout changes

Done when:
- docs match the code layout
- the binary name and folder names make sense together

## Archive rule

Old UX plans and one-off implementation notes move to `docs/archive/`.
New docs should focus on the current structure and the next steps.

## Recommended order

1. freeze the baseline with tests
2. split cmd/bootstrap
3. split git/graph
4. split app files
5. clean up UI/theme
6. update docs
7. run full build and tests
