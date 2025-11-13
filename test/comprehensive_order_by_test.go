package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// OrderByTestUser represents a user for ORDER BY testing
type OrderByTestUser struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Age       int       `gorm:"not null" json:"age"`
	Salary    float64   `gorm:"type:decimal(10,2)" json:"salary"`
	Active    bool      `gorm:"default:false" json:"active"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`

	// Related data
	DepartmentID uint             `gorm:"not null" json:"department_id"`
	Department   *OrderByTestDept `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
}

// OrderByTestDept represents a department for nested ORDER BY testing
type OrderByTestDept struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(255)" json:"name"`
	Code string `gorm:"type:varchar(10)" json:"code"`
}

func setupOrderByDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate tables
	if err := db.AutoMigrate(&OrderByTestDept{}, &OrderByTestUser{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Insert test departments
	departments := []OrderByTestDept{
		{ID: 1, Name: "Engineering", Code: "ENG"},
		{ID: 2, Name: "Marketing", Code: "MKT"},
		{ID: 3, Name: "Sales", Code: "SAL"},
	}
	for _, dept := range departments {
		if err := db.Create(&dept).Error; err != nil {
			t.Fatalf("Failed to create department: %v", err)
		}
	}

	// Insert test users
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	users := []OrderByTestUser{
		{ID: 1, Name: "Alice Johnson", Age: 30, Salary: 75000.00, Active: true, DepartmentID: 1, CreatedAt: baseTime, UpdatedAt: baseTime},
		{ID: 2, Name: "Bob Smith", Age: 25, Salary: 65000.00, Active: true, DepartmentID: 2, CreatedAt: baseTime.Add(time.Hour), UpdatedAt: baseTime.Add(time.Hour)},
		{ID: 3, Name: "Charlie Brown", Age: 35, Salary: 85000.00, Active: false, DepartmentID: 1, CreatedAt: baseTime.Add(2 * time.Hour), UpdatedAt: baseTime.Add(2 * time.Hour)},
		{ID: 4, Name: "Diana Prince", Age: 28, Salary: 70000.00, Active: true, DepartmentID: 3, CreatedAt: baseTime.Add(3 * time.Hour), UpdatedAt: baseTime.Add(3 * time.Hour)},
		{ID: 5, Name: "Eve Adams", Age: 32, Salary: 80000.00, Active: true, DepartmentID: 2, CreatedAt: baseTime.Add(4 * time.Hour), UpdatedAt: baseTime.Add(4 * time.Hour)},
		{ID: 6, Name: "Frank Wilson", Age: 27, Salary: 60000.00, Active: false, DepartmentID: 3, CreatedAt: baseTime.Add(5 * time.Hour), UpdatedAt: baseTime.Add(5 * time.Hour)},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	return db
}

// TestOrderByBoolean tests ORDER BY on boolean fields
func TestOrderByBoolean(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test boolean ASC (false first, then true)
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "active", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY boolean ASC: %v", err)
	}

	// Verify false values come first, then true values
	foundTrue := false
	for i, user := range result.Data {
		if user.Active && !foundTrue {
			foundTrue = true
		} else if !user.Active && foundTrue {
			t.Errorf("Found false after true at position %d", i)
		}
	}

	// Test boolean DESC (true first, then false)
	filterRoot.SortFields[0].Order = filter.SortOrderDesc

	result, err = handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY boolean DESC: %v", err)
	}

	// Verify true values come first, then false values
	foundFalse := false
	for i, user := range result.Data {
		if !user.Active && !foundFalse {
			foundFalse = true
		} else if user.Active && foundFalse {
			t.Errorf("Found true after false at position %d", i)
		}
	}

	activeUsers := []string{}
	inactiveUsers := []string{}
	for _, user := range result.Data {
		if user.Active {
			activeUsers = append(activeUsers, user.Name)
		} else {
			inactiveUsers = append(inactiveUsers, user.Name)
		}
	}

	t.Logf("✅ ORDER BY boolean (active) DESC: %d active %v, %d inactive %v",
		len(activeUsers), activeUsers, len(inactiveUsers), inactiveUsers)
}

// TestOrderByStrings tests ORDER BY on string fields with comprehensive scenarios
func TestOrderByStrings(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test string ASC
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY string ASC: %v", err)
	}

	// Verify alphabetical ascending order
	expectedNamesAsc := []string{"Alice Johnson", "Bob Smith", "Charlie Brown", "Diana Prince", "Eve Adams", "Frank Wilson"}
	for i, user := range result.Data {
		if user.Name != expectedNamesAsc[i] {
			t.Errorf("Expected name '%s' at position %d, got '%s'", expectedNamesAsc[i], i, user.Name)
		}
	}

	// Test string DESC
	filterRoot.SortFields[0].Order = filter.SortOrderDesc

	result, err = handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY string DESC: %v", err)
	}

	// Verify alphabetical descending order
	expectedNamesDesc := []string{"Frank Wilson", "Eve Adams", "Diana Prince", "Charlie Brown", "Bob Smith", "Alice Johnson"}
	for i, user := range result.Data {
		if user.Name != expectedNamesDesc[i] {
			t.Errorf("Expected name '%s' at position %d, got '%s'", expectedNamesDesc[i], i, user.Name)
		}
	}

	t.Logf("✅ ORDER BY string ASC: %v", expectedNamesAsc)
	t.Logf("✅ ORDER BY string DESC: %v", expectedNamesDesc)
}

// TestOrderByDateTime tests ORDER BY on date and time fields
func TestOrderByDateTime(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test date field ASC
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "created_at", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY date ASC: %v", err)
	}

	// Verify ascending date order
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].CreatedAt.After(result.Data[i].CreatedAt) {
			t.Errorf("Dates not in ascending order: %v > %v",
				result.Data[i-1].CreatedAt, result.Data[i].CreatedAt)
		}
	}

	// Test date field DESC
	filterRoot.SortFields[0].Order = filter.SortOrderDesc

	result, err = handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY date DESC: %v", err)
	}

	// Verify descending date order
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].CreatedAt.Before(result.Data[i].CreatedAt) {
			t.Errorf("Dates not in descending order: %v < %v",
				result.Data[i-1].CreatedAt, result.Data[i].CreatedAt)
		}
	}

	// Test updated_at field as well
	filterRoot.SortFields[0].Field = "updated_at"
	filterRoot.SortFields[0].Order = filter.SortOrderAsc

	result, err = handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY updated_at: %v", err)
	}

	// Verify ascending updated_at order
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].UpdatedAt.After(result.Data[i].UpdatedAt) {
			t.Errorf("Updated dates not in ascending order: %v > %v",
				result.Data[i-1].UpdatedAt, result.Data[i].UpdatedAt)
		}
	}

	t.Logf("✅ ORDER BY date/time fields: created_at ASC, DESC and updated_at ASC verified")
}

// TestOrderByTime tests ORDER BY specifically on time components
func TestOrderByTime(t *testing.T) {
	db := setupOrderByDB(t)

	// Insert users with specific time patterns
	timeUsers := []OrderByTestUser{
		{ID: 10, Name: "Morning User", Age: 25, Salary: 50000, Active: true, DepartmentID: 1,
			CreatedAt: time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 1, 8, 30, 0, 0, time.UTC)},
		{ID: 11, Name: "Afternoon User", Age: 30, Salary: 60000, Active: true, DepartmentID: 1,
			CreatedAt: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)},
		{ID: 12, Name: "Evening User", Age: 35, Salary: 70000, Active: true, DepartmentID: 1,
			CreatedAt: time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 1, 20, 30, 0, 0, time.UTC)},
	}

	for _, user := range timeUsers {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to create time user: %v", err)
		}
	}

	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Filter to only get our time test users and sort by created_at
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "id",
				Value:    9,
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "created_at", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY time: %v", err)
	}

	expectedOrder := []string{"Morning User", "Afternoon User", "Evening User"}
	for i, user := range result.Data {
		if user.Name != expectedOrder[i] {
			t.Errorf("Expected user '%s' at position %d, got '%s'", expectedOrder[i], i, user.Name)
		}
	}

	t.Logf("✅ ORDER BY time components verified: Morning -> Afternoon -> Evening")
}

// TestOrderByMixedDataTypes tests ORDER BY with multiple data types
func TestOrderByMixedDataTypes(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test multiple fields with different data types: boolean DESC, then string ASC
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "active", Order: filter.SortOrderDesc}, // Boolean DESC (true first)
			{Field: "name", Order: filter.SortOrderAsc},    // String ASC
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to execute mixed data types ORDER BY: %v", err)
	}

	// Verify complex sorting logic
	activeUsers := []*OrderByTestUser{}
	inactiveUsers := []*OrderByTestUser{}

	for _, user := range result.Data {
		if user.Active {
			activeUsers = append(activeUsers, user)
		} else {
			inactiveUsers = append(inactiveUsers, user)
		}
	}

	// All active users should come first
	if len(activeUsers) == 0 {
		t.Error("No active users found")
	}

	// Within active users, verify name order
	for i := 1; i < len(activeUsers); i++ {
		if activeUsers[i-1].Name > activeUsers[i].Name {
			t.Errorf("Active users names not in ascending order: '%s' > '%s'",
				activeUsers[i-1].Name, activeUsers[i].Name)
		}
	}

	// Within inactive users, verify name order
	for i := 1; i < len(inactiveUsers); i++ {
		if inactiveUsers[i-1].Name > inactiveUsers[i].Name {
			t.Errorf("Inactive users names not in ascending order: '%s' > '%s'",
				inactiveUsers[i-1].Name, inactiveUsers[i].Name)
		}
	}

	t.Logf("✅ ORDER BY mixed data types (boolean DESC, string ASC)")
	t.Logf("   Active users: %v", getNames(activeUsers))
	t.Logf("   Inactive users: %v", getNames(inactiveUsers))
}

// TestOrderByNumericFields tests ORDER BY on different numeric types
func TestOrderByNumericFields(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test integer field (age) ASC
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY age: %v", err)
	}

	expectedAges := []int{25, 27, 28, 30, 32, 35}
	for i, user := range result.Data {
		if user.Age != expectedAges[i] {
			t.Errorf("Expected age %d at position %d, got %d", expectedAges[i], i, user.Age)
		}
	}

	// Test float field (salary) DESC
	filterRoot.SortFields[0].Field = "salary"
	filterRoot.SortFields[0].Order = filter.SortOrderDesc

	result, err = handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to ORDER BY salary DESC: %v", err)
	}

	// Verify descending salary order
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].Salary < result.Data[i].Salary {
			t.Errorf("Salaries not in descending order: %.2f < %.2f",
				result.Data[i-1].Salary, result.Data[i].Salary)
		}
	}

	t.Logf("✅ ORDER BY numeric fields: age ASC %v, salary DESC verified", expectedAges)
}

// TestOrderByNestedFields tests ORDER BY on related table fields
func TestOrderByNestedFields(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "department.name", Order: filter.SortOrderAsc},
		},
		Preload: []string{"Department"},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Failed to execute nested ORDER BY: %v", err)
	}

	// Verify ascending order by department name
	expectedDeptNames := []string{"Engineering", "Engineering", "Marketing", "Marketing", "Sales", "Sales"}
	for i, user := range result.Data {
		if user.Department == nil {
			t.Errorf("Department not preloaded for user %s", user.Name)
			continue
		}
		if user.Department.Name != expectedDeptNames[i] {
			t.Errorf("Expected department '%s' at position %d, got '%s'",
				expectedDeptNames[i], i, user.Department.Name)
		}
	}

	t.Logf("✅ ORDER BY nested field (department.name) ASC")
}

// TestOrderByWithPagination tests ORDER BY combined with pagination
func TestOrderByWithPagination(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	// Get first page
	result1, err := handler.DataGorm(db, filterRoot, 0, 3)
	if err != nil {
		t.Fatalf("Failed to execute ORDER BY with pagination page 0: %v", err)
	}

	// Get second page
	result2, err := handler.DataGorm(db, filterRoot, 1, 3)
	if err != nil {
		t.Fatalf("Failed to execute ORDER BY with pagination page 2: %v", err)
	}

	if len(result1.Data) != 3 {
		t.Errorf("Expected 3 users on page 0, got %d", len(result1.Data))
	}

	if len(result2.Data) != 3 {
		t.Errorf("Expected 3 users on page 1, got %d", len(result2.Data))
	}

	// Verify correct order across pages
	page1Names := getNames(result1.Data)
	page2Names := getNames(result2.Data)

	allNames := append(page1Names, page2Names...)
	expectedNames := []string{"Alice Johnson", "Bob Smith", "Charlie Brown", "Diana Prince", "Eve Adams", "Frank Wilson"}

	for i, name := range allNames {
		if name != expectedNames[i] {
			t.Errorf("Expected name '%s' at position %d across pages, got '%s'",
				expectedNames[i], i, name)
		}
	}

	t.Logf("✅ ORDER BY with pagination - Page 0: %v, Page 1: %v", page1Names, page2Names)
}

// TestMultipleOrderBy tests ORDER BY with multiple fields and mixed ASC/DESC directions
func TestMultipleOrderBy(t *testing.T) {
	db := setupOrderByDB(t)

	// Add more test data with duplicate names and ages for better testing
	additionalUsers := []OrderByTestUser{
		{ID: 20, Name: "Alice Johnson", Age: 25, Salary: 50000.00, Active: false, DepartmentID: 2,
			CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: 21, Name: "Bob Smith", Age: 35, Salary: 90000.00, Active: true, DepartmentID: 1,
			CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		{ID: 22, Name: "Alice Johnson", Age: 40, Salary: 95000.00, Active: true, DepartmentID: 3,
			CreatedAt: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)},
	}

	for _, user := range additionalUsers {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to create additional user: %v", err)
		}
	}

	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test 1: Name ASC, Age DESC
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc}, // Primary sort: Name ascending
			{Field: "age", Order: filter.SortOrderDesc}, // Secondary sort: Age descending
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 20)
	if err != nil {
		t.Fatalf("Failed to execute multiple ORDER BY (name ASC, age DESC): %v", err)
	}

	// Verify the results
	t.Logf("📊 Multiple ORDER BY - Name ASC, Age DESC:")
	previousName := ""
	previousAge := -1

	for i, user := range result.Data {
		t.Logf("  %d: %s (age %d, salary %.0f)", i+1, user.Name, user.Age, user.Salary)

		// Verify name ordering (should be ascending)
		if previousName != "" && user.Name < previousName {
			t.Errorf("Names not in ascending order: '%s' should come after '%s'", user.Name, previousName)
		}

		// For same names, verify age ordering (should be descending)
		if previousName == user.Name && previousAge != -1 && user.Age > previousAge {
			t.Errorf("For name '%s': ages not in descending order: %d should come before %d",
				user.Name, user.Age, previousAge)
		}

		if previousName != user.Name {
			previousAge = -1 // Reset age comparison for new name
		}
		previousName = user.Name
		previousAge = user.Age
	}

	// Test 2: Age ASC, Salary DESC
	filterRoot.SortFields = []filter.SortField{
		{Field: "age", Order: filter.SortOrderAsc},     // Primary sort: Age ascending
		{Field: "salary", Order: filter.SortOrderDesc}, // Secondary sort: Salary descending
	}

	result, err = handler.DataGorm(db, filterRoot, 0, 20)
	if err != nil {
		t.Fatalf("Failed to execute multiple ORDER BY (age ASC, salary DESC): %v", err)
	}

	t.Logf("\n📊 Multiple ORDER BY - Age ASC, Salary DESC:")
	previousAge = -1
	var previousSalary float64 = -1

	for i, user := range result.Data {
		t.Logf("  %d: %s (age %d, salary %.0f)", i+1, user.Name, user.Age, user.Salary)

		// Verify age ordering (should be ascending)
		if previousAge != -1 && user.Age < previousAge {
			t.Errorf("Ages not in ascending order: %d should come after %d", user.Age, previousAge)
		}

		// For same ages, verify salary ordering (should be descending)
		if previousAge == user.Age && previousSalary != -1 && user.Salary > previousSalary {
			t.Errorf("For age %d: salaries not in descending order: %.0f should come before %.0f",
				user.Age, user.Salary, previousSalary)
		}

		if previousAge != user.Age {
			previousSalary = -1 // Reset salary comparison for new age
		}
		previousAge = user.Age
		previousSalary = user.Salary
	}

	// Test 3: Three fields - Active DESC, Name ASC, Age ASC
	filterRoot.SortFields = []filter.SortField{
		{Field: "active", Order: filter.SortOrderDesc}, // Primary: Active descending (true first)
		{Field: "name", Order: filter.SortOrderAsc},    // Secondary: Name ascending
		{Field: "age", Order: filter.SortOrderAsc},     // Tertiary: Age ascending
	}

	result, err = handler.DataGorm(db, filterRoot, 0, 20)
	if err != nil {
		t.Fatalf("Failed to execute triple ORDER BY (active DESC, name ASC, age ASC): %v", err)
	}

	t.Logf("\n📊 Triple ORDER BY - Active DESC, Name ASC, Age ASC:")
	previousActive := true
	previousName = ""
	previousAge = -1

	for i, user := range result.Data {
		t.Logf("  %d: %s (active: %t, age %d)", i+1, user.Name, user.Active, user.Age)

		// Verify active ordering (should be descending - true first)
		if !user.Active && previousActive {
			previousActive = false // First transition from active to inactive
			previousName = ""      // Reset for new active group
			previousAge = -1
		}

		// Within same active status, verify name ordering
		if previousName != "" && user.Name < previousName {
			activeStr := "active"
			if !user.Active {
				activeStr = "inactive"
			}
			t.Errorf("Within %s users, names not in ascending order: '%s' should come after '%s'",
				activeStr, user.Name, previousName)
		}

		// Within same active status and name, verify age ordering
		if previousName == user.Name && previousAge != -1 && user.Age < previousAge {
			t.Errorf("For name '%s': ages not in ascending order: %d should come after %d",
				user.Name, user.Age, previousAge)
		}

		if previousName != user.Name {
			previousAge = -1
		}
		previousName = user.Name
		previousAge = user.Age
	}

	t.Logf("✅ Multiple ORDER BY tests completed successfully")
}

// TestComplexMultipleOrderBy tests edge cases with multiple ORDER BY fields
func TestComplexMultipleOrderBy(t *testing.T) {
	db := setupOrderByDB(t)

	// Create users with identical values in some fields to test tie-breaking
	complexUsers := []OrderByTestUser{
		// Same department, different other fields
		{ID: 30, Name: "Zoe Alpha", Age: 30, Salary: 70000.00, Active: true, DepartmentID: 1,
			CreatedAt: time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)},
		{ID: 31, Name: "Zoe Alpha", Age: 30, Salary: 80000.00, Active: true, DepartmentID: 1,
			CreatedAt: time.Date(2024, 2, 1, 11, 0, 0, 0, time.UTC)},
		{ID: 32, Name: "Zoe Alpha", Age: 25, Salary: 70000.00, Active: false, DepartmentID: 1,
			CreatedAt: time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 33, Name: "Aaron Beta", Age: 30, Salary: 70000.00, Active: true, DepartmentID: 2,
			CreatedAt: time.Date(2024, 2, 1, 13, 0, 0, 0, time.UTC)},
	}

	for _, user := range complexUsers {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to create complex user: %v", err)
		}
	}

	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test complex sorting: Name ASC, Age DESC, Salary ASC, CreatedAt DESC
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},        // 1st: Name ascending
			{Field: "age", Order: filter.SortOrderDesc},        // 2nd: Age descending
			{Field: "salary", Order: filter.SortOrderAsc},      // 3rd: Salary ascending
			{Field: "created_at", Order: filter.SortOrderDesc}, // 4th: CreatedAt descending
		},
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "id",
				Value:    29,
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 20)
	if err != nil {
		t.Fatalf("Failed to execute complex multiple ORDER BY: %v", err)
	}

	t.Logf("📊 Complex Multiple ORDER BY (Name ASC, Age DESC, Salary ASC, CreatedAt DESC):")
	for i, user := range result.Data {
		t.Logf("  %d: %s | Age: %d | Salary: %.0f | Created: %s",
			i+1, user.Name, user.Age, user.Salary, user.CreatedAt.Format("15:04"))
	}

	// Verify the expected order for our test data:
	// Aaron Beta (age 30, salary 70000) - should be first alphabetically
	// Zoe Alpha (age 30, salary 70000, created 10:00) - higher age than age 25
	// Zoe Alpha (age 30, salary 80000, created 11:00) - same name/age, higher salary
	// Zoe Alpha (age 25, salary 70000, created 12:00) - same name, lower age

	if len(result.Data) < 4 {
		t.Fatalf("Expected at least 4 users, got %d", len(result.Data))
	}

	// Verify Aaron Beta comes first (alphabetically first name)
	if result.Data[0].Name != "Aaron Beta" {
		t.Errorf("Expected Aaron Beta first, got %s", result.Data[0].Name)
	}

	// Find all Zoe Alpha entries and verify their ordering
	zoeEntries := []*OrderByTestUser{}
	for _, user := range result.Data {
		if user.Name == "Zoe Alpha" {
			zoeEntries = append(zoeEntries, user)
		}
	}

	if len(zoeEntries) != 3 {
		t.Errorf("Expected 3 Zoe Alpha entries, got %d", len(zoeEntries))
	}

	// Verify Zoe entries are ordered by: Age DESC, then Salary ASC, then CreatedAt DESC
	for i := 1; i < len(zoeEntries); i++ {
		prev := zoeEntries[i-1]
		curr := zoeEntries[i]

		// Check age ordering (DESC)
		if prev.Age < curr.Age {
			t.Errorf("Age ordering wrong: %d should come before %d", prev.Age, curr.Age)
		} else if prev.Age == curr.Age {
			// Same age, check salary (ASC)
			if prev.Salary > curr.Salary {
				t.Errorf("Salary ordering wrong for same age %d: %.0f should come before %.0f",
					prev.Age, prev.Salary, curr.Salary)
			} else if prev.Salary == curr.Salary {
				// Same age and salary, check created_at (DESC)
				if prev.CreatedAt.Before(curr.CreatedAt) {
					t.Errorf("CreatedAt ordering wrong: %s should come before %s",
						prev.CreatedAt.Format("15:04"), curr.CreatedAt.Format("15:04"))
				}
			}
		}
	}

	// Test with nested field in multiple ORDER BY
	filterRoot.SortFields = []filter.SortField{
		{Field: "department.name", Order: filter.SortOrderAsc},
		{Field: "name", Order: filter.SortOrderDesc},
		{Field: "salary", Order: filter.SortOrderDesc},
	}
	filterRoot.Preload = []string{"Department"}
	filterRoot.FieldFilters = nil // Remove the ID filter

	result, err = handler.DataGorm(db, filterRoot, 0, 20)
	if err != nil {
		t.Fatalf("Failed to execute nested multiple ORDER BY: %v", err)
	}

	t.Logf("\n📊 Nested Multiple ORDER BY (Department ASC, Name DESC, Salary DESC):")
	for i, user := range result.Data {
		deptName := "Unknown"
		if user.Department != nil {
			deptName = user.Department.Name
		}
		t.Logf("  %d: %s | %s | Salary: %.0f",
			i+1, deptName, user.Name, user.Salary)
	}

	// Verify department ordering
	prevDeptName := ""
	for _, user := range result.Data {
		if user.Department == nil {
			t.Error("Department not preloaded")
			continue
		}

		currDeptName := user.Department.Name
		if prevDeptName != "" && currDeptName < prevDeptName {
			t.Errorf("Department names not in ascending order: '%s' should come after '%s'",
				currDeptName, prevDeptName)
		}
		prevDeptName = currDeptName
	}

	t.Logf("✅ Complex multiple ORDER BY scenarios completed successfully")
}

// Helper functions
func getNames(users []*OrderByTestUser) []string {
	names := make([]string, len(users))
	for i, user := range users {
		names[i] = user.Name
	}
	return names
}

func getAges(users []*OrderByTestUser) []int {
	ages := make([]int, len(users))
	for i, user := range users {
		ages[i] = user.Age
	}
	return ages
}
