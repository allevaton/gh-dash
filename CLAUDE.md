# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`gh-dash` is a `gh` CLI extension that ships a Bubble Tea–based terminal UI for GitHub (PRs, issues, notifications, and a per-repo view). Entry point is `gh-dash.go` → `cmd/root.go` → `internal/tui`. Module path is `github.com/dlvhdr/gh-dash/v4` (Go 1.25.8 per go.mod, though the devbox-pinned toolchain is 1.23 — defer to what's already installed in the devbox shell).

## Dev environment

- Tooling is managed by [Devbox](https://github.com/jetify-com/devbox). Run everything from inside `devbox shell` so versions of `go`, `golangci-lint` (2.10.1), `gofumpt`, `go-task`, `nerdfix`, etc. match CI.
- `prism` (test runner) and `gotip` (single-test helper) are installed by the devbox `init_hook` into `.devbox/go/bin`. If `task test` / `task test:one` aren't found, re-enter the devbox shell.

## Common commands

All driven by `Taskfile.yaml` (`task <name>`):

| Task | What it does |
| --- | --- |
| `task` (default) / `go run .` | Run the TUI |
| `task build` | `go build .` |
| `task install` | Build and install as `gh dash` extension (`gh ext install .`) |
| `task debug` | Run with `--debug`, writes to `./debug.log`. Use `task logs` in another pane to tail |
| `task debug:warn` | Same as debug but `LOG_LEVEL=warn` |
| `task dlv` | Headless `dlv` on `127.0.0.1:43000` |
| `task lint` | `golangci-lint run` (canonical CI lint) |
| `task lint:fix` | Lint with `--fix` (also runs `gofumpt`/`goimports`/`golines` formatters) |
| `task fmt` | `gofumpt -w` over tracked Go files |
| `task test` | `prism test ./...` |
| `task test:one` | `gotip` — interactive single-test picker |
| `task test:rerun` | `gotip --rerun` |
| `task check-nerd-font` / `task fix-nerd-font` | Validate/repair Nerd Font glyphs in source |
| `task docs` | Run the Astro docs site at `localhost:4321` (under `docs/`, pnpm) |

Stdlib-style single test: `go test ./internal/tui/components/prssection -run TestX -v`. CI runs `task lint`; running it locally before pushing avoids the maintainer-approval round-trip on fork PRs.

## Logging while debugging

```go
import log "charm.land/log/v2"
log.Debug("some message", "someVariable", someVariable)
```

`task debug` initializes `debug.log`; `task logs` tails it. `LOG_LEVEL` env var (`debug|info|warn|error`) controls level when `--debug` is set.

## Architecture

### Top-level flow

`gh-dash.go` → `cmd.Execute()` (Cobra + `charmbracelet/fang`) → `tui.NewModel(...)` → `tea.NewProgram(model).Run()`. Config resolution order (in `cmd/root.go` and `internal/config`):

1. `--config/-c` flag
2. `.gh-dash.yml` in the current git repo (if any)
3. `$GH_DASH_CONFIG`
4. `$XDG_CONFIG_HOME/gh-dash/config.yml`

The repo-aware code path comes from `internal/git.GetRepoInPwd()` and feeds `RepoPath` into the program context. The `FF_REPO_VIEW` feature flag (see `internal/config/feature_flags.go`) gates the per-repo view and the optional positional repo arg.

### Bubble Tea / Elm architecture

The TUI follows the Elm-style `Model` / `Update` / `View` triad with messages. Familiarity with [Bubble Tea](https://github.com/charmbracelet/bubbletea) is assumed. The code uses the v2 `charm.land/...` modules (note: not the older `github.com/charmbracelet/...` import paths for bubbletea/bubbles/lipgloss/log).

### Package layout (`internal/`)

- `config/` — koanf + YAML parser for the user's `.gh-dash.yml` / `config.yml`. `parser.go` defines `ViewType` (`prs`, `issues`, `notifications`, `repo`) and section configs. Validation via `go-playground/validator`. Templating in section filters uses `go-sprout` (e.g. `{{ nowModify "-3w" }}`).
- `data/` — GitHub API layer. PR/issue/notification fetching uses the GraphQL v4 client (`shurcooL/githubv4` via `cli/shurcooL-graphql`); the `gh` auth/transport comes from `cli/go-gh/v2`. Caching via `maypok86/otter`. `donestore.go` tracks "done" markers; `bookmarks.go` persists bookmarks.
- `git/` — local git helpers (`GetRepoInPwd`, branch lookups via `aymanbagabas/git-module`).
- `tui/` — the entire UI tree. The big picture:
  - `tui/ui.go` — root `Model`. Owns `tabs`, `sidebar`, the slice of `section.Section` for each view (`prs`, `issues`, `notifications`), the `repo` section, and the per-type sidebar/preview models (`prview`, `issueview`, `branchsidebar`, `notificationview`).
  - `tui/context/` — `ProgramContext` is the shared bag passed to every component (config, theme, terminal size, repo path, `StartTask` callback, version). Most components hold a `*context.ProgramContext`.
  - `tui/components/section/` — the abstract `Section` interface + `BaseModel` (table, search bar, prompt, pagination, fetch state). Concrete sections (`prssection`, `issuessection`, `notificationssection`, `reposection`) embed/compose this.
  - `tui/components/{prrow,issuerow,notificationrow,prview,issueview,...}` — row renderers and detail/sidebar views per entity type.
  - `tui/components/{table,tabs,footer,sidebar,prompt,search,inputbox,carousel,listviewport,branch,branchsidebar,cmp,cmpcontroller,tasks}` — generic UI primitives.
  - `tui/keys/` — keybinding maps per view; user overrides come from config.
  - `tui/theme/`, `tui/markdown/`, `tui/common/` — styling, glamour-rendered markdown, shared helpers (diff/labels/repopath).
  - `tui/tasks.go` — long-running task spinner orchestration; sections kick off fetches via `ctx.StartTask`.
- `utils/` — generic helpers and the template handler used to render section filter strings.

### Adding a new section type

A section type generally needs: a `data/` API function, a `components/<x>row` renderer, a `components/<x>section` package implementing `section.Section`, optional `tui/keys/<x>Keys.go`, and wiring in `tui/ui.go` (model fields, message routing, view switching) plus a config entry under `internal/config`.

## Style and lint expectations

- `golangci-lint` config in `.golangci.yml` enables `bodyclose`, `staticcheck`, `misspell`, `nolintlint`, `tparallel`, `whitespace`, etc., and disables `errcheck`/`ineffassign`/`unused`. Formatters enforced: `gofumpt`, `goimports`, `golines` (with `chain-split-dots`).
- `task lint:fix` is the fast path for line-length / import-order failures.
- `gotip` and `prism` are the test runners of choice — see Taskfile. Stdlib `go test` works fine for ad-hoc runs but CI uses the task targets.

## AI / contribution policy

`AI_POLICY.md` is strict: AI usage on contributions must be disclosed, the human contributor must fully understand all submitted code, no AI-generated media. The maintainer exemption applies only to maintainers. Treat anything you generate here as needing human review before it leaves the repo.

## Things that bite

- Imports use the `charm.land/...` v2 paths for bubbletea, bubbles, lipgloss, log, and glamour — don't auto-suggest the older `github.com/charmbracelet/...` paths for those packages (they exist for some sub-packages like `fang`, `x/ansi`, etc., but not for the core TUI libs here).
- The module is `v4` (`github.com/dlvhdr/gh-dash/v4`) — internal imports must include `/v4`.
- `task install` removes any installed `gh dash` extension before reinstalling from the local build (`gh ext remove dash` then `gh ext install .`). Don't run this on a coworker's box without warning.
- `.gh-dash.yml` at repo root is the project's *own* dashboard config (used when devs run `gh dash` here), not a fixture. Don't treat it as test data.
