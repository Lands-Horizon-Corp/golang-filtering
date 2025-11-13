package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Author represents a user who writes posts
type Author struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Posts []Post `gorm:"foreignKey:AuthorID" json:"posts"`
}

// Post represents a blog post with author and comments
type Post struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	AuthorID uint      `json:"author_id"`
	Author   Author    `gorm:"foreignKey:AuthorID" json:"author"`
	Comments []Comment `gorm:"foreignKey:PostID" json:"comments"`
}

// Comment represents a comment on a post
type Comment struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Content string `json:"content"`
	PostID  uint   `json:"post_id"`
}

// TestPreloadSingleRelation tests preloading a single related entity
func TestPreloadSingleRelation(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schemas
	if err := db.AutoMigrate(&Author{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test data
	author := Author{Name: "John Doe", Email: "john@example.com"}
	db.Create(&author)

	posts := []Post{
		{Title: "Go Programming", Content: "Learn Go", AuthorID: author.ID},
		{Title: "GORM Tutorial", Content: "Database ORM", AuthorID: author.ID},
	}
	db.Create(&posts)

	// Test without preload - Author should be empty
	handler := filter.NewFilter[Post](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
	}

	pageIndex := 0
	pageSize := 10
	result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 posts, got %d", result.TotalSize)
	}

	// Author should not be loaded
	if result.Data[0].Author.Name != "" {
		t.Error("Expected Author to not be loaded without preload")
	}

	// Test with preload - Author should be populated
	filterRootWithPreload := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		Preload:      []string{"Author"},
	}

	resultWithPreload, err := handler.DataGorm(db, filterRootWithPreload, pageIndex, pageSize)
	if err != nil {
		t.Fatalf("Error filtering with preload: %v", err)
	}

	if resultWithPreload.TotalSize != 2 {
		t.Errorf("Expected 2 posts, got %d", resultWithPreload.TotalSize)
	}

	// Author should be loaded
	if resultWithPreload.Data[0].Author.Name != "John Doe" {
		t.Errorf("Expected author name 'John Doe', got '%s'", resultWithPreload.Data[0].Author.Name)
	}

	if resultWithPreload.Data[0].Author.Email != "john@example.com" {
		t.Errorf("Expected author email 'john@example.com', got '%s'", resultWithPreload.Data[0].Author.Email)
	}
}

// TestPreloadMultipleRelations tests preloading multiple related entities
func TestPreloadMultipleRelations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Author{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test data
	author := Author{Name: "Jane Smith", Email: "jane@example.com"}
	db.Create(&author)

	post := Post{Title: "Test Post", Content: "Content", AuthorID: author.ID}
	db.Create(&post)

	comments := []Comment{
		{Content: "Great post!", PostID: post.ID},
		{Content: "Thanks for sharing", PostID: post.ID},
	}
	db.Create(&comments)

	// Test with multiple preloads
	handler := filter.NewFilter[Post](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		Preload:      []string{"Author", "Comments"},
	}

	pageIndex := 0
	pageSize := 10
	result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 post, got %d", result.TotalSize)
	}

	// Check Author is loaded
	if result.Data[0].Author.Name != "Jane Smith" {
		t.Errorf("Expected author 'Jane Smith', got '%s'", result.Data[0].Author.Name)
	}

	// Check Comments are loaded
	if len(result.Data[0].Comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(result.Data[0].Comments))
	}

	if result.Data[0].Comments[0].Content != "Great post!" {
		t.Errorf("Expected comment 'Great post!', got '%s'", result.Data[0].Comments[0].Content)
	}
}

// TestPreloadWithFiltering tests preload combined with filtering
func TestPreloadWithFiltering(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Author{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create multiple authors and posts
	author1 := Author{Name: "Alice", Email: "alice@example.com"}
	author2 := Author{Name: "Bob", Email: "bob@example.com"}
	db.Create(&author1)
	db.Create(&author2)

	posts := []Post{
		{Title: "Go Basics", Content: "Introduction", AuthorID: author1.ID},
		{Title: "Advanced Go", Content: "Deep dive", AuthorID: author1.ID},
		{Title: "Python Tutorial", Content: "Learn Python", AuthorID: author2.ID},
	}
	db.Create(&posts)

	// Filter for posts containing "Go" with author preloaded
	handler := filter.NewFilter[Post](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "title",
				Value:    "Go",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"Author"},
	}

	pageIndex := 0
	pageSize := 10
	result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 posts with 'Go' in title, got %d", result.TotalSize)
	}

	// Verify authors are loaded for filtered results
	for _, post := range result.Data {
		if post.Author.Name != "Alice" {
			t.Errorf("Expected author 'Alice', got '%s'", post.Author.Name)
		}
	}
}

// TestPreloadEmptyArray tests that empty preload array works correctly
func TestPreloadEmptyArray(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Author{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	author := Author{Name: "Test Author", Email: "test@example.com"}
	db.Create(&author)

	post := Post{Title: "Test", Content: "Content", AuthorID: author.ID}
	db.Create(&post)

	// Test with empty preload array
	handler := filter.NewFilter[Post](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		Preload:      []string{}, // Empty array
	}

	pageIndex := 0
	pageSize := 10
	result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 post, got %d", result.TotalSize)
	}

	// Author should not be loaded with empty preload array
	if result.Data[0].Author.Name != "" {
		t.Error("Expected Author to not be loaded with empty preload array")
	}
}

// TestPreloadWithSorting tests preload combined with sorting
func TestPreloadWithSorting(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Author{}, &Post{}, &Comment{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	author := Author{Name: "Charlie", Email: "charlie@example.com"}
	db.Create(&author)

	posts := []Post{
		{Title: "Zebra", Content: "Last", AuthorID: author.ID},
		{Title: "Apple", Content: "First", AuthorID: author.ID},
		{Title: "Mango", Content: "Middle", AuthorID: author.ID},
	}
	db.Create(&posts)

	// Sort by title with author preload
	handler := filter.NewFilter[Post](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		SortFields: []filter.SortField{
			{Field: "title", Order: filter.SortOrderAsc},
		},
		Preload: []string{"Author"},
	}

	pageIndex := 0
	pageSize := 10
	result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 posts, got %d", result.TotalSize)
	}

	// Check sorting
	if result.Data[0].Title != "Apple" {
		t.Errorf("Expected first post 'Apple', got '%s'", result.Data[0].Title)
	}

	if result.Data[2].Title != "Zebra" {
		t.Errorf("Expected last post 'Zebra', got '%s'", result.Data[2].Title)
	}

	// Check all have authors loaded
	for _, post := range result.Data {
		if post.Author.Name != "Charlie" {
			t.Errorf("Expected author 'Charlie', got '%s'", post.Author.Name)
		}
	}
}
