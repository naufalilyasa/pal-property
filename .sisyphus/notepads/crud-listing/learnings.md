
## CRUD Listing Implementation Learnings

### Backend Patterns
- **Atomic Increments**: Used `gorm.Expr("view_count + 1")` in `IncrementViewCount` to avoid lost updates during concurrent page views. This ensures the database handles the incrementation safely.
- **Partial Updates**: Implemented `Select(fields).Updates(listing)` pattern in the repository. This allows the service layer to specify exactly which fields should be persisted, preventing mass-assignment vulnerabilities and accidental overwrites of fields like `view_count` or `user_id` during a standard update.
- **Unique Slug Generation**: Integrated a slug generation utility that checks for uniqueness via the repository before persisting, ensuring SEO-friendly and unique URLs for every listing.
- **Layered Architecture Consistency**: Strictly followed the `handler -> service -> repository` flow. Domain errors (like `domain.ErrNotFound`) are translated at the repository level and handled globally at the HTTP layer.

### Security
- **Ownership Verification**: Implemented a `checkOwnership` helper in the service layer to ensure users can only update or delete their own listings, with an override for admin roles.
