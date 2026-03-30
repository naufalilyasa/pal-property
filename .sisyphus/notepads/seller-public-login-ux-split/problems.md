
- Resolved Task 4 polish: seller-protected deep links are no longer trapped by the dashboard layout with a generic `/dashboard` return path; the layout lets unauthenticated requests fall through so page-level seller `requireUser({ returnTo })` preserves the specific seller route destination, and `/dashboard/listings/new` is now explicitly guarded with seller intent too.

- Task 5 browser verification is environment-constrained in this worktree shell because `node`/`npm` are unavailable. Unit coverage exists for `/login`, `/seller/login`, and destination helpers, but Playwright/browser verification must be judged in the final review wave rather than locally executed here.
