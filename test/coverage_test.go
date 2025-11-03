package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestGormTimeConditions tests buildTimeCondition in GORM
func TestGormTimeConditions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	type TimeEvent struct {
		ID        uint      `gorm:"primarykey" json:"id"`
		Name      string    `json:"name"`
		EventTime time.Time `json:"event_time"`
	}

	if err := db.AutoMigrate(&TimeEvent{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	events := []TimeEvent{
		{ID: 1, Name: "Morning", EventTime: time.Date(2024, 1, 1, 8, 30, 0, 0, time.UTC)},
		{ID: 2, Name: "Noon", EventTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Afternoon", EventTime: time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC)},
		{ID: 4, Name: "Evening", EventTime: time.Date(2024, 1, 1, 18, 45, 0, 0, time.UTC)},
	}

	if err := db.Create(&events).Error; err != nil {
		t.Fatalf("Failed to create events: %v", err)
	}

	handler := filter.NewFilter[TimeEvent]()

	tests := []struct {
		name string
		mode filter.Mode
		time string
	}{
		{name: "Time equal", mode: filter.ModeEqual, time: "12:00:00"},
		{name: "Time after", mode: filter.ModeAfter, time: "14:00:00"},
		{name: "Time before", mode: filter.ModeBefore, time: "10:00:00"},
		{name: "Time not equal", mode: filter.ModeNotEqual, time: "12:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "event_time",
						Value:    tt.time,
						Mode:     tt.mode,
						DataType: filter.DataTypeTime,
					},
				},
			}

			result, err := handler.DataGorm(db, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}

			t.Logf("%s returned %d events", tt.name, result.TotalSize)
		})
	}

	// Test time range
	t.Run("Time range", func(t *testing.T) {
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

		result, err := handler.DataGorm(db, filterRoot, 1, 100)
		if err != nil {
			t.Fatalf("Time range failed: %v", err)
		}

		t.Logf("Time range returned %d events", result.TotalSize)
	})
}

// TestGormDateConditions tests buildDateCondition in GORM
func TestGormDateConditions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	type Document struct {
		ID        uint      `gorm:"primarykey" json:"id"`
		Title     string    `json:"title"`
		PublishAt time.Time `json:"publish_at"`
	}

	if err := db.AutoMigrate(&Document{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	docs := []Document{
		{ID: 1, Title: "Doc1", PublishAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Doc2", PublishAt: time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC)},
		{ID: 3, Title: "Doc3", PublishAt: time.Date(2024, 3, 10, 9, 0, 0, 0, time.UTC)},
	}

	if err := db.Create(&docs).Error; err != nil {
		t.Fatalf("Failed to create documents: %v", err)
	}

	handler := filter.NewFilter[Document]()

	tests := []struct {
		name string
		mode filter.Mode
		date string
	}{
		{name: "Date equal", mode: filter.ModeEqual, date: "2024-02-20"},
		{name: "Date not equal", mode: filter.ModeNotEqual, date: "2024-02-20"},
		{name: "Date before", mode: filter.ModeBefore, date: "2024-02-01"},
		{name: "Date after", mode: filter.ModeAfter, date: "2024-02-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "publish_at",
						Value:    tt.date,
						Mode:     tt.mode,
						DataType: filter.DataTypeDate,
					},
				},
			}

			result, err := handler.DataGorm(db, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}

			t.Logf("%s returned %d documents", tt.name, result.TotalSize)
		})
	}

	// Test date range
	t.Run("Date range", func(t *testing.T) {
		filterRoot := filter.Root{
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

		result, err := handler.DataGorm(db, filterRoot, 1, 100)
		if err != nil {
			t.Fatalf("Date range failed: %v", err)
		}

		t.Logf("Date range returned %d documents", result.TotalSize)
	})
}

// TestGormNumberConditionsEdgeCases tests number condition edge cases
func TestGormNumberConditionsEdgeCases(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	users := generateTestUsers()
	if err := db.Create(users).Error; err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	handler := filter.NewFilter[TestUser]()

	tests := []struct {
		name  string
		mode  filter.Mode
		value interface{}
	}{
		{name: "GT", mode: filter.ModeGT, value: 30},
		{name: "GTE", mode: filter.ModeGTE, value: 30},
		{name: "LT", mode: filter.ModeLT, value: 30},
		{name: "LTE", mode: filter.ModeLTE, value: 30},
		{name: "Equal", mode: filter.ModeEqual, value: 30},
		{name: "NotEqual", mode: filter.ModeNotEqual, value: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "age",
						Value:    tt.value,
						Mode:     tt.mode,
						DataType: filter.DataTypeNumber,
					},
				},
			}

			result, err := handler.DataGorm(db, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}

			t.Logf("%s with value %v returned %d users", tt.name, tt.value, result.TotalSize)
		})
	}

	// Test number range with different types
	t.Run("Number range", func(t *testing.T) {
		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field: "age",
					Value: filter.Range{
						From: 25,
						To:   35,
					},
					Mode:     filter.ModeRange,
					DataType: filter.DataTypeNumber,
				},
			},
		}

		result, err := handler.DataGorm(db, filterRoot, 1, 100)
		if err != nil {
			t.Fatalf("Number range failed: %v", err)
		}

		t.Logf("Number range returned %d users", result.TotalSize)
	})
}

// TestHybridEstimation tests estimateTableRows function
func TestHybridEstimation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	users := generateTestUsers()
	if err := db.Create(users).Error; err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	handler := filter.NewFilter[TestUser]()

	// Test with different thresholds and data sizes
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

	// Hybrid will internally call estimateTableRows
	result, err := handler.Hybrid(db, 1000, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Hybrid failed: %v", err)
	}

	t.Logf("Hybrid filter returned %d active users", result.TotalSize)

	// Test with low threshold to force database query
	result, err = handler.Hybrid(db, 1, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Hybrid with low threshold failed: %v", err)
	}

	t.Logf("Hybrid filter (low threshold) returned %d active users", result.TotalSize)

	// Test with high threshold to force in-memory query
	result, err = handler.Hybrid(db, 1000000, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Hybrid with high threshold failed: %v", err)
	}

	t.Logf("Hybrid filter (high threshold) returned %d active users", result.TotalSize)
}

// TestDateTimeWithDateTime tests parseDateTime function
func TestDateTimeWithDateTime(t *testing.T) {
	type Event struct {
		ID        uint      `json:"id"`
		Name      string    `json:"name"`
		EventTime time.Time `json:"event_time"`
	}

	events := []*Event{
		{ID: 1, Name: "Event1", EventTime: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		{ID: 2, Name: "Event2", EventTime: time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC)},
		{ID: 3, Name: "Event3", EventTime: time.Date(2024, 3, 10, 9, 15, 0, 0, time.UTC)},
	}

	handler := filter.NewFilter[Event]()

	// Test DateTime filtering with string value
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "event_time",
				Value:    "2024-02-20 14:45:00",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataQuery(events, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DateTime equal failed: %v", err)
	}

	t.Logf("DateTime equal returned %d events", result.TotalSize)

	// Test DateTime range
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "event_time",
				Value: filter.Range{
					From: "2024-01-01 00:00:00",
					To:   "2024-02-28 23:59:59",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err = handler.DataQuery(events, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DateTime range failed: %v", err)
	}

	t.Logf("DateTime range returned %d events", result.TotalSize)
}
