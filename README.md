<div align="center">

# TAbelhaHubCLI

**Generates post drafts for [tabelhahub](https://github.com/TAbelhaDev/tabelhahub) from what's actually happening** — recent commits, dirty work, project notes, and vagas the radar picked up.

**English** · [Português](README.pt-BR.md)

[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)

</div>

---

## What it is

A small Go CLI (`tahubcli`) with a single job: turn structured project/market data
into draft blog post suggestions for [tabelhahub](https://github.com/TAbelhaDev/tabelhahub),
the SvelteKit site. It has no state and no TUI — it reads JSON from stdin and
writes JSON to stdout, meant to sit in the middle of a `tabelhaglue` workflow
step (`taradar ipc projects.list --json | tahubcli ipc suggest --json | node
scripts/new-post.mjs`) or run standalone for testing.

## Installation

Requires Go 1.26+.

```bash
go install github.com/TAbelhaDev/tabelhahubcli@latest
```

That installs the binary as `tabelhahubcli` (matching the module name). To get
the short `tahubcli` name used throughout this README, build from source
instead:

```bash
git clone https://github.com/TAbelhaDev/tabelhahubcli.git
cd tabelhahubcli
go build -o tahubcli .
```

## Usage

```bash
tahubcli ipc suggest --json < payload.json
```

`payload.json` is either a bare array of projects, or an object with `projects`
and (optionally) `postings`:

```json
{
  "projects": [
    { "name": "tabelharadar", "dirty_count": 3,
      "last_commit_msg": "feat: suporte a grupos de projeto",
      "memory_notes": ["falta cobrir o caso de grupo vazio"] }
  ],
  "postings": [
    { "title": "Backend Go Pleno", "company": "Acme", "location": "Remoto",
      "score": 82, "url": "https://..." }
  ]
}
```

Each active project (one with a `last_commit_msg`, `dirty_count > 0`, or
`memory_notes`) becomes one post suggestion. When `postings` is present, one
extra "market summary" suggestion is appended, built from the top-N postings
by score (`n=5` by default, override with `n=` as an IPC filter). Output is
always an array of:

```json
{ "title": "...", "slug": "...", "summary": "...", "tags": ["..."], "placeholder": "..." }
```

`slug` never carries a date prefix and `placeholder` is markdown body only, no
frontmatter — both are added downstream by the hub's `scripts/new-post.mjs`.

## Development

```bash
go build ./...
go test ./...
```

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for the version history.

## License

[GNU AGPL-3.0](LICENSE) — free and open source. If you run a modified version of
this project, including as a network service, you also have to make the modified
source available under the same license.
