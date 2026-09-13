# shortcuts.codes

Personal site, Go backend, Markdown content, HTMX frontend, runs on rpi5.

## Stack

- Go 1.27, serves via `cmd/main.go`
- Content: Markdown files in `cmd/views/*.md`, parsed by `internal/markdown`
- Templates: `internal/template`, layout in `cmd/views/layout.html`
- Frontend: HTMX (`cmd/scripts/`), plain CSS (`cmd/css/style.css`)
- Build: Ko (Docker image), release-please for CD

## Commands

See `Makefile` for build/test/run targets.

## Conventions

- Edit content by editing the `.md` file in `cmd/views/`, not templates.
- Conventional commits (release-please parses them for versioning).
