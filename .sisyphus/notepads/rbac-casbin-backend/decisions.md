# Decisions

- Persisted roles stay `user` and `admin`; seller capability remains ownership-aware.
- Casbin model source of truth lives in `backend/pkg/authz/model.conf` and is embedded into the backend authz package.
- Policy storage is Postgres-backed in `casbin_rule`, created by `000006_casbin_policy` migrations rather than runtime schema creation.
- Authorization is hybrid: `middleware.RequirePermission` handles coarse route checks, while listing services do owner/admin resource checks through `service.AuthzService`.
- Protected requests continue to reload the current DB role every request so role changes take effect without re-login.
