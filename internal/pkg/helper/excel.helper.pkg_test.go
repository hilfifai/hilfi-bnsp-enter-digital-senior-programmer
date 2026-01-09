package helper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestCreateExcelFile(t *testing.T) {
	t.Run("successful excel file creation", func(t *testing.T) {
		filename := "test_output.xlsx"
		sheetName := "TestSheet"
		headers := []string{"Name", "Age", "City"}
		data := [][]interface{}{
			{"John", 30, "New York"},
			{"Jane", 25, "Los Angeles"},
			{"Bob", 35, "Chicago"},
		}

		// Clean up file after test
		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)

		// Verify file exists
		_, err = os.Stat(filename)
		assert.NoError(t, err)

		// Verify file content
		f, err := excelize.OpenFile(filename)
		assert.NoError(t, err)
		defer f.Close()

		// Check sheet name
		sheets := f.GetSheetList()
		assert.Contains(t, sheets, sheetName)

		// Check headers
		for i, header := range headers {
			cell := fmt.Sprintf("%c1", 'A'+i)
			value, err := f.GetCellValue(sheetName, cell)
			assert.NoError(t, err)
			assert.Equal(t, header, value)
		}

		// Check data
		for rowIndex, row := range data {
			for colIndex, expectedValue := range row {
				cell := fmt.Sprintf("%c%d", 'A'+colIndex, rowIndex+2)
				value, err := f.GetCellValue(sheetName, cell)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("%v", expectedValue), value)
			}
		}
	})

	t.Run("empty data", func(t *testing.T) {
		filename := "test_empty.xlsx"
		sheetName := "EmptySheet"
		headers := []string{"Col1", "Col2"}
		data := [][]interface{}{}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)

		// Verify file exists and has headers
		f, err := excelize.OpenFile(filename)
		assert.NoError(t, err)
		defer f.Close()

		// Check headers exist
		value, err := f.GetCellValue(sheetName, "A1")
		assert.NoError(t, err)
		assert.Equal(t, "Col1", value)

		value, err = f.GetCellValue(sheetName, "B1")
		assert.NoError(t, err)
		assert.Equal(t, "Col2", value)
	})

	t.Run("no headers", func(t *testing.T) {
		filename := "test_no_headers.xlsx"
		sheetName := "NoHeadersSheet"
		headers := []string{}
		data := [][]interface{}{
			{"value1", "value2"},
		}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)

		// Verify file exists
		f, err := excelize.OpenFile(filename)
		assert.NoError(t, err)
		defer f.Close()

		// Check data starts from row 2 (since no headers)
		value, err := f.GetCellValue(sheetName, "A2")
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("large dataset", func(t *testing.T) {
		filename := "test_large.xlsx"
		sheetName := "LargeSheet"
		headers := []string{"ID", "Name", "Value"}

		// Create large dataset
		data := make([][]interface{}, 1000)
		for i := 0; i < 1000; i++ {
			data[i] = []interface{}{i + 1, fmt.Sprintf("Name%d", i+1), i * 10}
		}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)

		// Verify some random data points
		f, err := excelize.OpenFile(filename)
		assert.NoError(t, err)
		defer f.Close()

		// Check last row
		value, err := f.GetCellValue(sheetName, "A1001")
		assert.NoError(t, err)
		assert.Equal(t, "1000", value)

		value, err = f.GetCellValue(sheetName, "B1001")
		assert.NoError(t, err)
		assert.Equal(t, "Name1000", value)
	})

	t.Run("special characters in data", func(t *testing.T) {
		filename := "test_special.xlsx"
		sheetName := "SpecialSheet"
		headers := []string{"Text", "Number", "Special"}
		data := [][]interface{}{
			{"Hello\nWorld", 123.45, "Special@#$%^&*()"},
			{"Unicode: 你好", -456, "Emoji: 😀🎉"},
		}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)

		// Verify special characters are preserved
		f, err := excelize.OpenFile(filename)
		assert.NoError(t, err)
		defer f.Close()

		value, err := f.GetCellValue(sheetName, "A2")
		assert.NoError(t, err)
		assert.Equal(t, "Hello\nWorld", value)

		value, err = f.GetCellValue(sheetName, "A3")
		assert.NoError(t, err)
		assert.Equal(t, "Unicode: 你好", value)
	})

	t.Run("invalid file path", func(t *testing.T) {
		filename := "/invalid/path/test.xlsx"
		sheetName := "TestSheet"
		headers := []string{"Col1"}
		data := [][]interface{}{{"value"}}

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("mixed data types", func(t *testing.T) {
		filename := "test_mixed.xlsx"
		sheetName := "MixedSheet"
		headers := []string{"String", "Int", "Float", "Bool", "Nil"}
		data := [][]interface{}{
			{"text", 42, 3.14, true, nil},
			{"", 0, 0.0, false, "not nil"},
		}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)

		// Verify mixed types are handled
		f, err := excelize.OpenFile(filename)
		assert.NoError(t, err)
		defer f.Close()

		// Check boolean value (Excel may capitalize boolean values)
		value, err := f.GetCellValue(sheetName, "D2")
		assert.NoError(t, err)
		assert.True(t, value == "true" || value == "TRUE")

		// Check nil value (should be empty)
		value, err = f.GetCellValue(sheetName, "E2")
		assert.NoError(t, err)
		// nil values are typically empty in Excel
		assert.Equal(t, "", value)
	})
}

func TestGetData(t *testing.T) {
	t.Run("successful HTTP GET", func(t *testing.T) {
		// Create test server
		expectedData := []byte("test response data")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			w.WriteHeader(http.StatusOK)
			w.Write(expectedData)
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
	})

	t.Run("JSON response", func(t *testing.T) {
		expectedJSON := `{"message":"hello","status":"ok"}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(expectedJSON))
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		assert.NoError(t, err)
		assert.Equal(t, []byte(expectedJSON), data)
	})

	t.Run("empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// No content
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		assert.NoError(t, err)
		assert.Empty(t, data)
	})

	t.Run("HTTP error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		// Function doesn't check status codes, just reads response
		assert.NoError(t, err)
		assert.Equal(t, []byte("Not Found"), data)
	})

	t.Run("large response", func(t *testing.T) {
		// Create large response data
		largeData := make([]byte, 1024*1024) // 1MB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(largeData)
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		assert.NoError(t, err)
		assert.Equal(t, largeData, data)
	})

	t.Run("invalid URL", func(t *testing.T) {
		data, err := GetData("invalid-url")

		assert.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("connection refused", func(t *testing.T) {
		// Use a port that's likely not in use
		data, err := GetData("http://localhost:99999")

		assert.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("timeout simulation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			// Note: This won't actually timeout with default client timeout
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("slow response"))
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		// Should still succeed as default timeout is usually sufficient
		assert.NoError(t, err)
		assert.Equal(t, []byte("slow response"), data)
	})

	t.Run("binary data", func(t *testing.T) {
		binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			w.Write(binaryData)
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		assert.NoError(t, err)
		assert.Equal(t, binaryData, data)
	})

	t.Run("custom headers in response", func(t *testing.T) {
		responseData := []byte("custom response")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Custom-Header", "custom-value")
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(responseData)
		}))
		defer server.Close()

		data, err := GetData(server.URL)

		assert.NoError(t, err)
		assert.Equal(t, responseData, data)
	})
}

func TestCreateExcelFileEdgeCases(t *testing.T) {
	t.Run("sheet name with spaces and special chars", func(t *testing.T) {
		filename := "test_special_sheet.xlsx"
		sheetName := "My Special Sheet! @#$"
		headers := []string{"Col1"}
		data := [][]interface{}{{"value"}}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)

		// Verify Excel handles special characters in sheet names
		f, err := excelize.OpenFile(filename)
		assert.NoError(t, err)
		defer f.Close()

		sheets := f.GetSheetList()
		// Excel might sanitize the sheet name
		assert.NotEmpty(t, sheets)
	})

	t.Run("very long sheet name", func(t *testing.T) {
		filename := "test_long_sheet.xlsx"
		// Excel has a limit on sheet name length (31 characters)
		sheetName := "VeryLongSheetNameThatExceedsNormalLimits"
		headers := []string{"Col1"}
		data := [][]interface{}{{"value"}}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		// Should handle long names (might truncate)
		assert.NoError(t, err)
		assert.Equal(t, filename, result)
	})

	t.Run("many columns", func(t *testing.T) {
		filename := "test_many_cols.xlsx"
		sheetName := "ManyCols"

		// Create many headers (test Excel column limit handling)
		headers := make([]string, 100)
		for i := 0; i < 100; i++ {
			headers[i] = fmt.Sprintf("Col%d", i+1)
		}

		data := [][]interface{}{make([]interface{}, 100)}
		for i := 0; i < 100; i++ {
			data[0][i] = fmt.Sprintf("value%d", i+1)
		}

		defer os.Remove(filename)

		result, err := CreateExcelFile(filename, sheetName, headers, data)

		assert.NoError(t, err)
		assert.Equal(t, filename, result)
	})
}

// Benchmark tests
func BenchmarkCreateExcelFile(b *testing.B) {
	filename := "bench_test.xlsx"
	sheetName := "BenchSheet"
	headers := []string{"ID", "Name", "Value", "Date"}
	data := [][]interface{}{
		{1, "Test Name", 123.45, "2024-01-01"},
		{2, "Another Name", 678.90, "2024-01-02"},
	}

	defer os.Remove(filename)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateExcelFile(filename, sheetName, headers, data)
	}
}

func BenchmarkGetData(b *testing.B) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("benchmark data"))
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetData(server.URL)
	}
}
