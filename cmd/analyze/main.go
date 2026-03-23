package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// CSV column indices (must match import script)
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
)

type row struct {
	date       string
	session    string
	exercise   string
	exType     string
	sets       string
	reps       string
	weight     string
	weightUnit string
	distance   string
	distUnit   string
	roundTimes string
	lineNum    int
}

func (r row) key() string { return r.exercise + "\x00" + r.exType }

func main() {
	filePath := flag.String("file", "", "Path to the CSV file (required)")
	flag.Parse()
	if *filePath == "" {
		log.Fatal("--file is required")
	}

	rows, err := readCSV(*filePath)
	if err != nil {
		log.Fatalf("failed to read CSV: %v", err)
	}

	// Group by exercise name + type
	grouped := map[string][]row{}
	var keys []string
	for _, r := range rows {
		k := r.key()
		if _, ok := grouped[k]; !ok {
			keys = append(keys, k)
		}
		grouped[k] = append(grouped[k], r)
	}
	sort.Strings(keys)

	issues := 0

	for _, k := range keys {
		group := grouped[k]
		name := group[0].exercise
		typ := group[0].exType

		// Determine which fields are EVER populated in this group
		everSets := anyNonEmpty(group, func(r row) string { return r.sets })
		everReps := anyNonEmpty(group, func(r row) string { return r.reps })
		everWeight := anyNonEmpty(group, func(r row) string { return r.weight })
		everDist := anyNonEmpty(group, func(r row) string { return r.distance })
		everTime := anyNonEmpty(group, func(r row) string { return r.roundTimes })

		// Check each row for missing fields that other rows in the group have
		var groupIssues []string
		for _, r := range group {
			var missing []string
			if everSets && r.sets == "" {
				missing = append(missing, "sets")
			}
			if everReps && r.reps == "" {
				missing = append(missing, "reps")
			}
			if everWeight && r.weight == "" {
				missing = append(missing, fmt.Sprintf("weight (others use %s)", canonicalWeight(group)))
			}
			if everDist && r.distance == "" {
				missing = append(missing, fmt.Sprintf("distance (others use %s%s)", canonicalDist(group), canonicalDistUnit(group)))
			}
			if everTime && r.roundTimes == "" {
				missing = append(missing, "round_times")
			}

			// Flag plank/timed-other exercises where time appears to be in the reps column
			if r.roundTimes == "" && r.reps != "" && r.weight == "" && r.distance == "" {
				if couldBeSeconds(r.reps) {
					missing = append(missing, fmt.Sprintf(
						"WARNING: reps=%s on a no-weight no-distance exercise — is this actually a duration? If so, move it to the round_times column (e.g. \"%ss\")",
						r.reps, r.reps,
					))
				}
			}

			if len(missing) > 0 {
				groupIssues = append(groupIssues,
					fmt.Sprintf("    line %-4d  %s  %-14s  missing: %s",
						r.lineNum, r.date, r.session, strings.Join(missing, ", ")))
				issues++
			}
		}

		if len(groupIssues) > 0 {
			fmt.Printf("\n%q (%s) — %d occurrence(s)\n", name, typ, len(group))
			for _, s := range groupIssues {
				fmt.Println(s)
			}
		}
	}

	fmt.Printf("\n---\nTotal issues found: %d\n", issues)
	if issues == 0 {
		fmt.Println("No inconsistencies detected.")
	}
}

func anyNonEmpty(group []row, f func(row) string) bool {
	for _, r := range group {
		if f(r) != "" {
			return true
		}
	}
	return false
}

// canonicalWeight returns the most common non-empty weight value in the group (with unit).
func canonicalWeight(group []row) string {
	counts := map[string]int{}
	for _, r := range group {
		if r.weight != "" {
			k := r.weight + r.weightUnit
			counts[k]++
		}
	}
	best, bestN := "", 0
	for k, n := range counts {
		if n > bestN {
			best, bestN = k, n
		}
	}
	return best
}

func canonicalDist(group []row) string {
	counts := map[string]int{}
	for _, r := range group {
		if r.distance != "" {
			counts[r.distance]++
		}
	}
	best, bestN := "", 0
	for k, n := range counts {
		if n > bestN {
			best, bestN = k, n
		}
	}
	return best
}

func canonicalDistUnit(group []row) string {
	for _, r := range group {
		if r.distUnit != "" {
			return r.distUnit
		}
	}
	return ""
}

// couldBeSeconds returns true if s looks like a plausible exercise duration in seconds
// (a small integer, e.g. 10–300).
func couldBeSeconds(s string) bool {
	v, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	return v >= 5 && v <= 600
}

func readCSV(path string) ([]row, error) {
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

	var rows []row
	for i, rec := range records[1:] {
		if len(rec) < 10 {
			continue
		}
		rows = append(rows, row{
			date:       field(rec, colDate),
			session:    field(rec, colSession),
			exercise:   field(rec, colExercise),
			exType:     field(rec, colType),
			sets:       field(rec, colSets),
			reps:       field(rec, colReps),
			weight:     field(rec, colWeight),
			weightUnit: field(rec, colWeightUnit),
			distance:   field(rec, colDistance),
			distUnit:   field(rec, colDistanceUnit),
			roundTimes: field(rec, colRoundTimes),
			lineNum:    i + 2, // +1 for header, +1 for 1-based
		})
	}
	return rows, nil
}

func field(rec []string, idx int) string {
	if idx < len(rec) {
		return strings.TrimSpace(rec[idx])
	}
	return ""
}
