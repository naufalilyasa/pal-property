# AGENTS.md — backend/cmd/

## OVERVIEW

Process entrypoints only. Each command wires config, logging, storage, and services for a different runtime mode.

## STRUCTURE

```text
cmd/
├── property-service/     # Fiber API bootstrap + DI wiring
├── listing-indexer/      # outbox worker + rebuild/rebuild-chat CLI
├── migrate/              # golang-migrate runner
└── seed-demo-listings/   # demo seed + search rebuild helper
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| API bootstrap | `property-service/main.go` | Fiber config, Goth store, route/service wiring |
| Search worker | `listing-indexer/main.go` | worker loop, rebuild, rebuild-chat |
| DB migrations | `migrate/main.go` | `up` by default, optional `down` |
| Demo seed flow | `seed-demo-listings/main.go` | wilayah/category/listing/media seed orchestration |

## CONVENTIONS

- Keep `cmd/` focused on wiring, env loading, logger init, and process lifecycle.
- Load runtime config through `pkg/config`; do not duplicate env parsing in subcommands.
- Use `listing-indexer` for operational Elasticsearch rebuilds instead of ad hoc scripts.
- Treat `seed-demo-listings` as a local/demo operator tool, not product business logic.
- Keep command-specific logging explicit; these binaries are long-running or operational entrypoints.

## ANTI-PATTERNS

- **NEVER** move business rules into `cmd/`; push them down into services/repositories.
- **NEVER** hardcode local-vs-docker hosts in code; read env and document mode-specific values.
- **NEVER** add a new binary without documenting its startup contract in parent AGENTS/README.
