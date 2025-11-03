package seeder

// SeederConfig holds configuration for data seeding
type SeederConfig struct {
	// Number of records to generate
	RecordCount int

	// Batch size for inserting records (to avoid memory issues)
	BatchSize int

	// Clear existing data before seeding
	ClearExisting bool

	// Show progress during seeding
	ShowProgress bool
}

// DefaultConfig returns default seeder configuration
func DefaultConfig() SeederConfig {
	return SeederConfig{
		RecordCount:   1000,
		BatchSize:     500,
		ClearExisting: false,
		ShowProgress:  true,
	}
}

// StressTestConfig returns configuration for stress testing (large dataset)
func StressTestConfig() SeederConfig {
	return SeederConfig{
		RecordCount:   10_000_000, // 10 million records
		BatchSize:     5_000,      // SQLite limit: ~999 variables / 6 fields = ~166 records max, use 5000 with chunking
		ClearExisting: true,
		ShowProgress:  true,
	}
}

// SmallTestConfig returns configuration for quick testing
func SmallTestConfig() SeederConfig {
	return SeederConfig{
		RecordCount:   100,
		BatchSize:     50,
		ClearExisting: true,
		ShowProgress:  false,
	}
}
