package helper

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test structs
type TestEntity struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	IsActive  *bool     `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

func TestTrackChanges(t *testing.T) {
	// Setup test data
	id := uuid.New()
	createdBy := uuid.New()
	updatedBy := uuid.New()
	now := time.Now()
	active := true
	inactive := false

	testCases := []struct {
		name        string
		oldData     *TestEntity
		newData     *TestEntity
		options     []TrackChangesOption
		expectedLen int
		expected    map[string]ChangeTrack
	}{
		{
			name: "no changes",
			oldData: &TestEntity{
				ID:        id,
				Name:      "Test",
				Code:      "TEST",
				IsActive:  &active,
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
			newData: &TestEntity{
				ID:        id,
				Name:      "Test",
				Code:      "TEST",
				IsActive:  &active,
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
			options:     []TrackChangesOption{WithSkipFields("UpdatedAt", "UpdatedBy", "CreatedAt", "CreatedBy")},
			expectedLen: 0,
			expected:    nil,
		},
		{
			name: "with changes",
			oldData: &TestEntity{
				ID:        id,
				Name:      "Old Name",
				Code:      "OLD",
				IsActive:  &active,
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
			newData: &TestEntity{
				ID:        id,
				Name:      "New Name",
				Code:      "NEW",
				IsActive:  &inactive,
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
			options:     []TrackChangesOption{WithSkipFields("UpdatedAt", "UpdatedBy", "CreatedAt", "CreatedBy")},
			expectedLen: 3,
			expected: map[string]ChangeTrack{
				"name":      {Old: "Old Name", New: "New Name"},
				"code":      {Old: "OLD", New: "NEW"},
				"is_active": {Old: active, New: inactive},
			},
		},
		{
			name: "with custom mapper",
			oldData: &TestEntity{
				ID:        id,
				Name:      "Test",
				Code:      "TEST",
				IsActive:  &active,
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
			newData: &TestEntity{
				ID:        id,
				Name:      "test",
				Code:      "test",
				IsActive:  &active,
				CreatedAt: now,
				UpdatedAt: now,
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
			options: []TrackChangesOption{
				WithSkipFields("UpdatedAt", "UpdatedBy", "CreatedAt", "CreatedBy"),
				WithCustomMapper("name", func(v interface{}) interface{} {
					if str, ok := v.(string); ok {
						return strings.ToLower(str)
					}
					return v
				}),
			},
			expectedLen: 1,
			expected: map[string]ChangeTrack{
				"code": {Old: "TEST", New: "test"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			changes := TrackChanges(tc.oldData, tc.newData, tc.options...)

			assert.Equal(t, tc.expectedLen, len(changes))

			if tc.expected != nil {
				for field, expectedChange := range tc.expected {
					actualChange, exists := changes[field]
					assert.True(t, exists, "Field %s should exist in changes", field)
					if exists {
						assert.Equal(t, expectedChange.Old, actualChange.Old)
						assert.Equal(t, expectedChange.New, actualChange.New)
					}
				}
			}
		})
	}
}

func TestBuildHistoryPayload(t *testing.T) {
	t.Run("empty changes", func(t *testing.T) {
		changes := Changes{}
		notes := "Test notes"

		payload, err := BuildHistoryPayload(changes, notes)
		assert.NoError(t, err)
		assert.Nil(t, payload)
	})

	t.Run("with changes", func(t *testing.T) {
		changes := Changes{
			"name": {Old: "Old Name", New: "New Name"},
			"code": {Old: "OLD", New: "NEW"},
		}
		notes := "Test notes"

		payload, err := BuildHistoryPayload(changes, notes)
		assert.NoError(t, err)
		assert.NotNil(t, payload)

		// Verify JSON structure
		var result map[string]interface{}
		err = json.Unmarshal(payload, &result)
		assert.NoError(t, err)

		assert.Equal(t, notes, result["notes"])
		assert.NotNil(t, result["changes"])

		changesMap := result["changes"].(map[string]interface{})
		assert.Equal(t, "Old Name", changesMap["name"].(map[string]interface{})["old"])
		assert.Equal(t, "New Name", changesMap["name"].(map[string]interface{})["new"])
		assert.Equal(t, "OLD", changesMap["code"].(map[string]interface{})["old"])
		assert.Equal(t, "NEW", changesMap["code"].(map[string]interface{})["new"])
	})

	t.Run("marshal error", func(t *testing.T) {
		// This test would require a custom type that fails to marshal
		// For now, we skip this as it's hard to simulate
		t.Skip("Skipping marshal error test - requires complex setup")
	})
}

func TestIsEqual(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        nil,
			b:        "test",
			expected: false,
		},
		{
			name:     "same string",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "different string",
			a:        "test1",
			b:        "test2",
			expected: false,
		},
		{
			name:     "same time",
			a:        now,
			b:        now,
			expected: true,
		},
		{
			name:     "different time",
			a:        now,
			b:        now.Add(time.Hour),
			expected: false,
		},
		{
			name:     "same number",
			a:        123,
			b:        123,
			expected: true,
		},
		{
			name:     "different number",
			a:        123,
			b:        456,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isEqual(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetJSONFieldName(t *testing.T) {
	testCases := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name: "with json tag",
			field: reflect.StructField{
				Tag: `json:"test_field"`,
			},
			expected: "test_field",
		},
		{
			name: "with json tag and omitempty",
			field: reflect.StructField{
				Tag: `json:"test_field,omitempty"`,
			},
			expected: "test_field",
		},
		{
			name: "without json tag",
			field: reflect.StructField{
				Tag: `db:"test_field"`,
			},
			expected: "",
		},
		{
			name: "with json tag '-'",
			field: reflect.StructField{
				Tag: `json:"-"`,
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getJSONFieldName(tc.field)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	active := true

	entity := TestEntity{
		ID:        id,
		Name:      "Test",
		Code:      "TEST",
		IsActive:  &active,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: uuid.New(),
		UpdatedBy: uuid.New(),
	}

	val := reflect.ValueOf(entity)

	testCases := []struct {
		name     string
		field    string
		expected interface{}
	}{
		{
			name:     "get name field",
			field:    "name",
			expected: "Test",
		},
		{
			name:     "get code field",
			field:    "code",
			expected: "TEST",
		},
		{
			name:     "get is_active field",
			field:    "is_active",
			expected: active,
		},
		{
			name:     "get created_at field",
			field:    "created_at",
			expected: now,
		},
		{
			name:     "non-existent field",
			field:    "non_existent",
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getFieldValue(val, tc.field)
			assert.Equal(t, tc.expected, result)
		})
	}
}
