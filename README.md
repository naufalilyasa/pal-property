# Pal Property

Property listing platform for Indonesia with a Go backend, a Next.js frontend, seller tooling, public property discovery, and Elasticsearch-powered search.

## Highlights

- **Seller dashboard** for creating, editing, and managing listings
- **Public browse + detail pages** with interactive maps and media galleries
- **Google OAuth** with backend-managed httpOnly sessions
- **Elasticsearch-backed search** with a dedicated listing indexer worker
- **Rich media support** for up to 10 images and 1 video per listing
- **Wilayah-backed location flow** for province/city/district/village selection
- **Backend RAG chatbot foundation** for property-specific customer support

## Quick Links

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Current Features](#current-features)
- [Repository Structure](#repository-structure)
- [Local Development](#local-development)
- [Environment Configuration](#environment-configuration)
- [Production Deployment](#production-deployment)
- [Testing](#testing)
- [API Notes](#api-notes)

## Overview

Pal Property is a modular-monolith real estate platform focused on:

- public property discovery
- seller listing management
- media-rich property publishing
- Elasticsearch-backed listing search
- Google OAuth authentication with backend-managed sessions

The backend is the source of truth for authentication, authorization, property data, media workflows, and search indexing. The frontend provides a seller dashboard and public browse/detail experiences.

This repository is organized as a practical production-oriented stack:

- **Go/Fiber backend** for auth, listings, media, indexing orchestration, and APIs
- **Next.js frontend** for public discovery and seller workflows
- **PostgreSQL + Redis + Elasticsearch** for persistence, session/memory support, and search

---

## Tech Stack

### Backend
- Go 1.26
- Fiber v3
- GORM
- PostgreSQL 17
- Redis 8
- Elasticsearch 8
- Goth Google OAuth
- Casbin authorization
- Cloudinary media storage

### Frontend
- Next.js 16 (App Router)
- React 19
- TypeScript
- Tailwind CSS v4
- TanStack Query
- React Hook Form
- Zod

### Testing
- Go test + Testify
- Vitest + Testing Library
- Playwright

---

## Current Features

### End-to-End Product Flow
- create seller-owned listings
- upload and manage listing media
- index active properties into Elasticsearch
- browse/search properties publicly
- inspect full listing detail pages with image gallery and map context

### Public Experience
- property browse page at `/listings`
- Elasticsearch-backed search with filters
- interactive listings map with clickable property markers
- property detail page with interactive gallery and zoomable lightbox
- localized property details and cleaner pricing display

### Seller Experience
- Google login with seller/public split UX
- seller dashboard
- tile-based seller listing inventory
- create and edit listing flows
- wilayah-backed location selection:
  - province
  - city/regency
  - district
  - village/subdistrict
- image and video management
- seller actions for:
  - edit
  - mark sold
  - archive
  - delete

### Media & Search
- up to 10 listing images
- single listing video
- primary image selection and image ordering
- dedicated listing indexer worker
- rebuild flow for search indexes

### AI / RAG Backend Foundation
- backend-only chatbot API foundation
- native Gemini Go SDK integration
- Redis session memory foundation
- dedicated chat retrieval index foundation

> The chatbot layer currently exists as a backend foundation and integration surface. A production chat UI can be built on top of it later.

---

## Repository Structure

```text
pal-property/
├── backend/
│   ├── cmd/
│   │   ├── property-service/
│   │   ├── listing-indexer/
│   │   └── migrate/
│   ├── db/migrations/
│   ├── internal/
│   │   ├── domain/
│   │   ├── dto/
│   │   ├── handler/
│   │   ├── repository/
│   │   └── service/
│   └── pkg/
├── frontend/
│   ├── app/
│   ├── components/
│   ├── e2e/
│   ├── features/
│   └── lib/
├── deploy/
├── docker-compose.yml
└── docker-compose.prod.yml
```

## Core Runtime Services

For local or VPS-based backend deployment, the main runtime services are:

- PostgreSQL
- Redis
- Elasticsearch
- backend API
- listing-indexer worker

---

## Architecture Notes

### Source of Truth
- PostgreSQL is the source of truth for application data.
- The Go backend is the source of truth for auth and session state.

### Search
- Public listing search is powered by Elasticsearch.
- Search indexing follows a DB-backed outbox/job pattern.
- `listing-indexer` consumes and projects listing changes into Elasticsearch.

### Auth
- Backend-issued httpOnly cookies
- Google OAuth through backend callback flow
- frontend does not own auth tokens in local storage/session storage

### Media
- image/video upload flows go through backend APIs
- Cloudinary is used as the storage provider

---

## Local Development

## Prerequisites

- Go 1.26+
- Node.js 20+
- Docker + Docker Compose

## 1. Start Infrastructure

```bash
docker compose up -d
```

This starts:
- PostgreSQL
- Redis
- Elasticsearch
- backend
- listing-indexer

> Note: local PostgreSQL is exposed on host port `5433`.

## 2. Backend

```bash
cd backend
go run ./cmd/property-service
```

Useful commands:

```bash
go run ./cmd/migrate/main.go
go test ./... -count=1
go build ./...
go vet ./...
```

## 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Useful commands:

```bash
npm test
npm run build
npm run test:e2e
```

---

## Environment Configuration

The project expects separate backend and frontend environment files.

### Backend
Use:
- `.env`
- `.env.docker`
- `.env.production`

Key areas:
- DB
- Redis
- Elasticsearch
- OAuth
- JWT keys
- Cloudinary
- chat/RAG settings

### Frontend
Use:
- `.env.local`
- `.env.production`

Important values:
- `API_BASE_URL`
- `NEXT_PUBLIC_API_BASE_URL`

Do not commit real secrets.

---

## Production Deployment

Recommended production split for a small VPS:

- **Frontend**: Vercel
- **Backend stack**: VPS
  - backend
  - listing-indexer
  - PostgreSQL
  - Redis
  - Elasticsearch
  - Caddy

Included production assets:

- `backend/Dockerfile.prod`
- `docker-compose.prod.yml`
- `backend/.env.production.example`
- `frontend/.env.production.example`
- `deploy/Caddyfile`

This hybrid deployment is recommended for small VPS instances because Elasticsearch is the heaviest service in the stack, while Next.js frontend hosting is much easier to operate on Vercel.

Bring up the production backend stack with:

```bash
docker compose -f docker-compose.prod.yml --env-file backend/.env.production up -d --build
```

Recommended follow-up:

```bash
docker compose -f docker-compose.prod.yml --env-file backend/.env.production run --rm listing-indexer rebuild
docker compose -f docker-compose.prod.yml --env-file backend/.env.production run --rm listing-indexer rebuild-chat
```

---

## Testing

### Backend

```bash
cd backend
go test ./... -count=1
```

### Frontend Unit Tests

```bash
cd frontend
npm test
```

### Frontend Browser Tests

```bash
cd frontend
npm run test:e2e
```

---

## API Notes

The backend uses a common response envelope:

```json
{
  "success": true,
  "message": "...",
  "data": {},
  "trace_id": "..."
}
```

Relevant API groups include:
- `/auth/*`
- `/api/listings/*`
- `/api/search/listings`
- `/api/me/listings`
- `/api/regions/*`
- `/api/chat/messages`

## Why This Project Exists

Pal Property is designed to provide a focused foundation for an Indonesia-specific property platform:

- strong backend ownership of auth and business rules
- search-oriented public discovery
- seller-first listing operations
- production-friendly infrastructure choices

It is intentionally built as an extensible base for future buyer workflows, richer recommendation systems, and AI-assisted property support.

---

## Status

This repository is actively evolving. The current focus is on:

- seller workflows
- public property discovery
- richer search and map UX
- production deployment hardening
- backend RAG/chatbot capabilities

---

## License

MIT
