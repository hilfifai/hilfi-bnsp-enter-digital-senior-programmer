package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"api-jellyfish/internal/pkg/logger/v2"
)

// ChangeTrack merepresentasikan perubahan satu field
type ChangeTrack struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

// Changes map[string]ChangeTrack
type Changes map[string]ChangeTrack

// TrackChangesConfig konfigurasi untuk track changes
type TrackChangesConfig struct {
	SkipFields    map[string]bool                          // field yang di-skip
	CustomMappers map[string]func(interface{}) interface{} // custom mapper untuk field tertentu
}

// TrackChangesOptions functional options pattern
type TrackChangesOption func(*TrackChangesConfig)

// WithSkipFields set fields yang di-skip
func WithSkipFields(fields ...string) TrackChangesOption {
	return func(c *TrackChangesConfig) {
		if c.SkipFields == nil {
			c.SkipFields = make(map[string]bool)
		}
		for _, field := range fields {
			c.SkipFields[field] = true
		}
	}
}

// WithCustomMapper set custom mapper untuk field tertentu
func WithCustomMapper(field string, mapper func(interface{}) interface{}) TrackChangesOption {
	return func(c *TrackChangesConfig) {
		if c.CustomMappers == nil {
			c.CustomMappers = make(map[string]func(interface{}) interface{})
		}
		c.CustomMappers[field] = mapper
	}
}

// TrackChanges fungsi utama untuk melacak perubahan
func TrackChanges(oldData, newData interface{}, opts ...TrackChangesOption) Changes {
	config := &TrackChangesConfig{}
	for _, opt := range opts {
		opt(config)
	}

	changes := make(Changes)

	oldVal := reflect.ValueOf(oldData)
	newVal := reflect.ValueOf(newData)

	// Handle pointer
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
		return changes
	}

	oldType := oldVal.Type()
	newType := newVal.Type()

	// Gabungkan semua field dari kedua struct
	fieldMap := make(map[string]bool)
	for i := 0; i < oldVal.NumField(); i++ {
		fieldName := getJSONFieldName(oldType.Field(i))
		if fieldName != "" {
			fieldMap[fieldName] = true
		}
	}
	for i := 0; i < newVal.NumField(); i++ {
		fieldName := getJSONFieldName(newType.Field(i))
		if fieldName != "" {
			fieldMap[fieldName] = true
		}
	}

	// Check changes untuk setiap field
	for fieldName := range fieldMap {
		// Skip field yang di-config
		if config.SkipFields != nil && config.SkipFields[fieldName] {
			continue
		}

		oldField := getFieldValue(oldVal, fieldName)
		newField := getFieldValue(newVal, fieldName)

		// Apply custom mapper jika ada
		if config.CustomMappers != nil {
			if mapper, exists := config.CustomMappers[fieldName]; exists {
				oldField = mapper(oldField)
				newField = mapper(newField)
			}
		}

		if !isEqual(oldField, newField) {
			changes[fieldName] = ChangeTrack{
				Old: oldField,
				New: newField,
			}
		}
	}

	return changes
}

// BuildHistoryPayload membangun payload history dari changes
func BuildHistoryPayload(changes Changes, notes string) ([]byte, error) {
	if len(changes) == 0 {
		return nil, nil
	}

	payload := map[string]interface{}{
		"notes":   notes,
		"changes": changes,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to build history payload")
		return nil, fmt.Errorf("failed to build history payload: %v", err)
	}

	return data, nil
}

// Helper functions
func getJSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return ""
	}
	// Ambil nama field sebelum koma (jika ada omitempty, dll)
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}

func getFieldValue(val reflect.Value, fieldName string) interface{} {
	if !val.IsValid() {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if getJSONFieldName(field) == fieldName {
			fieldVal := val.Field(i)

			// Handle pointer fields
			if fieldVal.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					return nil
				}
				return fieldVal.Elem().Interface()
			}

			return fieldVal.Interface()
		}
	}
	return nil
}

func isEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle time.Time comparison
	if ta, ok := a.(time.Time); ok {
		if tb, ok := b.(time.Time); ok {
			return ta.Equal(tb)
		}
		return false
	}

	// Handle basic types
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
