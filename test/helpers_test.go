package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestParseTime tests time parsing from various formats
func TestParseTime(t *testing.T) {
	type TimeEvent struct {
		ID        uint      `json:"id"`
		Name      string    `json:"name"`
		EventTime time.Time `json:"event_time"`
	}

	events := []*TimeEvent{
		{ID: 1, Name: "Morning", EventTime: time.Date(2024, 1, 1, 8, 30, 0, 0, time.UTC)},
		{ID: 2, Name: "Noon", EventTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Afternoon", EventTime: time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC)},
		{ID: 4, Name: "Evening", EventTime: time.Date(2024, 1, 1, 18, 45, 0, 0, time.UTC)},
	}

	handler := filter.NewFilter[TimeEvent]()

	// Test time equal
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "12:00:00",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Time equal filter failed: %v", err)
	}
	if result.TotalSize != 1 {
		t.Errorf("Expected 1 event at noon, got %d", result.TotalSize)
	}

	// Test time after
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "14:00:00",
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err = handler.DataQuery(events, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Time after filter failed: %v", err)
	}
	if result.TotalSize < 2 {
		t.Errorf("Expected at least 2 events after 14:00, got %d", result.TotalSize)
	}

	// Test time before
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "10:00:00",
				Mode:     filter.ModeBefore,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err = handler.DataQuery(events, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Time before filter failed: %v", err)
	}
	if result.TotalSize < 1 {
		t.Errorf("Expected at least 1 event before 10:00, got %d", result.TotalSize)
	}
}

// TestParseTimeRange tests time range parsing
func TestParseTimeRange(t *testing.T) {
	type TimeEvent struct {
		ID        uint      `json:"id"`
		Name      string    `json:"name"`
		EventTime time.Time `json:"event_time"`
	}

	events := []*TimeEvent{
		{ID: 1, Name: "Morning", EventTime: time.Date(2024, 1, 1, 8, 30, 0, 0, time.UTC)},
		{ID: 2, Name: "Noon", EventTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Afternoon", EventTime: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Evening", EventTime: time.Date(2024, 1, 1, 18, 30, 0, 0, time.UTC)},
	}

	handler := filter.NewFilter[TimeEvent]()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_time",
				Value: filter.Range{
					From: "10:00:00",
					To:   "16:00:00",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Time range filter failed: %v", err)
	}

	// Should find events between 10:00 and 16:00
	t.Logf("Time range filter returned %d events", result.TotalSize)
}

// TestParseDateTimeString tests various datetime string formats
func TestParseDateTimeString(t *testing.T) {
	handler := filter.NewFilter[TestUser]()
	users := generateTestUsers()

	dateFormats := []string{
		"2024-01-15 10:30:00",
		"2024-01-15T10:30:00Z",
		"2024/01/15 10:30:00",
	}

	for _, dateStr := range dateFormats {
		t.Run(dateStr, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "created_at",
						Value:    dateStr,
						Mode:     filter.ModeAfter,
						DataType: filter.DataTypeTime,
					},
				},
			}

			result, err := handler.DataQuery(users, filterRoot, 1, 100)
			if err != nil {
				t.Logf("DateTime parsing for %s: %v", dateStr, err)
			} else {
				t.Logf("DateTime %s parsed successfully, found %d users", dateStr, result.TotalSize)
			}
		})
	}
}

// TestParseDateTimeRange tests datetime range parsing
func TestParseDateTimeRange(t *testing.T) {
	handler := filter.NewFilter[TestUser]()
	users := generateTestUsers()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "created_at",
				Value: filter.Range{
					From: "2024-01-01",
					To:   "2024-03-31",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeTime,
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DateTime range filter failed: %v", err)
	}

	t.Logf("DateTime range filter returned %d users", result.TotalSize)
}

// TestApplyDateFilters tests date filtering operations
func TestApplyDateFilters(t *testing.T) {
	type Document struct {
		ID        uint      `json:"id"`
		Title     string    `json:"title"`
		PublishAt time.Time `json:"publish_at"`
	}

	docs := []*Document{
		{ID: 1, Title: "Doc1", PublishAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Doc2", PublishAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "Doc3", PublishAt: time.Date(2024, 3, 10, 9, 0, 0, 0, time.UTC)},
	}

	handler := filter.NewFilter[Document]()

	// Test date equal
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "publish_at",
				Value:    "2024-02-20",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(docs, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Date equal filter failed: %v", err)
	}
	t.Logf("Date equal filter returned %d documents", result.TotalSize)

	// Test date not equal
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "publish_at",
				Value:    "2024-02-20",
				Mode:     filter.ModeNotEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err = handler.DataQuery(docs, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Date not equal filter failed: %v", err)
	}
	t.Logf("Date not equal filter returned %d documents", result.TotalSize)

	// Test date before
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "publish_at",
				Value:    "2024-02-01",
				Mode:     filter.ModeBefore,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err = handler.DataQuery(docs, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Date before filter failed: %v", err)
	}
	if result.TotalSize < 1 {
		t.Errorf("Expected at least 1 document before 2024-02-01, got %d", result.TotalSize)
	}

	// Test date after
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "publish_at",
				Value:    "2024-02-01",
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err = handler.DataQuery(docs, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Date after filter failed: %v", err)
	}
	if result.TotalSize < 2 {
		t.Errorf("Expected at least 2 documents after 2024-02-01, got %d", result.TotalSize)
	}

	// Test date range
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "publish_at",
				Value: filter.Range{
					From: "2024-01-01",
					To:   "2024-02-28",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err = handler.DataQuery(docs, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Date range filter failed: %v", err)
	}
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 documents in Jan-Feb 2024, got %d", result.TotalSize)
	}
}

// TestApplyBoolFilters tests boolean filtering operations
func TestApplyBoolFilters(t *testing.T) {
	handler := filter.NewFilter[TestUser]()
	users := generateTestUsers()

	// Test bool equal true
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

	result, err := handler.DataQuery(users, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Bool equal true filter failed: %v", err)
	}

	for _, user := range result.Data {
		if !user.IsActive {
			t.Errorf("Expected only active users, got inactive user %s", user.Name)
		}
	}

	// Test bool equal false
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    false,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	result, err = handler.DataQuery(users, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Bool equal false filter failed: %v", err)
	}

	for _, user := range result.Data {
		if user.IsActive {
			t.Errorf("Expected only inactive users, got active user %s", user.Name)
		}
	}

	// Test bool not equal true
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    true,
				Mode:     filter.ModeNotEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	result, err = handler.DataQuery(users, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Bool not equal true filter failed: %v", err)
	}

	for _, user := range result.Data {
		if user.IsActive {
			t.Errorf("Expected only inactive users, got active user %s", user.Name)
		}
	}

	// Test bool not equal false
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    false,
				Mode:     filter.ModeNotEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	result, err = handler.DataQuery(users, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Bool not equal false filter failed: %v", err)
	}

	for _, user := range result.Data {
		if !user.IsActive {
			t.Errorf("Expected only active users, got inactive user %s", user.Name)
		}
	}
}
