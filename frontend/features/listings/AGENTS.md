# AGENTS.md — frontend/features/listings/

## OVERVIEW

Owns listing-specific UI, server reads, forms, and image workflows across seller and public surfaces.

## STRUCTURE

```text
features/listings/
├── components/  # cards, tables, refresh/filter widgets
├── forms/       # RHF + Zod listing form flow
├── images/      # image action helpers and route patterns
└── server/      # server-side listing reads
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Public listing reads | `server/get-search-listings.ts`, `server/get-listing-by-slug.ts` | server-only |
| Seller listing reads | `server/get-seller-listings.ts` | `/auth/me/listings` boundary |
| Listing form | `forms/listing-form.tsx`, `forms/listing-schema.ts` | RHF + Zod + field mapping |
| Image actions | `images/api.ts` | backend-owned image endpoints only |
| Public search UI | `components/listings-map-panel.tsx`, `components/search-listing-card*.tsx` | search map + card rhythm |

## CONVENTIONS

- Keep backend DTO mapping close to the feature.
- Use RHF + Zod for non-trivial listing forms.
- Keep image ordering/primary/delete/upload flows backend-authoritative.
- Keep public-search query normalization inside `features/listings/server/get-search-listings.ts` and related route components.
- Render backend image URLs with `next/image`; do not invent direct Cloudinary upload flows.

## ANTI-PATTERNS

- **NEVER** split listing schema rules across unrelated files.
- **NEVER** bypass backend image endpoints for seller media actions.
- **NEVER** put public-query contract logic into generic global utils.
