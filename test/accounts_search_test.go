package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Account model with many columns
type Account struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	Name            string     `json:"name" gorm:"index"`
	Email           string     `json:"email" gorm:"uniqueIndex"`
	Phone           string     `json:"phone"`
	Address         string     `json:"address"`
	City            string     `json:"city"`
	State           string     `json:"state"`
	Country         string     `json:"country"`
	PostalCode      string     `json:"postal_code"`
	CompanyName     string     `json:"company_name"`
	JobTitle        string     `json:"job_title"`
	Department      string     `json:"department"`
	ManagerName     string     `json:"manager_name"`
	EmployeeID      string     `json:"employee_id"`
	Salary          float64    `json:"salary"`
	Commission      float64    `json:"commission"`
	Status          string     `json:"status"`
	AccountType     string     `json:"account_type"`
	Industry        string     `json:"industry"`
	Website         string     `json:"website"`
	TaxID           string     `json:"tax_id"`
	CreditLimit     float64    `json:"credit_limit"`
	Balance         float64    `json:"balance"`
	IsActive        bool       `json:"is_active"`
	IsVerified      bool       `json:"is_verified"`
	IsPremium       bool       `json:"is_premium"`
	LastLoginAt     time.Time  `json:"last_login_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	BirthDate       time.Time  `json:"birth_date"`
	HireDate        time.Time  `json:"hire_date"`
	TerminationDate *time.Time `json:"termination_date"`
	Notes           string     `json:"notes"`
	Preferences     string     `json:"preferences"`
	Metadata        string     `json:"metadata"`
}

func setupAccountDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Account{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Insert test accounts with various names
	now := time.Now()
	accounts := []Account{
		{
			ID: 1, Name: "John Smith", Email: "john.smith@example.com",
			Phone: "555-0101", Address: "123 Main St", City: "New York",
			State: "NY", Country: "USA", PostalCode: "10001",
			CompanyName: "Acme Corp", JobTitle: "Developer", Department: "IT",
			Status: "active", IsActive: true, CreatedAt: now,
		},
		{
			ID: 2, Name: "Jane Johnson", Email: "jane.johnson@example.com",
			Phone: "555-0102", Address: "456 Oak Ave", City: "Los Angeles",
			State: "CA", Country: "USA", PostalCode: "90001",
			CompanyName: "Tech Inc", JobTitle: "Manager", Department: "Sales",
			Status: "active", IsActive: true, CreatedAt: now,
		},
		{
			ID: 3, Name: "Bob Wilson", Email: "bob.wilson@example.com",
			Phone: "555-0103", Address: "789 Pine Rd", City: "Chicago",
			State: "IL", Country: "USA", PostalCode: "60601",
			CompanyName: "Global Ltd", JobTitle: "Analyst", Department: "Finance",
			Status: "active", IsActive: true, CreatedAt: now,
		},
		{
			ID: 4, Name: "Alice Brown", Email: "alice.brown@example.com",
			Phone: "555-0104", Address: "321 Elm St", City: "Houston",
			State: "TX", Country: "USA", PostalCode: "77001",
			CompanyName: "Startup Co", JobTitle: "Designer", Department: "Marketing",
			Status: "inactive", IsActive: false, CreatedAt: now,
		},
		{
			ID: 5, Name: "Charlie Davis", Email: "charlie.davis@example.com",
			Phone: "555-0105", Address: "654 Maple Dr", City: "Phoenix",
			State: "AZ", Country: "USA", PostalCode: "85001",
			CompanyName: "Enterprise Inc", JobTitle: "Engineer", Department: "IT",
			Status: "active", IsActive: true, CreatedAt: now,
		},
		{
			ID: 6, Name: "Diana Miller", Email: "diana.miller@example.com",
			Phone: "555-0106", Address: "987 Cedar Ln", City: "Philadelphia",
			State: "PA", Country: "USA", PostalCode: "19101",
			CompanyName: "Solutions LLC", JobTitle: "Consultant", Department: "Operations",
			Status: "active", IsActive: true, CreatedAt: now,
		},
		{
			ID: 7, Name: "John Anderson", Email: "john.anderson@example.com",
			Phone: "555-0107", Address: "147 Birch Way", City: "San Antonio",
			State: "TX", Country: "USA", PostalCode: "78201",
			CompanyName: "Systems Corp", JobTitle: "Administrator", Department: "IT",
			Status: "active", IsActive: true, CreatedAt: now,
		},
		{
			ID: 8, Name: "Emily Johnson", Email: "emily.johnson@example.com",
			Phone: "555-0108", Address: "258 Spruce Ct", City: "San Diego",
			State: "CA", Country: "USA", PostalCode: "92101",
			CompanyName: "Digital Agency", JobTitle: "Developer", Department: "IT",
			Status: "active", IsActive: true, CreatedAt: now,
		},
	}

	db.Create(&accounts)
	return db
}

// TestAccountSearchEqual tests exact name match
func TestAccountSearchEqual(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John Smith",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 account, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 && result.Data[0].Name != "John Smith" {
		t.Errorf("Expected 'John Smith', got '%s'", result.Data[0].Name)
	}
} // TestAccountSearchNotEqual tests not equal
func TestAccountSearchNotEqual(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John Smith",
				Mode:     filter.ModeNotEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Should return all except John Smith (7 accounts)
	if result.TotalSize != 7 {
		t.Errorf("Expected 7 accounts, got %d", result.TotalSize)
	}
}

// TestAccountSearchContains tests contains substring
func TestAccountSearchContains(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Should return: John Smith, Jane Johnson, John Anderson, Emily Johnson (4 accounts)
	if result.TotalSize != 4 {
		t.Errorf("Expected 4 accounts containing 'John', got %d", result.TotalSize)
	}
}

// TestAccountSearchNotContains tests not contains
func TestAccountSearchNotContains(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "Johnson",
				Mode:     filter.ModeNotContains,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Should exclude Jane Johnson and Emily Johnson (6 accounts)
	if result.TotalSize != 6 {
		t.Errorf("Expected 6 accounts not containing 'Johnson', got %d", result.TotalSize)
	}
}

// TestAccountSearchStartsWith tests starts with prefix
func TestAccountSearchStartsWith(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John",
				Mode:     filter.ModeStartsWith,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Should return: John Smith, John Anderson (2 accounts)
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 accounts starting with 'John', got %d", result.TotalSize)
	}

	for _, account := range result.Data {
		if account.Name[:4] != "John" {
			t.Errorf("Account name '%s' does not start with 'John'", account.Name)
		}
	}
}

// TestAccountSearchEndsWith tests ends with suffix
func TestAccountSearchEndsWith(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "Smith",
				Mode:     filter.ModeEndsWith,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Should return: John Smith (1 account)
	if result.TotalSize != 1 {
		t.Errorf("Expected 1 account ending with 'Smith', got %d", result.TotalSize)
	}

	if len(result.Data) > 0 && result.Data[0].Name != "John Smith" {
		t.Errorf("Expected 'John Smith', got '%s'", result.Data[0].Name)
	}
}

// TestAccountSearchIsEmpty tests empty name
func TestAccountSearchIsEmpty(t *testing.T) {
	db := setupAccountDB(t)

	// Insert account with empty name
	db.Create(&Account{
		ID: 99, Name: "", Email: "empty@example.com",
		Status: "active", IsActive: true,
	})

	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "",
				Mode:     filter.ModeIsEmpty,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 account with empty name, got %d", result.TotalSize)
	}
}

// TestAccountSearchIsNotEmpty tests non-empty name
func TestAccountSearchIsNotEmpty(t *testing.T) {
	db := setupAccountDB(t)

	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "",
				Mode:     filter.ModeIsNotEmpty,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// All 8 accounts have non-empty names
	if result.TotalSize != 8 {
		t.Errorf("Expected 8 accounts with non-empty names, got %d", result.TotalSize)
	}
}

// TestAccountSearchCaseInsensitive tests case insensitive search
func TestAccountSearchCaseInsensitive(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	testCases := []struct {
		name     string
		value    string
		mode     filter.Mode
		expected int
	}{
		{"Equal lowercase", "john smith", filter.ModeEqual, 1},
		{"Equal uppercase", "JOHN SMITH", filter.ModeEqual, 1},
		{"Equal mixedcase", "JoHn SmItH", filter.ModeEqual, 1},
		{"Contains lowercase", "john", filter.ModeContains, 4},
		{"Contains uppercase", "JOHN", filter.ModeContains, 4},
		{"StartsWith lowercase", "john", filter.ModeStartsWith, 2},
		{"StartsWith uppercase", "JOHN", filter.ModeStartsWith, 2},
		{"EndsWith lowercase", "smith", filter.ModeEndsWith, 1},
		{"EndsWith uppercase", "SMITH", filter.ModeEndsWith, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "name",
						Value:    tc.value,
						Mode:     tc.mode,
						DataType: filter.DataTypeText,
					},
				},
			}

			result, err := handler.DataGorm(db, filterRoot, 1, 10)
			if err != nil {
				t.Fatalf("Error searching: %v", err)
			}

			if result.TotalSize != tc.expected {
				t.Errorf("%s: expected %d accounts, got %d", tc.name, tc.expected, result.TotalSize)
			}
		})
	}
}

// TestAccountSearchWithPagination tests name search with pagination
func TestAccountSearchWithPagination(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "o", // Contains 'o'
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	// Get first page with 3 items
	result, err := handler.DataGorm(db, filterRoot, 0, 3)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	if len(result.Data) != 3 {
		t.Errorf("Expected 3 accounts on page 0, got %d", len(result.Data))
	}

	// Get second page
	result2, err := handler.DataGorm(db, filterRoot, 1, 3)
	if err != nil {
		t.Fatalf("Error searching page 2: %v", err)
	}

	// Verify different results on different pages
	if len(result2.Data) > 0 && result.Data[0].ID == result2.Data[0].ID {
		t.Error("Page 1 and Page 2 should have different accounts")
	}
}

// TestAccountSearchWithSorting tests name search with sorting
func TestAccountSearchWithSorting(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "status",
				Value:    "active",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Verify ascending order
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].Name > result.Data[i].Name {
			t.Errorf("Names not in ascending order: %s > %s",
				result.Data[i-1].Name, result.Data[i].Name)
		}
	}
}

// TestAccountSearchMultipleConditions tests multiple search conditions
func TestAccountSearchMultipleConditions(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	// Search for accounts with name containing "John" AND status = "active"
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "status",
				Value:    "active",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Should return active accounts containing 'John'
	for _, account := range result.Data {
		if account.Status != "active" {
			t.Errorf("Account %s status is not active", account.Name)
		}
	}
}

// TestAccountSearchORLogic tests OR logic between conditions
func TestAccountSearchORLogic(t *testing.T) {
	db := setupAccountDB(t)
	handler := filter.NewFilter[Account](filter.GolangFilteringConfig{})

	// Search for accounts with name starting with "John" OR "Alice"
	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John",
				Mode:     filter.ModeStartsWith,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "name",
				Value:    "Alice",
				Mode:     filter.ModeStartsWith,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error searching: %v", err)
	}

	// Should return: John Smith, John Anderson, Alice Brown (3 accounts)
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 accounts, got %d", result.TotalSize)
	}
}
