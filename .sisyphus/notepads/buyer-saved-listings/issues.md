
# Issues
- 2026-04-??: contains endpoint previously bubbled service limit errors as HTTP 500; added handler-side validation and dedicated test coverage so client errors are returned instead.
- 2026-03-29: contains endpoint rejects empty listing_ids, so the client helper now returns an empty result instead of calling the API to prevent 400 responses while still honoring the shared limit.
- 2026-03-29: Single-resource vs repeated-list `data-testid` values cannot be inferred reliably from render location alone, so the reusable button needs an explicit scope flag to preserve both selector contracts.
- 2026-03-29: real browser QA for public save interactions is locally data-blocked unless a real active listing exists; `/listings` rendered an empty state from the live search API and dummy detail fallbacks intentionally use a non-interactive save button because they have no real listing ID.
- 2026-03-29: `npm run test:e2e -- --grep "saved listings"` is currently blocked by an already-running `next dev` process (`pid 2865048`, port 3000) holding `frontend/.next/dev/lock`, so Playwright's configured webServer for port 3100 exits before the saved-listings spec can boot.
- 2026-03-29: The original `test:e2e` npm script swallowed forwarded CLI args because `npm run ... -- --grep ...` appended them to the trailing `rm` command instead of `playwright test`; wrapping the script in `sh -c ... "$@"` restores targeted Playwright runs.
- 2026-03-29: Manual public save verification remains data-blocked because the live `/api/search/listings` returns zero items and the fallback detail page renders a non-interactive save button until seeded data exists.
