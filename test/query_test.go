package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestUser represents a test user model
type TestUser struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	IsActive  bool      `json:"is_active"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// generateTestUsers creates sample test users
func generateTestUsers() []*TestUser {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return []*TestUser{
		{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 25, IsActive: true, Role: "admin", CreatedAt: baseTime},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 30, IsActive: true, Role: "user", CreatedAt: baseTime.AddDate(0, 1, 0)},
		{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Age: 35, IsActive: false, Role: "user", CreatedAt: baseTime.AddDate(0, 2, 0)},
		{ID: 4, Name: "Alice Brown", Email: "alice@example.com", Age: 28, IsActive: true, Role: "moderator", CreatedAt: baseTime.AddDate(0, 3, 0)},
		{ID: 5, Name: "Charlie Wilson", Email: "charlie@example.com", Age: 42, IsActive: true, Role: "admin", CreatedAt: baseTime.AddDate(0, 4, 0)},
		{ID: 6, Name: "Diana Prince", Email: "diana@example.com", Age: 33, IsActive: false, Role: "user", CreatedAt: baseTime.AddDate(0, 5, 0)},
		{ID: 7, Name: "John Smith", Email: "johnsmith@example.com", Age: 29, IsActive: true, Role: "user", CreatedAt: baseTime.AddDate(0, 6, 0)},
		{ID: 8, Name: "Eve Adams", Email: "eve@example.com", Age: 26, IsActive: true, Role: "moderator", CreatedAt: baseTime.AddDate(0, 7, 0)},
		{ID: 9, Name: "Frank Miller", Email: "frank@example.com", Age: 38, IsActive: false, Role: "user", CreatedAt: baseTime.AddDate(0, 8, 0)},
		{ID: 10, Name: "Grace Lee", Email: "grace@example.com", Age: 31, IsActive: true, Role: "admin", CreatedAt: baseTime.AddDate(0, 9, 0)},
	}
}

// TestFilterHandler_NewFilter tests filter handler creation
func TestFilterHandler_NewFilter(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
}

// TestFilterHandler_DataQuery_EmptyData tests filtering empty data
func TestFilterHandler_DataQuery_EmptyData(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	result, err := handler.DataQuery([]*TestUser{}, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalSize != 0 {
		t.Errorf("Expected total size 0, got %d", result.TotalSize)
	}
	if len(result.Data) != 0 {
		t.Errorf("Expected empty data, got %d items", len(result.Data))
	}
}

// TestFilterHandler_DataQuery_NoFilters tests with no filters applied
func TestFilterHandler_DataQuery_NoFilters(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalSize != len(users) {
		t.Errorf("Expected total size %d, got %d", len(users), result.TotalSize)
	}
	if len(result.Data) != len(users) {
		t.Errorf("Expected %d items, got %d", len(users), len(result.Data))
	}
}

// TestFilterHandler_ModeEqual tests equal filter mode
func TestFilterHandler_ModeEqual(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "role",
				Value:    "admin",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find 3 admins: John Doe, Charlie Wilson, Grace Lee
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 admins, got %d", result.TotalSize)
	}

	for _, user := range result.Data {
		if user.Role != "admin" {
			t.Errorf("Expected all users to be admins, got role %s", user.Role)
		}
	}
}

// TestFilterHandler_ModeNotEqual tests not equal filter mode
func TestFilterHandler_ModeNotEqual(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "role",
				Value:    "user",
				Mode:     filter.ModeNotEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find all non-users (admins and moderators)
	for _, user := range result.Data {
		if user.Role == "user" {
			t.Errorf("Expected no users with role 'user', got %s", user.Name)
		}
	}
}

// TestFilterHandler_ModeContains tests contains filter mode
func TestFilterHandler_ModeContains(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

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

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find 3 users: John Doe, Bob Johnson, John Smith
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 users with 'John' in name, got %d", result.TotalSize)
	}

	for _, user := range result.Data {
		if !containsIgnoreCase(user.Name, "John") {
			t.Errorf("Expected all names to contain 'John', got %s", user.Name)
		}
	}
}

// TestFilterHandler_ModeStartsWith tests starts with filter mode
func TestFilterHandler_ModeStartsWith(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

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

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find 2 users: John Doe, John Smith
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 users starting with 'John', got %d", result.TotalSize)
	}
}

// TestFilterHandler_ModeEndsWith tests ends with filter mode
func TestFilterHandler_ModeEndsWith(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "email",
				Value:    "example.com",
				Mode:     filter.ModeEndsWith,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// All users should have example.com emails
	if result.TotalSize != len(users) {
		t.Errorf("Expected all %d users, got %d", len(users), result.TotalSize)
	}
}

// TestFilterHandler_ModeGT tests greater than filter mode
func TestFilterHandler_ModeGT(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users older than 30
	for _, user := range result.Data {
		if user.Age <= 30 {
			t.Errorf("Expected age > 30, got %d for %s", user.Age, user.Name)
		}
	}
}

// TestFilterHandler_ModeGTE tests greater than or equal filter mode
func TestFilterHandler_ModeGTE(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users 30 or older
	for _, user := range result.Data {
		if user.Age < 30 {
			t.Errorf("Expected age >= 30, got %d for %s", user.Age, user.Name)
		}
	}
}

// TestFilterHandler_ModeLT tests less than filter mode
func TestFilterHandler_ModeLT(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeLT,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users younger than 30
	for _, user := range result.Data {
		if user.Age >= 30 {
			t.Errorf("Expected age < 30, got %d for %s", user.Age, user.Name)
		}
	}
}

// TestFilterHandler_ModeLTE tests less than or equal filter mode
func TestFilterHandler_ModeLTE(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeLTE,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users 30 or younger
	for _, user := range result.Data {
		if user.Age > 30 {
			t.Errorf("Expected age <= 30, got %d for %s", user.Age, user.Name)
		}
	}
}

// TestFilterHandler_ModeRange tests range filter mode
func TestFilterHandler_ModeRange(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    filter.Range{From: 28, To: 35},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users between 28 and 35
	for _, user := range result.Data {
		if user.Age < 28 || user.Age > 35 {
			t.Errorf("Expected age between 28-35, got %d for %s", user.Age, user.Name)
		}
	}
}

// TestFilterHandler_BoolFilter tests boolean filter
func TestFilterHandler_BoolFilter(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find only active users
	for _, user := range result.Data {
		if !user.IsActive {
			t.Errorf("Expected all users to be active, got %s (inactive)", user.Name)
		}
	}
}

// TestFilterHandler_LogicAnd tests AND logic
func TestFilterHandler_LogicAnd(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find active users who are 30 or older
	for _, user := range result.Data {
		if !user.IsActive || user.Age < 30 {
			t.Errorf("Expected active users aged >= 30, got %s (active=%v, age=%d)", user.Name, user.IsActive, user.Age)
		}
	}
}

// TestFilterHandler_LogicOr tests OR logic
func TestFilterHandler_LogicOr(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "role",
				Value:    "admin",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "role",
				Value:    "moderator",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find admins OR moderators
	for _, user := range result.Data {
		if user.Role != "admin" && user.Role != "moderator" {
			t.Errorf("Expected admin or moderator, got role %s for %s", user.Role, user.Name)
		}
	}
}

// TestFilterHandler_Pagination tests pagination
func TestFilterHandler_Pagination(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	// Page 0 (first page in 0-based indexing), size 3
	result, err := handler.DataQuery(users, filterRoot, 0, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.PageIndex != 0 {
		t.Errorf("Expected page index 0 (0-based), got %d", result.PageIndex)
	}
	if result.PageSize != 3 {
		t.Errorf("Expected page size 3, got %d", result.PageSize)
	}
	if len(result.Data) != 3 {
		t.Errorf("Expected 3 items on page 0, got %d", len(result.Data))
	}
	if result.TotalSize != len(users) {
		t.Errorf("Expected total size %d, got %d", len(users), result.TotalSize)
	}
	expectedPages := (len(users) + 3 - 1) / 3
	if result.TotalPage != expectedPages {
		t.Errorf("Expected %d total pages, got %d", expectedPages, result.TotalPage)
	}
}

// TestFilterHandler_Sorting tests sorting functionality
func TestFilterHandler_Sorting(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderDesc},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be sorted by age descending
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].Age < result.Data[i].Age {
			t.Errorf("Expected descending age order, got %d before %d", result.Data[i-1].Age, result.Data[i].Age)
		}
	}
}

// TestFilterHandler_MultipleSorts tests multiple sort fields
func TestFilterHandler_MultipleSorts(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "is_active", Order: filter.SortOrderDesc},
			{Field: "age", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// First should be sorted by is_active (true first), then by age ascending
	if len(result.Data) < 2 {
		t.Fatal("Expected at least 2 results")
	}
}

// Helper function for case-insensitive contains check
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		indexIgnoreCase(s, substr) >= 0)
}

func indexIgnoreCase(s, substr string) int {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
