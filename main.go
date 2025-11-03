package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint `gorm:"primarykey"`
	Name      string
	Age       int
	Email     string
	IsActive  bool
	CreatedAt time.Time
}

func main() {
	// Initialize database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto migrate
	db.AutoMigrate(&User{})

	// Seed some test data
	seedData(db)

	// Create filter handler
	filterHandler := filter.NewFilter[User]()

	// Define filters
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "name",
				Value:          "John",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
			{
				Field:          "age",
				Value:          18,
				Mode:           filter.FilterModeGTE,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.FilterSortOrderDesc},
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}

	// Execute filtering with pagination
	result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== Filter Results ===\n")
	fmt.Printf("Total: %d, Pages: %d, Current Page: %d/%d\n",
		result.TotalSize, result.TotalPage, result.PageIndex, result.TotalPage)
	fmt.Printf("Found %d users on this page\n\n", len(result.Data))

	for i, user := range result.Data {
		fmt.Printf("%d. Name: %s, Age: %d, Email: %s, Active: %v\n",
			i+1, user.Name, user.Age, user.Email, user.IsActive)
	}
}

func seedData(db *gorm.DB) {
	// Check if data already exists
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		return
	}

	users := []User{
		{Name: "John Doe", Age: 25, Email: "john@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Jane Smith", Age: 30, Email: "jane@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "John Smith", Age: 22, Email: "johnsmith@example.com", IsActive: false, CreatedAt: time.Now()},
		{Name: "Bob Johnson", Age: 35, Email: "bob@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Alice Wonder", Age: 28, Email: "alice@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Johnny Walker", Age: 40, Email: "johnny@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Sarah Connor", Age: 32, Email: "sarah@example.com", IsActive: false, CreatedAt: time.Now()},
		{Name: "John Connor", Age: 18, Email: "johnconnor@example.com", IsActive: true, CreatedAt: time.Now()},
	}

	db.Create(&users)
	fmt.Println("Seeded test data")
}
