# AGENTS.md — backend/internal/dto/

## OVERVIEW

Transport contracts between HTTP handlers and the outside world. Request validation tags and response payload shapes live here.

## STRUCTURE

```text
dto/
├── request/   # incoming payloads and validation tags
└── response/  # serialized response shapes
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Listing request contract | `request/listing_request.go` | validation tags + pointer semantics |
| Listing response contract | `response/listing_response.go` | public payload + nested category/image shape |
| Auth/category payloads | `request/*`, `response/*` | follow same naming pattern |

## CONVENTIONS

- Request structs carry validation tags used by handlers.
- Use pointer fields when “missing” and “explicit null/zero” must stay distinct.
- Response structs stay transport-focused; avoid service/repository behavior here.
- Keep naming explicit: `CreateXRequest`, `UpdateXRequest`, `XResponse`.

## ANTI-PATTERNS

- **NEVER** add business logic or persistence behavior here.
- **NEVER** leak provider-specific details unless intentionally part of the API contract.
