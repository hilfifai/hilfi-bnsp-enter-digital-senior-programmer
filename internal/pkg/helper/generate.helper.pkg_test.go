package helper

import (
	_type "api-jellyfish/internal/common/type"
	"fmt"
	"net/http"
	"testing"
)

// TestGenerateID - Unit playground untuk fungsi GenerateID
func TestGenerateID(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "Berhasil generate ID dengan panjang 16 karakter",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := GenerateID()

			if tt.expectError {
				if err == nil {
					t.Errorf("GenerateID() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateID() unexpected error: %v", err)
				return
			}

			if id == "" {
				t.Errorf("GenerateID() returned empty string")
			}

			if len(id) != 16 {
				t.Errorf("GenerateID() expected length 16, got %d", len(id))
			}

			// Test bahwa ID hanya mengandung karakter yang diizinkan
			for _, char := range id {
				found := false
				for _, allowedChar := range urlAlphabet {
					if char == allowedChar {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GenerateID() contains invalid character: %c", char)
				}
			}
		})
	}
}

// TestGenerateID_Uniqueness - Test untuk memastikan ID yang dihasilkan unik
func TestGenerateID_Uniqueness(t *testing.T) {
	generatedIDs := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Errorf("GenerateID() unexpected error: %v", err)
			return
		}

		if generatedIDs[id] {
			t.Errorf("GenerateID() generated duplicate ID: %s", id)
			return
		}

		generatedIDs[id] = true
	}
}

// TestParseResponse - Unit playground untuk fungsi ParseResponse
func TestParseResponse(t *testing.T) {
	tests := []struct {
		name          string
		inputResponse *_type.Response
		expectedCode  int
		expectedMsg   string
		description   string
	}{
		{
			name: "Response dengan code valid dan message kosong",
			inputResponse: &_type.Response{
				Code:    http.StatusOK,
				Message: "",
			},
			expectedCode: http.StatusOK,
			expectedMsg:  "Success",
			description:  "Seharusnya mengisi message default untuk status OK",
		},
		{
			name: "Response dengan code valid dan message sudah ada",
			inputResponse: &_type.Response{
				Code:    http.StatusCreated,
				Message: "Custom Message",
			},
			expectedCode: http.StatusCreated,
			expectedMsg:  "Custom Message",
			description:  "Seharusnya tidak mengubah message yang sudah ada",
		},
		{
			name: "Response dengan code invalid (terlalu rendah)",
			inputResponse: &_type.Response{
				Code:    100,
				Message: "",
			},
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "Internal Server Error",
			description:  "Seharusnya mengubah code menjadi Internal Server Error",
		},
		{
			name: "Response dengan code invalid (terlalu tinggi)",
			inputResponse: &_type.Response{
				Code:    600,
				Message: "",
			},
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "Internal Server Error",
			description:  "Seharusnya mengubah code menjadi Internal Server Error",
		},
		{
			name: "Response dengan berbagai status code",
			inputResponse: &_type.Response{
				Code:    http.StatusBadRequest,
				Message: "",
			},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Bad Request",
			description:  "Seharusnya mengisi message default untuk Bad Request",
		},
		{
			name: "Response dengan status Unauthorized",
			inputResponse: &_type.Response{
				Code:    http.StatusUnauthorized,
				Message: "",
			},
			expectedCode: http.StatusUnauthorized,
			expectedMsg:  "Unauthorized",
			description:  "Seharusnya mengisi message default untuk Unauthorized",
		},
		{
			name: "Response dengan status Forbidden",
			inputResponse: &_type.Response{
				Code:    http.StatusForbidden,
				Message: "",
			},
			expectedCode: http.StatusForbidden,
			expectedMsg:  "Forbidden",
			description:  "Seharusnya mengisi message default untuk Forbidden",
		},
		{
			name: "Response dengan status Not Found",
			inputResponse: &_type.Response{
				Code:    http.StatusNotFound,
				Message: "",
			},
			expectedCode: http.StatusNotFound,
			expectedMsg:  "Not Found",
			description:  "Seharusnya mengisi message default untuk Not Found",
		},
		{
			name: "Response dengan status Method Not Allowed",
			inputResponse: &_type.Response{
				Code:    http.StatusMethodNotAllowed,
				Message: "",
			},
			expectedCode: http.StatusMethodNotAllowed,
			expectedMsg:  "Method Not Allowed",
			description:  "Seharusnya mengisi message default untuk Method Not Allowed",
		},
		{
			name: "Response dengan status Service Unavailable",
			inputResponse: &_type.Response{
				Code:    http.StatusServiceUnavailable,
				Message: "",
			},
			expectedCode: http.StatusServiceUnavailable,
			expectedMsg:  "Service Unavailable",
			description:  "Seharusnya mengisi message default untuk Service Unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseResponse(tt.inputResponse)

			if result.Code != tt.expectedCode {
				t.Errorf("ParseResponse() code = %v, expected %v - %s",
					result.Code, tt.expectedCode, tt.description)
			}

			if result.Message != tt.expectedMsg {
				t.Errorf("ParseResponse() message = %v, expected %v - %s",
					result.Message, tt.expectedMsg, tt.description)
			}
		})
	}
}

// TestParseResponse_NilInput - Test untuk input nil
func TestParseResponse_NilInput(t *testing.T) {
	result := ParseResponse(nil)
	if result != nil {
		t.Errorf("ParseResponse(nil) expected nil, got %v", result)
	}
}

// TestStringPtr - Unit playground untuk fungsi StringPtr
func TestStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "String kosong",
			input:    "",
			expected: "",
		},
		{
			name:     "String dengan konten",
			input:    "playground string",
			expected: "playground string",
		},
		{
			name:     "String dengan karakter khusus",
			input:    "playground@123!#",
			expected: "playground@123!#",
		},
		{
			name:     "String dengan spasi",
			input:    "   playground   ",
			expected: "   playground   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringPtr(tt.input)

			if result == nil {
				t.Errorf("StringPtr() returned nil pointer")
				return
			}

			if *result != tt.expected {
				t.Errorf("StringPtr() = %v, expected %v", *result, tt.expected)
			}
		})
	}
}

// TestStringNilPtr - Unit playground untuk fungsi StringNilPtr
func TestStringNilPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "String kosong - seharusnya return nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "String dengan konten - seharusnya return pointer",
			input:    "playground string",
			expected: nil, // akan di-check secara manual
		},
		{
			name:     "String dengan spasi - seharusnya return pointer",
			input:    "   ",
			expected: nil, // akan di-check secara manual
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringNilPtr(tt.input)

			if tt.input == "" {
				if result != nil {
					t.Errorf("StringNilPtr() for empty string expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("StringNilPtr() for non-empty string expected pointer, got nil")
				} else if *result != tt.input {
					t.Errorf("StringNilPtr() = %v, expected %v", *result, tt.input)
				}
			}
		})
	}
}

// TestIntPtr - Unit playground untuk fungsi IntPtr
func TestIntPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "Integer positif",
			input:    42,
			expected: 42,
		},
		{
			name:     "Integer negatif",
			input:    -42,
			expected: -42,
		},
		{
			name:     "Integer nol",
			input:    0,
			expected: 0,
		},
		{
			name:     "Integer maksimum",
			input:    2147483647,
			expected: 2147483647,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntPtr(tt.input)

			if result == nil {
				t.Errorf("IntPtr() returned nil pointer")
				return
			}

			if *result != tt.expected {
				t.Errorf("IntPtr() = %v, expected %v", *result, tt.expected)
			}
		})
	}
}

// TestInt64Ptr - Unit playground untuk fungsi Int64Ptr
func TestInt64Ptr(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{
			name:     "Int64 positif",
			input:    42,
			expected: 42,
		},
		{
			name:     "Int64 negatif",
			input:    -42,
			expected: -42,
		},
		{
			name:     "Int64 nol",
			input:    0,
			expected: 0,
		},
		{
			name:     "Int64 maksimum",
			input:    9223372036854775807,
			expected: 9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Int64Ptr(tt.input)

			if result == nil {
				t.Errorf("Int64Ptr() returned nil pointer")
				return
			}

			if *result != tt.expected {
				t.Errorf("Int64Ptr() = %v, expected %v", *result, tt.expected)
			}
		})
	}
}

// TestBoolPtr - Unit playground untuk fungsi BoolPtr
func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected bool
	}{
		{
			name:     "Boolean true",
			input:    true,
			expected: true,
		},
		{
			name:     "Boolean false",
			input:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolPtr(tt.input)

			if result == nil {
				t.Errorf("BoolPtr() returned nil pointer")
				return
			}

			if *result != tt.expected {
				t.Errorf("BoolPtr() = %v, expected %v", *result, tt.expected)
			}
		})
	}
}

// Benchmark tests untuk performance testing
func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateID()
		if err != nil {
			b.Fatalf("GenerateID() failed: %v", err)
		}
	}
}

func BenchmarkParseResponse(b *testing.B) {
	response := &_type.Response{
		Code:    http.StatusOK,
		Message: "",
	}

	for i := 0; i < b.N; i++ {
		ParseResponse(response)
	}
}

func BenchmarkStringPtr(b *testing.B) {
	testString := "playground string"
	for i := 0; i < b.N; i++ {
		StringPtr(testString)
	}
}

func BenchmarkIntPtr(b *testing.B) {
	testInt := 42
	for i := 0; i < b.N; i++ {
		IntPtr(testInt)
	}
}

func BenchmarkBoolPtr(b *testing.B) {
	testBool := true
	for i := 0; i < b.N; i++ {
		BoolPtr(testBool)
	}
}

// TestMap - Unit test untuk fungsi Map
func TestMap(t *testing.T) {
	t.Run("string to int conversion", func(t *testing.T) {
		input := []string{"1", "2", "3", "4", "5"}
		expected := []int{1, 2, 3, 4, 5}

		result := Map(input, func(s string) int {
			val, _ := StringToInt(s)
			return val
		})

		if len(result) != len(expected) {
			t.Errorf("Map() length = %v, expected %v", len(result), len(expected))
		}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("Map() result[%d] = %v, expected %v", i, v, expected[i])
			}
		}
	})

	t.Run("int to string conversion", func(t *testing.T) {
		input := []int{10, 20, 30}
		expected := []string{"10", "20", "30"}

		result := Map(input, func(i int) string {
			return TypeToString(i)
		})

		if len(result) != len(expected) {
			t.Errorf("Map() length = %v, expected %v", len(result), len(expected))
		}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("Map() result[%d] = %v, expected %v", i, v, expected[i])
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []string{}
		result := Map(input, func(s string) int {
			return len(s)
		})

		if len(result) != 0 {
			t.Errorf("Map() with empty slice should return empty slice, got length %v", len(result))
		}
	})

	t.Run("single element", func(t *testing.T) {
		input := []string{"hello"}
		expected := []int{5}

		result := Map(input, func(s string) int {
			return len(s)
		})

		if len(result) != 1 {
			t.Errorf("Map() length = %v, expected 1", len(result))
		}

		if result[0] != expected[0] {
			t.Errorf("Map() result[0] = %v, expected %v", result[0], expected[0])
		}
	})

	t.Run("string length mapping", func(t *testing.T) {
		input := []string{"a", "bb", "ccc", "dddd"}
		expected := []int{1, 2, 3, 4}

		result := Map(input, func(s string) int {
			return len(s)
		})

		if len(result) != len(expected) {
			t.Errorf("Map() length = %v, expected %v", len(result), len(expected))
		}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("Map() result[%d] = %v, expected %v", i, v, expected[i])
			}
		}
	})

	t.Run("complex transformation", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		input := []Person{
			{Name: "Alice", Age: 25},
			{Name: "Bob", Age: 30},
			{Name: "Charlie", Age: 35},
		}

		expected := []string{"Alice-25", "Bob-30", "Charlie-35"}

		result := Map(input, func(p Person) string {
			return fmt.Sprintf("%s-%d", p.Name, p.Age)
		})

		if len(result) != len(expected) {
			t.Errorf("Map() length = %v, expected %v", len(result), len(expected))
		}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("Map() result[%d] = %v, expected %v", i, v, expected[i])
			}
		}
	})

	t.Run("boolean mapping", func(t *testing.T) {
		input := []int{0, 1, 2, 0, 3}
		expected := []bool{false, true, true, false, true}

		result := Map(input, func(i int) bool {
			return i != 0
		})

		if len(result) != len(expected) {
			t.Errorf("Map() length = %v, expected %v", len(result), len(expected))
		}

		for i, v := range result {
			if v != expected[i] {
				t.Errorf("Map() result[%d] = %v, expected %v", i, v, expected[i])
			}
		}
	})

	t.Run("large slice", func(t *testing.T) {
		// Create large input slice
		input := make([]int, 10000)
		for i := 0; i < 10000; i++ {
			input[i] = i
		}

		result := Map(input, func(i int) int {
			return i * 2
		})

		if len(result) != 10000 {
			t.Errorf("Map() length = %v, expected 10000", len(result))
		}

		// Check some random values
		if result[0] != 0 {
			t.Errorf("Map() result[0] = %v, expected 0", result[0])
		}

		if result[5000] != 10000 {
			t.Errorf("Map() result[5000] = %v, expected 10000", result[5000])
		}

		if result[9999] != 19998 {
			t.Errorf("Map() result[9999] = %v, expected 19998", result[9999])
		}
	})
}

// BenchmarkMap - Benchmark test untuk fungsi Map
func BenchmarkMap(b *testing.B) {
	input := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Map(input, func(x int) int {
			return x * 2
		})
	}
}
