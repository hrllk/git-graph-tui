# git-graph-tui

`git-graph-tui` is a terminal UI for inspecting a Git repository without leaving the shell.

It focuses on the workflows I actually use in day-to-day work:

- browse local branches, remotes, and tags
- inspect the commit graph
- switch branches
- create branches
- fetch, pull, merge, rebase, reset, and push

## MVP Scope

The graph view now renders from **local branches only**.

That means the graph is built from `refs/heads/*` instead of `--all`, so remote-only branches do not add extra lanes to the main graph. The graph still shows the current branch decoration, branch labels, and local tracking context.

## Requirements

- Go 1.25.6 or newer
- Git

## Run

```bash
go run ./cmd/git-graph-tui
```

## Build

```bash
go build -o git-graph-tui ./cmd/git-graph-tui
```

## Keyboard

- `1` local branches
- `2` remotes
- `3` tags
- `4` graph
- `tab` / `shift+tab` switch sections
- `up` / `down` / `j` / `k` move
- `enter` inspect or execute the current action
- `f` fetch
- `q` quit

## Notes

- The graph uses Git's own ASCII graph output for the main render path.
- When Git does not provide that output, the app falls back to its internal lane renderer.
- This repository is currently an MVP. The UI and command set are intentionally small and opinionated.

