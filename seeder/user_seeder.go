package seeder

import (
	"fmt"
	"time"

	"github.com/jaswdr/faker"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"index"`
	Age       int       `gorm:"index"`
	Email     string    `gorm:"index"`
	IsActive  bool      `gorm:"index"`
	CreatedAt time.Time `gorm:"index"`
}

// UserSeeder generates fake user data
type UserSeeder struct {
	db     *gorm.DB
	faker  faker.Faker
	config SeederConfig
}

// NewUserSeeder creates a new user seeder
func NewUserSeeder(db *gorm.DB, config SeederConfig) *UserSeeder {
	return &UserSeeder{
		db:     db,
		faker:  faker.New(),
		config: config,
	}
}

// Seed generates and inserts fake users into the database
func (s *UserSeeder) Seed() error {
	// Clear existing data if configured
	if s.config.ClearExisting {
		if s.config.ShowProgress {
			fmt.Println("Clearing existing users...")
		}
		if err := s.db.Exec("DELETE FROM users").Error; err != nil {
			return fmt.Errorf("failed to clear users: %w", err)
		}
	}

	totalRecords := s.config.RecordCount
	batchSize := s.config.BatchSize

	if s.config.ShowProgress {
		fmt.Printf("Seeding %d users in batches of %d...\n", totalRecords, batchSize)
	}

	startTime := time.Now()

	for i := 0; i < totalRecords; i += batchSize {
		remaining := totalRecords - i
		currentBatchSize := batchSize
		if remaining < batchSize {
			currentBatchSize = remaining
		}

		users := s.generateUsers(currentBatchSize)

		if err := s.db.Create(&users).Error; err != nil {
			return fmt.Errorf("failed to insert batch at index %d: %w", i, err)
		}

		if s.config.ShowProgress && (i+currentBatchSize)%10000 == 0 {
			elapsed := time.Since(startTime)
			progress := float64(i+currentBatchSize) / float64(totalRecords) * 100
			fmt.Printf("Progress: %.2f%% (%d/%d) - Elapsed: %s\n",
				progress, i+currentBatchSize, totalRecords, elapsed.Round(time.Second))
		}
	}

	elapsed := time.Since(startTime)

	if s.config.ShowProgress {
		fmt.Printf("✓ Seeded %d users in %s\n", totalRecords, elapsed.Round(time.Millisecond))
		fmt.Printf("  Average: %.2f records/second\n", float64(totalRecords)/elapsed.Seconds())
	}

	return nil
}

// generateUsers creates a batch of fake users
func (s *UserSeeder) generateUsers(count int) []User {
	users := make([]User, count)

	firstNames := []string{"John", "Jane", "Bob", "Alice", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}

	for i := 0; i < count; i++ {
		// Mix of realistic patterns
		firstName := firstNames[s.faker.IntBetween(0, len(firstNames)-1)]
		lastName := lastNames[s.faker.IntBetween(0, len(lastNames)-1)]

		// Add some variation
		if s.faker.IntBetween(1, 10) > 7 {
			firstName = s.faker.Person().FirstName()
			lastName = s.faker.Person().LastName()
		}

		name := fmt.Sprintf("%s %s", firstName, lastName)

		// Generate email from name
		emailPrefix := s.faker.Internet().Slug()
		email := fmt.Sprintf("%s@%s", emailPrefix, s.faker.Internet().FreeEmailDomain())

		// Age distribution: mostly adults (18-65), some elderly (65-90), few teens (13-18)
		age := s.generateAge()

		// 80% active users
		isActive := s.faker.IntBetween(1, 10) <= 8

		// Created dates spread over the last 3 years
		daysAgo := s.faker.IntBetween(0, 365*3)
		createdAt := time.Now().AddDate(0, 0, -daysAgo)

		users[i] = User{
			Name:      name,
			Age:       age,
			Email:     email,
			IsActive:  isActive,
			CreatedAt: createdAt,
		}
	}

	return users
}

// generateAge returns a realistic age distribution
func (s *UserSeeder) generateAge() int {
	roll := s.faker.IntBetween(1, 100)

	switch {
	case roll <= 10: // 10% teens
		return s.faker.IntBetween(13, 17)
	case roll <= 75: // 65% young adults
		return s.faker.IntBetween(18, 35)
	case roll <= 90: // 15% middle aged
		return s.faker.IntBetween(36, 55)
	case roll <= 98: // 8% seniors
		return s.faker.IntBetween(56, 75)
	default: // 2% elderly
		return s.faker.IntBetween(76, 90)
	}
}

// GetStats returns statistics about the seeded data
func (s *UserSeeder) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	var totalCount int64
	if err := s.db.Model(&User{}).Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_users"] = totalCount

	// Active count
	var activeCount int64
	if err := s.db.Model(&User{}).Where("is_active = ?", true).Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active_users"] = activeCount
	stats["inactive_users"] = totalCount - activeCount

	// Age distribution
	var avgAge float64
	if err := s.db.Model(&User{}).Select("AVG(age)").Scan(&avgAge).Error; err != nil {
		return nil, err
	}
	stats["average_age"] = fmt.Sprintf("%.2f", avgAge)

	var minAge, maxAge int
	s.db.Model(&User{}).Select("MIN(age)").Scan(&minAge)
	s.db.Model(&User{}).Select("MAX(age)").Scan(&maxAge)
	stats["min_age"] = minAge
	stats["max_age"] = maxAge

	return stats, nil
}
