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
			FilterDataType: filter.FilterDataTypeText,
			Field:          "name",
			Mode:           filter.FilterModeContains,
			Value:          "John",
		},
		{
			FilterDataType: filter.FilterDataTypeNumber,
			Field:          "age",
			Mode:           filter.FilterModeGTE,
			Value:          18,
		},
		{
			FilterDataType: filter.FilterDataTypeDate,
			Field:          "created_at",
			Mode:           filter.FilterModeAfter,
			Value:          "2024-01-01",
		},
		{
			FilterDataType: filter.FilterDataTypeBool,
			Field:          "is_active",
			Mode:           filter.FilterModeEqual,
			Value:          true,
		},
		{
			FilterDataType: filter.FilterDataTypeText,
			Field:          "friend.name",
			Mode:           filter.FilterModeContains,
			Value:          "Alice",
		},
		{
			FilterDataType: filter.FilterDataTypeNumber,
			Field:          "height",
			Mode:           filter.FilterModeRange,
			Value:          filter.FilterRange{From: 160.0, To: 180.0},
		},
	},
	Logic: filter.FilterLogicAnd,
	SortFields: []filter.SortField{
		{Field: "age", Order: "asc"},
		{Field: "name", Order: "desc"},
	},
}

// Sample users data with friends, birthdays, and height
var sampleUsers = []*models.User{
	{ID: 1, Name: "John Doe", Email: "john@example.com", Age: 25, Height: 175.5, IsActive: true, CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), Birthday: time.Date(1999, 3, 15, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 4, Name: "Alice Williams"}},
	{ID: 2, Name: "Jane Smith", Email: "jane@example.com", Age: 30, Height: 165.0, IsActive: true, CreatedAt: time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC), Birthday: time.Date(1994, 7, 22, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 5, Name: "Charlie Brown"}},
	{ID: 3, Name: "Bob Johnson", Email: "bob@example.com", Age: 35, Height: 182.3, IsActive: false, CreatedAt: time.Date(2023, 12, 10, 8, 15, 0, 0, time.UTC), Birthday: time.Date(1989, 11, 5, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 1, Name: "John Doe"}},
	{ID: 4, Name: "Alice Williams", Email: "alice@test.com", Age: 22, Height: 168.7, IsActive: true, CreatedAt: time.Date(2024, 3, 5, 16, 20, 0, 0, time.UTC), Birthday: time.Date(2002, 1, 10, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 9, Name: "Grace Lee"}},
	{ID: 5, Name: "Charlie Brown", Email: "charlie@example.com", Age: 28, Height: 178.0, IsActive: true, CreatedAt: time.Date(2024, 1, 25, 9, 0, 0, 0, time.UTC), Birthday: time.Date(1996, 5, 18, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 2, Name: "Jane Smith"}},
	{ID: 6, Name: "Diana Prince", Email: "diana@example.com", Age: 32, Height: 172.5, IsActive: false, CreatedAt: time.Date(2023, 11, 30, 11, 30, 0, 0, time.UTC), Birthday: time.Date(1992, 9, 3, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 7, Name: "Eve Davis"}},
	{ID: 7, Name: "Eve Davis", Email: "eve@test.com", Age: 27, Height: 160.2, IsActive: true, CreatedAt: time.Date(2024, 2, 14, 13, 45, 0, 0, time.UTC), Birthday: time.Date(1997, 12, 25, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 6, Name: "Diana Prince"}},
	{ID: 8, Name: "Frank Miller", Email: "frank@example.com", Age: 40, Height: 185.8, IsActive: false, CreatedAt: time.Date(2023, 10, 20, 7, 30, 0, 0, time.UTC), Birthday: time.Date(1984, 4, 8, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 3, Name: "Bob Johnson"}},
	{ID: 9, Name: "Grace Lee", Email: "grace@example.com", Age: 26, Height: 163.4, IsActive: true, CreatedAt: time.Date(2024, 3, 10, 15, 10, 0, 0, time.UTC), Birthday: time.Date(1998, 8, 14, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 10, Name: "Henry Wilson"}},
	{ID: 10, Name: "Henry Wilson", Email: "henry@test.com", Age: 33, Height: 180.1, IsActive: true, CreatedAt: time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC), Birthday: time.Date(1991, 6, 30, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 1, Name: "John Doe"}},
	{ID: 11, Name: "Ivy Chen", Email: "ivy@example.com", Age: 24, Height: 158.9, IsActive: true, CreatedAt: time.Date(2024, 4, 12, 10, 15, 0, 0, time.UTC), Birthday: time.Date(2000, 2, 20, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 4, Name: "Alice Williams"}},
	{ID: 12, Name: "Jack Robinson", Email: "jack@test.com", Age: 29, Height: 177.6, IsActive: false, CreatedAt: time.Date(2023, 9, 8, 14, 30, 0, 0, time.UTC), Birthday: time.Date(1995, 10, 12, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 2, Name: "Jane Smith"}},
	{ID: 13, Name: "Kate Anderson", Email: "kate@example.com", Age: 31, Height: 170.3, IsActive: true, CreatedAt: time.Date(2024, 5, 18, 11, 45, 0, 0, time.UTC), Birthday: time.Date(1993, 3, 7, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 7, Name: "Eve Davis"}},
	{ID: 14, Name: "Leo Martinez", Email: "leo@example.com", Age: 23, Height: 173.2, IsActive: true, CreatedAt: time.Date(2024, 6, 22, 9, 20, 0, 0, time.UTC), Birthday: time.Date(2001, 11, 28, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 9, Name: "Grace Lee"}},
	{ID: 15, Name: "Mia Thompson", Email: "mia@test.com", Age: 36, Height: 166.8, IsActive: false, CreatedAt: time.Date(2023, 8, 15, 16, 0, 0, 0, time.UTC), Birthday: time.Date(1988, 5, 5, 0, 0, 0, 0, time.UTC), Friend: models.UserFriend{ID: 3, Name: "Bob Johnson"}},
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
