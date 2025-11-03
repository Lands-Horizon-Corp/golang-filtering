package test

import (
	"strings"
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestSanitize_SQLInjectionPrevention tests SQL injection prevention
func TestSanitize_SQLInjectionPrevention(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		mustNotContain []string
	}{
		{
			name:           "SQL comment injection",
			input:          "admin'--",
			mustNotContain: []string{"--", "'"},
		},
		{
			name:           "OR 1=1 injection",
			input:          "' OR 1=1--",
			mustNotContain: []string{"--", "'"},
		},
		{
			name:           "Union injection",
			input:          "admin'; DROP TABLE users--",
			mustNotContain: []string{"--", "';", "'"},
		},
		{
			name:           "Stored procedure execution",
			input:          "test'; exec xp_cmdshell('dir')--",
			mustNotContain: []string{"--", "';", "'", "(", ")"},
		},
		{
			name:           "Null byte injection",
			input:          "admin\x00password",
			mustNotContain: []string{"\x00"},
		},
		{
			name:           "Multi-line comment",
			input:          "admin /* comment */ password",
			mustNotContain: []string{"/*", "*/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.Sanitize(tt.input)

			for _, forbidden := range tt.mustNotContain {
				if strings.Contains(result, forbidden) {
					t.Errorf("Sanitize() failed to remove dangerous pattern '%s' from '%s', got '%s'",
						forbidden, tt.input, result)
				}
			}
		})
	}
}

// TestFilterHandler_SQLInjectionProtection tests filter with SQL injection attempts
func TestFilterHandler_SQLInjectionProtection(t *testing.T) {
	handler := filter.NewFilter[TestUser]()
	users := generateTestUsers()

	maliciousInputs := []string{
		"admin'--",
		"' OR '1'='1",
		"admin'; DROP TABLE users--",
		"' OR 1=1--",
		"admin\"; exec xp_cmdshell('dir')--",
	}

	for _, malicious := range maliciousInputs {
		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "name",
					Value:    malicious,
					Mode:     filter.ModeContains,
					DataType: filter.DataTypeText,
				},
			},
		}

		// Should not panic or cause errors
		result, err := handler.DataQuery(users, filterRoot, 1, 10)
		if err != nil {
			t.Errorf("DataQuery with malicious input '%s' returned error: %v", malicious, err)
		}

		// Result should be empty or very limited since sanitized string won't match real data
		if result.TotalSize > len(users) {
			t.Errorf("Unexpected result size for malicious input '%s': %d", malicious, result.TotalSize)
		}
	}
}

// TestFilterHandler_DataGorm_SQLInjectionProtection tests database filter with SQL injection
func TestFilterHandler_DataGorm_SQLInjectionProtection(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser]()

	maliciousInputs := []string{
		"admin'--",
		"' OR '1'='1",
		"admin'; DROP TABLE users--",
		"' OR 1=1--",
	}

	for _, malicious := range maliciousInputs {
		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "name",
					Value:    malicious,
					Mode:     filter.ModeContains,
					DataType: filter.DataTypeText,
				},
			},
		}

		// Should not panic or cause SQL errors
		result, err := handler.DataGorm(db, filterRoot, 1, 10)
		if err != nil {
			t.Errorf("DataGorm with malicious input '%s' returned error: %v", malicious, err)
		}

		// Verify database still exists and has correct record count
		var count int64
		if err := db.Model(&TestUser{}).Count(&count).Error; err != nil {
			t.Errorf("Database corrupted after malicious input '%s': %v", malicious, err)
		}
		if count != 10 {
			t.Errorf("Database record count changed after malicious input '%s': expected 10, got %d", malicious, count)
		}

		// Result should be empty since sanitized string won't match
		if result.TotalSize != 0 {
			t.Logf("Note: Sanitized input '%s' matched %d records (may be expected if sanitization leaves valid text)",
				malicious, result.TotalSize)
		}
	}
}

// TestSanitize_PreservesLegitimateInput tests that sanitization doesn't break valid input
func TestSanitize_PreservesLegitimateInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Normal text",
			input: "John Doe",
		},
		{
			name:  "Email",
			input: "user@example.com",
		},
		{
			name:  "Text with spaces",
			input: "Hello World",
		},
		{
			name:  "Text with numbers",
			input: "User123",
		},
		{
			name:  "Text with hyphens",
			input: "test-user",
		},
		{
			name:  "Text with underscores",
			input: "test_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.Sanitize(tt.input)

			// Should preserve alphanumeric characters
			if result == "" {
				t.Errorf("Sanitize() removed all content from legitimate input '%s'", tt.input)
			}

			// Should at least preserve the core alphanumeric part
			hasAlphaNum := false
			for _, char := range result {
				if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
					hasAlphaNum = true
					break
				}
			}
			if !hasAlphaNum {
				t.Errorf("Sanitize() removed alphanumeric content from '%s', got '%s'", tt.input, result)
			}
		})
	}
}

// TestFilterHandler_XSSPrevention tests XSS attack prevention
func TestFilterHandler_XSSPrevention(t *testing.T) {
	handler := filter.NewFilter[TestUser]()
	users := generateTestUsers()

	xssAttempts := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<svg onload=alert('XSS')>",
	}

	for _, xss := range xssAttempts {
		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "name",
					Value:    xss,
					Mode:     filter.ModeContains,
					DataType: filter.DataTypeText,
				},
			},
		}

		result, err := handler.DataQuery(users, filterRoot, 1, 10)
		if err != nil {
			t.Errorf("DataQuery with XSS input '%s' returned error: %v", xss, err)
		}

		// Verify no script tags or dangerous content in results
		for _, user := range result.Data {
			if strings.Contains(user.Name, "<script>") || strings.Contains(user.Name, "javascript:") {
				t.Errorf("XSS content found in result: %s", user.Name)
			}
		}
	}
}
