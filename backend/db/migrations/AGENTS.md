# AGENTS.md — backend/db/migrations/

## OVERVIEW

SQL schema history for golang-migrate. Files are ordered, paired, and should stay reversible.

## WHERE TO LOOK

| Need | Location | Notes |
|------|----------|-------|
| Base schema | `000001_init_schema.*.sql` | initial tables and constraints |
| Category seed conventions | `000004_category_fk_set_null_and_seed.*.sql` | fixed UUID seed data |
| Listing image schema | `000005_listing_images_cloudinary_metadata.*.sql` | image metadata additions |

## CONVENTIONS

- Add new migrations as paired `NNNNNN_name.up.sql` and `.down.sql` files.
- Keep down migrations meaningful when practical.
- Preserve deterministic seed IDs when tests depend on them.
- Favor additive forward migrations over editing already-applied history.

## ANTI-PATTERNS

- **NEVER** change old migrations casually after they are in shared history.
- **NEVER** add app-only assumptions without matching domain/repository changes.
