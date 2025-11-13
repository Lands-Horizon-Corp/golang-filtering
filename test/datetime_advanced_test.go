package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Appointment model for advanced datetime testing
type Appointment struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `json:"title"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Duration    int       `json:"duration"` // in minutes
}

// TestGormDateTimeEqual tests datetime equality with GORM
func TestGormDateTimeEqual(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Appointment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	targetDateTime := time.Date(2025, 11, 5, 14, 30, 0, 0, time.UTC)

	appointments := []Appointment{
		{ID: 1, Title: "Meeting 1", ScheduledAt: targetDateTime, Duration: 60},
		{ID: 2, Title: "Meeting 2", ScheduledAt: targetDateTime.Add(time.Hour), Duration: 30},
		{ID: 3, Title: "Meeting 3", ScheduledAt: targetDateTime, Duration: 45},
		{ID: 4, Title: "Meeting 4", ScheduledAt: targetDateTime.Add(-time.Hour), Duration: 90},
	}

	db.Create(&appointments)

	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "scheduled_at",
				Value:    "2025-11-05T14:30:00Z",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 appointments at exact time, got %d", result.TotalSize)
	}
}

// TestGormDateTimeNotEqual tests datetime not equal with GORM
func TestGormDateTimeNotEqual(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Appointment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	targetDate := time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC)

	appointments := []Appointment{
		{ID: 1, Title: "App1", StartDate: targetDate, Duration: 30},
		{ID: 2, Title: "App2", StartDate: targetDate.AddDate(0, 0, 1), Duration: 60},
		{ID: 3, Title: "App3", StartDate: targetDate.AddDate(0, 0, -1), Duration: 45},
		{ID: 4, Title: "App4", StartDate: targetDate, Duration: 90},
	}

	db.Create(&appointments)

	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "start_date",
				Value:    "2025-11-05",
				Mode:     filter.ModeNotEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 appointments not on Nov 5, got %d", result.TotalSize)
	}
}

// TestGormTimeNotEqual tests time not equal with GORM
func TestGormTimeNotEqual(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Appointment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	targetTime := time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)

	appointments := []Appointment{
		{ID: 1, Title: "App1", StartTime: targetTime, Duration: 30},
		{ID: 2, Title: "App2", StartTime: time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC), Duration: 60},
		{ID: 3, Title: "App3", StartTime: time.Date(0, 1, 1, 16, 0, 0, 0, time.UTC), Duration: 45},
		{ID: 4, Title: "App4", StartTime: targetTime, Duration: 90},
	}

	db.Create(&appointments)

	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "start_time",
				Value:    "14:30:00",
				Mode:     filter.ModeNotEqual,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 appointments not at 14:30, got %d", result.TotalSize)
	}
}

// TestGormTimeGT tests time greater than with GORM
func TestGormTimeGT(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Appointment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	appointments := []Appointment{
		{ID: 1, Title: "Morning", StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC), Duration: 60},
		{ID: 2, Title: "Lunch", StartTime: time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC), Duration: 60},
		{ID: 3, Title: "Afternoon", StartTime: time.Date(0, 1, 1, 15, 0, 0, 0, time.UTC), Duration: 90},
		{ID: 4, Title: "Evening", StartTime: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC), Duration: 120},
	}

	db.Create(&appointments)

	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "start_time",
				Value:    "14:00:00",
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 appointments after 14:00, got %d", result.TotalSize)
	}
}

// TestDateTimeRangeOverlap tests finding appointments within a date range
func TestDateTimeRangeOverlap(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{
			ID:        1,
			Title:     "Weekly Meeting",
			StartDate: time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Title:     "Conference",
			StartDate: time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 11, 7, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        3,
			Title:     "Workshop",
			StartDate: time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 11, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:        4,
			Title:     "Team Building",
			StartDate: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	// Filter for appointments starting in first half of November
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "start_date",
				Value: filter.Range{
					From: "2025-11-01",
					To:   "2025-11-10",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 appointments in first half of November, got %d", result.TotalSize)
	}
}

// TestTimeRangeWithSeconds tests time filtering with second precision
func TestTimeRangeWithSeconds(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{ID: 1, Title: "App1", StartTime: time.Date(0, 1, 1, 14, 29, 55, 0, time.UTC)},
		{ID: 2, Title: "App2", StartTime: time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)},
		{ID: 3, Title: "App3", StartTime: time.Date(0, 1, 1, 14, 30, 15, 0, time.UTC)},
		{ID: 4, Title: "App4", StartTime: time.Date(0, 1, 1, 14, 30, 30, 0, time.UTC)},
		{ID: 5, Title: "App5", StartTime: time.Date(0, 1, 1, 14, 30, 45, 0, time.UTC)},
		{ID: 6, Title: "App6", StartTime: time.Date(0, 1, 1, 14, 31, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "start_time",
				Value: filter.Range{
					From: "14:30:00",
					To:   "14:30:30",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 appointments in 30-second window, got %d", result.TotalSize)
	}
}

// TestMultipleDateFiltersWithOR tests combining date filters with OR logic
func TestMultipleDateFiltersWithOR(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{ID: 1, Title: "Jan1", StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Feb14", StartDate: time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "Mar15", StartDate: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Title: "Jul4", StartDate: time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Title: "Dec25", StartDate: time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)},
	}

	// Find appointments on New Year's Day OR Christmas
	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "start_date",
				Value:    "2025-01-01",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
			{
				Field:    "start_date",
				Value:    "2025-12-25",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 appointments on holidays, got %d", result.TotalSize)
	}
}

// TestDateAndTimeFiltersWithAND tests combining date and time filters
func TestDateAndTimeFiltersWithAND(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{
			ID:        1,
			Title:     "Morning Nov 5",
			StartDate: time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Title:     "Afternoon Nov 5",
			StartDate: time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:        3,
			Title:     "Morning Nov 6",
			StartDate: time.Date(2025, 11, 6, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:        4,
			Title:     "Afternoon Nov 6",
			StartDate: time.Date(2025, 11, 6, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC),
		},
	}

	// Find afternoon appointments on Nov 5
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "start_date",
				Value:    "2025-11-05",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
			{
				Field:    "start_time",
				Value:    "12:00:00",
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 afternoon appointment on Nov 5, got %d", result.TotalSize)
	}
}

// TestDateSortingDescending tests sorting dates in descending order
func TestDateSortingDescending(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{ID: 1, Title: "Middle", StartDate: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Latest", StartDate: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "Earliest", StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Title: "Mid-Late", StartDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		SortFields: []filter.SortField{
			{Field: "start_date", Order: filter.SortOrderDesc},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error sorting: %v", err)
	}

	if len(result.Data) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(result.Data))
	}

	// Verify descending order
	if result.Data[0].ID != 2 || result.Data[3].ID != 3 {
		t.Error("Dates not sorted in descending order")
	}
}

// TestTimeSortingAscending tests sorting times in ascending order
func TestTimeSortingAscending(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{ID: 1, Title: "Afternoon", StartTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Evening", StartTime: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "Morning", StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)},
		{ID: 4, Title: "Noon", StartTime: time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		SortFields: []filter.SortField{
			{Field: "start_time", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error sorting: %v", err)
	}

	if len(result.Data) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(result.Data))
	}

	// Verify ascending order
	if result.Data[0].ID != 3 || result.Data[3].ID != 2 {
		t.Error("Times not sorted in ascending order")
	}
}

// TestDateRangeWithPagination tests date filtering with pagination
func TestDateRangeWithPagination(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	// Create 30 appointments throughout November
	appointments := make([]*Appointment, 30)
	for i := 0; i < 30; i++ {
		appointments[i] = &Appointment{
			ID:        uint(i + 1),
			Title:     "Appointment " + string(rune(i+1)),
			StartDate: time.Date(2025, 11, i+1, 0, 0, 0, 0, time.UTC),
		}
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "start_date",
				Value: filter.Range{
					From: "2025-11-01",
					To:   "2025-11-30",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
		SortFields: []filter.SortField{
			{Field: "start_date", Order: filter.SortOrderAsc},
		},
	}

	// Test first page
	result1, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error on page 1: %v", err)
	}

	if result1.TotalSize != 30 {
		t.Errorf("Expected total 30, got %d", result1.TotalSize)
	}

	if len(result1.Data) != 10 {
		t.Errorf("Expected 10 items on page 1, got %d", len(result1.Data))
	}

	// Test second page
	result2, err := handler.DataQuery(appointments, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error on page 2: %v", err)
	}

	if len(result2.Data) != 10 {
		t.Errorf("Expected 10 items on page 2, got %d", len(result2.Data))
	}

	// Verify different data on different pages
	if result1.Data[0].ID == result2.Data[0].ID {
		t.Error("Page 1 and Page 2 should have different data")
	}
}

// TestDateBeforeAndAfter tests combining before and after filters
func TestDateBeforeAndAfter(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{ID: 1, Title: "Jan", StartDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Mar", StartDate: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "May", StartDate: time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Title: "Jul", StartDate: time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Title: "Sep", StartDate: time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 6, Title: "Nov", StartDate: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)},
	}

	// Find appointments between March and September (exclusive)
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "start_date",
				Value:    "2025-03-01",
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeDate,
			},
			{
				Field:    "start_date",
				Value:    "2025-10-01",
				Mode:     filter.ModeBefore,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should include March 15, May 15, July 15, Sep 15
	if result.TotalSize != 4 {
		t.Errorf("Expected 4 appointments between March and Sept, got %d", result.TotalSize)
	}
}

// TestLeapYearDate tests filtering on February 29th
func TestLeapYearDate(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{ID: 1, Title: "Feb28-2024", StartDate: time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Feb29-2024", StartDate: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "Mar1-2024", StartDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Title: "Feb28-2025", StartDate: time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "start_date",
				Value:    "2024-02-29",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering leap day: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 appointment on leap day, got %d", result.TotalSize)
	}
}

// TestDateTimeWithDifferentTimezones tests datetime filtering across timezones

// TestEmptyDateTimeValues tests filtering with nil/zero datetime values

// TestDateTimeComplexCombination tests complex datetime filter combinations
func TestDateTimeComplexCombination(t *testing.T) {
	handler := filter.NewFilter[Appointment](filter.GolangFilteringConfig{})

	appointments := []*Appointment{
		{
			ID:        1,
			Title:     "Morning Early Nov",
			StartDate: time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
			Duration:  60,
		},
		{
			ID:        2,
			Title:     "Afternoon Mid Nov",
			StartDate: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC),
			Duration:  90,
		},
		{
			ID:        3,
			Title:     "Evening Late Nov",
			StartDate: time.Date(2025, 11, 25, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC),
			Duration:  120,
		},
		{
			ID:        4,
			Title:     "Morning Mid Nov",
			StartDate: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
			StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
			Duration:  60,
		},
	}

	// Find appointments between Nov 10-20 that start before noon
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "start_date",
				Value: filter.Range{
					From: "2025-11-10",
					To:   "2025-11-20",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
			{
				Field:    "start_time",
				Value:    "12:00:00",
				Mode:     filter.ModeLT,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(appointments, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should only match appointment 4 (Nov 15 at 9am)
	if result.TotalSize != 1 {
		t.Errorf("Expected 1 morning appointment mid-Nov, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 && result.Data[0].ID != 4 {
		t.Error("Wrong appointment matched")
	}
}
