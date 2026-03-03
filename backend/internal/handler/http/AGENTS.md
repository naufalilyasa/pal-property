# AGENTS.md — backend/internal/handler/http/

## OVERVIEW

Fiber v3 HTTP handlers. Thin layer — validate input, call service, return response. No business logic.

## PATTERN

```go
type AuthHandler struct {
    svc service.AuthService  // interface, not concrete
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}

func (h *AuthHandler) GetMe(c fiber.Ctx) error {
    userID := c.Locals("user_id").(uuid.UUID)
    user, err := h.svc.GetMe(c.Context(), userID)
    if err != nil {
        return err  // global error handler maps domain errors → HTTP status
    }
    return c.JSON(fiber.Map{"success": true, "data": user})
}
```

## CONVENTIONS

- **Extract user ID**: `c.Locals("user_id").(uuid.UUID)` — set by `middleware.Protected()`
- **Return errors**: `return err` — global error handler in `router.go` maps `domain.Err*` → HTTP codes
- **Cookies**: set via `c.Cookie(&fiber.Cookie{Name: "access_token", HTTPOnly: true, SameSite: "Lax", Secure: isProduction})`
- **Bind JSON**: `c.Bind().JSON(&req)` (Fiber v3 binder API)
- **OAuth bridge**: `adaptor.HTTPHandlerFunc(gothic.BeginAuthHandler)` wraps net/http handlers for Fiber

## RESPONSE SHAPE

Success:
```json
{"success": true, "data": {...}}
```
Error (via global handler):
```json
{"success": false, "message": "...", "data": null, "trace_id": "uuid"}
```

## TEST PATTERN

- Suite: `type AuthHandlerSuite struct { suite.Suite }`
- Uses `testcontainers-go` for real Postgres + Redis
- Config injected via `config.Env = config.AppConfig{...}` in `SetupSuite()`
- Run: `go test ./internal/handler/http/... -run TestAuthHandlerSuite -v`

## ANTI-PATTERNS

- **NEVER** call repository directly from handler
- **NEVER** put `if/else` business logic here — delegate to service
- **NEVER** construct domain entities here — use DTOs
