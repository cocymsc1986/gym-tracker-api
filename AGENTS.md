# Agent Guide

A practical reference for AI coding agents (and humans new to the repo)
working on tickets against `gym-tracker-api`. Read `README.md` first
for the product/architecture overview; this file documents the
patterns, conventions, and workflow.

---

## Stack at a glance

- **Language:** Go (`go 1.20` in `go.mod`; CI builds on 1.21)
- **Router:** `github.com/gorilla/mux`
- **AWS SDK:** `github.com/aws/aws-sdk-go` (v1, not v2)
- **Lambda adapter:** `github.com/akrylysov/algnhsa` — lets the same
  `http.Handler` serve Lambda or a local HTTP listener
- **Storage:** DynamoDB (two tables, `Workouts-{env}` and
  `Exercises-{env}`; one GSI on exercise type)
- **Auth:** AWS Cognito User Pool (validated by token on every
  protected request)
- **IaC:** Terraform under `terraform/`
- **Testing:** stdlib `testing`, hand-rolled mocks (no testify, no
  gomock)

---

## Layered architecture

```
HTTP request
  → handler   (internal/handlers)         parse path/body, call service
  → service   (internal/services)         validate + business logic
  → repo iface (internal/repository)      contract
  → DynamoDB  (internal/repository/db)    aws-sdk-go calls
```

Rules:

1. **Handlers stay thin.** Decode the request, call one service
   method, format the response. No business logic, no DB calls.
2. **Services own validation and orchestration.** They never import
   `aws-sdk-go` or anything DynamoDB-specific. They depend only on
   `repository.WorkoutRepository` / `repository.ExerciseRepository`
   interfaces from `internal/repository/interfaces.go`.
3. **Repositories are the only place that touches DynamoDB.** When you
   add a new query, add it to the interface in
   `internal/repository/interfaces.go` first, then implement it in
   `internal/repository/db/`.
4. **Models hold their own `Validate()`** (`internal/models/workout.go`,
   `internal/models/exercise.go`). Services call `Validate()` before
   writes; handlers do not.
5. **Dependency injection is wired in `cmd/api/main.go`** in
   `setupHandlers()`. New service/handler? Construct it there.

---

## Conventions

### File and package naming

- Packages match their directory: `handlers`, `services`,
  `repository`, `db`, `models`, `middleware`, `utils`.
- Concrete repo implementations are named `Dynamo<Entity>Repository`
  with a `NewDynamo<Entity>Repository` constructor.
- Service interfaces are exported (`WorkoutService`), implementations
  are lower-case (`workoutService`) and only constructed via
  `NewWorkoutService`.

### HTTP shape

- All workout/exercise routes are scoped to `{userId}` in the path.
  The user ID is the Cognito `sub`. There is no implicit "current
  user" — callers pass it explicitly.
- Successful create → `201`, successful update/get → `200`, delete →
  `204 No Content`.
- Error responses use `utils.WriteErrorResponse` which formats
  `{"error": "..."}` JSON and picks a status code based on the error
  type (`HTTPError`, `validator.ValidationErrors`, else 500). When
  introducing a new "expected" error, either return an `HTTPError`
  with an explicit status or extend `WriteErrorResponse`.
- Auth handlers in `internal/handlers/auth_handler.go` predate the
  service layer and write JSON directly. Don't model new code on
  them — copy the pattern from `workout_handler.go` instead.

### Errors

- Sentinel errors live in `internal/models/errors.go`. Reuse them
  (`ErrWorkoutNotFound`, `ErrExerciseNotFound`, ...) instead of
  inventing parallel ones. Service code uses
  `errors.Is(err, models.ErrXxx)` — see
  `workout_service_test.go:233`.
- Repository methods wrap underlying AWS errors with `fmt.Errorf("...: %w", err)`.
- When a service expects "not found", it currently relies on the repo
  returning either an error or a nil item (see `AddExerciseToWorkout`
  branching on `workout == nil`). Mirror whichever style is already
  used for that entity.

### Validation

- Add field-level rules as struct tags AND in `Validate()`.
  `internal/utils/utils.go` knows how to format
  `*validator.ValidationErrors`, but service code currently only
  calls the hand-written `Validate()` methods. If you add struct-tag
  validation, also call the validator from the service layer or
  document why not.

### IDs and timestamps

- `utils.GenerateUUID()` mints new IDs. Use it for both workouts and
  exercises (the import tool uses `uuid.New().String()` directly,
  which is equivalent).
- `utils.GetCurrentTime()` returns UTC. Use it for `CreatedAt` etc.
- The CreatedAt fallback in `DynamoWorkoutRepository.Create` exists
  for the import script's benefit; handlers always pre-set it.

### CORS

- Allowed origins come from the comma-separated
  `CORS_ALLOWED_ORIGINS` env var, with sensible localhost defaults.
- The middleware supports `*`, exact matches, `*.domain.com` wildcards,
  and treats `https://www.foo.com` as equivalent to `https://foo.com`.
  Don't reimplement origin logic — extend `cors.go` if a new case is
  needed, and add a table-driven test next to `cors_test.go`.

---

## Adding a new endpoint (checklist)

Use this when a ticket asks for "add a way to do X". Touch files in
this order so each step compiles before the next:

1. **Model** (if new fields are needed) — extend the struct, JSON
   tags, dynamodbav tags, and `Validate()`.
2. **Repository interface** — add the method to
   `internal/repository/interfaces.go`.
3. **DynamoDB implementation** — implement the method in
   `internal/repository/db/*.go`. Use `GetItem` / `Query` /
   `dynamodbattribute.UnmarshalMap` consistent with existing methods.
   Remember to add a Global Secondary Index in
   `terraform/dynamodb.tf` if you need a new query pattern.
4. **Service** — add a method to the service interface and
   implementation. Validate inputs, call the repo, wrap errors.
5. **Service tests** — extend the existing `mockXxxRepo` in
   `internal/services/*_test.go` to satisfy any new interface
   method, then add `Test<Method>_Success` and
   `Test<Method>_<FailureMode>` cases following the table of existing
   tests.
6. **Handler** — pull path vars / body, call the service, write
   response with `utils.WriteJSONResponse` / `utils.WriteErrorResponse`.
7. **Route** — register the route in `cmd/api/main.go`. Wrap with
   `authMiddleware.Authenticate` unless it's intentionally public.
8. **Run `go test ./...`** before pushing.

If the change requires new infrastructure (a Cognito attribute, a
DynamoDB GSI, a new env var), update `terraform/` in the same PR and
make sure the new env var is added to both `lambda.tf` and
`.vscode/launch.json` (for local dev).

---

## Testing patterns

`internal/services/workout_service_test.go` is the reference. The
shape:

```go
type mockWorkoutRepo struct {
    workout *models.Workout
    err     error
    updated *models.Workout  // capture writes when you want to assert on them
}

func (m *mockWorkoutRepo) GetByID(...) (*models.Workout, error) { return m.workout, m.err }
// ... satisfy every method on the interface

func TestSomething_Success(t *testing.T) {
    svc := NewWorkoutService(&mockWorkoutRepo{workout: sampleWorkout()})
    got, err := svc.GetWorkout("user-1", "workout-1")
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if got.WorkoutID != "workout-1" { t.Errorf(...) }
}
```

Conventions:

- One file per service / middleware: `<thing>_test.go` next to the
  source.
- Hand-rolled mocks satisfying the repo interface. **No mocking
  libraries.** If you add a method to a repo interface, you must add a
  stub on every mock in `*_test.go` or the tests stop compiling.
- Table tests welcome — see `cors_test.go` for a clean example.
- Cover the happy path, validation failures, and the repo-error path.
- Handler / repository / Lambda glue is **not** unit tested. Don't
  retroactively add HTTP tests unless the ticket asks — keep changes
  scoped.

Run a single package: `go test ./internal/services/...`.
Run a single test: `go test ./internal/services -run TestCalculateRPM_KM`.

---

## Common commands

```bash
go test ./...                          # full test suite (also runs in CI)
go run ./cmd/api                       # local HTTP server on :8080 (needs AWS env)
DEV_MODE=true go run ./cmd/api         # local HTTP server, in-memory, no AWS deps
./build.sh                             # produce lambda.zip (local packaging)
docker build -t gym-tracker-api:dev .  # build the dev container
docker run --rm -p 8080:8080 gym-tracker-api:dev
go run ./cmd/import --user-id ... --file ... --env test --dry-run
go run ./cmd/analyze --file ...
go run ./cmd/reset  --user-id ... --env test --dry-run
go fmt ./...
go vet ./...
```

---

## Dev mode (`DEV_MODE=true`)

Used by QA agents and local hacking. All dev-only code is isolated:

- `cmd/api/dev.go` — `runDev()`, stub auth handlers, seed data. Only
  reached when `DEV_MODE=true`; `main.go` has a 4-line branch at the
  top that returns into it.
- `internal/repository/memory/` — `WorkoutRepository` and
  `ExerciseRepository` implementations of the same interfaces the
  Dynamo repos satisfy. Thread-safe (`sync.RWMutex`), store value
  copies (no aliasing), mirror Dynamo error messages ("workout not
  found" etc.) so service-layer behavior is consistent.

Behavior contract dev mode must preserve:

- `Authorization` header is **accepted but not validated**. Pass-through
  middleware only.
- All `/auth/*` routes return shape-correct JSON. `/auth/signin`
  returns a real `handlers.AuthResponse` plus a `user_id: "dev-user"`
  field for callers that don't decode tokens.
- Seed data is created on boot under `DevUserID = "dev-user"`.
  Currently 2 workouts and 3 exercises (weights / cardio /
  body_weight). The exact contents can change but a QA agent should
  always be able to `GET /workouts/dev-user` and find at least one
  workout immediately after start.
- Data resets when the process restarts. No persistence.

Rules when changing dev mode:

1. **Never make production code depend on `cmd/api/dev.go` or
   `internal/repository/memory`.** Those packages are only reachable
   from the dev branch. If you find yourself wanting to share code
   between dev wiring and prod wiring, prefer to extract a helper
   into a neutral package rather than have prod import dev code.
2. **Keep the dev branch self-contained.** The whole `runDev()` flow
   should boot to a serving HTTP listener with `DEV_MODE=true` and
   nothing else set.
3. **Route parity matters.** Every route registered in `main.go` must
   also be registered in `runDev()`, otherwise dev becomes a
   misleading test target. If you add a new route in `main.go`, mirror
   it in `runDev()`.
4. **Don't add real Cognito calls to dev handlers.** The point is no
   network dependencies. If you need richer auth behavior in dev, fake
   it in `dev.go`.

The Dockerfile (`Dockerfile` at the repo root) bakes `DEV_MODE=true`
into the image so `docker run -p 8080:8080 gym-tracker-api:dev` works
with no flags. Don't repurpose this image for prod — there's no
`bootstrap` binary and no Lambda adapter wiring.

---

## Branching & PR workflow

- The default branch is `main`. Merges to `main` auto-deploy to the
  `test` environment via `.github/workflows/deploy.yml`.
- Feature branches: anything except `main` / `develop` triggers
  `build.yaml` (test + build). Don't push directly to `main`.
- If you're an automated agent operating under a `claude/*` branch,
  stay on the branch the harness assigns you. Create the branch
  locally if it doesn't exist; push with `git push -u origin <branch>`.
- Production deploys are **manual only** — `workflow_dispatch` with
  `environment=production`. Don't try to automate this without
  explicit ask.

When opening a PR, summarize:

1. The behavior change (user-visible)
2. Which layers were touched (handler / service / repo / terraform)
3. Whether infra changes are needed (new env var, GSI, IAM permission)

---

## Gotchas worth knowing

- **`build.sh` produces `main`, the deploy workflow produces
  `bootstrap`.** The Lambda runtime is `provided.al2`, which **requires**
  the binary to be named `bootstrap`. If you're testing a Lambda
  package by hand, rename `main` → `bootstrap` before zipping.
- **`godotenv.Load("../../.env")`** in `cmd/api/main.go` resolves
  relative to the CWD. `go run ./cmd/api` from the repo root will not
  find a root-level `.env` via that path. Either use the VS Code
  launch config (which uses `envFile`) or export vars on the shell.
- **`AuthMiddleware` calls Cognito on every request** with
  `GetUser`. There's no local JWT verification, so it costs an API
  call per request and depends on network reachability of Cognito.
  Don't be surprised by latency in local tests against real Cognito.
- **`Exercise.RPM` is opt-in.** Pass `"storeRpm": true` in the
  create/update body to have the service calculate and persist it.
  The formula assumes 6.2 metres = 1 revolution
  (`services/exercise_service.go:69`).
- **`exercise_respository.go`** — the filename has a typo
  (`respository`). Don't fix it in a drive-by; rename only as part of
  a ticket that touches it for other reasons (it would churn imports
  for no behavior change).
- **`ListByName` queries the `ExerciseTypeIndex` GSI on the `name`
  field** (`exercise_respository.go:178-205`). The GSI is only on
  `ExerciseType` in terraform, so this query as-written may not be
  performant or correct for all records — flag this in any ticket
  that exercises the path.
- **`CreatedAt` on update** — `Workout.Update` does a full `PutItem`
  with whatever the caller sent. If the handler doesn't preserve
  `CreatedAt` from the existing record, it will be wiped. Worth
  double-checking when adding update logic.

---

## When in doubt

- Mirror the workout flow — it's the most consistent example of the
  layered pattern.
- The auth handler is **not** the pattern to copy (it pre-dates the
  refactor).
- `REFACTORING_PLAN.md` describes the original design intent; treat
  it as historical context, not a spec.
- Keep changes scoped to the ticket. Don't refactor neighbouring code
  unless the ticket asks for it.
