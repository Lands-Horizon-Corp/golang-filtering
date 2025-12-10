package query_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ----------------------
// MODELS
// ----------------------
type Animal struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Type      string `gorm:"size:50"` // Land, Air, Water
	HabitatID uint
	Habitat   Habitat
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Habitat struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"size:100;not null"`
	Type    string `gorm:"size:50"` // Land, Air, Water
	Animals []Animal
}

type Predator struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	PreyID    uint
	Prey      Animal
	HabitatID uint
	Habitat   Habitat
}

// ----------------------
// PAGINATION RESULT
// ----------------------
type PaginationResult[T any] struct {
	PageIndex int
	PageSize  int
	TotalSize int
	TotalPage int
	Data      []*T
}

// ----------------------
// SEED DATA
// ----------------------
func seedData(db *gorm.DB) {
	habitats := []Habitat{
		{Name: "Forest", Type: "Land"},
		{Name: "Sky", Type: "Air"},
		{Name: "Ocean", Type: "Water"},
	}
	db.Create(&habitats)

	animals := []Animal{
		{Name: "Elephant", Type: "Land", HabitatID: 1},
		{Name: "Tiger", Type: "Land", HabitatID: 1},
		{Name: "Eagle", Type: "Air", HabitatID: 2},
		{Name: "Parrot", Type: "Air", HabitatID: 2},
		{Name: "Shark", Type: "Water", HabitatID: 3},
		{Name: "Salmon", Type: "Water", HabitatID: 3},
	}
	db.Create(&animals)

	predators := []Predator{
		{Name: "Lion", PreyID: 2, HabitatID: 1},
		{Name: "Falcon", PreyID: 4, HabitatID: 2},
		{Name: "Orca", PreyID: 6, HabitatID: 3},
	}
	db.Create(&predators)

	// Soft delete one
	db.Delete(&animals[5])
}

// ----------------------
// DB HELPER (added here)
// ----------------------
func animalTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// migrate
	if err := db.AutoMigrate(&Animal{}, &Habitat{}, &Predator{}); err != nil {
		return nil, err
	}

	// seed
	seedData(db)

	// RETURN DB ALREADY SCOPED TO ANIMAL
	return db.Model(&Animal{}), nil
}

// ----------------------
// TEST SUITE
// ----------------------
func TestRawAllMethods(t *testing.T) {

	// ⬇⬇⬇ Use the helper
	db, err := animalTestDB()
	if err != nil {
		t.Fatal(err)
	}

	p := query.NewPagination[Animal](query.PaginationConfig{
		Verbose: true,
	})

	// RawPagination
	paginated, _ := p.RawPagination(db, 0, 2, "Habitat")
	fmt.Println("RawPagination:", paginated.Data)

	// RawFind
	rawFind, _ := p.RawFind(db, "Habitat")
	fmt.Println("RawFind:", rawFind)

	// RawCount
	count, _ := p.RawCount(db)
	fmt.Println("RawCount:", count)

	// RawFindLock
	rawLock, _ := p.RawFindLock(db, "Habitat")
	fmt.Println("RawFindLock:", rawLock)

	// RawFindOne
	one, _ := p.RawFindOne(db, "Habitat")
	fmt.Println("RawFindOne:", one)

	// RawFindOneWithLock
	oneLock, _ := p.RawFindOneWithLock(db, "Habitat")
	fmt.Println("RawFindOneWithLock:", oneLock)

	// RawExists
	exists, _ := p.RawExists(db)
	fmt.Println("RawExists:", exists)

	// RawExistsIncludingDeleted
	existsDeleted, _ := p.RawExistsIncludingDeleted(db)
	fmt.Println("RawExistsIncludingDeleted:", existsDeleted)

	// RawGetMax / Min
	maxID, _ := p.RawGetMax(db, "id")
	fmt.Println("RawGetMax ID:", maxID)
	minID, _ := p.RawGetMin(db, "id")
	fmt.Println("RawGetMin ID:", minID)

	// RawGetMaxLock / MinLock
	maxLock, _ := p.RawGetMaxLock(db, "id")
	fmt.Println("RawGetMaxLock ID:", maxLock)
	minLock, _ := p.RawGetMinLock(db, "id")
	fmt.Println("RawGetMinLock ID:", minLock)

	// RawTabular
	rawTabular, _ := p.RawTabular(db, func(a *Animal) map[string]any {
		return map[string]any{
			"Name":    a.Name,
			"Type":    a.Type,
			"Habitat": a.Habitat.Name,
		}
	}, "Habitat")
	fmt.Println("RawTabular CSV length:", len(rawTabular))

	// RawFindIncludeDeleted
	includeDeleted, _ := p.RawFindIncludeDeleted(db, "Habitat")
	fmt.Println("RawFindIncludeDeleted:", includeDeleted)

	// RawFindLockIncludeDeleted
	includeDeletedLock, _ := p.RawFindLockIncludeDeleted(db, "Habitat")
	fmt.Println("RawFindLockIncludeDeleted:", includeDeletedLock)
}
