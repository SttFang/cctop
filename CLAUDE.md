# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is cctop

A terminal monitoring dashboard for Claude Code usage, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). It reads `~/.claude/stats-cache.json` and session JSONL files from `~/.claude/projects/`, indexes them into a local SQLite database (`~/.claude/cctop.db`), and presents an interactive TUI with four tabs: Overview, Sessions, Analytics, and Tools.

## Commands

```bash
make build          # Build binary to bin/cctop (CGO_ENABLED=0)
make test           # Run all tests with -race
make lint           # Run golangci-lint
make coverage       # Generate coverage report
make run            # Build and run

go test ./internal/store/ -v -run TestGetSessions   # Run a single test
```

Version info is injected via ldflags (`pkg/version`).

## Architecture

```
main.go                     Entry point: parse config → create App → run Bubble Tea program
internal/
  config/
    config.go               CLI flags, paths (~/.claude dir), validation
    pricing.go              Per-model token pricing table, cost computation
  data/
    types.go                Data structures for stats-cache.json and session JSONL messages
    statscache.go           Parse stats-cache.json
    session_parser.go       Parse individual session JSONL files into SessionSummary
  store/
    store.go                SQLite wrapper (modernc.org/sqlite, pure Go, no CGO)
    schema.go               DB schema: sessions, tool_calls, sync_state tables
    sync.go                 Incremental sync: scan projects dir, parse changed JSONL → upsert DB
    queries.go              Session listing/filtering/sorting, tool frequency queries
  metrics/
    aggregator.go           Derive Overview metrics from StatsCache; formatting helpers
  tui/
    app.go                  Root Bubble Tea model (AppModel), Init/Update/View, tab state
    render.go               Top-level render: header + content (dispatches to views) + statusbar
    common/                 Theme, layout computation, key bindings
    components/             Reusable UI: statcard, panel, header, statusbar, barchart, linechart, heatmap
    views/                  Tab views: overview, sessions (table+detail), analytics, tools
```

**Data flow**: On startup and every refresh interval (default 30s), the app (1) loads `stats-cache.json` → computes Overview metrics, and (2) incrementally syncs session JSONL files into SQLite (skipping unchanged files via mtime/size). Tab-specific queries run against the SQLite store on tab switch or page navigation.

## Key patterns

- SQLite uses pure-Go driver `modernc.org/sqlite` — no CGO needed, builds cross-platform with `CGO_ENABLED=0`.
- Session sync is incremental: `sync_state` table tracks file mtime/size to skip unmodified files.
- Model pricing is a static lookup table in `config/pricing.go` keyed by model name substrings. When adding new models, add an entry to `PricingTable`.
- Test fixtures live in `testdata/` (sample JSONL files, stats-cache.json). Store tests use `":memory:"` SQLite.
