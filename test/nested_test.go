package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestNestedStructFiltering tests filtering on nested struct fields
func TestNestedStructFiltering(t *testing.T) {
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		ZipCode string `json:"zip_code"`
	}

	type Person struct {
		ID      uint    `json:"id"`
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	people := []*Person{
		{ID: 1, Name: "Alice", Address: Address{Street: "123 Main St", City: "New York", ZipCode: "10001"}},
		{ID: 2, Name: "Bob", Address: Address{Street: "456 Oak Ave", City: "Los Angeles", ZipCode: "90001"}},
		{ID: 3, Name: "Charlie", Address: Address{Street: "789 Pine Rd", City: "New York", ZipCode: "10002"}},
		{ID: 4, Name: "Diana", Address: Address{Street: "321 Elm St", City: "Chicago", ZipCode: "60601"}},
	}

	maxDepth := 3
	handler := filter.NewFilter[Person](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth, // Allow 3 levels of nesting
	})

	// Test nested field filtering
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "address.city",
				Value:    "New York",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataQuery(people, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Nested field filter failed: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 people in New York, got %d", result.TotalSize)
	}

	for _, person := range result.Data {
		if person.Address.City != "New York" {
			t.Errorf("Expected city New York, got %s", person.Address.City)
		}
	}
}

// TestNestedFieldSorting tests sorting on nested struct fields
func TestNestedFieldSorting(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}

	type Person struct {
		ID      uint    `json:"id"`
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	people := []*Person{
		{ID: 1, Name: "Alice", Address: Address{City: "New York"}},
		{ID: 2, Name: "Bob", Address: Address{City: "Los Angeles"}},
		{ID: 3, Name: "Charlie", Address: Address{City: "Chicago"}},
	}

	maxDepth := 3
	handler := filter.NewFilter[Person](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth, // Allow 3 levels of nesting
	})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "address.city", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.DataQuery(people, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Nested field sort failed: %v", err)
	}

	if len(result.Data) != 3 {
		t.Errorf("Expected 3 people, got %d", len(result.Data))
	}

	// Verify sorted by city
	expectedOrder := []string{"Chicago", "Los Angeles", "New York"}
	for i, person := range result.Data {
		if person.Address.City != expectedOrder[i] {
			t.Errorf("Expected city %s at position %d, got %s", expectedOrder[i], i, person.Address.City)
		}
	}
}

// TestDeeplyNestedFields tests multi-level nested structures
func TestDeeplyNestedFields(t *testing.T) {
	type Contact struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	type Address struct {
		City    string  `json:"city"`
		Contact Contact `json:"contact"`
	}

	type Person struct {
		ID      uint    `json:"id"`
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	people := []*Person{
		{
			ID:   1,
			Name: "Alice",
			Address: Address{
				City:    "New York",
				Contact: Contact{Email: "alice@example.com", Phone: "555-0001"},
			},
		},
		{
			ID:   2,
			Name: "Bob",
			Address: Address{
				City:    "Los Angeles",
				Contact: Contact{Email: "bob@example.com", Phone: "555-0002"},
			},
		},
	}

	maxDepth := 3
	handler := filter.NewFilter[Person](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth, // Allow 3 levels of nesting
	})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "address.contact.email",
				Value:    "alice",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataQuery(people, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Deeply nested field filter failed: %v", err)
	}

	t.Logf("Deeply nested filter returned %d people with alice email", result.TotalSize)
}
