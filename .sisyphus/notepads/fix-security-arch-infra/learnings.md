- Used `caarlos0/env` to replace `viper` which prevents issues parsing environment variables and makes validation much easier.
- Learned that RedPanda is a 1-to-1 API compatible Kafka distribution that has better performance characteristics while lacking Zookeeper dependency, which made `docker-compose` cleaner.
- AES-256-GCM encryption added effectively wraps tokens securely. Key management via environment var is critical here.
- CORS in Fiber router requires `AllowCredentials: true` to correctly use HTTP-only cookies in a cross-origin setting.
- Documented that the canonical DB SSL env key is `DB_SSL_MODE` across `backend/pkg/config/config.go`, `.env-example`, `.env`, and `.env.docker`, and that tracked OAuth callback examples now match `/auth/oauth/google/callback` alongside the handler and Postman collection.
- 2026-03-19: Tunet docs show a migration-safe rollout that writes plaintext+ciphertext in `dual` mode and reads ciphertext first, only touching plaintext-only rows when falling back before moving to `enc_only`.
- 2026-03-19: Password hashing guide proves the standard pattern: try the modern hash, treat failure as genuine plaintext only when the stored value clearly isn't hashed, then upgrade the row on success.
- 2026-03-19: Auth repo now tries `crypto.Decrypt`, falls back to legacy plaintext only for guarded plaintext-like failures, and keeps malformed or tampered ciphertexts as errors; tests cover encrypted, plaintext, base64-like plaintext, and invalid inputs.
- 2026-03-19: Heavy cleanup in `auth_service.go` keeps domain sentinels intact for `GetMe`/`CompleteAuth` and refresh-token validation while still wrapping token-generation/cache failures; added regression tests for the new semantics so we never reintroduce ad-hoc `errors.New` noise.
- 2026-03-19: Synced the AGENTS docs so they treat Redpanda as the active broker, highlight `caarlos0/env/v11`, affirm encrypted OAuth provider tokens, and call out that prices stay `int64`, which matches the current backend reality.
- 2026-03-19: Sanitized the local `backend/.env.docker` template by replacing OAuth/JWT values with placeholders to keep developer config handling safer.
- 2026-03-19: Registered the broker comment as Redpanda to keep infra notes aligned with the live messaging stack.
- 2026-03-19: Added a config regression test that pins `LoadConfig` to the documented `DB_SSL_MODE` so future tweaks can’t drift back to the old key.

- 2026-03-24: Refresh-token cache validation now treats redis.Nil as unauthorized while bubbling other cache errors so handler surfaces infra failures instead of masking them.
- 2026-03-24: Hardened the OAuth plaintext fallback so truncated/corrupted ciphertext can’t misreport as legacy plaintext and added a regression test that enforces the new boundary.
- 2026-03-24: Finalized the legacy plaintext guard by validating raw plaintext-like shape (no whitespace/control characters) before accepting corrupted base64 input, keeping true legacy tokens safe while rejecting obvious garbage.
* 2026-03-23T19:48:21Z: Runtime QA re-run after cache/plaintext fixes. Verified refreshed error handling distinguishes redis.Nil vs general infra failure and plaintext fallback still preserves decrypt-on-read while rejecting malformed ciphertext; `go build`, targeted service, and repo tests pass.
