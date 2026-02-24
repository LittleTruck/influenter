# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Influenter is an AI-powered collaboration case management system for Taiwanese creators (KOL/influencers). It syncs Gmail emails, uses AI to classify and extract case info, and organizes them into trackable projects.

## Development Commands

### Full stack (Docker)
```bash
make dev              # Start all services (PostgreSQL, Redis, backend API, worker, frontend)
make down             # Stop all services
make restart          # Restart all services
make logs-api         # View backend API logs
make logs-worker      # View background worker logs
```

### Frontend only
```bash
cd frontend && npm install    # Install dependencies
cd frontend && npm run dev    # Dev server on :3000
cd frontend && npm run build
cd frontend && npm run lint
cd frontend && npm run typecheck
```

### Backend only
```bash
cd backend && go test ./... -v       # Run all tests
cd backend && go run ./cmd/server/main.go   # API server on :8080
cd backend && go run ./cmd/worker/main.go   # Background worker
```

### Database migrations
```bash
make migrate-up                           # Run pending migrations
make migrate-down                         # Rollback latest migration
make migrate-create NAME=migration_name   # Create new migration
make db-reset                             # Drop and recreate database
```

## Architecture

### Tech Stack
- **Frontend**: Nuxt 4 (Vue 3) + Nuxt UI + TypeScript + Pinia
- **Backend**: Go 1.24 + Gin + GORM + PostgreSQL + Redis
- **Background jobs**: Asynq (Redis-based task queue)
- **AI**: OpenAI GPT-4 via function calling for structured extraction
- **Auth**: Google OAuth 2.0 + JWT

### Request Flow
```
Frontend (:3000) → REST API (:8080) → PostgreSQL (:5432)
                                    → Redis (:6379) → Asynq Worker
                                    → Gmail API (OAuth2)
                                    → OpenAI API
```

### Backend Structure (`backend/`)
- `cmd/server/main.go` — API entry point with all route definitions
- `cmd/worker/main.go` — Background worker entry point
- `cmd/migrate/main.go` — Database migration CLI
- `internal/api/` — HTTP handlers (auth, emails, gmail, cases)
- `internal/services/` — Business logic (auth, gmail/, openai/)
- `internal/models/` — GORM models (User, Email, Case, OAuthAccount)
- `internal/middleware/` — Auth middleware, logging
- `internal/workers/` — Asynq task handlers (email sync)
- `internal/config/` — Config loading from env vars
- `migrations/` — SQL migration files (up/down pairs)

### Frontend Structure (`frontend/app/`)
- `pages/` — File-based routing: `cases/[id].vue`, `emails/[id].vue`, `calendar.vue`, `settings/`
- `stores/` — Pinia stores: auth, emails, cases, caseFields, collaborationItems
- `composables/` — Reusable logic: useAuth, useCases, useCaseFields, useCalendar, useEmailSanitizer, useErrorHandler
- `components/` — Organized by feature: `cases/`, `emails/`, `calendar/`, `settings/`, `base/`, `common/`, `ui/`
- `middleware/` — Route guards: `auth.ts` (protected), `guest.ts` (login only)
- `types/` — TypeScript type definitions

### API Routes
All protected routes require `Authorization: Bearer <JWT>` header. Base path: `/api/v1`.

Key route groups:
- `POST /auth/google` — Google OAuth login
- `GET/PATCH /emails/:id`, `POST /emails/:id/send-reply`
- `GET/POST /cases`, `GET /cases/:id`, `POST /cases/:id/draft-reply`
- `POST /gmail/sync`, `GET /gmail/status`

Swagger docs available at `http://localhost:8080/swagger/index.html`.

### Database
PostgreSQL with GORM. UUID primary keys. Migrations in `backend/migrations/` as numbered SQL files. Key models: User, OAuthAccount (encrypted OAuth tokens), Email (with AI analysis fields), Case (with status enum: to_confirm, in_progress, completed, cancelled, other).

## Key Patterns

- **Frontend auth**: JWT stored in localStorage, validated by `auth.ts` middleware. Google login via `vue3-google-login` plugin.
- **Email sync**: User triggers sync → task enqueued to Redis via Asynq → worker fetches from Gmail API → stores in DB → triggers OpenAI analysis. Auto-sync every 5 minutes.
- **AI services** (`backend/internal/services/openai/`): Use OpenAI function calling to extract structured data (brand, contact, amounts, deadlines) from emails, classify emails, and generate draft replies.
- **Navigation from cases to emails**: Email links from case pages include `?from_case=<caseId>` query param so the email detail page's back button returns to the originating case.

## Environment Variables

See `.env.example`. Key vars: `JWT_SECRET`, `ENCRYPTION_KEY`, `DB_*`, `REDIS_ADDR`, `OPENAI_API_KEY`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`, `NUXT_PUBLIC_API_BASE`.

## Language

This project's UI and comments are in Traditional Chinese (zh-TW). Commit messages and code identifiers are in English.
