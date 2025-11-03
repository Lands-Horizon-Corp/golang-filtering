package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Lands-Horizon-Corp/golang-filtering/seeder"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Command line flags
	recordCount := flag.Int("count", 1000, "Number of records to generate")
	batchSize := flag.Int("batch", 500, "Batch size for inserting records")
	clearExisting := flag.Bool("clear", false, "Clear existing data before seeding")
	showProgress := flag.Bool("progress", true, "Show progress during seeding")
	showStats := flag.Bool("stats", true, "Show statistics after seeding")
	preset := flag.String("preset", "", "Use preset config: small, default, stress")

	flag.Parse()

	// Initialize database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&seeder.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Determine configuration
	var config seeder.SeederConfig

	switch *preset {
	case "small":
		config = seeder.SmallTestConfig()
		fmt.Println("Using SMALL preset (100 records)")
	case "stress":
		config = seeder.StressTestConfig()
		fmt.Println("Using STRESS TEST preset (1,000,000 records)")
	case "default":
		config = seeder.DefaultConfig()
		fmt.Println("Using DEFAULT preset (1,000 records)")
	default:
		// Use custom flags
		config = seeder.SeederConfig{
			RecordCount:   *recordCount,
			BatchSize:     *batchSize,
			ClearExisting: *clearExisting,
			ShowProgress:  *showProgress,
		}
		fmt.Printf("Using CUSTOM config (%d records)\n", *recordCount)
	}

	fmt.Println()

	// Create seeder
	userSeeder := seeder.NewUserSeeder(db, config)

	// Run seeding
	if err := userSeeder.Seed(); err != nil {
		log.Fatal("Seeding failed:", err)
	}

	// Show statistics
	if *showStats {
		fmt.Println("\n=== Database Statistics ===")
		stats, err := userSeeder.GetStats()
		if err != nil {
			log.Fatal("Failed to get stats:", err)
		}

		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	fmt.Println("\n✓ Seeding completed successfully!")
}
