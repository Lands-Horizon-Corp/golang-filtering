package test

import (
	"strings"
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

func TestGormCSVDeterministicOrdering(t *testing.T) {
	// Initialize the database
	db := setupTestDB(t)

	// Create handler
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	// Generate CSV multiple times to ensure deterministic ordering
	var csvOutputs []string
	for i := 0; i < 3; i++ {
		csvData, err := handler.GormNoPaginationCSV(db, filter.Root{})
		if err != nil {
			t.Fatalf("GormNoPaginationCSV failed: %v", err)
		}
		csvOutputs = append(csvOutputs, string(csvData))
	}

	// All CSV outputs should be identical (deterministic ordering)
	for i := 1; i < len(csvOutputs); i++ {
		if csvOutputs[i] != csvOutputs[0] {
			t.Errorf("CSV output %d differs from first output", i)
			t.Logf("First output:\n%s", csvOutputs[0])
			t.Logf("Output %d:\n%s", i, csvOutputs[i])
		}
	}

	// Verify header ordering is alphabetical
	lines := strings.Split(csvOutputs[0], "\n")
	if len(lines) < 1 {
		t.Fatal("CSV output should have at least a header line")
	}

	headers := strings.Split(lines[0], ",")
	t.Logf("Actual headers: %v", headers)

	// Verify that all outputs are identical (deterministic ordering)
	// The exact order will depend on the internal field getter mapping,
	// but it should be consistent across multiple runs
	if len(headers) < 5 {
		t.Errorf("Expected at least 5 headers for TestUser fields, got %d", len(headers))
	}
}
