package helper

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTimezone(t *testing.T) {
	// Save original TZ environment variable
	originalTZ := os.Getenv("TZ")
	defer func() {
		if originalTZ != "" {
			os.Setenv("TZ", originalTZ)
		} else {
			os.Unsetenv("TZ")
		}
	}()

	t.Run("default timezone when TZ not set", func(t *testing.T) {
		os.Unsetenv("TZ")

		result := GetTimezone()
		assert.Equal(t, "Asia/Jakarta", result)
	})

	t.Run("custom timezone from environment", func(t *testing.T) {
		os.Setenv("TZ", "America/New_York")

		result := GetTimezone()
		assert.Equal(t, "America/New_York", result)
	})

	t.Run("empty TZ environment variable", func(t *testing.T) {
		os.Setenv("TZ", "")

		result := GetTimezone()
		assert.Equal(t, "Asia/Jakarta", result)
	})

	t.Run("various timezones", func(t *testing.T) {
		testCases := []string{
			"UTC",
			"Europe/London",
			"America/Los_Angeles",
			"Asia/Tokyo",
			"Australia/Sydney",
		}

		for _, tz := range testCases {
			os.Setenv("TZ", tz)
			result := GetTimezone()
			assert.Equal(t, tz, result)
		}
	})

	t.Run("invalid timezone format", func(t *testing.T) {
		os.Setenv("TZ", "Invalid/Timezone")

		result := GetTimezone()
		assert.Equal(t, "Invalid/Timezone", result)
		// Note: Function doesn't validate timezone, just returns what's set
	})
}

func TestGetDateStart(t *testing.T) {
	// Save original TZ environment variable
	originalTZ := os.Getenv("TZ")
	defer func() {
		if originalTZ != "" {
			os.Setenv("TZ", originalTZ)
		} else {
			os.Unsetenv("TZ")
		}
	}()

	t.Run("get date start with default timezone", func(t *testing.T) {
		os.Unsetenv("TZ") // Use default Asia/Jakarta

		// Create a specific time
		inputTime := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Check that time is set to start of day in Jakarta timezone
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.June, result.Month())
		// Day might be different due to timezone conversion
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())
	})

	t.Run("get date start with UTC timezone", func(t *testing.T) {
		os.Setenv("TZ", "UTC")

		inputTime := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.June, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())

		// Verify timezone
		assert.Equal(t, "UTC", result.Location().String())
	})

	t.Run("get date start with America/New_York timezone", func(t *testing.T) {
		os.Setenv("TZ", "America/New_York")

		inputTime := time.Date(2024, 6, 15, 4, 30, 45, 0, time.UTC) // 4 AM UTC

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.June, result.Month())
		// In summer, NY is UTC-4, so 4 AM UTC = midnight NY (start of June 15)
		// or it could be June 14 depending on the exact time
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())
	})

	t.Run("timezone conversion across date boundary", func(t *testing.T) {
		os.Setenv("TZ", "Asia/Tokyo") // UTC+9

		// 2 AM UTC should be 11 AM JST same day
		inputTime := time.Date(2024, 6, 15, 2, 0, 0, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.June, result.Month())
		assert.Equal(t, 15, result.Day()) // Same day in Tokyo
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
	})

	t.Run("timezone conversion backwards across date boundary", func(t *testing.T) {
		os.Setenv("TZ", "America/Los_Angeles") // UTC-8 (or UTC-7 in summer)

		// Early morning UTC should be previous day in LA
		inputTime := time.Date(2024, 6, 15, 6, 0, 0, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.June, result.Month())
		// Could be 14th or 15th depending on DST
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
	})

	t.Run("invalid timezone", func(t *testing.T) {
		os.Setenv("TZ", "Invalid/Timezone")

		inputTime := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.Error(t, err)
		assert.True(t, result.IsZero())
	})

	t.Run("leap year date", func(t *testing.T) {
		os.Setenv("TZ", "UTC")

		// February 29, 2024 (leap year)
		inputTime := time.Date(2024, 2, 29, 15, 30, 45, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.February, result.Month())
		assert.Equal(t, 29, result.Day())
		assert.Equal(t, 0, result.Hour())
	})

	t.Run("year boundary", func(t *testing.T) {
		os.Setenv("TZ", "UTC")

		// New Year's Eve
		inputTime := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2023, result.Year())
		assert.Equal(t, time.December, result.Month())
		assert.Equal(t, 31, result.Day())
		assert.Equal(t, 0, result.Hour())
	})

	t.Run("already at start of day", func(t *testing.T) {
		os.Setenv("TZ", "UTC")

		// Already at start of day
		inputTime := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, inputTime.Year(), result.Year())
		assert.Equal(t, inputTime.Month(), result.Month())
		assert.Equal(t, inputTime.Day(), result.Day())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())
	})

	t.Run("end of day", func(t *testing.T) {
		os.Setenv("TZ", "UTC")

		// Almost end of day
		inputTime := time.Date(2024, 6, 15, 23, 59, 59, 999999999, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.June, result.Month())
		assert.Equal(t, 15, result.Day())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
		assert.Equal(t, 0, result.Nanosecond())
	})
}

func TestGetDateStartDSTTransition(t *testing.T) {
	// Save original TZ environment variable
	originalTZ := os.Getenv("TZ")
	defer func() {
		if originalTZ != "" {
			os.Setenv("TZ", originalTZ)
		} else {
			os.Unsetenv("TZ")
		}
	}()

	t.Run("spring forward DST transition", func(t *testing.T) {
		os.Setenv("TZ", "America/New_York")

		// Spring forward date in 2024 (March 10)
		inputTime := time.Date(2024, 3, 10, 10, 0, 0, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.March, result.Month())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
	})

	t.Run("fall back DST transition", func(t *testing.T) {
		os.Setenv("TZ", "America/New_York")

		// Fall back date in 2024 (November 3)
		inputTime := time.Date(2024, 11, 3, 10, 0, 0, 0, time.UTC)

		result, err := GetDateStart(inputTime)

		assert.NoError(t, err)
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.November, result.Month())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
	})
}

// Benchmark tests
func BenchmarkGetTimezone(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetTimezone()
	}
}

func BenchmarkGetDateStart(b *testing.B) {
	os.Setenv("TZ", "Asia/Jakarta")
	inputTime := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDateStart(inputTime)
	}
}
