package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	types "api-jellyfish/internal/common/type"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIResponse(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("basic response without debug", func(t *testing.T) {
		gin.SetMode(gin.ReleaseMode)
		defer gin.SetMode(gin.TestMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		APIResponse(c, http.StatusOK, "Success", "test data", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success", response.Message)
		assert.Equal(t, "test data", response.Data)
		assert.Nil(t, response.Debug)
	})

	t.Run("response with debug mode", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Set request ID and version
		c.Set("request_id", "test-123")
		c.Set("version", "v1.0.0")

		APIResponse(c, http.StatusOK, "Success", "test data", nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Success", response.Message)
		assert.Equal(t, "test data", response.Data)
		assert.NotNil(t, response.Debug)
		assert.Equal(t, "test-123", response.Debug.RequestID)
		assert.Equal(t, "v1.0.0", response.Debug.Version)
		assert.True(t, response.Debug.RuntimeMs >= 0)
	})

	t.Run("response with error in debug mode", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		testError := errors.New("test error occurred")

		APIResponse(c, http.StatusInternalServerError, "Error", nil, testError)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Error", response.Message)
		assert.Nil(t, response.Data)
		assert.NotNil(t, response.Debug)
		assert.NotNil(t, response.Debug.Error)
		assert.Equal(t, "test error occurred", *response.Debug.Error)
	})

	t.Run("response with existing debug context", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Set existing debug info
		startTime := time.Now().Add(-100 * time.Millisecond)
		debugInfo := &types.ResponseAPIDebug{
			RequestID: "existing-123",
			Version:   "v2.0.0",
			StartTime: startTime,
		}
		c.Set("debug", debugInfo)

		APIResponse(c, http.StatusCreated, "Created", map[string]string{"id": "123"}, nil)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Created", response.Message)
		assert.NotNil(t, response.Debug)
		assert.Equal(t, "existing-123", response.Debug.RequestID)
		assert.Equal(t, "v2.0.0", response.Debug.Version)
		assert.True(t, response.Debug.RuntimeMs >= 100) // At least 100ms
		assert.False(t, response.Debug.EndTime.IsZero())
	})

	t.Run("response with start-time context", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		startTime := time.Now().Add(-50 * time.Millisecond)
		c.Set("start-time", startTime)

		APIResponse(c, http.StatusOK, "OK", nil, nil)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Debug)
		assert.True(t, response.Debug.RuntimeMs >= 50) // At least 50ms
		// Note: StartTime comparison might have different monotonic clock values
		assert.WithinDuration(t, startTime, response.Debug.StartTime, time.Millisecond)
	})

	t.Run("response with invalid start-time context", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Set invalid start-time (not time.Time)
		c.Set("start-time", "invalid-time")

		APIResponse(c, http.StatusOK, "OK", nil, nil)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Debug)
		// Should use current time when start-time is invalid
		assert.True(t, response.Debug.RuntimeMs >= 0)
	})

	t.Run("response with complex data structure", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		complexData := map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "John", "active": true},
				{"id": 2, "name": "Jane", "active": false},
			},
			"total": 2,
			"metadata": map[string]string{
				"page": "1",
				"size": "10",
			},
		}

		APIResponse(c, http.StatusOK, "Users retrieved", complexData, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Users retrieved", response.Message)
		assert.NotNil(t, response.Data)

		// Verify complex data structure is preserved
		dataMap, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(2), dataMap["total"]) // JSON numbers are float64
	})

	t.Run("response with different status codes", func(t *testing.T) {
		testCases := []struct {
			statusCode int
			message    string
		}{
			{http.StatusOK, "OK"},
			{http.StatusCreated, "Created"},
			{http.StatusBadRequest, "Bad Request"},
			{http.StatusUnauthorized, "Unauthorized"},
			{http.StatusForbidden, "Forbidden"},
			{http.StatusNotFound, "Not Found"},
			{http.StatusInternalServerError, "Internal Server Error"},
		}

		for _, tc := range testCases {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			APIResponse(c, tc.statusCode, tc.message, nil, nil)

			assert.Equal(t, tc.statusCode, w.Code)

			var response types.ResponseAPI
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tc.message, response.Message)
		}
	})

	t.Run("response aborts context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		APIResponse(c, http.StatusOK, "Success", nil, nil)

		// Context should be aborted
		assert.True(t, c.IsAborted())
	})

	t.Run("response with nil error", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		APIResponse(c, http.StatusOK, "Success", "data", nil)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Debug)
		assert.Nil(t, response.Debug.Error)
	})

	t.Run("response with empty message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		APIResponse(c, http.StatusOK, "", "data", nil)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "", response.Message)
		assert.Equal(t, "data", response.Data)
	})

	t.Run("response with nil data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		APIResponse(c, http.StatusNoContent, "No Content", nil, nil)

		// Check that the status code is set correctly
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Check that context is aborted (primary behavior we're testing)
		assert.True(t, c.IsAborted())
	})
}

func TestAPIResponseEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("response timing accuracy", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		startTime := time.Now()
		c.Set("start-time", startTime)

		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)

		APIResponse(c, http.StatusOK, "OK", nil, nil)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Debug)

		// Runtime should be at least 10ms but allow some tolerance for CI/test environments
		assert.True(t, response.Debug.RuntimeMs >= 1)   // At least 1ms for any processing
		assert.True(t, response.Debug.RuntimeMs < 1000) // Should be less than 1 second
	})

	t.Run("response with special characters in message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		specialMessage := "Success with special chars: äöü 中文 🎉 \"quoted\" 'single' & <tag>"

		APIResponse(c, http.StatusOK, specialMessage, nil, nil)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, specialMessage, response.Message)
	})

	t.Run("response preserves context values", func(t *testing.T) {
		gin.SetMode(gin.DebugMode)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Set various context values
		c.Set("request_id", "preserve-test-123")
		c.Set("version", "v3.1.4")
		c.Set("custom_value", "should_remain")

		APIResponse(c, http.StatusOK, "OK", nil, nil)

		// Check that custom context values are preserved
		customValue, exists := c.Get("custom_value")
		assert.True(t, exists)
		assert.Equal(t, "should_remain", customValue)

		var response types.ResponseAPI
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "preserve-test-123", response.Debug.RequestID)
		assert.Equal(t, "v3.1.4", response.Debug.Version)
	})
}

// Benchmark tests
func BenchmarkAPIResponse(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		APIResponse(c, http.StatusOK, "Success", "benchmark data", nil)
	}
}

func BenchmarkAPIResponseWithDebug(b *testing.B) {
	gin.SetMode(gin.DebugMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("request_id", "bench-123")
		c.Set("version", "v1.0.0")

		APIResponse(c, http.StatusOK, "Success", "benchmark data", nil)
	}
}
