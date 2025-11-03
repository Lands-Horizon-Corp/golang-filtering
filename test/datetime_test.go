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
	handler := filter.NewFilter[Event]()

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

			_, err := handler.DataQuery(events, filterRoot, 1, 10)
			if tc.shouldParse && err != nil {
				t.Errorf("Expected format %s to parse, got error: %v", tc.name, err)
			}
		})
	}
}

// TestTimeFormats tests various time format parsing
func TestTimeFormats(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

			_, err := handler.DataQuery(events, filterRoot, 1, 10)
			if tc.shouldParse && err != nil {
				t.Errorf("Expected time format %s to parse, got error: %v", tc.name, err)
			}
		})
	}
}

// TestDateModeEqual tests date equality filtering
func TestDateModeEqual(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results, got %d", result.TotalSize)
	}
}

// TestDateModeBefore tests date before filtering
func TestDateModeBefore(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results before date, got %d", result.TotalSize)
	}
}

// TestDateModeAfter tests date after filtering
func TestDateModeAfter(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results after date, got %d", result.TotalSize)
	}
}

// TestDateModeGTE tests date greater than or equal
func TestDateModeGTE(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 results (including today), got %d", result.TotalSize)
	}
}

// TestDateModeLTE tests date less than or equal
func TestDateModeLTE(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 results (including today), got %d", result.TotalSize)
	}
}

// TestDateModeRange tests date range filtering
func TestDateModeRange(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 results in November range, got %d", result.TotalSize)
	}
}

// TestTimeModeEqual tests time equality
func TestTimeModeEqual(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results, got %d", result.TotalSize)
	}
}

// TestTimeModeBefore tests time before filtering
func TestTimeModeBefore(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results before noon, got %d", result.TotalSize)
	}
}

// TestTimeModeAfter tests time after filtering
func TestTimeModeAfter(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results after 13:00, got %d", result.TotalSize)
	}
} // TestTimeModeRange tests time range filtering
func TestTimeModeRange(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 4 {
		t.Errorf("Expected 4 results in work hours, got %d", result.TotalSize)
	}
}

// TestDateTimeEdgeCases tests edge cases
func TestDateTimeEdgeCases(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

			result, err := handler.DataQuery(events, filterRoot, 1, 10)
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
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 results, got %d", result.TotalSize)
	}
}

// TestDateTimeSorting tests datetime sorting
func TestDateTimeSorting(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
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
	handler := filter.NewFilter[Event]()

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

	result, err := handler.DataQuery(events, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 result with nanoseconds, got %d", result.TotalSize)
	}
}

// TestInvalidDateTimeFormats tests error handling
func TestInvalidDateTimeFormats(t *testing.T) {
	handler := filter.NewFilter[Event]()

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

			_, err := handler.DataQuery(events, filterRoot, 1, 10)
			if err == nil {
				t.Errorf("Expected error for %s, got none", tc.name)
			}
		})
	}
}
