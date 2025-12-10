package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/Lands-Horizon-Corp/golang-filtering/registry"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AccountType string

const (
	AccountTypeFines     AccountType = "fines"
	AccountTypeInterest  AccountType = "interest"
	AccountTypeSVFLedger AccountType = "svf_ledger"
)

type AdjustmentEntry struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	BranchID       uuid.UUID
	CurrencyID     uuid.UUID
	Type           AccountType
	Amount         float64
	UpdatedAt      time.Time
}

func TestRegistryArrFindWithFilters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&AdjustmentEntry{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	orgID := uuid.New()
	branchID := uuid.New()
	currencyID := uuid.New()

	entries := []AdjustmentEntry{
		{ID: uuid.New(), OrganizationID: orgID, BranchID: branchID, CurrencyID: currencyID, Type: AccountTypeFines, Amount: 100, UpdatedAt: time.Now().Add(-time.Hour)},
		{ID: uuid.New(), OrganizationID: orgID, BranchID: branchID, CurrencyID: currencyID, Type: AccountTypeInterest, Amount: 50, UpdatedAt: time.Now()},
		{ID: uuid.New(), OrganizationID: orgID, BranchID: branchID, CurrencyID: currencyID, Type: AccountTypeSVFLedger, Amount: 75, UpdatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: uuid.New(), OrganizationID: orgID, BranchID: branchID, CurrencyID: currencyID, Type: "other", Amount: 200, UpdatedAt: time.Now()},
	}
	if err := db.Create(&entries).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[AdjustmentEntry, AdjustmentEntry, any]{
		Database: db,
		Resource: func(d *AdjustmentEntry) *AdjustmentEntry { return d },
	})

	ctx := context.Background()

	// Map FilterSQL to ArrFilterSQL
	arrFilters := []query.ArrFilterSQL{
		{Field: "organization_id", Op: query.ModeEqual, Value: orgID},
		{Field: "branch_id", Op: query.ModeEqual, Value: branchID},
		{Field: "currency_id", Op: query.ModeEqual, Value: currencyID},
		{Field: "type", Op: query.ModeInside, Value: []AccountType{
			AccountTypeFines,
			AccountTypeInterest,
			AccountTypeSVFLedger,
		}},
	}

	arrSorts := []query.ArrFilterSortSQL{
		{Field: "updated_at", Order: query.SortOrderDesc},
	}

	res, err := r.ArrFind(ctx, arrFilters, arrSorts)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 3)

	// Ensure the results are sorted by UpdatedAt descending
	assert.True(t, res[0].UpdatedAt.After(res[1].UpdatedAt) || res[0].UpdatedAt.Equal(res[1].UpdatedAt))
	assert.True(t, res[1].UpdatedAt.After(res[2].UpdatedAt) || res[1].UpdatedAt.Equal(res[2].UpdatedAt))
}
