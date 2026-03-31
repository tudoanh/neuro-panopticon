package scanner

import (
	"fmt"
	"testing"
	"time"

	"neuropanopticon/backend/models"
)

func TestComputeScore_NoFindings(t *testing.T) {
	score := ComputeScore(nil)
	if score.Score != 100 {
		t.Errorf("expected score 100 with no findings, got %d", score.Score)
	}
	if score.Findings != 0 {
		t.Errorf("expected 0 findings, got %d", score.Findings)
	}
}

func TestComputeScore_SingleCritical(t *testing.T) {
	findings := []models.Finding{
		{ID: "test-1", Severity: models.SeverityCritical},
	}
	score := ComputeScore(findings)
	if score.Score != 75 {
		t.Errorf("expected score 75 (100 - 25 critical), got %d", score.Score)
	}
	if score.Critical != 1 {
		t.Errorf("expected 1 critical, got %d", score.Critical)
	}
}

func TestComputeScore_MixedSeverities(t *testing.T) {
	findings := []models.Finding{
		{ID: "c1", Severity: models.SeverityCritical},
		{ID: "h1", Severity: models.SeverityHigh},
		{ID: "m1", Severity: models.SeverityMedium},
		{ID: "l1", Severity: models.SeverityLow},
		{ID: "i1", Severity: models.SeverityInfo},
	}
	// 100 - 25 - 15 - 8 - 3 - 0 = 49
	score := ComputeScore(findings)
	if score.Score != 49 {
		t.Errorf("expected score 49, got %d", score.Score)
	}
	if score.Findings != 5 {
		t.Errorf("expected 5 findings, got %d", score.Findings)
	}
	if score.Critical != 1 || score.High != 1 || score.Medium != 1 || score.Low != 1 {
		t.Errorf("severity counts wrong: c=%d h=%d m=%d l=%d",
			score.Critical, score.High, score.Medium, score.Low)
	}
}

func TestComputeScore_FloorAtZero(t *testing.T) {
	var findings []models.Finding
	for i := 0; i < 10; i++ {
		findings = append(findings, models.Finding{
			ID:       fmt.Sprintf("c%d", i),
			Severity: models.SeverityCritical,
		})
	}
	// 10 * 25 = 250 penalty, but floor is 0
	score := ComputeScore(findings)
	if score.Score != 0 {
		t.Errorf("expected score 0 (floor), got %d", score.Score)
	}
}

func TestComputeScore_DeduplicatesByID(t *testing.T) {
	findings := []models.Finding{
		{ID: "dup-1", Severity: models.SeverityHigh},
		{ID: "dup-1", Severity: models.SeverityHigh}, // duplicate
		{ID: "dup-2", Severity: models.SeverityMedium},
	}
	// Should only count dup-1 once: 100 - 15 - 8 = 77
	score := ComputeScore(findings)
	if score.Score != 77 {
		t.Errorf("expected score 77 (deduped), got %d", score.Score)
	}
	if score.Findings != 2 {
		t.Errorf("expected 2 findings (deduped), got %d", score.Findings)
	}
}

func TestComputeScore_MaxScoreIs100(t *testing.T) {
	score := ComputeScore(nil)
	if score.MaxScore != 100 {
		t.Errorf("expected max_score 100, got %d", score.MaxScore)
	}
}

func TestComputeScore_SetsLastUpdated(t *testing.T) {
	before := time.Now()
	score := ComputeScore(nil)
	after := time.Now()
	if score.LastUpdated.Before(before) || score.LastUpdated.After(after) {
		t.Errorf("last_updated should be between test start and end")
	}
}
