package query_test

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Animal struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Type      string `gorm:"size:50"`
	HabitatID uint
	Habitat   Habitat
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Habitat struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"size:100;not null"`
	Type    string `gorm:"size:50"`
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

	db.Delete(&animals[5])
}

func animalTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Animal{}, &Habitat{}, &Predator{}); err != nil {
		return nil, err
	}

	seedData(db)
	return db.Model(&Animal{}), nil
}

func TestRawAllMethods(t *testing.T) {

	db, err := animalTestDB()
	if err != nil {
		t.Fatal(err)
	}

	p := query.NewPagination[Animal](query.PaginationConfig{
		Verbose: true,
	})

	paginated, _ := p.RawPagination(db, 0, 2, "Habitat")
	fmt.Println("RawPagination:", paginated.Data)

	rawFind, _ := p.RawFind(db, "Habitat")
	fmt.Println("RawFind:", len(rawFind))

	count, _ := p.RawCount(db)
	fmt.Println("RawCount:", count)

	rawLock, _ := p.RawFindLock(db, "Habitat")
	fmt.Println("RawFindLock:", len(rawLock))

	one, _ := p.RawFindOne(db, "Habitat")
	fmt.Println("RawFindOne:", one)

	oneLock, _ := p.RawFindOneWithLock(db, "Habitat")
	fmt.Println("RawFindOneWithLock:", oneLock)

	exists, _ := p.RawExists(db)
	fmt.Println("RawExists:", exists)

	existsDeleted, _ := p.RawExistsIncludingDeleted(db)
	fmt.Println("RawExistsIncludingDeleted:", existsDeleted)

	maxID, _ := p.RawGetMax(db, "id")
	minID, _ := p.RawGetMin(db, "id")
	fmt.Println("RawGetMax ID:", maxID)
	fmt.Println("RawGetMin ID:", minID)

	maxLock, _ := p.RawGetMaxLock(db, "id")
	minLock, _ := p.RawGetMinLock(db, "id")
	fmt.Println("RawGetMaxLock ID:", maxLock)
	fmt.Println("RawGetMinLock ID:", minLock)

	rawTabular, _ := p.RawTabular(db, func(a *Animal) map[string]any {
		return map[string]any{
			"Name":    a.Name,
			"Type":    a.Type,
			"Habitat": a.Habitat.Name,
		}
	}, "Habitat")
	fmt.Println("RawTabular length:", len(rawTabular))

	includeDeleted, _ := p.RawFindIncludeDeleted(db, "Habitat")
	fmt.Println("RawFindIncludeDeleted:", len(includeDeleted))

	includeDeletedLock, _ := p.RawFindLockIncludeDeleted(db, "Habitat")
	fmt.Println("RawFindLockIncludeDeleted:", len(includeDeletedLock))
}

func TestRawPaginationComplex(t *testing.T) {

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&Animal{}, &Habitat{}, &Predator{}); err != nil {
		t.Fatal(err)
	}
	seedData(db)

	p := query.NewPagination[Animal](query.PaginationConfig{
		Verbose: true,
	})
	db = db.Model(&Animal{})

	dbQuery := db.
		Table("animals").
		Joins("JOIN habitats ON habitats.id = animals.habitat_id").
		Where("habitats.name = ?", "Forest").
		Order("animals.name ASC")
	res, err := p.RawPagination(
		dbQuery,
		0,
		10,
		"Habitat",
	)
	if err != nil {
		t.Fatalf("RawPagination failed: %s", err)
	}
	if len(res.Data) == 0 {
		t.Fatalf("expected results but got 0")
	}
	for _, a := range res.Data {
		if a.Habitat.Name != "Forest" {
			t.Fatalf("expected Forest habitat, got %s", a.Habitat.Name)
		}
	}
	t.Logf("SUCCESS — Got %d animals from Forest", len(res.Data))
}

func TestPaginationRaw(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&Animal{}, &Habitat{}, &Predator{}); err != nil {
		t.Fatal(err)
	}
	seedData(db)

	e := echo.New()
	req := httptest.NewRequest("GET", "/?pageIndex=0&pageSize=2", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	p := query.NewPagination[Animal](query.PaginationConfig{
		Verbose: true,
	})

	db = db.Model(&Animal{})

	rawQuery := func(d *gorm.DB) *gorm.DB {
		return d.
			Joins("JOIN habitats ON habitats.id = animals.habitat_id").
			Where("habitats.name = ?", "Forest").
			Order("animals.name ASC")
	}

	result, err := p.PaginationRaw(db, ctx, rawQuery, "Habitat")
	if err != nil {
		t.Fatalf("PaginationRaw failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected some animals but got 0")
	}

	for _, a := range result.Data {
		if a.Habitat.Name != "Forest" {
			t.Fatalf("expected Forest habitat, got %s", a.Habitat.Name)
		}
	}

	t.Logf("SUCCESS — got %d animals via PaginationRaw", len(result.Data))
}
