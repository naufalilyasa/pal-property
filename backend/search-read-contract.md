# Search Read Contract Proposal

## Purpose

Future buyer-facing listing discovery should read from Elasticsearch instead of OLTP listing tables.

## Proposed Public Endpoint

- `GET /api/search/listings`

## Proposed Query Surface

- `q`
- `transaction_type`
- `category_id`
- `location_province`
- `location_city`
- `price_min`, `price_max`
- `page`, `limit`
- `sort` (`relevance`, `newest`, `price_asc`, `price_desc`)

## Response Shape

Use the standard backend envelope:

```json
{
  "success": true,
  "message": "Search results fetched successfully.",
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "limit": 20,
    "total_pages": 0
  },
  "trace_id": "..."
}
```

## Search Document Fields

- `id`
- `user_id`
- `category_id`
- `category { id, name, slug }`
- `title`
- `slug`
- `description_excerpt`
- `transaction_type`
- `price`
- `currency`
- `location_province`
- `location_city`
- `location_district`
- `status`
- `is_featured`
- `primary_image_url`
- `image_urls`
- `created_at`
- `updated_at`
- `deleted_at`

## Visibility Rule

- Public search must only return public-safe statuses.
- Seller/admin search, if added later, should be separate routes or explicit authenticated modes.

## Notes

- This is a contract proposal only; the route is not implemented yet.
- Current eventing/indexing work provides the write-side foundation for this future read path.
