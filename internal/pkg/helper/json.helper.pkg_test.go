package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	types "api-jellyfish/internal/common/type"
	"api-jellyfish/internal/pkg/pagination"
	"api-jellyfish/internal/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Setup validation for tests that use validation functions
	if err := validation.Setup(); err != nil {
		panic("Failed to setup validation in tests: " + err.Error())
	}
}

type Dummy struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func setupGinTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) types.ResponseAPI {
	var response types.ResponseAPI
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	return response
}

// Original JSON helper function tests
func TestJSONToString(t *testing.T) {
	obj := Dummy{Name: "John", Age: 30}
	jsonStr, err := JSONToString(obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if jsonStr != `{"name":"John","age":30}` {
		t.Errorf("Unexpected JSON string: %s", jsonStr)
	}
}

func TestJSONToStruct(t *testing.T) {
	obj := Dummy{Name: "Jane", Age: 25}
	result, err := JSONToStruct[Dummy](obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil || result.Name != "Jane" || result.Age != 25 {
		t.Errorf("Unexpected result: %+v", result)
	}
}

func TestJSONToByte(t *testing.T) {
	obj := Dummy{Name: "Alice", Age: 22}
	jsonBytes, err := JSONToByte(obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := []byte(`{"name":"Alice","age":22}`)
	if string(jsonBytes) != string(expected) {
		t.Errorf("Unexpected JSON bytes: %s", string(jsonBytes))
	}
}

// Tests for new response helper functions

func TestAPISuccess(t *testing.T) {
	c, w := setupGinTest()
	testData := map[string]string{"result": "ok"}

	APISuccess(c, "Operation successful", testData)

	assert.Equal(t, http.StatusOK, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Operation successful", response.Message)
	assert.NotNil(t, response.Data)
}

func TestAPISuccess_EmptyMessage(t *testing.T) {
	c, w := setupGinTest()

	APISuccess(c, "", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Success", response.Message)
}

func TestAPICreated(t *testing.T) {
	c, w := setupGinTest()
	testData := map[string]int{"id": 123}

	APICreated(c, "Resource created", testData)

	assert.Equal(t, http.StatusCreated, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Resource created", response.Message)
	assert.NotNil(t, response.Data)
}

func TestAPIBadRequest(t *testing.T) {
	c, w := setupGinTest()
	testError := errors.New("validation error")

	APIBadRequest(c, "Invalid input", testError)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Invalid input", response.Message)
	assert.Nil(t, response.Data)
}

func TestAPIUnauthorized(t *testing.T) {
	c, w := setupGinTest()

	APIUnauthorized(c, "", errors.New("token expired"))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Unauthorized", response.Message)
}

func TestAPINotFound(t *testing.T) {
	c, w := setupGinTest()

	APINotFound(c, "User not found", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "User not found", response.Message)
}

func TestAPIInternalServerError(t *testing.T) {
	c, w := setupGinTest()

	APIInternalServerError(c, "", errors.New("database error"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Internal server error", response.Message)
}

func TestAPIValidationError(t *testing.T) {
	c, w := setupGinTest()

	// Create a mock validation error
	mockErr := errors.New("validation failed")

	APIValidationError(c, "Custom validation message", mockErr)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Custom validation message", response.Message)
	assert.NotNil(t, response.Data)
}

func TestAPIPaginated(t *testing.T) {
	c, w := setupGinTest()
	testData := []map[string]string{
		{"name": "Item 1"},
		{"name": "Item 2"},
	}

	APIPaginated(c, "", testData, 1, 10, 25)

	assert.Equal(t, http.StatusOK, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Data retrieved successfully", response.Message)

	// Check if response.Data contains standard pagination structure
	dataMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, dataMap, "data")
	assert.Contains(t, dataMap, "total")
	assert.Contains(t, dataMap, "page")
	assert.Contains(t, dataMap, "page_size")
	assert.Contains(t, dataMap, "total_pages")

	// Verify pagination values
	assert.Equal(t, float64(1), dataMap["page"])
	assert.Equal(t, float64(10), dataMap["page_size"])
	assert.Equal(t, float64(25), dataMap["total"])
	assert.Equal(t, float64(3), dataMap["total_pages"])
}

func TestAPIErrorFromService(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
	}{
		{"not found error", errors.New("not_found"), http.StatusNotFound},
		{"unauthorized error", errors.New("unauthorized"), http.StatusUnauthorized},
		{"forbidden error", errors.New("forbidden"), http.StatusForbidden},
		{"conflict error", errors.New("conflict"), http.StatusConflict},
		{"validation error", errors.New("validation_failed"), http.StatusBadRequest},
		{"generic error", errors.New("some other error"), http.StatusInternalServerError},
		{"nil error", nil, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupGinTest()

			APIErrorFromService(c, tt.error)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPICRUDHelpers(t *testing.T) {
	testData := map[string]interface{}{"id": 1, "name": "test"}

	t.Run("APICreateSuccess", func(t *testing.T) {
		c, w := setupGinTest()
		APICreateSuccess(c, "User", testData)
		assert.Equal(t, http.StatusCreated, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "User created successfully", response.Message)
	})

	t.Run("APIUpdateSuccess", func(t *testing.T) {
		c, w := setupGinTest()
		APIUpdateSuccess(c, "User", testData)
		assert.Equal(t, http.StatusOK, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "User updated successfully", response.Message)
	})

	t.Run("APIDeleteSuccess", func(t *testing.T) {
		c, w := setupGinTest()
		APIDeleteSuccess(c, "User")
		assert.Equal(t, http.StatusOK, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "User deleted successfully", response.Message)
	})

	t.Run("APIGetSuccess", func(t *testing.T) {
		c, w := setupGinTest()
		APIGetSuccess(c, "User", testData)
		assert.Equal(t, http.StatusOK, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "User retrieved successfully", response.Message)
	})
}

func TestAPIFieldValidationHelpers(t *testing.T) {
	t.Run("APIRequiredField", func(t *testing.T) {
		c, w := setupGinTest()
		APIRequiredField(c, "email")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "Field 'email' is required", response.Message)
	})

	t.Run("APIInvalidField", func(t *testing.T) {
		c, w := setupGinTest()
		APIInvalidField(c, "email", "invalid format")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "Field 'email' is invalid: invalid format", response.Message)
	})
}

func TestAPIAuthHelpers(t *testing.T) {
	t.Run("APITokenExpired", func(t *testing.T) {
		c, w := setupGinTest()
		APITokenExpired(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "Token has expired", response.Message)
	})

	t.Run("APIInvalidToken", func(t *testing.T) {
		c, w := setupGinTest()
		APIInvalidToken(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "Invalid or malformed token", response.Message)
	})

	t.Run("APIInsufficientPermissions", func(t *testing.T) {
		c, w := setupGinTest()
		APIInsufficientPermissions(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "Insufficient permissions to perform this action", response.Message)
	})
}

func TestAPIUtilityHelpers(t *testing.T) {
	t.Run("APIRateLimitExceeded", func(t *testing.T) {
		c, w := setupGinTest()
		APIRateLimitExceeded(c)
		assert.Equal(t, 429, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "Too many requests, please try again later", response.Message)
	})

	t.Run("APIMaintenanceMode", func(t *testing.T) {
		c, w := setupGinTest()
		APIMaintenanceMode(c)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		response := parseResponse(t, w)
		assert.Equal(t, "Service is currently under maintenance", response.Message)
	})
}

func TestAPIPaginatedFromStruct(t *testing.T) {
	c, w := setupGinTest()
	testData := []map[string]string{
		{"name": "Item 1"},
		{"name": "Item 2"},
	}

	paginatedResponse := pagination.NewPaginatedResponse(testData, 25, 1, 10)
	APIPaginatedFromStruct(c, "Custom paginated message", paginatedResponse)

	assert.Equal(t, http.StatusOK, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Custom paginated message", response.Message)

	// Check if response.Data contains standard pagination structure
	dataMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, dataMap, "data")
	assert.Contains(t, dataMap, "total")
	assert.Contains(t, dataMap, "page")
	assert.Contains(t, dataMap, "page_size")
	assert.Contains(t, dataMap, "total_pages")

	// Verify pagination values
	assert.Equal(t, float64(1), dataMap["page"])
	assert.Equal(t, float64(10), dataMap["page_size"])
	assert.Equal(t, float64(25), dataMap["total"])
	assert.Equal(t, float64(3), dataMap["total_pages"])
}

func TestAPIPaginatedWithCustomMeta(t *testing.T) {
	c, w := setupGinTest()
	testData := []map[string]string{
		{"name": "Item 1"},
		{"name": "Item 2"},
	}

	meta := map[string]interface{}{
		"query_time": "50ms",
		"cached":     true,
		"version":    "1.0",
	}

	APIPaginatedWithCustomMeta(c, "", testData, 2, 5, 15, meta)

	assert.Equal(t, http.StatusOK, w.Code)
	response := parseResponse(t, w)
	assert.Equal(t, "Data retrieved successfully", response.Message)

	// Check if response.Data contains standard pagination structure with meta
	dataMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, dataMap, "data")
	assert.Contains(t, dataMap, "total")
	assert.Contains(t, dataMap, "page")
	assert.Contains(t, dataMap, "page_size")
	assert.Contains(t, dataMap, "total_pages")
	assert.Contains(t, dataMap, "meta")

	// Verify pagination values
	assert.Equal(t, float64(2), dataMap["page"])
	assert.Equal(t, float64(5), dataMap["page_size"])
	assert.Equal(t, float64(15), dataMap["total"])
	assert.Equal(t, float64(3), dataMap["total_pages"])

	// Verify meta data
	metaMap, ok := dataMap["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "50ms", metaMap["query_time"])
	assert.Equal(t, true, metaMap["cached"])
	assert.Equal(t, "1.0", metaMap["version"])
}

func TestAPINoContent(t *testing.T) {
	c, w := setupGinTest()

	APINoContent(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}
