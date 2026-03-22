# Workout Data Import Script

Imports historical workout data from a CSV file into DynamoDB, creating one
Exercise record per CSV row and one Workout record per unique date + session pair.

---

## Prerequisites

- Go 1.20+
- AWS credentials with read/write access to the DynamoDB tables:
  - `Workouts-{env}` (e.g. `Workouts-prod`)
  - `Exercises-{env}` (e.g. `Exercises-prod`)

Set the following environment variables before running:

```bash
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
```

---

## CSV Format

The CSV must have the following header row and columns (in order):

| Column          | Description                                        | Stored as                     |
|-----------------|----------------------------------------------------|-------------------------------|
| `date`          | Workout date (`YYYY-MM-DD`)                        | `Workout.Date`                |
| `session`       | Session name (e.g. `Session 1`)                    | `Workout.Name`                |
| `exercise`      | Exercise name                                      | `Exercise.Name`               |
| `type`          | Exercise type: `cardio`, `weights`, or `other`     | `Exercise.ExerciseType`       |
| `sets`          | Number of sets (integer)                           | `Exercise.Sets` length        |
| `reps`          | Reps per set (integer)                             | `Exercise.Sets[*].Reps`       |
| `weight`        | Weight value (number); `0` for bodyweight          | `Exercise.Sets[*].Weight`     |
| `weight_unit`   | Weight unit, e.g. `kg`                             | `Exercise.Sets[*].Unit`       |
| `distance`      | Distance value (number)                            | `Exercise.Distance`           |
| `distance_unit` | Distance unit, e.g. `km` or `m`                   | `Exercise.DistanceUnit`       |
| `round_times`   | Round/set times (see notes below)                  | `Exercise.Time` (cardio only) |
| `effort`        | Effort level (`Easy`, `Moderate`, `Hard`)          | *(not stored)*                |
| `notes`         | Free-text notes                                    | *(not stored)*                |

---

## Flags

| Flag         | Default  | Description                                                        |
|--------------|----------|--------------------------------------------------------------------|
| `--user-id`  | required | Cognito UserID (sub) to assign all imported data to               |
| `--file`     | required | Path to the CSV file                                               |
| `--env`      | `prod`   | DynamoDB table environment suffix (`prod` or `test`)              |
| `--dry-run`  | `false`  | Parse and print what would be written without touching DynamoDB   |

---

## Usage

### 1. Dry run — verify parsing without writing anything

```bash
go run cmd/import/main.go \
  --user-id <your-cognito-sub> \
  --file /path/to/workouts.csv \
  --env prod \
  --dry-run
```

Check the output: confirm the number of workout sessions and exercises matches
what you expect from the CSV.

### 2. Import into the test environment first

```bash
AWS_REGION=us-east-1 \
AWS_ACCESS_KEY_ID=... \
AWS_SECRET_ACCESS_KEY=... \
go run cmd/import/main.go \
  --user-id <your-cognito-sub> \
  --file /path/to/workouts.csv \
  --env test
```

### 3. Verify via the API

```
GET /workouts/{userId}    → should list all imported workout sessions
GET /exercises/{userId}   → should list all imported exercises
```

### 4. Import into prod once satisfied

```bash
AWS_REGION=us-east-1 \
AWS_ACCESS_KEY_ID=... \
AWS_SECRET_ACCESS_KEY=... \
go run cmd/import/main.go \
  --user-id <your-cognito-sub> \
  --file /path/to/workouts.csv \
  --env prod
```

---

## Data Mapping Notes

**Sets and reps**
Every exercise with `sets > 0` gets a `Sets` array of that length, where each
entry holds the weight, unit, and reps for one set. Bodyweight exercises (no
weight in the CSV) are stored with `weight = 0` so the rep data is preserved.

**`round_times` → `Time` (cardio only)**
For cardio exercises, `round_times` is parsed into an average duration in
seconds and stored in `Exercise.Time`. Multiple values separated by ` / ` are
each parsed then averaged:

| CSV value          | Parsed as                       |
|--------------------|---------------------------------|
| `57s`              | 57 s                            |
| `6:05m`            | 6 min 5 s = 365 s               |
| `6:15`             | 6 min 15 s = 375 s              |
| `57s / 59s`        | avg(57, 59) = 58 s              |
| `6:05m / 6:15`     | avg(365, 375) = 370 s           |

`round_times` is ignored for `weights` and `other` exercise types.

**`effort` and `notes`**
These columns are not mapped to any field in the current data model and are
silently dropped during import.
