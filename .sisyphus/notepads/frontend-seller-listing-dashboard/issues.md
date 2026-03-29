# Issues

- `frontend/README.md` still contains create-next-app boilerplate and Vercel starter copy. This did not block the seller-dashboard plan acceptance criteria, so it remains a follow-up documentation cleanup item outside the implemented scope.
- Local browser verification for buyer-saved-listings hit live-data blockers: `/listings` returns 0 results because the frontend fetch to `http://127.0.0.1:8080/api/categories` was blocked by CORS when the Go service is not running, so no listing cards (and save controls) can render without the backend.
- Clicking the detail-page save control (e.g., `/listings/test`) as an anonymous user never redirected to `/login`; the button tries to bootstrap `/auth/me` via the same 8080 API client and bails out in the promise catch, leaving the page on `/listings/test` with no error text. The lacking backend/CORS makes the anonymous save flow untestable until the Go auth/data service is available.
