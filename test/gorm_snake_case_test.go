package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MemberProfile represents a user profile (snake_case naming test)
type MemberProfile struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	FirstName string `gorm:"type:varchar(255)" json:"first_name"`
	LastName  string `gorm:"type:varchar(255)" json:"last_name"`
	Email     string `gorm:"type:varchar(255)" json:"email"`
}

// AccountTransaction represents a transaction with member_profile relation
type AccountTransaction struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `gorm:"not null" json:"created_at"`
	MemberProfileID uint           `gorm:"not null" json:"member_profile_id"`
	MemberProfile   *MemberProfile `gorm:"foreignKey:MemberProfileID" json:"member_profile,omitempty"`
	Amount          float64        `gorm:"type:decimal" json:"amount"`
	Description     string         `gorm:"type:varchar(255)" json:"description"`
}

func setupSnakeCaseDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&MemberProfile{}, &AccountTransaction{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create member profiles
	profiles := []MemberProfile{
		{ID: 1, FirstName: "John", LastName: "Doe", Email: "john@example.com"},
		{ID: 2, FirstName: "Jane", LastName: "Smith", Email: "jane@example.com"},
		{ID: 3, FirstName: "Bob", LastName: "Johnson", Email: "bob@example.com"},
	}
	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			t.Fatalf("Failed to create profile: %v", err)
		}
	}

	// Create transactions
	transactions := []AccountTransaction{
		{ID: 1, MemberProfileID: 1, Amount: 100.00, Description: "Payment", CreatedAt: time.Now()},
		{ID: 2, MemberProfileID: 1, Amount: 50.00, Description: "Refund", CreatedAt: time.Now()},
		{ID: 3, MemberProfileID: 2, Amount: 200.00, Description: "Payment", CreatedAt: time.Now()},
		{ID: 4, MemberProfileID: 3, Amount: 75.00, Description: "Payment", CreatedAt: time.Now()},
	}
	for _, tx := range transactions {
		if err := db.Create(&tx).Error; err != nil {
			t.Fatalf("Failed to create transaction: %v", err)
		}
	}

	return db
}

// TestSnakeCaseNestedFilter tests filtering with snake_case relation names
func TestSnakeCaseNestedFilter(t *testing.T) {
	db := setupSnakeCaseDB(t)
	handler := filter.NewFilter[AccountTransaction]()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "member_profile.first_name",
				Value:    "John",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"MemberProfile"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by member_profile.first_name: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 transactions for John, got %d", result.TotalSize)
	}

	for _, tx := range result.Data {
		if tx.MemberProfile == nil {
			t.Error("MemberProfile was not preloaded")
			continue
		}
		if tx.MemberProfile.FirstName != "John" {
			t.Errorf("Expected first_name 'John', got '%s'", tx.MemberProfile.FirstName)
		}
		t.Logf("Transaction %d: %s - %s %s", tx.ID, tx.Description, tx.MemberProfile.FirstName, tx.MemberProfile.LastName)
	}
}

// TestSnakeCaseNestedFilterEmail tests filtering by email field
func TestSnakeCaseNestedFilterEmail(t *testing.T) {
	db := setupSnakeCaseDB(t)
	handler := filter.NewFilter[AccountTransaction]()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "member_profile.email",
				Value:    "jane@example.com",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"MemberProfile"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by member_profile.email: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 transaction for jane@example.com, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		tx := result.Data[0]
		if tx.MemberProfile.Email != "jane@example.com" {
			t.Errorf("Expected email 'jane@example.com', got '%s'", tx.MemberProfile.Email)
		}
	}
}

// TestSnakeCaseNestedFilterContains tests contains mode with snake_case
func TestSnakeCaseNestedFilterContains(t *testing.T) {
	db := setupSnakeCaseDB(t)
	handler := filter.NewFilter[AccountTransaction]()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "member_profile.last_name",
				Value:    "son",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"MemberProfile"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by member_profile.last_name contains: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 transaction with 'son' in last_name, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		tx := result.Data[0]
		if tx.MemberProfile.LastName != "Johnson" {
			t.Errorf("Expected last_name 'Johnson', got '%s'", tx.MemberProfile.LastName)
		}
	}
}

// TestSnakeCaseNestedSorting tests sorting by snake_case relation field
func TestSnakeCaseNestedSorting(t *testing.T) {
	db := setupSnakeCaseDB(t)
	handler := filter.NewFilter[AccountTransaction]()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "member_profile.first_name", Order: filter.SortOrderAsc},
		},
		Preload: []string{"MemberProfile"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to sort by member_profile.first_name: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 total transactions, got %d", result.TotalSize)
	}

	// Verify sorted by first_name: Bob, Jane, John, John
	expectedOrder := []string{"Bob", "Jane", "John", "John"}
	for i, tx := range result.Data {
		if tx.MemberProfile == nil {
			t.Error("MemberProfile was not preloaded")
			continue
		}
		if tx.MemberProfile.FirstName != expectedOrder[i] {
			t.Errorf("Expected first_name '%s' at position %d, got '%s'", expectedOrder[i], i, tx.MemberProfile.FirstName)
		}
		t.Logf("Position %d: %s - %s", i, tx.Description, tx.MemberProfile.FirstName)
	}
}

// TestSnakeCaseMultipleFilters tests combining main table and snake_case relation filters
func TestSnakeCaseMultipleFilters(t *testing.T) {
	db := setupSnakeCaseDB(t)
	handler := filter.NewFilter[AccountTransaction]()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "member_profile.first_name",
				Value:    "John",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "amount",
				Value:    75.0,
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeNumber,
			},
		},
		Preload: []string{"MemberProfile"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter with multiple conditions: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 transaction (John with amount > 75), got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		tx := result.Data[0]
		if tx.Amount != 100.00 {
			t.Errorf("Expected amount 100.00, got %.2f", tx.Amount)
		}
		if tx.MemberProfile.FirstName != "John" {
			t.Errorf("Expected first_name 'John', got '%s'", tx.MemberProfile.FirstName)
		}
	}
}

// TestSnakeCaseOrLogic tests OR logic with snake_case relations
func TestSnakeCaseOrLogic(t *testing.T) {
	db := setupSnakeCaseDB(t)
	handler := filter.NewFilter[AccountTransaction]()

	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "member_profile.first_name",
				Value:    "John",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "member_profile.first_name",
				Value:    "Jane",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"MemberProfile"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter with OR logic: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 transactions (John or Jane), got %d", result.TotalSize)
	}

	for _, tx := range result.Data {
		if tx.MemberProfile == nil {
			t.Error("MemberProfile was not preloaded")
			continue
		}
		if tx.MemberProfile.FirstName != "John" && tx.MemberProfile.FirstName != "Jane" {
			t.Errorf("Expected first_name John or Jane, got '%s'", tx.MemberProfile.FirstName)
		}
	}
}
