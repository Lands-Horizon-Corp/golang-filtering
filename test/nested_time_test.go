package test

import (
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TimeOfDay is a custom type that stores time in SQLite as TEXT
type TimeOfDay struct {
	time.Time
}

// Scan implements sql.Scanner for reading from database
func (t *TimeOfDay) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("TimeOfDay: cannot scan type %T into TimeOfDay", value)
	}

	parsed, err := time.Parse("15:04:05", str)
	if err != nil {
		return fmt.Errorf("TimeOfDay: cannot parse %s: %w", str, err)
	}

	t.Time = parsed
	return nil
}

// Value implements driver.Valuer for writing to database
func (t TimeOfDay) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time.Format("15:04:05"), nil
}

// WorkShift represents a work shift schedule
type WorkShift struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	StartTime TimeOfDay `gorm:"type:varchar(8)" json:"start_time"` // Stored as HH:MM:SS in SQLite
	EndTime   TimeOfDay `gorm:"type:varchar(8)" json:"end_time"`   // Stored as HH:MM:SS in SQLite
}

// EmployeeAttendance represents attendance with nested work_shift
// NOTE: WorkShift is a pointer - DataQuery nested filtering has limitations with pointer fields
// Use GORM (DataGorm) for reliable nested filtering on pointer-based relationships
type EmployeeAttendance struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	EmployeeName string     `gorm:"type:varchar(255)" json:"employee_name"`
	Date         time.Time  `gorm:"type:date" json:"date"`
	WorkShiftID  uint       `gorm:"not null" json:"work_shift_id"`
	WorkShift    *WorkShift `gorm:"foreignKey:WorkShiftID" json:"work_shift,omitempty"`
	CheckIn      string     `gorm:"type:varchar(8)" json:"check_in"`  // HH:MM:SS format
	CheckOut     string     `gorm:"type:varchar(8)" json:"check_out"` // HH:MM:SS format
}

func setupNestedTimeDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&WorkShift{}, &EmployeeAttendance{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create work shifts with different times
	parseTime := func(timeStr string) TimeOfDay {
		t, _ := time.Parse("15:04:05", timeStr)
		return TimeOfDay{Time: t}
	}

	shifts := []WorkShift{
		{
			ID:        1,
			Name:      "Morning Shift",
			StartTime: parseTime("08:00:00"),
			EndTime:   parseTime("16:00:00"),
		},
		{
			ID:        2,
			Name:      "Afternoon Shift",
			StartTime: parseTime("14:00:00"),
			EndTime:   parseTime("22:00:00"),
		},
		{
			ID:        3,
			Name:      "Night Shift",
			StartTime: parseTime("22:00:00"),
			EndTime:   parseTime("06:00:00"),
		},
		{
			ID:        4,
			Name:      "Early Morning",
			StartTime: parseTime("06:00:00"),
			EndTime:   parseTime("14:00:00"),
		},
	}

	for _, shift := range shifts {
		if err := db.Create(&shift).Error; err != nil {
			t.Fatalf("Failed to create shift: %v", err)
		}
	}

	// Create attendance records
	today := time.Now()
	attendances := []EmployeeAttendance{
		{
			ID:           1,
			EmployeeName: "Alice",
			Date:         today,
			WorkShiftID:  1,
			CheckIn:      "08:15:00",
			CheckOut:     "16:30:00",
		},
		{
			ID:           2,
			EmployeeName: "Bob",
			Date:         today,
			WorkShiftID:  2,
			CheckIn:      "14:10:00",
			CheckOut:     "22:05:00",
		},
		{
			ID:           3,
			EmployeeName: "Charlie",
			Date:         today,
			WorkShiftID:  3,
			CheckIn:      "22:00:00",
			CheckOut:     "06:15:00",
		},
		{
			ID:           4,
			EmployeeName: "Diana",
			Date:         today,
			WorkShiftID:  1,
			CheckIn:      "08:00:00",
			CheckOut:     "16:00:00",
		},
		{
			ID:           5,
			EmployeeName: "Eve",
			Date:         today,
			WorkShiftID:  4,
			CheckIn:      "06:05:00",
			CheckOut:     "14:10:00",
		},
	}

	for _, attendance := range attendances {
		if err := db.Create(&attendance).Error; err != nil {
			t.Fatalf("Failed to create attendance: %v", err)
		}
	}

	return db
}

// ============ GORM Tests ============

// TestGormNestedTimeEqual tests filtering by exact time on nested relation
func TestGormNestedTimeEqual(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "08:00:00",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeTime,
			},
		},
		Preload: []string{"WorkShift"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time equal: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 attendances with 08:00 start time, got %d", result.TotalSize)
	}

	for _, att := range result.Data {
		if att.WorkShift == nil {
			t.Error("WorkShift was not preloaded")
			continue
		}
		t.Logf("Found: %s - Shift: %s (Start: %s)", att.EmployeeName, att.WorkShift.Name, att.WorkShift.StartTime)
	}
}

// TestGormNestedTimeAfter tests filtering by time after on nested relation
func TestGormNestedTimeAfter(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "15:00:00",
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeTime,
			},
		},
		Preload: []string{"WorkShift"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time after: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 attendance with start time after 15:00, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		att := result.Data[0]
		if att.WorkShift.Name != "Night Shift" {
			t.Errorf("Expected Night Shift, got %s", att.WorkShift.Name)
		}
	}
}

// TestGormNestedTimeBefore tests filtering by time before on nested relation
func TestGormNestedTimeBefore(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "10:00:00",
				Mode:     filter.ModeLT,
				DataType: filter.DataTypeTime,
			},
		},
		Preload: []string{"WorkShift"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time before: %v", err)
	}

	// Should match Morning Shift (08:00) and Early Morning (06:00)
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 attendances with start time before 10:00, got %d", result.TotalSize)
	}

	for _, att := range result.Data {
		if att.WorkShift != nil {
			t.Logf("Found: %s - Shift: %s", att.EmployeeName, att.WorkShift.Name)
		}
	}
}

// TestGormNestedTimeRange tests time range filtering on nested relation
func TestGormNestedTimeRange(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "work_shift.start_time",
				Value: filter.Range{
					From: "06:00:00",
					To:   "14:00:00",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
		Preload: []string{"WorkShift"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time range: %v", err)
	}

	// Should match Early Morning (06:00), Morning Shift (08:00), Afternoon Shift (14:00)
	if result.TotalSize != 4 {
		t.Errorf("Expected 4 attendances with start time between 06:00 and 14:00, got %d", result.TotalSize)
	}
}

// TestGormNestedTimeSorting tests sorting by nested time field
func TestGormNestedTimeSorting(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "work_shift.start_time", Order: filter.SortOrderAsc},
		},
		Preload: []string{"WorkShift"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to sort by work_shift.start_time: %v", err)
	}

	if result.TotalSize != 5 {
		t.Errorf("Expected 5 total attendances, got %d", result.TotalSize)
	}

	// Verify sorted by start_time: 06:00, 08:00, 08:00, 14:00, 22:00
	expectedShifts := []string{"Early Morning", "Morning Shift", "Morning Shift", "Afternoon Shift", "Night Shift"}
	for i, att := range result.Data {
		if att.WorkShift == nil {
			t.Error("WorkShift was not preloaded")
			continue
		}
		if att.WorkShift.Name != expectedShifts[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expectedShifts[i], att.WorkShift.Name)
		}
		t.Logf("Position %d: %s - %s (Start: %s)", i, att.EmployeeName, att.WorkShift.Name, att.WorkShift.StartTime)
	}
}

// TestGormNestedTimeMultipleFilters tests combining nested time with other filters
func TestGormNestedTimeMultipleFilters(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "08:00:00",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeTime,
			},
			{
				Field:    "employee_name",
				Value:    "Alice",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"WorkShift"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter with multiple conditions: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 attendance (Alice at 08:00), got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		att := result.Data[0]
		if att.EmployeeName != "Alice" {
			t.Errorf("Expected employee Alice, got %s", att.EmployeeName)
		}
	}
}

// ============ DataQuery Tests ============

func setupNestedTimeDataQuery(_ *testing.T) []*EmployeeAttendance {
	parseTime := func(timeStr string) TimeOfDay {
		t, _ := time.Parse("15:04:05", timeStr)
		return TimeOfDay{Time: t}
	}

	shifts := []*WorkShift{
		{
			ID:        1,
			Name:      "Morning Shift",
			StartTime: parseTime("08:00:00"),
			EndTime:   parseTime("16:00:00"),
		},
		{
			ID:        2,
			Name:      "Afternoon Shift",
			StartTime: parseTime("14:00:00"),
			EndTime:   parseTime("22:00:00"),
		},
		{
			ID:        3,
			Name:      "Night Shift",
			StartTime: parseTime("22:00:00"),
			EndTime:   parseTime("06:00:00"),
		},
		{
			ID:        4,
			Name:      "Early Morning",
			StartTime: parseTime("06:00:00"),
			EndTime:   parseTime("14:00:00"),
		},
	}

	today := time.Now()
	return []*EmployeeAttendance{
		{
			ID:           1,
			EmployeeName: "Alice",
			Date:         today,
			WorkShiftID:  1,
			WorkShift:    shifts[0],
			CheckIn:      "08:15:00",
			CheckOut:     "16:30:00",
		},
		{
			ID:           2,
			EmployeeName: "Bob",
			Date:         today,
			WorkShiftID:  2,
			WorkShift:    shifts[1],
			CheckIn:      "14:10:00",
			CheckOut:     "22:05:00",
		},
		{
			ID:           3,
			EmployeeName: "Charlie",
			Date:         today,
			WorkShiftID:  3,
			WorkShift:    shifts[2],
			CheckIn:      "22:00:00",
			CheckOut:     "06:15:00",
		},
		{
			ID:           4,
			EmployeeName: "Diana",
			Date:         today,
			WorkShiftID:  1,
			WorkShift:    shifts[0],
			CheckIn:      "08:00:00",
			CheckOut:     "16:00:00",
		},
		{
			ID:           5,
			EmployeeName: "Eve",
			Date:         today,
			WorkShiftID:  4,
			WorkShift:    shifts[3],
			CheckIn:      "06:05:00",
			CheckOut:     "14:10:00",
		},
	}
}

// TestDataQueryNestedTimeEqual tests in-memory filtering by exact time on nested relation
func TestDataQueryNestedTimeEqual(t *testing.T) {
	data := setupNestedTimeDataQuery(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "08:00:00",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(data, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time equal: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 attendances with 08:00 start time, got %d", result.TotalSize)
	}

	for _, att := range result.Data {
		expectedTime, _ := time.Parse("15:04:05", "08:00:00")
		if !att.WorkShift.StartTime.Equal(expectedTime) {
			t.Errorf("Expected start time 08:00:00, got %s", att.WorkShift.StartTime.Format("15:04:05"))
		}
	}
}

// TestDataQueryNestedTimeAfter tests in-memory filtering by time after
func TestDataQueryNestedTimeAfter(t *testing.T) {
	data := setupNestedTimeDataQuery(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "15:00:00",
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(data, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time after: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 attendance with start time after 15:00, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		if result.Data[0].WorkShift.Name != "Night Shift" {
			t.Errorf("Expected Night Shift, got %s", result.Data[0].WorkShift.Name)
		}
	}
}

// TestDataQueryNestedTimeBefore tests in-memory filtering by time before
func TestDataQueryNestedTimeBefore(t *testing.T) {
	data := setupNestedTimeDataQuery(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "10:00:00",
				Mode:     filter.ModeBefore,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(data, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time before: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 attendances with start time before 10:00, got %d", result.TotalSize)
	}
}

// TestDataQueryNestedTimeRange tests in-memory time range filtering
func TestDataQueryNestedTimeRange(t *testing.T) {
	data := setupNestedTimeDataQuery(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "work_shift.start_time",
				Value: filter.Range{
					From: "06:00:00",
					To:   "14:00:00",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(data, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by work_shift.start_time range: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 attendances with start time between 06:00 and 14:00, got %d", result.TotalSize)
	}
}

// TestDataQueryNestedTimeSorting tests in-memory sorting by nested time
func TestDataQueryNestedTimeSorting(t *testing.T) {
	data := setupNestedTimeDataQuery(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "work_shift.start_time", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataQuery(data, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to sort by work_shift.start_time: %v", err)
	}

	expectedShifts := []string{"Early Morning", "Morning Shift", "Morning Shift", "Afternoon Shift", "Night Shift"}
	for i, att := range result.Data {
		if att.WorkShift.Name != expectedShifts[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expectedShifts[i], att.WorkShift.Name)
		}
	}
}

// ============ Hybrid Tests ============

// TestHybridNestedTimeSmallDataset tests hybrid filtering with small dataset (uses DataQuery)
func TestHybridNestedTimeSmallDataset(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "08:00:00",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeTime,
			},
		},
		Preload: []string{"WorkShift"},
	}

	// Large threshold to force DataQuery path
	result, err := handler.Hybrid(db, 1000, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed hybrid filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 attendances, got %d", result.TotalSize)
	}

	for _, att := range result.Data {
		if att.WorkShift == nil {
			t.Error("WorkShift was not preloaded")
			continue
		}
		t.Logf("Hybrid (DataQuery): %s - %s", att.EmployeeName, att.WorkShift.Name)
	}
}

// TestHybridNestedTimeLargeThreshold tests hybrid with large threshold (uses DataGorm)
func TestHybridNestedTimeLargeThreshold(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "work_shift.start_time",
				Value:    "15:00:00",
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeTime,
			},
		},
		Preload: []string{"WorkShift"},
	}

	// Small threshold to force DataGorm path
	result, err := handler.Hybrid(db, 2, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed hybrid filtering: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 attendance, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		if result.Data[0].WorkShift.Name != "Night Shift" {
			t.Errorf("Expected Night Shift, got %s", result.Data[0].WorkShift.Name)
		}
		t.Logf("Hybrid (DataGorm): %s - %s", result.Data[0].EmployeeName, result.Data[0].WorkShift.Name)
	}
}

// TestHybridNestedTimeConsistency tests that hybrid returns same results as GORM and DataQuery
func TestHybridNestedTimeConsistency(t *testing.T) {
	db := setupNestedTimeDB(t)
	handler := filter.NewFilter[EmployeeAttendance](filter.GolangFilteringConfig{})

	testCases := []struct {
		name      string
		fieldName string
		value     any
		mode      filter.Mode
		dataType  filter.DataType
	}{
		{"Equal", "work_shift.start_time", "08:00:00", filter.ModeEqual, filter.DataTypeTime},
		{"After", "work_shift.start_time", "12:00:00", filter.ModeGT, filter.DataTypeTime},
		{"Before", "work_shift.start_time", "10:00:00", filter.ModeLT, filter.DataTypeTime},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    tc.fieldName,
						Value:    tc.value,
						Mode:     tc.mode,
						DataType: tc.dataType,
					},
				},
				Preload: []string{"WorkShift"},
			}

			// Test DataGorm
			gormResult, err := handler.DataGorm(db, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("DataGorm failed: %v", err)
			}

			// Test Hybrid (should use DataGorm path)
			hybridResult, err := handler.Hybrid(db, 2, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("Hybrid failed: %v", err)
			}

			if gormResult.TotalSize != hybridResult.TotalSize {
				t.Errorf("Inconsistent results: GORM=%d, Hybrid=%d", gormResult.TotalSize, hybridResult.TotalSize)
			}

			t.Logf("%s: GORM=%d, Hybrid=%d", tc.name, gormResult.TotalSize, hybridResult.TotalSize)
		})
	}
}
