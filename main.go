package main

import (
	"example/filter"
	"example/models"
	"example/tools"
	"fmt"
	"log"
	"math/rand"
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

// Generate fake users
func generateFakeUsers(count int) []*models.User {
	firstNames := []string{"John", "Jane", "Bob", "Alice", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry",
		"Ivy", "Jack", "Kate", "Leo", "Mia", "Noah", "Olivia", "Peter", "Quinn", "Rachel",
		"Sam", "Tina", "Uma", "Victor", "Wendy", "Xavier", "Yara", "Zack"}

	lastNames := []string{"Doe", "Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez",
		"Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson",
		"Martin", "Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis"}

	users := make([]*models.User, count)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		name := fmt.Sprintf("%s %s", firstName, lastName)
		email := fmt.Sprintf("%s.%s@example.com",
			firstName[:1]+firstName[1:],
			lastName)

		age := rand.Intn(50) + 18             // 18-67
		height := 150.0 + rand.Float64()*40.0 // 150-190 cm
		isActive := rand.Intn(2) == 1

		// Random date between 2023 and 2024
		daysAgo := rand.Intn(730)
		createdAt := time.Now().AddDate(0, 0, -daysAgo)

		// Random birthday between 1960 and 2005
		birthYear := 1960 + rand.Intn(45)
		birthMonth := time.Month(rand.Intn(12) + 1)
		birthDay := rand.Intn(28) + 1
		birthday := time.Date(birthYear, birthMonth, birthDay, 0, 0, 0, 0, time.UTC)

		// Random friend
		friendFirstName := firstNames[rand.Intn(len(firstNames))]
		friendLastName := lastNames[rand.Intn(len(lastNames))]
		friend := models.UserFriend{
			ID:   rand.Intn(1000) + 1,
			Name: fmt.Sprintf("%s %s", friendFirstName, friendLastName),
		}

		users[i] = &models.User{
			ID:        i + 1,
			Name:      name,
			Email:     email,
			Age:       age,
			Height:    height,
			IsActive:  isActive,
			CreatedAt: createdAt,
			Birthday:  birthday,
			Friend:    friend,
		}
	}

	return users
}

func main() {
	tools.RunCLI("models/")

	// Generate fake users - CHANGE THIS NUMBER
	userCount := 10_000_000 // Change this to generate more or fewer users
	sampleUsers := generateFakeUsers(userCount)

	fmt.Printf("Generated %d fake users\n", userCount)

	filter, err := filter.NewFilter[models.User]("User")
	if err != nil {
		log.Fatal(err)
	}

	// Time the filter operation
	startTime := time.Now()
	result, err := filter.FilterData(sampleUsers, sampleFilterRootAnd, func(processed, total int, percentage float32) {
		fmt.Printf("\rProgress: %d/%d (%.2f%%)", processed, total, percentage)
	})
	elapsed := time.Since(startTime)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nFiltered results: %d users\n", result.TotalSize)
	fmt.Println(result)
	fmt.Printf("Filter execution time: %v\n", elapsed)
	fmt.Printf("Time per record: %v\n", elapsed/time.Duration(userCount))

}
