package query_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Profession struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Category  string `gorm:"size:100"`
	Employees []Employee
}

type Employee struct {
	ID           uint   `gorm:"primaryKey"`
	FirstName    string `gorm:"size:100;not null"`
	LastName     string `gorm:"size:100;not null"`
	ProfessionID uint
	Profession   Profession
	Department   string `gorm:"size:100"`
	HiredAt      time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type PaginationResult[T any] struct {
	PageIndex int
	PageSize  int
	TotalSize int
	TotalPage int
	Data      []*T
}

func seedHiringData(db *gorm.DB) {
	professions := []Profession{
		{Name: "Software Engineer", Category: "Engineering"},
		{Name: "Backend Engineer", Category: "Engineering"},
		{Name: "Frontend Engineer", Category: "Engineering"},
		{Name: "Construction Engineer", Category: "Engineering"},
		{Name: "Architect", Category: "Engineering"},
		{Name: "Surgeon", Category: "Medical"},
		{Name: "Doctor", Category: "Medical"},
		{Name: "Philosopher", Category: "Academic"},
		{Name: "Researcher", Category: "Academic"},
		{Name: "Professor", Category: "Academic"},
	}
	db.Create(&professions)

	employees := []Employee{
		{FirstName: "Alice", LastName: "Smith", ProfessionID: 1, Department: "SE", HiredAt: time.Now().AddDate(-2, 0, 0)},
		{FirstName: "Bob", LastName: "Johnson", ProfessionID: 2, Department: "BE", HiredAt: time.Now().AddDate(-3, 0, 0)},
		{FirstName: "Carol", LastName: "Williams", ProfessionID: 3, Department: "FE", HiredAt: time.Now().AddDate(-1, 0, 0)},
		{FirstName: "David", LastName: "Brown", ProfessionID: 4, Department: "Construction", HiredAt: time.Now().AddDate(-5, 0, 0)},
		{FirstName: "Eve", LastName: "Jones", ProfessionID: 5, Department: "Architect", HiredAt: time.Now().AddDate(-4, 0, 0)},

		// Medical
		{FirstName: "Frank", LastName: "Miller", ProfessionID: 6, Department: "Surgery", HiredAt: time.Now().AddDate(-6, 0, 0)},
		{FirstName: "Grace", LastName: "Davis", ProfessionID: 7, Department: "General Medicine", HiredAt: time.Now().AddDate(-2, 0, 0)},
		{FirstName: "Hank", LastName: "Garcia", ProfessionID: 6, Department: "Surgery", HiredAt: time.Now().AddDate(-1, 0, 0)},

		// Academic
		{FirstName: "Ivy", LastName: "Martinez", ProfessionID: 8, Department: "Philosophy", HiredAt: time.Now().AddDate(-10, 0, 0)},
		{FirstName: "Jack", LastName: "Rodriguez", ProfessionID: 9, Department: "Research", HiredAt: time.Now().AddDate(-8, 0, 0)},
		{FirstName: "Karen", LastName: "Lopez", ProfessionID: 10, Department: "Teaching", HiredAt: time.Now().AddDate(-12, 0, 0)},
	}

	db.Create(&employees)

	db.Delete(&employees[10])
}

func hiringTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Profession{}, &Employee{}); err != nil {
		return nil, err
	}
	seedHiringData(db)
	return db.Model(&Employee{}), nil
}
func TestPaginationRawHiring(t *testing.T) {
	db, err := hiringTestDB()
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	req := httptest.NewRequest("GET", "/?pageIndex=0&pageSize=5", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	p := query.NewPagination[Employee](query.PaginationConfig{
		Verbose: true,
	})

	rawQuery := func(d *gorm.DB) *gorm.DB {
		return d.
			Joins("JOIN professions ON professions.id = employees.profession_id").
			Where("professions.category = ?", "Engineering").
			Where(
				d.Where("employees.department LIKE ?", "SE").
					Or("employees.department LIKE ?", "BE").
					Or("employees.department LIKE ?", "FE"),
			).
			Order("professions.name ASC").
			Order("employees.hired_at DESC")
	}

	result, err := p.PaginationRaw(db, ctx, rawQuery, "Profession")
	if err != nil {
		t.Fatalf("PaginationRaw failed: %v", err)
	}

	if len(result.Data) == 0 {
		t.Fatal("expected some engineering employees but got 0")
	}

	for _, e := range result.Data {
		if e.Profession.Category != "Engineering" {
			t.Fatalf("expected Engineering category, got %s", e.Profession.Category)
		}
	}

	t.Logf("SUCCESS — got %d engineering employees via PaginationRaw", len(result.Data))
}
