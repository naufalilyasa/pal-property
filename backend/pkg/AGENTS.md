# AGENTS.md — backend/pkg/

## OVERVIEW

Reusable support packages shared across the backend. This layer now includes the concrete Cloudinary adapter and provider-agnostic media asset types used by listing images.

## PACKAGES

| Package | Purpose |
|---------|---------|
| `config/` | environment parsing and validation |
| `crypto/` | AES-256-GCM helpers |
| `cloudinary/` | concrete listing-image storage adapter |
| `mediaasset/` | provider-agnostic upload/destroy request/response types |
| `middleware/` | shared Fiber auth/role middleware |
| `logger/` | Zap logger setup |
| `utils/jwt/` | RS256 JWT helpers |
| `utils/slug/` | slug generation |
| `utils/response.go` | JSON success envelopes |

## CONFIG

- `config.LoadConfig()` parses env vars, validates required values, decodes JWT/AES secrets, and validates Cloudinary env combinations.
- Listing-image env guidance:
  - `CLOUDINARY_ENABLED=false` allows the app to start without image credentials.
  - If any Cloudinary credential is set, all three must be set together.
  - If `CLOUDINARY_ENABLED=true`, `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, and `CLOUDINARY_API_SECRET` are all required.

## ANTI-PATTERNS

- **NEVER** import `internal/` packages from `pkg/`.
- **NEVER** add config fields without updating env guidance files.
