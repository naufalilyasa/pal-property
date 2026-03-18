# Learnings

- The refreshed root AGENTS needed one extra truthfulness pass after implementation because the repo does not actually contain `workers/`, `infra/`, or `deploy/` directories yet.
- Root guidance must use exact grep-friendly wording for plan verification, especially around auth authority, Axios prohibition, fake storage, and missing `.cursor/rules` files.
