package scanner

import (
	"time"

	"neuropanopticon/backend/models"
)

// Penalty points per severity level. The score starts at 100 and is reduced
// by penalties for each finding. The floor is 0.
const (
	penaltyCritical = 25
	penaltyHigh     = 15
	penaltyMedium   = 8
	penaltyLow      = 3
	penaltyInfo     = 0
)

// ComputeScore calculates a SecurityScore from a set of findings.
// The algorithm starts at 100 (perfect) and deducts points per finding
// based on severity. Duplicate finding IDs are counted only once.
func ComputeScore(findings []models.Finding) models.SecurityScore {
	seen := make(map[string]bool)
	score := models.SecurityScore{
		Score:       100,
		MaxScore:    100,
		LastUpdated: time.Now(),
	}

	totalPenalty := 0

	for _, f := range findings {
		// Deduplicate by finding ID
		if seen[f.ID] {
			continue
		}
		seen[f.ID] = true

		score.Findings++

		switch f.Severity {
		case models.SeverityCritical:
			score.Critical++
			totalPenalty += penaltyCritical
		case models.SeverityHigh:
			score.High++
			totalPenalty += penaltyHigh
		case models.SeverityMedium:
			score.Medium++
			totalPenalty += penaltyMedium
		case models.SeverityLow:
			score.Low++
			totalPenalty += penaltyLow
		default:
			// info — no penalty
		}
	}

	score.Score = 100 - totalPenalty
	if score.Score < 0 {
		score.Score = 0
	}

	return score
}
