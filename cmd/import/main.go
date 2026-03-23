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
	// colEffort = 11  // mapped to Effort/Level — not used
	// colNotes  = 12  // no matching model field
)

// bodyWeightExercises is the set of exercise names (lower-cased) that map to the
// body_weight type regardless of what the CSV type column says.
var bodyWeightExercises = map[string]bool{
	"push ups":   true,
	"push-ups":   true,
	"sit ups":    true,
	"sit-ups":    true,
	"plank":      true,
	"in and out": true,
	"burpees":    true,
	"lunges":     true,
	"back lunges": false, // weighted — keep as weights when weight>0
}

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

	// --- Build nearest-time index for exercises missing round_times ---
	allRows := flattenRows(groups)
	nearestTimes := buildNearestTimeIndex(allRows)

	// --- Import ---
	totalWorkouts := 0
	totalExercises := 0
	rowIdx := 0

	for _, group := range groups {
		exerciseIDs := make([]string, 0)

		for _, row := range group.rows {
			exercises := buildExercises(row, nearestTimes[rowIdx])
			rowIdx++

			for _, exercise := range exercises {
				if *dryRun {
					setsDesc := fmt.Sprintf("sets=%d", len(exercise.Sets))
					if len(exercise.Sets) > 0 {
						s := exercise.Sets[0]
						setsDesc += fmt.Sprintf("[reps=%d dur=%ds weight=%.1f%s]", s.Reps, s.Duration, s.Weight, s.Unit)
					}
					fmt.Printf("  [exercise] %-30s %-12s %s dist=%.0f%s time=%ds\n",
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

// buildExercises converts one CSV row into one or more Exercise records.
// Multi-set cardio rows are split into one Exercise per set.
// Warm Up (Cardio) rows are dropped (returns nil).
// nearestTime is used to fill in Time for exercises that have no round_times (e.g. Ski Erg cals).
func buildExercises(row []string, nearestTime int) []*models.Exercise {
	name := field(row, colExercise)
	csvType := field(row, colType)

	// Drop warm-up rows
	if strings.EqualFold(name, "warm up (cardio)") || strings.EqualFold(name, "warm up") {
		return nil
	}

	sets := parseInt(field(row, colSets))
	reps := parseInt(field(row, colReps))
	weight := parseFloat(field(row, colWeight))
	weightUnit := field(row, colWeightUnit)
	distance := parseFloat(field(row, colDistance))
	distanceUnit := field(row, colDistanceUnit)
	roundTimes := field(row, colRoundTimes)

	exerciseType := resolveType(name, csvType, weight)

	switch exerciseType {
	case models.ExerciseTypeCardio:
		return buildCardioExercises(name, sets, reps, weight, weightUnit, distance, distanceUnit, roundTimes, nearestTime)
	case models.ExerciseTypeBodyWeight:
		return buildBodyWeightExercises(name, sets, reps, distance, distanceUnit, roundTimes)
	default: // weights, other
		return buildWeightExercises(name, exerciseType, sets, reps, weight, weightUnit, distance, distanceUnit)
	}
}

// resolveType determines the effective ExerciseType for a row.
// CSV type is used as-is for weights/other, but:
//   - "other" exercises that are bodyweight exercises → "body_weight"
//   - "other" with distance and no weight → "body_weight"
func resolveType(name, csvType string, weight float64) string {
	lower := strings.ToLower(name)
	switch csvType {
	case models.ExerciseTypeWeights:
		return models.ExerciseTypeWeights
	case models.ExerciseTypeCardio:
		return models.ExerciseTypeCardio
	case models.ExerciseTypeBodyWeight:
		return models.ExerciseTypeBodyWeight
	case models.ExerciseTypeOther:
		if bodyWeightExercises[lower] || (weight == 0) {
			return models.ExerciseTypeBodyWeight
		}
		return models.ExerciseTypeOther
	default:
		return csvType
	}
}

// buildCardioExercises handles cardio rows.
// If round_times is present: split into one Exercise per time value.
// Ski Erg with reps (calories): reps → distance, unit = "cal"; nearestTime fills in Time when no round_times.
// Weighted Walk: combine sets into one row, multiply distance by sets.
func buildCardioExercises(name string, sets, reps int, weight float64, weightUnit string, distance float64, distanceUnit, roundTimes string, nearestTime int) []*models.Exercise {
	lower := strings.ToLower(name)

	// Weighted Walk: combine sets into a single row with total distance
	if lower == "weighted walk" {
		effectiveSets := sets
		if effectiveSets == 0 {
			effectiveSets = 1
		}
		e := newExercise(name, models.ExerciseTypeCardio)
		e.Distance = distance * float64(effectiveSets)
		e.DistanceUnit = distanceUnit
		if weight > 0 {
			e.Sets = []models.WeightItem{{Weight: weight, Unit: weightUnit}}
		}
		return []*models.Exercise{e}
	}

	// Ski Erg with reps = calories (no distance column populated)
	if lower == "ski erg" && reps > 0 && distance == 0 {
		return splitPerSet(name, models.ExerciseTypeCardio, sets, func(i int, e *models.Exercise) {
			e.Distance = float64(reps)
			e.DistanceUnit = "cal"
			e.Time = nearestTime // filled from nearest Ski Erg row that has round_times
		})
	}

	// Cardio with round_times: split one Exercise per time
	if roundTimes != "" {
		times := parseRoundTimeList(roundTimes)
		result := make([]*models.Exercise, 0, len(times))
		for _, t := range times {
			e := newExercise(name, models.ExerciseTypeCardio)
			e.Distance = distance
			e.DistanceUnit = distanceUnit
			e.Time = t
			if weight > 0 {
				e.Sets = []models.WeightItem{{Weight: weight, Unit: weightUnit}}
			}
			result = append(result, e)
		}
		return result
	}

	// Cardio with sets but no round_times: split into per-set rows
	if sets > 1 {
		return splitPerSet(name, models.ExerciseTypeCardio, sets, func(i int, e *models.Exercise) {
			e.Distance = distance
			e.DistanceUnit = distanceUnit
			if weight > 0 {
				e.Sets = []models.WeightItem{{Weight: weight, Unit: weightUnit}}
			}
		})
	}

	// Single-set cardio
	e := newExercise(name, models.ExerciseTypeCardio)
	e.Distance = distance
	e.DistanceUnit = distanceUnit
	if reps > 0 && distance == 0 {
		// reps = calories (e.g. Ski Erg)
		e.Distance = float64(reps)
		e.DistanceUnit = "cal"
	}
	return []*models.Exercise{e}
}

// buildBodyWeightExercises handles body_weight rows.
// Plank: if round_times present, split per time value with Duration.
//        if reps present, treat reps as duration (seconds) per set.
// Other: split into one Exercise per set, each with Sets=[WeightItem{Reps}].
// Distance-based (lunges): split per set, each with Distance.
func buildBodyWeightExercises(name string, sets, reps int, distance float64, distanceUnit, roundTimes string) []*models.Exercise {
	lower := strings.ToLower(name)
	isPlank := lower == "plank"

	if isPlank {
		if roundTimes != "" {
			// One exercise; each set gets its own Duration from round_times
			times := parseRoundTimeList(roundTimes)
			items := make([]models.WeightItem, len(times))
			for i, t := range times {
				items[i] = models.WeightItem{Duration: t}
			}
			e := newExercise(name, models.ExerciseTypeBodyWeight)
			e.Sets = items
			return []*models.Exercise{e}
		}
		// reps = duration in seconds; N identical sets
		if reps > 0 {
			effectiveSets := sets
			if effectiveSets == 0 {
				effectiveSets = 1
			}
			items := make([]models.WeightItem, effectiveSets)
			for i := range items {
				items[i] = models.WeightItem{Duration: reps}
			}
			e := newExercise(name, models.ExerciseTypeBodyWeight)
			e.Sets = items
			return []*models.Exercise{e}
		}
	}

	// Distance-based bodyweight exercise (e.g. Lunges with distance)
	if distance > 0 {
		effectiveSets := sets
		if effectiveSets == 0 {
			effectiveSets = 1
		}
		return splitPerSet(name, models.ExerciseTypeBodyWeight, effectiveSets, func(i int, e *models.Exercise) {
			e.Distance = distance
			e.DistanceUnit = distanceUnit
			e.Sets = []models.WeightItem{{Reps: reps}}
		})
	}

	// Standard reps-based bodyweight exercise
	effectiveSets := sets
	if effectiveSets == 0 {
		effectiveSets = 1
	}
	return []*models.Exercise{buildRepsExercise(name, models.ExerciseTypeBodyWeight, effectiveSets, reps, 0, "")}
}

// buildWeightExercises handles weights and other rows.
// These are kept as a single Exercise with Sets []WeightItem.
func buildWeightExercises(name, exerciseType string, sets, reps int, weight float64, weightUnit string, distance float64, distanceUnit string) []*models.Exercise {
	e := newExercise(name, exerciseType)
	e.Distance = distance
	e.DistanceUnit = distanceUnit

	if sets > 0 {
		items := make([]models.WeightItem, sets)
		for i := range items {
			items[i] = models.WeightItem{
				Weight: weight,
				Unit:   weightUnit,
				Reps:   reps,
			}
		}
		e.Sets = items
	} else if weight > 0 {
		// Weighted exercise with no explicit set count — preserve weight
		e.Sets = []models.WeightItem{{Weight: weight, Unit: weightUnit, Reps: reps}}
	} else if reps > 0 {
		// No sets, no weight — single-set rep-only entry
		e.Sets = []models.WeightItem{{Reps: reps}}
	}

	return []*models.Exercise{e}
}

// buildRepsExercise creates a single Exercise with n identical WeightItem sets.
func buildRepsExercise(name, exerciseType string, sets, reps int, weight float64, weightUnit string) *models.Exercise {
	e := newExercise(name, exerciseType)
	items := make([]models.WeightItem, sets)
	for i := range items {
		items[i] = models.WeightItem{Weight: weight, Unit: weightUnit, Reps: reps}
	}
	e.Sets = items
	return e
}

// splitPerSet creates `n` Exercise records, calling configure(i, e) on each.
func splitPerSet(name, exerciseType string, n int, configure func(i int, e *models.Exercise)) []*models.Exercise {
	result := make([]*models.Exercise, n)
	for i := 0; i < n; i++ {
		e := newExercise(name, exerciseType)
		configure(i, e)
		result[i] = e
	}
	return result
}

func newExercise(name, exerciseType string) *models.Exercise {
	return &models.Exercise{
		ExerciseID:   uuid.New().String(),
		Name:         name,
		ExerciseType: exerciseType,
	}
}

// flattenRows returns all CSV rows across all groups as a single ordered slice.
func flattenRows(groups []workoutGroup) [][]string {
	var all [][]string
	for _, g := range groups {
		all = append(all, g.rows...)
	}
	return all
}

// buildNearestTimeIndex scans all rows and returns a map of rowIndex → nearest averaged time (seconds)
// for rows that have no round_times value. Only rows for the same exercise name with a non-empty
// round_times are considered as candidates. Rows that already have round_times are excluded.
func buildNearestTimeIndex(rows [][]string) map[int]int {
	type timeAt struct{ idx, secs int }
	byName := map[string][]timeAt{}

	for i, row := range rows {
		rt := field(row, colRoundTimes)
		if rt == "" {
			continue
		}
		times := parseRoundTimeList(rt)
		if len(times) == 0 {
			continue
		}
		total := 0
		for _, t := range times {
			total += t
		}
		name := strings.ToLower(field(row, colExercise))
		byName[name] = append(byName[name], timeAt{i, total / len(times)})
	}

	result := map[int]int{}
	for i, row := range rows {
		if field(row, colRoundTimes) != "" {
			continue // row already has its own times
		}
		name := strings.ToLower(field(row, colExercise))
		entries := byName[name]
		if len(entries) == 0 {
			continue
		}
		nearest := entries[0]
		for _, e := range entries[1:] {
			if abs(i-e.idx) < abs(i-nearest.idx) {
				nearest = e
			}
		}
		result[i] = nearest.secs
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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

// parseRoundTimeList splits a round_times string on " / " and parses each value into seconds.
// E.g. "3:41m / 3:38m / 3:38m" → [221, 218, 218]
func parseRoundTimeList(s string) []int {
	parts := strings.Split(s, " / ")
	var result []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		result = append(result, parseSingleTime(p))
	}
	return result
}

// parseSingleTime parses a single time string into seconds.
// Formats: "57s", "6:05m", "6:15", "3:39s", "45s", "39"
func parseSingleTime(p string) int {
	if strings.Contains(p, ":") {
		// M:SS format — strip trailing letter then split on ":"
		p = strings.TrimRight(p, "ms")
		halves := strings.SplitN(p, ":", 2)
		return parseInt(halves[0])*60 + parseInt(halves[1])
	}
	// Bare seconds — strip trailing letter
	p = strings.TrimRight(p, "ms")
	return parseInt(p)
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
