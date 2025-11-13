package example

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	Name         string      `json:"name"`
	ID           uint        `json:"id" gorm:"primaryKey"`
	FirstName    string      `json:"first_name"`
	LastName     string      `json:"last_name"`
	Email        string      `json:"email"`
	Age          int         `json:"age"`
	IsActive     bool        `json:"is_active"`
	Role         string      `json:"role"`
	Salary       float64     `json:"salary"`
	DepartmentID uint        `json:"department_id"`
	Department   *Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	CreatedAt    time.Time   `json:"created_at"`
}

type Department struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func main() {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Migrate tables
	db.AutoMigrate(&Department{}, &User{})

	// Create sample departments
	engineering := &Department{ID: 1, Name: "Engineering", Code: "ENG"}
	marketing := &Department{ID: 2, Name: "Marketing", Code: "MKT"}
	hr := &Department{ID: 3, Name: "Human Resources", Code: "HR"}

	db.Create([]Department{*engineering, *marketing, *hr})

	// Create sample users
	users := []*User{
		{ID: 1, FirstName: "John", LastName: "Doe", Email: "john.doe@company.com", Age: 30, IsActive: true, Role: "senior_developer", Salary: 85000.00, DepartmentID: 1, CreatedAt: time.Now().AddDate(-2, 0, 0)},
		{ID: 2, FirstName: "Jane", LastName: "Smith", Email: "jane.smith@company.com", Age: 28, IsActive: true, Role: "marketing_manager", Salary: 75000.00, DepartmentID: 2, CreatedAt: time.Now().AddDate(-1, -6, 0)},
		{ID: 3, FirstName: "Bob", LastName: "Johnson", Email: "bob.johnson@company.com", Age: 35, IsActive: false, Role: "developer", Salary: 70000.00, DepartmentID: 1, CreatedAt: time.Now().AddDate(-3, 0, 0)},
		{ID: 4, FirstName: "Alice", LastName: "Brown", Email: "alice.brown@company.com", Age: 26, IsActive: true, Role: "hr_coordinator", Salary: 55000.00, DepartmentID: 3, CreatedAt: time.Now().AddDate(-1, 0, 0)},
		{ID: 5, FirstName: "Charlie", LastName: "Wilson", Email: "charlie.wilson@company.com", Age: 32, IsActive: true, Role: "team_lead", Salary: 90000.00, DepartmentID: 1, CreatedAt: time.Now().AddDate(-4, 0, 0)},
	}

	for _, user := range users {
		db.Create(user)
	}

	// Create handler with automatic field getters
	handler := filter.NewFilter[User](filter.GolangFilteringConfig{})

	fmt.Println("=== Custom CSV Export Examples ===\n")

	// Example 1: In-Memory Custom CSV with Field Transformation
	fmt.Println("1. In-Memory Custom CSV with Field Transformation:")
	filterRoot := filter.Root{
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				DataType: filter.DataTypeBool,
				Mode:     filter.ModeEqual,
				Value:    true,
			},
		},
		Logic: filter.LogicAnd,
	}

	csvData, err := handler.DataQueryNoPageCSVCustom(users, filterRoot, func(user *User) map[string]any {
		return map[string]any{
			"Employee ID":   user.ID,
			"Full Name":     fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			"Email Address": user.Email,
			"Age":           user.Age,
			"Position":      strings.ReplaceAll(strings.Title(user.Role), "_", " "),
			"Annual Salary": fmt.Sprintf("$%.2f", user.Salary),
			"Status":        map[bool]string{true: "Active", false: "Inactive"}[user.IsActive],
			"Hire Date":     user.CreatedAt.Format("January 2, 2006"),
		}
	})

	if err != nil {
		log.Fatal("DataQueryNoPageCSVCustom failed:", err)
	}

	fmt.Println(string(csvData))
	fmt.Println()

	// Example 2: Database Custom CSV with Department Information
	fmt.Println("2. Database Custom CSV with Department Information (with preloading):")

	// Preload department information for the custom CSV
	db.Preload("Department").Find(&users)

	csvData, err = handler.GormNoPaginationCSVCustom(db.Preload("Department"), filterRoot, func(user *User) map[string]any {
		departmentName := "Unknown"
		departmentCode := "N/A"
		if user.Department != nil {
			departmentName = user.Department.Name
			departmentCode = user.Department.Code
		}

		return map[string]any{
			"ID":         user.ID,
			"Name":       fmt.Sprintf("%s, %s", user.LastName, user.FirstName),
			"Email":      user.Email,
			"Department": fmt.Sprintf("%s (%s)", departmentName, departmentCode),
			"Role":       strings.ToUpper(strings.ReplaceAll(user.Role, "_", " ")),
			"Salary":     fmt.Sprintf("$%,.0f", user.Salary),
		}
	})

	if err != nil {
		log.Fatal("GormNoPaginationCSVCustom failed:", err)
	}

	fmt.Println(string(csvData))
	fmt.Println()

	// Example 3: Hybrid Custom CSV with Preset Conditions
	fmt.Println("3. Hybrid Custom CSV with Preset Conditions (Engineering Department only):")

	type DepartmentFilter struct {
		DepartmentID uint `gorm:"column:department_id"`
	}

	presetConditions := &DepartmentFilter{DepartmentID: 1} // Engineering department

	csvData, err = handler.HybridCSVCustomWithPreset(db.Preload("Department"), presetConditions, 1000, filter.Root{Logic: filter.LogicAnd}, func(user *User) map[string]any {
		// Custom calculation for experience level
		yearsExperience := time.Since(user.CreatedAt).Hours() / (24 * 365)
		var experienceLevel string
		switch {
		case yearsExperience < 1:
			experienceLevel = "Junior"
		case yearsExperience < 3:
			experienceLevel = "Mid-Level"
		default:
			experienceLevel = "Senior"
		}

		return map[string]any{
			"Engineer ID":      user.ID,
			"Engineer Name":    fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			"Contact Email":    user.Email,
			"Technical Role":   strings.Title(strings.ReplaceAll(user.Role, "_", " ")),
			"Experience Level": experienceLevel,
			"Years at Company": fmt.Sprintf("%.1f", yearsExperience),
			"Current Status":   map[bool]string{true: "ACTIVE", false: "INACTIVE"}[user.IsActive],
		}
	})

	if err != nil {
		log.Fatal("HybridCSVCustomWithPreset failed:", err)
	}

	fmt.Println(string(csvData))
	fmt.Println()

	// Example 4: Custom CSV with Complex Business Logic
	fmt.Println("4. Custom CSV with Complex Business Logic (Performance Report):")

	csvData, err = handler.DataQueryNoPageCSVCustom(users, filter.Root{Logic: filter.LogicAnd}, func(user *User) map[string]any {
		// Calculate performance metrics based on salary and experience
		yearsExperience := time.Since(user.CreatedAt).Hours() / (24 * 365)
		salaryPerYear := user.Salary / (yearsExperience + 0.1) // Avoid division by zero

		var performanceRating string
		switch {
		case salaryPerYear > 50000:
			performanceRating = "Excellent"
		case salaryPerYear > 35000:
			performanceRating = "Good"
		case salaryPerYear > 25000:
			performanceRating = "Average"
		default:
			performanceRating = "Needs Improvement"
		}

		// Calculate bonus eligibility
		bonusEligible := user.IsActive && yearsExperience > 1 && user.Salary > 60000

		return map[string]any{
			"Employee":         fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			"Department ID":    user.DepartmentID,
			"Current Salary":   fmt.Sprintf("$%,.2f", user.Salary),
			"Years Experience": fmt.Sprintf("%.1f", yearsExperience),
			"Salary Per Year":  fmt.Sprintf("$%,.0f", salaryPerYear),
			"Performance":      performanceRating,
			"Bonus Eligible":   map[bool]string{true: "YES", false: "NO"}[bonusEligible],
			"Report Date":      time.Now().Format("2006-01-02 15:04:05"),
		}
	})

	if err != nil {
		log.Fatal("Complex custom CSV failed:", err)
	}

	fmt.Println(string(csvData))

	fmt.Println("\n=== Custom CSV Features Demonstrated ===")
	fmt.Println("✓ Field transformation and formatting")
	fmt.Println("✓ Custom column names and headers")
	fmt.Println("✓ Calculated fields and business logic")
	fmt.Println("✓ Nested field access (with preloading)")
	fmt.Println("✓ Preset conditions with custom mapping")
	fmt.Println("✓ Deterministic column ordering (alphabetical)")
	fmt.Println("✓ RFC 4180 compliant CSV escaping")
	fmt.Println("✓ Compatible with in-memory, database, and hybrid strategies")
}
