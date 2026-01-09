package helper

import (
	"api-jellyfish/internal/pkg/logger"
	"errors"
	"io"
	"log"
	"testing"
)

// init - Setup logger untuk testing environment
func init() {
	// Setup logger dengan output ke ioutil.Discard untuk testing
	logger.Error = log.New(io.Discard, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Info = log.New(io.Discard, "[INFO]\t", log.Ldate|log.Ltime)
	logger.Warning = log.New(io.Discard, "[WARNING]\t", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Debug = log.New(io.Discard, "[DEBUG]\t", log.Ldate|log.Ltime|log.Lshortfile)
	logger.HTTP = log.New(io.Discard, "[HTTP]\t", log.Ldate|log.Ltime)
}

// TestHandleAppError - Unit playground untuk fungsi HandleAppError
func TestHandleAppError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		function    string
		step        string
		fatal       bool
		expectError bool
		description string
	}{
		{
			name:        "Error nil dengan fatal false",
			err:         nil,
			function:    "TestFunction",
			step:        "TestStep",
			fatal:       false,
			expectError: false,
			description: "Seharusnya return nil ketika error adalah nil",
		},
		{
			name:        "Error nil dengan fatal true",
			err:         nil,
			function:    "TestFunction",
			step:        "TestStep",
			fatal:       true,
			expectError: false,
			description: "Seharusnya return nil ketika error adalah nil meskipun fatal true",
		},
		{
			name:        "Error dengan fatal false",
			err:         errors.New("playground error"),
			function:    "TestFunction",
			step:        "TestStep",
			fatal:       false,
			expectError: true,
			description: "Seharusnya return error ketika ada error dan fatal false",
		},
		{
			name:        "Error dengan fatal true",
			err:         errors.New("fatal playground error"),
			function:    "TestFunction",
			step:        "TestStep",
			fatal:       true,
			expectError: true,
			description: "Seharusnya return error ketika ada error dan fatal true",
		},
		{
			name:        "Error dengan function dan step kosong",
			err:         errors.New("empty function step error"),
			function:    "",
			step:        "",
			fatal:       false,
			expectError: true,
			description: "Seharusnya handle error meskipun function dan step kosong",
		},
		{
			name:        "Error dengan karakter khusus",
			err:         errors.New("error with special chars: @#$%^&*()"),
			function:    "Special@Function",
			step:        "Special#Step",
			fatal:       false,
			expectError: true,
			description: "Seharusnya handle error dengan karakter khusus",
		},
		{
			name:        "Error dengan spasi di function dan step",
			err:         errors.New("error with spaces"),
			function:    "Test Function",
			step:        "Test Step",
			fatal:       true,
			expectError: true,
			description: "Seharusnya handle error dengan spasi di function dan step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleAppError(tt.err, tt.function, tt.step, tt.fatal)

			if tt.expectError {
				if result == nil {
					t.Errorf("HandleAppError() expected error but got nil - %s", tt.description)
				} else if result != tt.err {
					t.Errorf("HandleAppError() returned different error - %s", tt.description)
				}
			} else {
				if result != nil {
					t.Errorf("HandleAppError() expected nil but got error: %v - %s", result, tt.description)
				}
			}
		})
	}
}

// TestHandleAppError_ErrorTypes - Test untuk berbagai tipe error
func TestHandleAppError_ErrorTypes(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		function    string
		step        string
		fatal       bool
		description string
	}{
		{
			name:        "Custom error type",
			err:         &CustomError{Message: "custom error message"},
			function:    "CustomFunction",
			step:        "CustomStep",
			fatal:       false,
			description: "Seharusnya handle custom error type",
		},
		{
			name:        "Wrapped error",
			err:         errors.New("wrapped: " + errors.New("inner error").Error()),
			function:    "WrappedFunction",
			step:        "WrappedStep",
			fatal:       true,
			description: "Seharusnya handle wrapped error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleAppError(tt.err, tt.function, tt.step, tt.fatal)

			if result == nil {
				t.Errorf("HandleAppError() expected error but got nil - %s", tt.description)
			}

			if result != tt.err {
				t.Errorf("HandleAppError() returned different error - %s", tt.description)
			}
		})
	}
}

// TestHandleAppError_EdgeCases - Test untuk edge cases
func TestHandleAppError_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		function    string
		step        string
		fatal       bool
		description string
	}{
		{
			name:        "Error dengan function sangat panjang",
			err:         errors.New("long function error"),
			function:    "ThisIsAVeryLongFunctionNameThatExceedsNormalLengthForTestingPurposesAndShouldBeHandledCorrectly",
			step:        "TestStep",
			fatal:       false,
			description: "Seharusnya handle function dengan nama sangat panjang",
		},
		{
			name:        "Error dengan step sangat panjang",
			err:         errors.New("long step error"),
			function:    "TestFunction",
			step:        "ThisIsAVeryLongStepNameThatExceedsNormalLengthForTestingPurposesAndShouldBeHandledCorrectly",
			fatal:       true,
			description: "Seharusnya handle step dengan nama sangat panjang",
		},
		{
			name:        "Error dengan unicode characters",
			err:         errors.New("error with unicode: 你好世界"),
			function:    "UnicodeFunction",
			step:        "UnicodeStep",
			fatal:       false,
			description: "Seharusnya handle error dengan unicode characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HandleAppError(tt.err, tt.function, tt.step, tt.fatal)

			if result == nil {
				t.Errorf("HandleAppError() expected error but got nil - %s", tt.description)
			}

			if result != tt.err {
				t.Errorf("HandleAppError() returned different error - %s", tt.description)
			}
		})
	}
}

// CustomError - Custom error type untuk testing
type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

// Benchmark tests untuk performance testing
func BenchmarkHandleAppError_NoError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HandleAppError(nil, "BenchFunction", "BenchStep", false)
	}
}

func BenchmarkHandleAppError_WithError(b *testing.B) {
	testError := errors.New("benchmark error")
	for i := 0; i < b.N; i++ {
		HandleAppError(testError, "BenchFunction", "BenchStep", false)
	}
}

func BenchmarkHandleAppError_FatalError(b *testing.B) {
	testError := errors.New("fatal benchmark error")
	for i := 0; i < b.N; i++ {
		HandleAppError(testError, "BenchFunction", "BenchStep", true)
	}
}
