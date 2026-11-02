# Weather Watchlist

Search for cities, save them, and see current weather and a 7-day forecast for each one.

- Backend: Go (`net/http`) + SQLite
- Frontend: React, TypeScript, Vite, TanStack Query, Tailwind + shadcn/ui
- Data: [Open-Meteo](https://open-meteo.com) (no API key)

## Run

Requires Go 1.26+ and Node 22.12+.

```sh
make dev-backend    # API on :8080
make install        # once
make dev-frontend   # UI on http://localhost:5173
```

Checks (same as CI): `make lint`, `make test`, `make build`. `make lint` needs golangci-lint v2.
`make install` also enables a pre-commit hook that runs `make lint`.

## API

| Method | Path                           |                         |
| ------ | ------------------------------ | ----------------------- |
| GET    | `/api/places?q=hel`            | Search places           |
| GET    | `/api/locations`               | List saved locations    |
| POST   | `/api/locations`               | Save a location         |
| DELETE | `/api/locations/{id}`          | Remove a location       |
| GET    | `/api/locations/{id}/forecast` | Current + 7-day forecast |

## Structure

```
backend/
  cmd/server/      main.go (wiring), routes.go (all routes)
  internal/
    watchlist/     feature: search, save, list, remove
    forecast/      feature: forecast for a saved location
    openmeteo/     Open-Meteo client
    database/      SQLite + migrations
    httpx/         JSON, errors, middleware
    apperr/        error type
frontend/src/
  components/      PlaceSearch, LocationCard, ui/ (shadcn)
  hooks.ts         React Query hooks
  api.ts           fetch client
```

Backend packages are split by feature. Inside a feature, files are the layers:
`handler.go → service.go → store.go`, with `types.go` for types and validation.
In CI, depguard (`backend/.golangci.yml`) checks imports per file kind, e.g. `service.go`
can't import `net/http`. On the frontend, components go through `hooks.ts` and never call `api.ts` directly,
enforced by oxlint.

## Next steps

- Users and auth: the watchlist is global; scope it per user.
- Production setup: frontend on S3 + CloudFront, backend on ECS Fargate; CloudFront routes
  `/api/*` to the backend, so it stays one origin with no CORS. PostgreSQL instead of SQLite.
- Routing: when a second page is needed, add a router, make `App.tsx` the layout, and move
  the watchlist into `pages/`. `api.ts`, `hooks.ts` and `components/` don't change.
- Observability: request IDs and metrics.
- API contract: OpenAPI spec with generated TypeScript types.
