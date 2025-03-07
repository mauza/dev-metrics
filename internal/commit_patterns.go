package internal

import (
	"math/rand"
	"strings"
	"time"
)

type DeveloperPersona struct {
	Name           string
	WorkStartHour  int
	WorkEndHour    int
	Timezone       string
	CommitFreq     string // "frequent", "moderate", "sparse"
	CodingStyle    map[string]float64
	CommonPatterns []string
}

type CommitPattern struct {
	Timestamp   time.Time
	NumFiles    int
	ChangeType  string
	CommitType  string
	Description string
}

type CommitPatternGenerator struct {
	personas    map[string]DeveloperPersona
	commitTypes map[string]struct {
		FileCountRange [2]int
		Changes        []string
	}
	projectPatterns *ProjectPatternGenerator
}

func NewCommitPatternGenerator() *CommitPatternGenerator {
	g := &CommitPatternGenerator{
		personas: map[string]DeveloperPersona{
			"early_bird": {
				Name:          "early_bird",
				WorkStartHour: 6,
				WorkEndHour:   14,
				Timezone:      "America/New_York",
				CommitFreq:    "frequent",
				CodingStyle: map[string]float64{
					"refactor": 0.3,
					"feature":  0.2,
					"docs":     0.2,
					"fix":      0.3,
				},
				CommonPatterns: []string{
					"Refactor {component} for better maintainability",
					"Optimize {component} performance",
					"Update documentation for {component}",
				},
			},
			"night_owl": {
				Name:          "night_owl",
				WorkStartHour: 14,
				WorkEndHour:   22,
				Timezone:      "America/Los_Angeles",
				CommitFreq:    "moderate",
				CodingStyle: map[string]float64{
					"feature":  0.4,
					"fix":      0.3,
					"test":     0.2,
					"refactor": 0.1,
				},
				CommonPatterns: []string{
					"Add {feature} to {component}",
					"Fix edge case in {component}",
					"Implement {feature}",
				},
			},
			// ... add balanced persona similarly
		},
		commitTypes: map[string]struct {
			FileCountRange [2]int
			Changes        []string
		}{
			"feature": {
				FileCountRange: [2]int{2, 5},
				Changes:        []string{"add_feature", "enhance_feature", "implement_feature"},
			},
			"fix": {
				FileCountRange: [2]int{1, 3},
				Changes:        []string{"fix_bug", "handle_edge_case", "improve_error_handling"},
			},
			// ... add other commit types similarly
		},
	}
	g.projectPatterns = NewProjectPatternGenerator()
	return g
}

func (g *CommitPatternGenerator) GeneratePatterns(startDate, endDate time.Time, personaName string) []CommitPattern {
	if personaName == "" {
		// Select random persona
		personas := make([]string, 0, len(g.personas))
		for k := range g.personas {
			personas = append(personas, k)
		}
		personaName = personas[rand.Intn(len(personas))]
	}

	persona := g.personas[personaName]

	// Generate sprint cycles
	sprintCycles := g.projectPatterns.GenerateSprintCycles(startDate, endDate)

	var patterns []CommitPattern
	for _, cycle := range sprintCycles {
		// Adjust commit frequency based on sprint intensity
		adjustedFreq := g.adjustFrequency(persona.CommitFreq, cycle.Intensity)

		// Generate commits for each day in the cycle
		for current := cycle.StartDate; current.Before(cycle.EndDate); current = current.AddDate(0, 0, 1) {
			if current.Weekday() < 6 { // Skip weekends
				dayCommits := g.generateDayCommits(current, &persona, cycle, adjustedFreq)
				patterns = append(patterns, dayCommits...)
			}
		}
	}

	return patterns
}

func (g *CommitPatternGenerator) adjustFrequency(baseFreq string, intensity float64) string {
	frequencies := []string{"sparse", "moderate", "frequent"}
	var baseIndex int

	// Find current frequency index
	for i, freq := range frequencies {
		if freq == baseFreq {
			baseIndex = i
			break
		}
	}

	if intensity > 0.8 {
		if baseIndex < len(frequencies)-1 {
			return frequencies[baseIndex+1]
		}
	} else if intensity < 0.4 {
		if baseIndex > 0 {
			return frequencies[baseIndex-1]
		}
	}
	return baseFreq
}

func (g *CommitPatternGenerator) generateDayCommits(
	date time.Time,
	persona *DeveloperPersona,
	sprint SprintCycle,
	adjustedFreq string,
) []CommitPattern {
	var patterns []CommitPattern

	// Determine number of commits for the day
	numCommits := g.getCommitCount(adjustedFreq)

	// Generate commit times
	commitTimes := g.generateCommitTimes(date, numCommits, persona.WorkStartHour, persona.WorkEndHour, persona.Timezone)

	// Adjust commit types based on sprint phase
	codingStyle := g.adjustCodingStyle(persona.CodingStyle, sprint.Phase)

	// Generate commits with adjusted style
	for _, commitTime := range commitTimes {
		commitType := g.selectCommitType(codingStyle)
		commitInfo := g.commitTypes[commitType]

		// Add safety check for file count range
		fileCountDiff := commitInfo.FileCountRange[1] - commitInfo.FileCountRange[0] + 1
		if fileCountDiff <= 0 {
			fileCountDiff = 1
		}
		numFiles := rand.Intn(fileCountDiff) + commitInfo.FileCountRange[0]

		// Add safety check for empty Changes slice
		changeType := "unknown"
		if len(commitInfo.Changes) > 0 {
			changeType = commitInfo.Changes[rand.Intn(len(commitInfo.Changes))]
		}

		patterns = append(patterns, CommitPattern{
			Timestamp:   commitTime,
			NumFiles:    numFiles,
			ChangeType:  changeType,
			CommitType:  commitType,
			Description: g.generateCommitDescription(persona, commitType, sprint.FocusAreas),
		})
	}

	return patterns
}

func (g *CommitPatternGenerator) getCommitCount(frequency string) int {
	ranges := map[string][2]int{
		"frequent": {8, 15},
		"moderate": {4, 8},
		"sparse":   {1, 4},
	}
	r := ranges[frequency]
	return rand.Intn(r[1]-r[0]+1) + r[0]
}

func (g *CommitPatternGenerator) generateCommitTimes(
	date time.Time,
	numCommits int,
	startHour int,
	endHour int,
	timezone string,
) []time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	workMinutes := (endHour - startHour) * 60
	lunchStart := startHour + 4 // Lunch after 4 hours of work

	times := make([]time.Time, 0, numCommits)
	usedMinutes := make(map[int]bool)

	for len(times) < numCommits {
		// Generate a random minute in the workday
		minute := rand.Intn(workMinutes)
		hour := startHour + (minute / 60)
		min := minute % 60

		// Skip lunch hour
		if hour == lunchStart {
			continue
		}

		if !usedMinutes[minute] {
			commitTime := time.Date(
				date.Year(), date.Month(), date.Day(),
				hour, min, rand.Intn(60), 0, loc,
			)
			times = append(times, commitTime)
			usedMinutes[minute] = true
		}
	}

	return times
}

func (g *CommitPatternGenerator) adjustCodingStyle(
	style map[string]float64,
	phase ProjectPhase,
) map[string]float64 {
	adjusted := make(map[string]float64)
	for k, v := range style {
		adjusted[k] = v
	}

	switch phase {
	case PhaseFeatureDev:
		adjusted["feature"] = max(0.4, adjusted["feature"])
	case PhaseStabilization:
		adjusted["fix"] = max(0.4, adjusted["fix"])
	case PhaseRelease:
		adjusted["docs"] = max(0.3, adjusted["docs"])
	}

	return adjusted
}

func (g *CommitPatternGenerator) selectCommitType(style map[string]float64) string {
	total := 0.0
	for _, weight := range style {
		total += weight
	}

	r := rand.Float64() * total
	current := 0.0

	for commitType, weight := range style {
		current += weight
		if r <= current {
			return commitType
		}
	}

	// Fallback to first type
	for commitType := range style {
		return commitType
	}
	return "feature"
}

func (g *CommitPatternGenerator) generateCommitDescription(
	persona *DeveloperPersona,
	commitType string,
	focusAreas []string,
) string {
	if len(focusAreas) == 0 {
		return "Update codebase"
	}

	area := focusAreas[rand.Intn(len(focusAreas))]
	parts := strings.Split(area, "/")
	component := parts[0]
	feature := "feature"
	if len(parts) > 1 {
		feature = parts[1]
	}

	// Add safety check for empty CommonPatterns
	if len(persona.CommonPatterns) == 0 {
		return "Update " + component
	}

	pattern := persona.CommonPatterns[rand.Intn(len(persona.CommonPatterns))]

	// Replace placeholders
	pattern = strings.ReplaceAll(pattern, "{component}", component)
	pattern = strings.ReplaceAll(pattern, "{feature}", feature)
	pattern = strings.ReplaceAll(pattern, "{issue}",
		[]string{"memory leak", "performance", "edge case"}[rand.Intn(3)])

	return pattern
}

// Helper function for Go <1.21
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
