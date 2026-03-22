# Reset Script

Deletes all exercises and workouts for a given user from DynamoDB. Useful for
wiping a test user's data before re-running the import script.

---

## Prerequisites

- Go 1.20+
- AWS credentials with read/write access to the DynamoDB tables:
  - `Workouts-{env}`
  - `Exercises-{env}`

```bash
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
```

---

## Flags

| Flag        | Default  | Description                                                      |
|-------------|----------|------------------------------------------------------------------|
| `--user-id` | required | Cognito UserID (sub) whose data will be deleted                 |
| `--env`     | `prod`   | DynamoDB table environment suffix (`prod` or `test`)            |
| `--dry-run` | `false`  | List what would be deleted without touching DynamoDB            |

---

## Usage

### 1. Dry run — preview what will be deleted

```bash
go run cmd/reset/main.go \
  --user-id <your-cognito-sub> \
  --env test \
  --dry-run
```

### 2. Delete all data for the user

```bash
AWS_REGION=us-east-1 \
AWS_ACCESS_KEY_ID=... \
AWS_SECRET_ACCESS_KEY=... \
go run cmd/reset/main.go \
  --user-id <your-cognito-sub> \
  --env test
```

### 3. Re-run the import

```bash
AWS_REGION=us-east-1 \
AWS_ACCESS_KEY_ID=... \
AWS_SECRET_ACCESS_KEY=... \
go run cmd/import/main.go \
  --user-id <your-cognito-sub> \
  --file /path/to/workouts.csv \
  --env test
```
