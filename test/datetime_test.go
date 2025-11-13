package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// Event model for datetime testing
type Event struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	EventDate   time.Time `json:"event_date"`
	EventTime   time.Time `json:"event_time"`
	CreatedAt   time.Time `json:"created_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

// TestDateTimeFormats tests various datetime format parsing
func TestDateTimeFormats(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	now := time.Date(2025, 11, 3, 14, 30, 45, 0, time.UTC)
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	events := []*Event{
		{ID: 1, Name: "Event1", CreatedAt: now},
		{ID: 2, Name: "Event2", CreatedAt: yesterday},
		{ID: 3, Name: "Event3", CreatedAt: tomorrow},
	}

	testCases := []struct {
		name        string
		dateString  string
		shouldParse bool
	}{
		{"RFC3339", "2025-11-03T14:30:45Z", true},
		{"RFC3339Nano", "2025-11-03T14:30:45.123456789Z", true},
		{"ISO with timezone", "2025-11-03T14:30:45-07:00", true},
		{"Space separator", "2025-11-03 14:30:45", true},
		{"Slash separator", "2025/11/03 14:30:45", true},
		{"Date only ISO", "2025-11-03", true},
		{"Date only slash", "2025/11/03", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "created_at",
						Value:    tc.dateString,
						Mode:     filter.ModeEqual,
						DataType: filter.DataTypeDate,
					},
				},
			}

			_, err := handler.DataQuery(events, filterRoot, 0, 10)
			if tc.shouldParse && err != nil {
				t.Errorf("Expected format %s to parse, got error: %v", tc.name, err)
			}
		})
	}
}

// TestTimeFormats tests various time format parsing
func TestTimeFormats(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	morning := time.Date(0, 1, 1, 9, 30, 0, 0, time.UTC)
	afternoon := time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)
	evening := time.Date(0, 1, 1, 18, 45, 30, 0, time.UTC)

	events := []*Event{
		{ID: 1, Name: "Morning", EventTime: morning},
		{ID: 2, Name: "Afternoon", EventTime: afternoon},
		{ID: 3, Name: "Evening", EventTime: evening},
	}

	testCases := []struct {
		name        string
		timeString  string
		shouldParse bool
	}{
		{"24h HH:MM:SS", "14:30:00", true},
		{"24h HH:MM", "14:30", true},
		{"Kitchen", "2:30PM", true},
		{"12h seconds", "2:30:00 PM", true},
		{"Nanoseconds", "14:30:00.123456789", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "event_time",
						Value:    tc.timeString,
						Mode:     filter.ModeEqual,
						DataType: filter.DataTypeTime,
					},
				},
			}

			_, err := handler.DataQuery(events, filterRoot, 0, 10)
			if tc.shouldParse && err != nil {
				t.Errorf("Expected time format %s to parse, got error: %v", tc.name, err)
			}
		})
	}
}

// TestDateModeEqual tests date equality filtering
func TestDateModeEqual(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	targetDate := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	events := []*Event{
		{ID: 1, Name: "Event1", EventDate: targetDate},
		{ID: 2, Name: "Event2", EventDate: targetDate.AddDate(0, 0, -1)},
		{ID: 3, Name: "Event3", EventDate: targetDate},
		{ID: 4, Name: "Event4", EventDate: targetDate.AddDate(0, 0, 1)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_date",
				Value:    "2025-11-03",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results, got %d", result.TotalSize)
	}
}

// TestDateModeBefore tests date before filtering
func TestDateModeBefore(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	baseDate := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	events := []*Event{
		{ID: 1, Name: "Past1", EventDate: baseDate.AddDate(0, 0, -5)},
		{ID: 2, Name: "Past2", EventDate: baseDate.AddDate(0, 0, -2)},
		{ID: 3, Name: "Today", EventDate: baseDate},
		{ID: 4, Name: "Future", EventDate: baseDate.AddDate(0, 0, 2)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_date",
				Value:    "2025-11-03",
				Mode:     filter.ModeBefore,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results before date, got %d", result.TotalSize)
	}
}

// TestDateModeAfter tests date after filtering
func TestDateModeAfter(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	baseDate := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	events := []*Event{
		{ID: 1, Name: "Past", EventDate: baseDate.AddDate(0, 0, -2)},
		{ID: 2, Name: "Today", EventDate: baseDate},
		{ID: 3, Name: "Future1", EventDate: baseDate.AddDate(0, 0, 2)},
		{ID: 4, Name: "Future2", EventDate: baseDate.AddDate(0, 0, 5)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_date",
				Value:    "2025-11-03",
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results after date, got %d", result.TotalSize)
	}
}

// TestDateModeGTE tests date greater than or equal
func TestDateModeGTE(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	baseDate := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	events := []*Event{
		{ID: 1, Name: "Past", EventDate: baseDate.AddDate(0, 0, -2)},
		{ID: 2, Name: "Today", EventDate: baseDate},
		{ID: 3, Name: "Future1", EventDate: baseDate.AddDate(0, 0, 2)},
		{ID: 4, Name: "Future2", EventDate: baseDate.AddDate(0, 0, 5)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_date",
				Value:    "2025-11-03",
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 results (including today), got %d", result.TotalSize)
	}
}

// TestDateModeLTE tests date less than or equal
func TestDateModeLTE(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	baseDate := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	events := []*Event{
		{ID: 1, Name: "Past1", EventDate: baseDate.AddDate(0, 0, -5)},
		{ID: 2, Name: "Past2", EventDate: baseDate.AddDate(0, 0, -2)},
		{ID: 3, Name: "Today", EventDate: baseDate},
		{ID: 4, Name: "Future", EventDate: baseDate.AddDate(0, 0, 2)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_date",
				Value:    "2025-11-03",
				Mode:     filter.ModeLTE,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 results (including today), got %d", result.TotalSize)
	}
}

// TestDateModeRange tests date range filtering
func TestDateModeRange(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Oct", EventDate: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Nov1", EventDate: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Nov5", EventDate: time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Nov10", EventDate: time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "Nov30", EventDate: time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Dec", EventDate: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_date",
				Value: filter.Range{
					From: "2025-11-01",
					To:   "2025-11-30",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 results in November range, got %d", result.TotalSize)
	}
}

// TestTimeModeEqual tests time equality
func TestTimeModeEqual(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	targetTime := time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)
	events := []*Event{
		{ID: 1, Name: "Morning", EventTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Target1", EventTime: targetTime},
		{ID: 3, Name: "Target2", EventTime: targetTime},
		{ID: 4, Name: "Evening", EventTime: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "14:30:00",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results, got %d", result.TotalSize)
	}
}

// TestTimeModeBefore tests time before filtering
func TestTimeModeBefore(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Early", EventTime: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Morning", EventTime: time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Noon", EventTime: time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Afternoon", EventTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "12:00:00",
				Mode:     filter.ModeBefore,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results before noon, got %d", result.TotalSize)
	}
}

// TestTimeModeAfter tests time after filtering
func TestTimeModeAfter(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Morning", EventTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Noon", EventTime: time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Afternoon", EventTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Evening", EventTime: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "13:00:00",
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results after 13:00, got %d", result.TotalSize)
	}
} // TestTimeModeRange tests time range filtering
func TestTimeModeRange(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Early", EventTime: time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "WorkStart", EventTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Lunch", EventTime: time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Afternoon", EventTime: time.Date(0, 1, 1, 15, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "WorkEnd", EventTime: time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Late", EventTime: time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_time",
				Value: filter.Range{
					From: "09:00:00",
					To:   "17:00:00",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 results in work hours, got %d", result.TotalSize)
	}
}

// TestDateTimeEdgeCases tests edge cases
func TestDateTimeEdgeCases(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	midnight := time.Date(2025, 11, 3, 0, 0, 0, 0, time.UTC)
	morning := time.Date(2025, 11, 4, 10, 30, 0, 0, time.UTC)
	leapDay := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	newYear := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	events := []*Event{
		{ID: 1, Name: "Midnight", CreatedAt: midnight},
		{ID: 2, Name: "Morning", CreatedAt: morning},
		{ID: 3, Name: "LeapDay", CreatedAt: leapDay},
		{ID: 4, Name: "NewYear", CreatedAt: newYear},
	}

	testCases := []struct {
		name     string
		value    string
		expected int
	}{
		{"Midnight", "2025-11-03T00:00:00Z", 1},
		{"Morning", "2025-11-04T10:30:00Z", 1},
		{"LeapDay", "2024-02-29T00:00:00Z", 1},
		{"NewYear", "2026-01-01", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "created_at",
						Value:    tc.value,
						Mode:     filter.ModeEqual,
						DataType: filter.DataTypeDate,
					},
				},
			}

			result, err := handler.DataQuery(events, filterRoot, 0, 10)
			if err != nil {
				t.Fatalf("Error filtering %s: %v", tc.name, err)
			}

			if result.TotalSize != tc.expected {
				t.Errorf("%s: expected %d, got %d", tc.name, tc.expected, result.TotalSize)
			}
		})
	}
}

// TestMultipleDateTimeFilters tests combining multiple filters
func TestMultipleDateTimeFilters(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{
			ID:        1,
			Name:      "Event1",
			EventDate: time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC),
			EventTime: time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Name:      "Event2",
			EventDate: time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC),
			EventTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:        3,
			Name:      "Event3",
			EventDate: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
			EventTime: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC),
		},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_date",
				Value:    "2025-11-01",
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeDate,
			},
			{
				Field:    "event_time",
				Value:    "12:00:00",
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results, got %d", result.TotalSize)
	}
}

// TestDateTimeSorting tests datetime sorting
func TestDateTimeSorting(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Event1", CreatedAt: time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Event2", CreatedAt: time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Event3", CreatedAt: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		SortFields: []filter.SortField{
			{Field: "created_at", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error sorting: %v", err)
	}

	if len(result.Data) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result.Data))
	}

	if result.Data[0].ID != 2 || result.Data[1].ID != 1 || result.Data[2].ID != 3 {
		t.Error("Events not sorted correctly by date")
	}
}

// TestTimeWithNanoseconds tests nanosecond precision
func TestTimeWithNanoseconds(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	baseTime := time.Date(0, 1, 1, 14, 30, 45, 123456789, time.UTC)
	events := []*Event{
		{ID: 1, Name: "Precise", EventTime: baseTime},
		{ID: 2, Name: "NoNano", EventTime: time.Date(0, 1, 1, 14, 30, 45, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "14:30:45.123456789",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 result with nanoseconds, got %d", result.TotalSize)
	}
}

// TestInvalidDateTimeFormats tests error handling
func TestInvalidDateTimeFormats(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Event1", CreatedAt: time.Now()},
	}

	invalidFormats := []struct {
		name     string
		value    string
		dataType filter.DataType
	}{
		{"Invalid date", "2025-13-45", filter.DataTypeDate},
		{"Invalid time", "25:99:99", filter.DataTypeTime},
		{"Garbage date", "not-a-date", filter.DataTypeDate},
		{"Garbage time", "not-a-time", filter.DataTypeTime},
	}

	for _, tc := range invalidFormats {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "created_at",
						Value:    tc.value,
						Mode:     filter.ModeEqual,
						DataType: tc.dataType,
					},
				},
			}

			_, err := handler.DataQuery(events, filterRoot, 0, 10)
			if err == nil {
				t.Errorf("Expected error for %s, got none", tc.name)
			}
		})
	}
}

// TestDateRangeComprehensive tests various date range scenarios
func TestDateRangeComprehensive(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	// Create events spanning across months
	events := []*Event{
		{ID: 1, Name: "Jan1", EventDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Jan15", EventDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Jan31", EventDate: time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Feb1", EventDate: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "Feb14", EventDate: time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Feb28", EventDate: time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)},
		{ID: 7, Name: "Mar1", EventDate: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 8, Name: "Mar15", EventDate: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 9, Name: "Mar31", EventDate: time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC)},
		{ID: 10, Name: "Apr1", EventDate: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
	}

	testCases := []struct {
		name     string
		from     string
		to       string
		expected int
		desc     string
	}{
		{"Single month", "2025-02-01", "2025-02-28", 3, "All of February"},
		{"Two months", "2025-01-15", "2025-02-15", 4, "Mid-Jan to mid-Feb"},
		{"Quarter", "2025-01-01", "2025-03-31", 9, "Q1 2025"},
		{"Single day", "2025-02-14", "2025-02-14", 1, "Valentine's Day only"},
		{"Boundary inclusive", "2025-01-01", "2025-01-31", 3, "All of January"},
		{"Cross month boundary", "2025-01-31", "2025-02-01", 2, "Month transition"},
		{"Full range", "2025-01-01", "2025-04-01", 10, "All events"},
		{"Empty range before", "2024-11-01", "2024-12-31", 0, "Before all events"},
		{"Empty range after", "2025-05-01", "2025-06-30", 0, "After all events"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field: "event_date",
						Value: filter.Range{
							From: tc.from,
							To:   tc.to,
						},
						Mode:     filter.ModeRange,
						DataType: filter.DataTypeDate,
					},
				},
			}

			result, err := handler.DataQuery(events, filterRoot, 0, 20)
			if err != nil {
				t.Fatalf("Error filtering %s: %v", tc.name, err)
			}

			if result.TotalSize != tc.expected {
				t.Errorf("%s (%s): expected %d results, got %d",
					tc.name, tc.desc, tc.expected, result.TotalSize)
			}
		})
	}
}

// TestTimeRangeComprehensive tests various time range scenarios
func TestTimeRangeComprehensive(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	// Create events throughout a day
	events := []*Event{
		{ID: 1, Name: "Midnight", EventTime: time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "1AM", EventTime: time.Date(0, 1, 1, 1, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "3AM", EventTime: time.Date(0, 1, 1, 3, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "6AM", EventTime: time.Date(0, 1, 1, 6, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "9AM", EventTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Noon", EventTime: time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 7, Name: "3PM", EventTime: time.Date(0, 1, 1, 15, 0, 0, 0, time.UTC)},
		{ID: 8, Name: "6PM", EventTime: time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)},
		{ID: 9, Name: "9PM", EventTime: time.Date(0, 1, 1, 21, 0, 0, 0, time.UTC)},
		{ID: 10, Name: "11PM", EventTime: time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC)},
	}

	testCases := []struct {
		name     string
		from     string
		to       string
		expected int
		desc     string
	}{
		{"Work hours", "09:00:00", "17:00:00", 3, "9AM to 5PM inclusive"},
		{"Morning", "06:00:00", "12:00:00", 3, "6AM to noon inclusive"},
		{"Afternoon", "12:00:00", "18:00:00", 3, "Noon to 6PM inclusive"},
		{"Evening", "18:00:00", "23:59:59", 3, "6PM to midnight"},
		{"Night hours", "00:00:00", "06:00:00", 4, "Midnight to 6AM"},
		{"Single hour", "12:00:00", "12:59:59", 1, "Noon hour"},
		{"Full day", "00:00:00", "23:59:59", 10, "All day"},
		{"Lunch time", "11:00:00", "13:00:00", 1, "11AM to 1PM"},
		{"Late night", "21:00:00", "23:59:59", 2, "9PM to midnight"},
		{"Early morning", "00:00:00", "03:00:00", 3, "Midnight to 3AM"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field: "event_time",
						Value: filter.Range{
							From: tc.from,
							To:   tc.to,
						},
						Mode:     filter.ModeRange,
						DataType: filter.DataTypeTime,
					},
				},
			}

			result, err := handler.DataQuery(events, filterRoot, 0, 20)
			if err != nil {
				t.Fatalf("Error filtering %s: %v", tc.name, err)
			}

			if result.TotalSize != tc.expected {
				t.Errorf("%s (%s): expected %d results, got %d",
					tc.name, tc.desc, tc.expected, result.TotalSize)
			}
		})
	}
}

// TestDateRangeWithTimezone tests date ranges with different timezones
func TestDateRangeWithTimezone(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	// Create events in different timezones
	utc := time.UTC
	est := time.FixedZone("EST", -5*3600)
	pst := time.FixedZone("PST", -8*3600)

	events := []*Event{
		{ID: 1, Name: "UTC", EventDate: time.Date(2025, 3, 1, 12, 0, 0, 0, utc)},
		{ID: 2, Name: "EST", EventDate: time.Date(2025, 3, 1, 12, 0, 0, 0, est)},
		{ID: 3, Name: "PST", EventDate: time.Date(2025, 3, 1, 12, 0, 0, 0, pst)},
		{ID: 4, Name: "UTC2", EventDate: time.Date(2025, 3, 15, 12, 0, 0, 0, utc)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_date",
				Value: filter.Range{
					From: "2025-03-01",
					To:   "2025-03-31",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// All events should be in March regardless of timezone
	if result.TotalSize < 4 {
		t.Errorf("Expected at least 4 results in March, got %d", result.TotalSize)
	}
}

// TestTimeRangeWithMinutesSeconds tests precise time ranges
func TestTimeRangeWithMinutesSeconds(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "14:29:00", EventTime: time.Date(0, 1, 1, 14, 29, 0, 0, time.UTC)},
		{ID: 2, Name: "14:30:00", EventTime: time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)},
		{ID: 3, Name: "14:30:30", EventTime: time.Date(0, 1, 1, 14, 30, 30, 0, time.UTC)},
		{ID: 4, Name: "14:31:00", EventTime: time.Date(0, 1, 1, 14, 31, 0, 0, time.UTC)},
		{ID: 5, Name: "14:35:00", EventTime: time.Date(0, 1, 1, 14, 35, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_time",
				Value: filter.Range{
					From: "14:30:00",
					To:   "14:31:00",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 results in 14:30-14:31 range, got %d", result.TotalSize)
	}
}

// TestDateRangeYearBoundary tests date ranges crossing year boundary
func TestDateRangeYearBoundary(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Dec20", EventDate: time.Date(2024, 12, 20, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Dec25", EventDate: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Dec31", EventDate: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Jan1", EventDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "Jan5", EventDate: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Jan10", EventDate: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_date",
				Value: filter.Range{
					From: "2024-12-25",
					To:   "2025-01-05",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 results crossing year boundary, got %d", result.TotalSize)
	}
}

// TestTimeRangeMidnightCrossing tests time ranges around midnight
func TestTimeRangeMidnightCrossing(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "22:00", EventTime: time.Date(0, 1, 1, 22, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "23:00", EventTime: time.Date(0, 1, 1, 23, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "23:30", EventTime: time.Date(0, 1, 1, 23, 30, 0, 0, time.UTC)},
		{ID: 4, Name: "00:00", EventTime: time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "01:00", EventTime: time.Date(0, 1, 1, 1, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "02:00", EventTime: time.Date(0, 1, 1, 2, 0, 0, 0, time.UTC)},
	}

	// Test late evening (cannot cross midnight in a single range)
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_time",
				Value: filter.Range{
					From: "22:00:00",
					To:   "23:59:59",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should include events from 22:00 through 23:30
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 results in late evening range, got %d", result.TotalSize)
	}

	// Test early morning
	filterRoot2 := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_time",
				Value: filter.Range{
					From: "00:00:00",
					To:   "02:00:00",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result2, err := handler.DataQuery(events, filterRoot2, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering early morning: %v", err)
	}

	// Should include midnight, 1AM, and 2AM
	if result2.TotalSize != 3 {
		t.Errorf("Expected 3 results in early morning range, got %d", result2.TotalSize)
	}
}

// TestDateRangeInvalidRange tests error handling for invalid date ranges
func TestDateRangeInvalidRange(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Event", EventDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
	}

	// From date after To date
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_date",
				Value: filter.Range{
					From: "2025-12-31",
					To:   "2025-01-01",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	_, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err == nil {
		t.Error("Expected error for invalid date range (from > to), got none")
	}
}

// TestTimeRangeInvalidRange tests error handling for invalid time ranges
func TestTimeRangeInvalidRange(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Event", EventTime: time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC)},
	}

	// From time after To time
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_time",
				Value: filter.Range{
					From: "18:00:00",
					To:   "09:00:00",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	_, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err == nil {
		t.Error("Expected error for invalid time range (from > to), got none")
	}
}

// TestDateRangeWithSorting tests date range with sorting
func TestDateRangeWithSorting(t *testing.T) {
	handler := filter.NewFilter[Event](filter.GolangFilteringConfig{})

	events := []*Event{
		{ID: 1, Name: "Event3", EventDate: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Event1", EventDate: time.Date(2025, 3, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Event2", EventDate: time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Event4", EventDate: time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)},
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_date",
				Value: filter.Range{
					From: "2025-03-01",
					To:   "2025-03-31",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
		SortFields: []filter.SortField{
			{Field: "event_date", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 0, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 results, got %d", result.TotalSize)
	}

	// Verify ascending order
	if len(result.Data) >= 2 {
		if result.Data[0].EventDate.After(result.Data[1].EventDate) {
			t.Error("Results not sorted in ascending order")
		}
	}
}
