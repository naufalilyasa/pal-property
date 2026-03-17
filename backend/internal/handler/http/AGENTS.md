# AGENTS.md — backend/internal/handler/http/

## OVERVIEW

Fiber v3 HTTP handlers. Thin transport only: parse params/body/files, read auth locals, call the service, return the shared response envelope.

## CONVENTIONS

- Extract auth state from `c.Locals("user_id")` and `c.Locals("user_role")`.
- Pass `c.Context()` into services.
- Bind JSON with `c.Bind().JSON(&req)`.
- For listing-image upload, read multipart input with `c.FormFile("file")` and delegate all validation/storage behavior to `ListingService.UploadImage`.
- Return domain errors directly so the global Fiber error handler can map them.

## TEST PATTERN

- `testify/suite` + testcontainers Postgres.
- Manual route registration in tests is acceptable when the suite only needs one handler family.
- Listing image coverage should use fake storage instead of live Cloudinary.

## ANTI-PATTERNS

- **NEVER** call repositories directly.
- **NEVER** put ownership or reorder business rules here.
- **NEVER** stream files to Cloudinary directly from handlers.
