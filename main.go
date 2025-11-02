package main

import (
	"example/filter"
	"example/models"
	"example/tools"
	"log"
	"time"
)

// Sample filter root with AND logic
var sampleFilterRootAnd = filter.FilterRoot{
	Filters: []filter.Filter{
		{
			DataType: filter.DataTypeText,
			Field:    "name",
			Mode:     string(filter.FilterModeContains),
			Value:    "John",
		},
		{
			DataType: filter.DataTypeNumber,
			Field:    "age",
			Mode:     string(filter.FilterModeGTE),
			Value:    18,
		},
		{
			DataType: filter.DataTypeDate,
			Field:    "created_at",
			Mode:     string(filter.FilterModeAfter),
			Value:    "2024-01-01",
		},
		{
			DataType: filter.DataTypeBool,
			Field:    "is_active",
			Mode:     string(filter.FilterModeEqual),
			Value:    true,
		},
		{
			DataType: filter.DataTypeText,
			Field:    "friend.name",
			Mode:     string(filter.FilterModeContains),
			Value:    "Alice",
		},
	},
	Logic: filter.FilterLogicAnd,
	SortFields: []filter.SortField{
		{Field: "age", Order: "asc"},
		{Field: "name", Order: "desc"},
	},
}

// Sample users data with friends
var sampleUsers = []*models.User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 25, IsActive: true, CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 4, Name: "Alice Williams"}},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 30, IsActive: true, CreatedAt: time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 5, Name: "Charlie Brown"}},
	{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Age: 35, IsActive: false, CreatedAt: time.Date(2023, 12, 10, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 1, Name: "John Doe"}},
	{ID: 4, Name: "Alice Williams", Email: "alice@test.com", Age: 22, IsActive: true, CreatedAt: time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 9, Name: "Grace Lee"}},
	{ID: 5, Name: "Charlie Brown", Email: "charlie@example.com", Age: 28, IsActive: true, CreatedAt: time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 2, Name: "Jane Smith"}},
	{ID: 6, Name: "Diana Prince", Email: "diana@example.com", Age: 32, IsActive: false, CreatedAt: time.Date(2023, 11, 30, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 7, Name: "Eve Davis"}},
	{ID: 7, Name: "Eve Davis", Email: "eve@test.com", Age: 27, IsActive: true, CreatedAt: time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 6, Name: "Diana Prince"}},
	{ID: 8, Name: "Frank Miller", Email: "frank@example.com", Age: 40, IsActive: false, CreatedAt: time.Date(2023, 10, 20, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 3, Name: "Bob Johnson"}},
	{ID: 9, Name: "Grace Lee", Email: "grace@example.com", Age: 26, IsActive: true, CreatedAt: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 10, Name: "Henry Wilson"}},
	{ID: 10, Name: "Henry Wilson", Email: "henry@test.com", Age: 33, IsActive: true, CreatedAt: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 1, Name: "John Doe"}},
}

func main() {
	tools.RunCLI("models/")

	filter, err := filter.NewFilter[models.User]("User")
	if err != nil {
		log.Fatal(err)
	}
	_, err = filter.FilterData(sampleUsers, sampleFilterRootAnd)
	if err != nil {
		log.Fatal(err)
	}
	// tools.PrintAsJSON(result)
}
