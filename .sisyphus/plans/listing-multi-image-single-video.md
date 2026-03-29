# Listing Multi Image + Single Video

## Context
- Extend current listing media flow to support multi-image upload plus one optional listing video.
- Preserve current edit-only seller media workflow and minimal UI drift.
- Product defaults: one optional video, same seller media section, 100MB / 60-second validation target, tests-after.

## Execution Order
1. listing-video schema/contracts
2. batch image upload backend support
3. single-video backend upload/delete + response shape
4. frontend media helpers/types
5. seller media section upgrade
6. public detail video rendering + full-stack QA
7. final verification wave

## Tasks
- [x] 1. Add listing-video schema and backend media contracts
- [x] 2. Extend the existing image API to accept batch image uploads
- [x] 3. Add backend single-video upload, delete, and listing response support
- [x] 4. Extend frontend listing media helpers and types for batch images and video
- [x] 5. Upgrade the seller media section for multi-image and single-video management
- [x] 6. Render optional listing video on public detail and complete full-stack media QA

## Final Verification Wave
- [x] F1. Plan Compliance Audit — oracle
- [x] F2. Code Quality Review — unspecified-high
- [x] F3. Real Manual QA — unspecified-high (+ playwright if UI)
- [x] F4. Scope Fidelity Check — deep

## Success Criteria
- Sellers can batch upload images in the existing edit flow
- Sellers can upload/delete at most one optional video per listing
- Existing image primary/reorder/delete behaviors still work
- Public detail can render the optional video without major design changes
