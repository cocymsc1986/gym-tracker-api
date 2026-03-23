package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gym-tracker-api/internal/models"
	repoDb "gym-tracker-api/internal/repository/db"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

// CSV column indices
const (
	colDate         = 0
	colSession      = 1
	colExercise     = 2
	colType         = 3
	colSets         = 4
	colReps         = 5
	colWeight       = 6
	colWeightUnit   = 7
	colDistance     = 8
	colDistanceUnit = 9
	colRoundTimes   = 10
	// colEffort = 11  // not mapped — Level left as zero value
	// colNotes  = 12  // no matching model field
)

type workoutGroup struct {
	date    string
	session string
	rows    [][]string
}

func main() {
	userID := flag.String("user-id", "", "Cognito UserID (sub) to assign the data to (required)")
	filePath := flag.String("file", "", "Path to the CSV file (required)")
	env := flag.String("env", "prod", "Environment suffix for DynamoDB table names (prod or test)")
	dryRun := flag.Bool("dry-run", false, "Parse and print what would be written without writing to DynamoDB")
	flag.Parse()

	if *userID == "" {
		log.Fatal("--user-id is required")
	}
	if *filePath == "" {
		log.Fatal("--file is required")
	}

	// --- AWS / DynamoDB setup ---
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewEnvCredentials(),
	}))
	dynamo := dynamodb.New(sess)

	workoutsTable := fmt.Sprintf("Workouts-%s", *env)
	exercisesTable := fmt.Sprintf("Exercises-%s", *env)

	workoutRepo := repoDb.NewDynamoWorkoutRepository(dynamo, workoutsTable)
	exerciseRepo := repoDb.NewDynamoExerciseRepository(dynamo, exercisesTable)

	// --- Parse CSV ---
	groups, err := parseCSV(*filePath)
	if err != nil {
		log.Fatalf("failed to parse CSV: %v", err)
	}

	fmt.Printf("Parsed %d workout sessions from CSV\n", len(groups))
	if *dryRun {
		fmt.Println("DRY RUN — no data will be written to DynamoDB")
	}

	// --- Import ---
	totalWorkouts := 0
	totalExercises := 0

	for _, group := range groups {
		exerciseIDs := make([]string, 0, len(group.rows))

		for _, row := range group.rows {
			exercise := buildExercise(row)

			if *dryRun {
				setsDesc := fmt.Sprintf("sets=%d", len(exercise.Sets))
				if len(exercise.Sets) > 0 {
					s := exercise.Sets[0]
					setsDesc += fmt.Sprintf("[reps=%d weight=%.1f%s]", s.Reps, s.Weight, s.Unit)
				}
				fmt.Printf("  [exercise] %-30s %-8s %s dist=%.0f%s time=%ds\n",
					exercise.Name, exercise.ExerciseType,
					setsDesc,
					exercise.Distance, exercise.DistanceUnit,
					exercise.Time)
			} else {
				if err := exerciseRepo.Create(*userID, exercise); err != nil {
					log.Printf("WARNING: failed to create exercise %q in workout %s/%s: %v",
						exercise.Name, group.date, group.session, err)
					continue
				}
			}

			exerciseIDs = append(exerciseIDs, exercise.ExerciseID)
			totalExercises++
		}

		workout := &models.Workout{
			UserID:    *userID,
			WorkoutID: uuid.New().String(),
			Name:      group.session,
			Date:      group.date,
			Exercises: exerciseIDs,
			CreatedAt: time.Now(),
		}

		if *dryRun {
			fmt.Printf("[workout] %s — %s (%d exercises)\n", group.date, group.session, len(exerciseIDs))
		} else {
			if err := workoutRepo.Create(workout); err != nil {
				log.Printf("WARNING: failed to create workout %s/%s: %v", group.date, group.session, err)
				continue
			}
		}

		totalWorkouts++
	}

	fmt.Printf("\nDone. %d workouts, %d exercises processed.\n", totalWorkouts, totalExercises)
}

// parseCSV reads the file and returns ordered workout groups preserving CSV row order.
func parseCSV(path string) ([]workoutGroup, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV has no data rows")
	}

	// skip header row (index 0)
	var (
		keyOrder []string
		groups   = map[string]*workoutGroup{}
	)

	for _, row := range records[1:] {
		if len(row) < 10 {
			continue
		}
		date := strings.TrimSpace(row[colDate])
		sess := strings.TrimSpace(row[colSession])
		key := date + "\x00" + sess

		if _, exists := groups[key]; !exists {
			keyOrder = append(keyOrder, key)
			groups[key] = &workoutGroup{date: date, session: sess}
		}
		groups[key].rows = append(groups[key].rows, row)
	}

	result := make([]workoutGroup, 0, len(keyOrder))
	for _, k := range keyOrder {
		result = append(result, *groups[k])
	}
	return result, nil
}

// buildExercise constructs a models.Exercise from one CSV row.
func buildExercise(row []string) *models.Exercise {
	name := field(row, colExercise)
	exerciseType := field(row, colType)

	sets := parseInt(field(row, colSets))
	reps := parseInt(field(row, colReps))
	weight := parseFloat(field(row, colWeight))
	weightUnit := field(row, colWeightUnit)
	distance := parseFloat(field(row, colDistance))
	distanceUnit := field(row, colDistanceUnit)
	roundTimes := field(row, colRoundTimes)

	exercise := &models.Exercise{
		ExerciseID:   uuid.New().String(),
		Name:         name,
		ExerciseType: exerciseType,
		Distance:     distance,
		DistanceUnit: distanceUnit,
		Time:         parseRoundTimes(roundTimes, exerciseType),
	}

	// Build Sets for any exercise where sets > 0 (weights, other bodyweight, cardio intervals).
	// If sets is blank/0 but weight or reps is present (e.g. a distance-based lunges entry
	// with weight but no explicit set count), preserve that data in a single item so it
	// is not silently dropped.
	if sets > 0 {
		items := make([]models.WeightItem, sets)
		for i := range items {
			items[i] = models.WeightItem{
				Weight: weight, // 0 for bodyweight exercises
				Unit:   weightUnit,
				Reps:   reps,
			}
		}
		exercise.Sets = items
	} else if weight > 0 || reps > 0 {
		exercise.Sets = []models.WeightItem{{
			Weight: weight,
			Unit:   weightUnit,
			Reps:   reps,
		}}
	}

	return exercise
}

// parseRoundTimes parses the round_times CSV field into seconds.
// Multiple round values separated by " / " are averaged.
// Formats: "57s", "6:05m", "6:15", "3:39s", "39" — colon means M:SS, else bare seconds.
func parseRoundTimes(s, exerciseType string) int {
	if s == "" {
		return 0
	}
	parts := strings.Split(s, " / ")
	total := 0
	count := 0
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var secs int
		if strings.Contains(p, ":") {
			// M:SS format — strip any trailing letter then split on ":"
			p = strings.TrimRight(p, "ms")
			halves := strings.SplitN(p, ":", 2)
			mins := parseInt(halves[0])
			seconds := parseInt(halves[1])
			secs = mins*60 + seconds
		} else {
			// bare seconds — strip trailing letter
			p = strings.TrimRight(p, "ms")
			secs = parseInt(p)
		}
		total += secs
		count++
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func field(row []string, idx int) string {
	if idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	// Handle float-formatted integers exported from spreadsheets (e.g. "3.0")
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f)
	}
	return 0
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
