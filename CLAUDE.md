# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

High-concurrency event ticket reservation system in Go. Guarantees zero double bookings through a three-layer defense: distributed locks (Redis), optimistic locking (PostgreSQL), and idempotency keys.

## Common Commands

```bash
# Development
make docker-up          # Start PostgreSQL (5433) + Redis (6379)
make migrate-up         # Apply database migrations
make run                # Start server at localhost:8080

# Testing
make test               # Run all tests with -race -cover
make test-coverage      # Generate coverage.html
make test-integration   # Integration tests (requires Docker)
go test -v -run TestName ./path/to/package  # Run single test

# Quality
make lint               # golangci-lint
make install-tools      # Install golangci-lint and migrate CLI

# Monitoring
make monitoring-up      # Start Prometheus + Grafana stack
```

## Architecture

**Clean Architecture** with dependency flow: `api/` → `application/` → `domain/` ← `infrastructure/`

```
cmd/api/main.go              # Entry point with DI container
internal/
  domain/                    # Pure business logic (no external deps)
    event/                   # Event entity, repository interface, errors
    seat/                    # Seat states: available → reserved → confirmed
    reservation/             # 15-minute expiration, idempotency key
    transaction/             # Tx and Manager interfaces
  application/               # Services with transaction boundaries
  infrastructure/
    postgres/                # sqlx-based repository implementations
    redis/                   # Distributed lock + seat cache (Lua scripts)
  api/handler/               # Echo handlers with Swagger annotations
  worker/                    # ExpiredReservationCleaner (auto-cancel)
db/migrations/               # golang-migrate SQL files
```

**Key principle**: Domain layer has zero external dependencies. Infrastructure implements domain interfaces.

## Three-Layer Defense Implementation

### 1. Distributed Lock (Redis)
Located in `internal/infrastructure/redis/distributed_lock.go`:
```go
lock, err := s.lockManager.AcquireLockWithRetry(ctx, lockKey, 10*time.Second, 3, 100*time.Millisecond)
defer lock.Release(ctx)  // Always defer release
```
Uses Lua script for atomic owner verification before delete.

### 2. Optimistic Locking (PostgreSQL)
Located in `internal/infrastructure/postgres/seat_repository.go`:
```go
query := `UPDATE seats SET status = 'reserved', version = version + 1
          WHERE id = ANY($1) AND status = 'available'`
// Check RowsAffected() - if != expected, return seat.ErrSeatAlreadyReserved
```

### 3. Idempotency Check
In reservation service - check `idempotency_key` before transaction:
```go
existing, err := s.reservationRepo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
if err == nil { return existing, nil }  // Return existing reservation
```

## Code Patterns

### Transaction Boundaries
**Application layer only** manages transactions. Repositories receive `*sqlx.Tx`:
```go
tx, _ := s.db.BeginTxx(ctx, nil)
defer tx.Rollback()  // Required
s.reservationRepo.Create(ctx, tx, res)
s.seatRepo.ReserveSeats(ctx, tx, seatIDs, res.ID)
tx.Commit()
```

### Domain Errors
Each domain package defines errors in `errors.go`. Use `errors.Is()` for comparison:
- `seat.ErrSeatAlreadyReserved`, `seat.ErrOptimisticLockConflict`
- `reservation.ErrReservationNotFound`, `event.ErrEventNotOpen`

### Conventions
- First parameter: `context.Context` for I/O functions
- SQL placeholders: `$1, $2` (PostgreSQL format)
- Array parameters: `pq.Array()`
- Logging: `zap` with `logger.With(zap.String(...))`
- Error wrapping: `fmt.Errorf("予約作成に失敗: %w", err)`

## Testing Strategy

| Layer | Coverage | Type |
|-------|----------|------|
| Domain | 100% | Unit tests (pure Go) |
| Application | 89% | Unit + scenario tests (mocked repos) |
| Handler | 91% | Unit tests (mocked services) |
| Infrastructure | - | E2E tests with real DB/Redis |

**Patterns used**:
- Table-driven tests
- `testify/mock` for repositories
- `sync.WaitGroup` for concurrency testing
- E2E tests in `e2e/reservation_flow_test.go`

**TDD workflow required** (Red → Green → Refactor):
1. Write one failing test
2. Implement minimum code to pass
3. Refactor while keeping tests green

## Database Migrations

Using golang-migrate. Files in `db/migrations/`:
```bash
make migrate-create    # Create new migration (prompts for name)
make migrate-up        # Apply migrations
make migrate-down      # Rollback last migration
make migrate-status    # Check current version
```

## API Design

Base path: `/api/v1`. Authentication via `X-User-ID` header (demo; use JWT in production).

Key endpoints:
- `POST /reservations` - Create reservation (requires idempotency key)
- `POST /reservations/:id/confirm` - Confirm within 15 minutes
- `POST /reservations/:id/cancel` - Cancel reservation
- `GET /metrics` - Prometheus metrics (intentionally public)
- `GET /swagger/*` - Swagger UI

## Monitoring

**Prometheus metrics** (`internal/pkg/metrics/metrics.go`):
- `http_requests_total` - Counter (method, path, status_code)
- `http_request_duration_seconds` - Histogram
- `reservations_total` - Counter (status: success/conflict/lock_failed)
- `distributed_lock_duration_seconds` - Histogram
- `active_reservations` - Gauge

## CI/CD

GitHub Actions (`.github/workflows/ci.yml`):
1. **lint** - golangci-lint
2. **security** - govulncheck
3. **test** - Full suite with PostgreSQL 16 + Redis 7 services
4. **build** - Compile binary

Tests run with `-race -short -p 1` to serialize package execution.

## Git Commit Guidelines

- Do not add Claude signatures (`🤖 Generated with Claude Code` or `Co-Authored-By`) to commit messages
- Write commit messages in Japanese
