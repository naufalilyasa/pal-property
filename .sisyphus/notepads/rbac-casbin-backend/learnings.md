# Learnings

- The worktree was created before the untracked RBAC plan existed in the main checkout, so the plan and Boulder state had to be copied/synced into the branch-local `.sisyphus/` workspace before execution could proceed cleanly.
- `github.com/casbin/gorm-adapter/v3` works with a migration-managed schema, but the clean production-safe path still requires explicit table creation in SQL before enforcer bootstrap.
- Handler integration suites needed local Casbin policy-table bootstrapping because they use `AutoMigrate` for app entities only and now compile against the new `Protected(db, authzService)` middleware signature.
