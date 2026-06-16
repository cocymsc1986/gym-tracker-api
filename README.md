# gym-tracker-api

A Go REST API for tracking gym workouts. Users record **workouts** (a named
session on a given date) and **exercises** (cardio, weights, body weight,
other) and link exercises into workouts. The API is deployed as an AWS
Lambda behind API Gateway, with DynamoDB for storage and Cognito for
authentication.

A web/mobile frontend consumes this API (see the `CORS_ALLOWED_ORIGINS`
variable in `terraform/lambda.tf`).

---

## What's in the box

- **Workout & exercise CRUD** keyed by Cognito user ID
- **Cognito-backed auth** (sign up, confirm, sign in, refresh, password
  reset) via `internal/handlers/auth_handler.go`
- **CORS middleware** that supports wildcard subdomains and a `www.` /
  apex equivalence (`internal/middleware/cors.go`)
- **RPM calculation** for cardio exercises (revolutions-per-minute, used
  by the frontend for pacing) — opt-in via `storeRpm` on create/update
- **Lambda + local HTTP** dual-mode entry point — the same binary runs
  under Lambda (`AWS_LAMBDA_RUNTIME_API` set) or as a standalone HTTP
  server on `PORT` (default `8080`)
- **Companion CLI tools** in `cmd/` for CSV import, dry-run analysis,
  and per-user data reset

---

## Architecture

Layered Go service:

```
cmd/api/main.go            HTTP/Lambda entrypoint, route wiring, DI
└─ internal/
   ├─ handlers/            HTTP request/response only
   ├─ middleware/          auth (Cognito) and CORS
   ├─ services/            business logic + validation
   ├─ repository/
   │  ├─ interfaces.go     repository contracts
   │  └─ db/               DynamoDB implementations
   ├─ models/              domain types, errors, Validate() methods
   └─ utils/               JSON I/O helpers, error responses, UUID
```

Dependencies flow downward only: handlers depend on services, services
depend on `repository.*Repository` interfaces. The DynamoDB
implementations live in `internal/repository/db` and are injected at
startup in `cmd/api/main.go`. This makes services and handlers easy to
unit-test against fake repositories — see the `mockWorkoutRepo` and
`mockExerciseRepo` patterns in `internal/services/*_test.go`.

### Routes

All workout/exercise routes are protected by `AuthMiddleware`, which
validates the `Authorization: Bearer <token>` header against Cognito.

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/auth/signup` | Create a Cognito user |
| POST | `/auth/confirm` | Confirm signup with email code |
| POST | `/auth/signin` | Exchange email/password for tokens |
| POST | `/auth/refresh` | Refresh an access token |
| POST | `/auth/reset` | Start password reset |
| POST | `/auth/reset/confirm` | Confirm password reset with code |
| GET | `/workouts/{userId}` | List workouts |
| GET | `/workouts/{userId}/{workoutId}` | Get a workout |
| POST | `/workouts/{userId}` | Create a workout |
| PUT | `/workouts/{userId}/{workoutId}` | Update a workout |
| DELETE | `/workouts/{userId}/{workoutId}` | Delete a workout |
| GET | `/workouts/{userId}/{workoutId}/exercises` | List exercise IDs in a workout |
| POST | `/workouts/{userId}/{workoutId}/exercises/{exerciseId}` | Add an exercise to a workout |
| DELETE | `/workouts/{userId}/{workoutId}/exercises/{exerciseId}` | Remove an exercise from a workout |
| GET | `/exercises/{userId}` | List exercises |
| GET | `/exercises/{userId}/{exerciseId}` | Get an exercise |
| GET | `/exercises/{userId}/name/{exerciseName}` | List exercises by name |
| POST | `/exercises/{userId}` | Create an exercise (body: `{ ...exercise, "storeRpm": bool }`) |
| PUT | `/exercises/{userId}/{exerciseId}` | Update an exercise |
| DELETE | `/exercises/{userId}/{exerciseId}` | Delete an exercise |

### Data model

`Workout` (`internal/models/workout.go`) is the parent record. It owns a
list of `Exercises` which is a slice of **exercise IDs**, not embedded
exercises. Exercises are stored independently in their own DynamoDB
table.

`Exercise` (`internal/models/exercise.go`) supports four types:
`weights`, `cardio`, `body_weight`, `other`. Weight/rep data lives in
`Sets []WeightItem`. Cardio exercises additionally use `Time`,
`Distance`, `DistanceUnit`, and optionally `RPM`.

DynamoDB tables (`terraform/dynamodb.tf`):

- `Workouts-{env}` — hash `UserID`, range `WorkoutID`
- `Exercises-{env}` — hash `UserID`, range `ExerciseID`, GSI
  `ExerciseTypeIndex` on `ExerciseType`

---

## Running locally

### Prerequisites

- Go 1.20+ (CI uses 1.21)
- AWS credentials with read/write access to the DynamoDB and Cognito
  resources you point at
- An existing Cognito user pool & client, or a deployed dev environment
  to borrow IDs from

### Configuration

The API reads everything from env vars and (optionally) a `.env` file
at the repo root. `cmd/api/main.go` calls `godotenv.Load("../../.env")`
which resolves relative to where the binary runs — typically with
`go run ./cmd/api` from the repo root, it will look up two levels. The
VS Code launch configs in `.vscode/launch.json` are the easiest way to
inject env vars; otherwise set them on the shell.

Required:

| Variable | Notes |
|----------|-------|
| `AWS_REGION` | e.g. `us-east-1` |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | Resolved via `credentials.NewEnvCredentials()` |
| `DYNAMO_TABLE_WORKOUTS` | e.g. `Workouts-test` |
| `DYNAMO_TABLE_EXERCISES` | e.g. `Exercises-test` |
| `COGNITO_CLIENT_ID` | Cognito app client id |

Optional:

| Variable | Default | Notes |
|----------|---------|-------|
| `PORT` | `8080` | HTTP port when not running under Lambda |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173,capacitor://localhost` | Comma-separated allowlist; supports `*.domain` and `*` |
| `COGNITO_USER_POOL_ID` | – | Set in Lambda env; not currently read by the Go code |

### Run

```bash
go run ./cmd/api
```

Then hit `http://localhost:8080/...`. Auth endpoints work as-is; for
protected routes, sign in first to grab an access token, then send it
as `Authorization: Bearer <token>`.

### Tests

```bash
go test ./...
```

The service layer is the test-rich layer — handlers and the DynamoDB
repos are not unit-tested in this repo.

### Build (Lambda artifact)

```bash
./build.sh
```

Produces `main` (Linux amd64) and `lambda.zip`. The deploy workflow
builds a `bootstrap` binary instead — `provided.al2` runtime requires
the binary be named `bootstrap`.

---

## Dev mode (no AWS dependencies)

The API can run fully self-contained — no AWS credentials, no DynamoDB,
no Cognito — by setting `DEV_MODE=true`. In dev mode:

- Repositories are **in-memory** (`internal/repository/memory`). Data
  resets on every process start.
- The `Authorization` header is **not checked** — any request reaches
  the handlers.
- `/auth/*` endpoints return **shaped stub responses** so callers can
  exercise the full sign-in flow without a real Cognito.
- Data is **seeded at boot** under the well-known user ID `dev-user`
  (2 workouts, 3 exercises across all three exercise types).
- `/auth/signin` returns a `user_id` field in addition to the standard
  `AuthResponse`, so callers don't need to decode the token.

Production code paths are untouched: when `DEV_MODE` is unset, behavior
is identical to before.

### Run with Go directly

```bash
DEV_MODE=true go run ./cmd/api
# Dev server listening on :8080 (seeded user: dev-user)
```

### Run with Docker (recommended for QA agents)

```bash
docker build -t gym-tracker-api:dev .
docker run --rm -p 8080:8080 gym-tracker-api:dev
```

The image bakes `DEV_MODE=true`, `PORT=8080`, and `CORS_ALLOWED_ORIGINS=*`
as defaults; override any of them with `-e`. The Dockerfile uses a
multi-stage Go build (1.21-alpine → alpine:3.19) and exposes a
healthcheck against `/exercises/dev-user`.

### Smoke test the dev API

```bash
# Get a (stub) token + user id
curl -s -X POST http://localhost:8080/auth/signin \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@test","password":"x"}'
# {"access_token":"dev-access-token","refresh_token":"dev-refresh-token",
#  "token_type":"Bearer","expires_in":3600,"user_id":"dev-user"}

# List seeded workouts
curl -s http://localhost:8080/workouts/dev-user

# List seeded exercises
curl -s http://localhost:8080/exercises/dev-user

# Create a new workout (Authorization header is accepted but not required)
curl -s -X POST http://localhost:8080/workouts/dev-user \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer dev-access-token' \
  -d '{"name":"My Workout","date":"2026-06-16"}'
```

To test against a real DynamoDB / Cognito backend, leave `DEV_MODE`
unset and supply the env vars listed in the [Configuration](#configuration)
section above.

---

## Companion tools

All three live under `cmd/` and ship with their own READMEs.

| Tool | Purpose |
|------|---------|
| `cmd/import` | Bulk-import a CSV of historical workouts into DynamoDB (`go run ./cmd/import --user-id <sub> --file workouts.csv --env test`). See `cmd/import/README.md` for the CSV schema and field mapping. |
| `cmd/analyze` | Dry-scan a CSV for inconsistencies before importing — flags rows missing fields that peers in the same exercise group have. |
| `cmd/reset` | Delete every workout and exercise for a given user (`--user-id`, `--env`). Useful for re-running imports on a test user. Supports `--dry-run`. |

---

## Deployment

CI/CD lives in `.github/workflows/`:

- **`build.yaml`** — runs `go test` and a Linux build on every push to
  any branch except `main` and `develop`.
- **`deploy.yml`** — auto-deploys `main` → `test`, and manual dispatch
  → `test` or `production`. Builds the `bootstrap` binary, packages it,
  and runs `terraform apply` against the matching environment's
  `tfvars` file.

Terraform layout:

- `terraform/*.tf` — shared resources (API Gateway, Lambda, DynamoDB,
  Cognito, IAM)
- `terraform/environments/{test,prod}/` — per-env tfvars and backend
  config (the deploy job copies the chosen `backend.tf` into
  `terraform/` before `terraform init`)
- `terraform/remote-state/` — bootstrap stack for the S3 backend

Secrets the deploy job needs (set on the GitHub `test` /
`production` environments): `AWS_ACCESS_KEY_ID`,
`AWS_SECRET_ACCESS_KEY`, `CORS_ALLOWED_ORIGINS`.

---

## Project layout

```
.
├── cmd/
│   ├── api/           main HTTP/Lambda entrypoint
│   ├── import/        CSV → DynamoDB importer
│   ├── analyze/       CSV consistency checker
│   └── reset/         delete-by-user utility
├── internal/
│   ├── handlers/      HTTP handlers
│   ├── middleware/    auth, CORS
│   ├── services/      business logic + Validate()
│   ├── repository/    interfaces + DynamoDB impls
│   ├── models/        domain types, errors
│   └── utils/         JSON, UUID, time helpers
├── terraform/         IaC for AWS resources
├── .github/workflows/ CI: build + deploy
└── build.sh           local Lambda packaging
```

See `AGENTS.md` for the patterns and conventions to follow when making
changes.
