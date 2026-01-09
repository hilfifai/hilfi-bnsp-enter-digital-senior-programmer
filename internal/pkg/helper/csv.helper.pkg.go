package helper

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func GetFieldNames(record map[string]interface{}) []string {
	var fields []string
	for field := range record {
		fields = append(fields, field)
	}
	return fields
}

func GenerateHeaders(fields []string) []string {
	headers := make([]string, len(fields))
	for i, field := range fields {
		headers[i] = FormatHeaderName(field)
	}
	return headers
}

func FormatHeaderName(fieldName string) string {
	name := strings.ReplaceAll(fieldName, "_", " ")
	name = strings.Title(strings.ToLower(name))
	name = strings.ReplaceAll(name, "Id", "ID")
	name = strings.ReplaceAll(name, "Sap", "SAP")
	return name
}

func MapToCSVRecord(record map[string]interface{}, fields []string) []string {
	csvRecord := make([]string, len(fields))

	// Ensure fields are sorted according to the already sorted fields
	for i, field := range fields {
		value := record[field]
		csvRecord[i] = FormatValue(value)
	}

	return csvRecord
}

func FormatCSVRow(row []interface{}) []string {
	record := make([]string, len(row))
	for i, value := range row {
		record[i] = FormatValue(value)
	}
	return record
}

func FormatValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int, int64, int32, float64, float32:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case *int, *int64, *float64:
		val := reflect.ValueOf(value)
		if val.IsNil() {
			return ""
		}
		return fmt.Sprintf("%v", val.Elem().Interface())
	default:
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
		return fmt.Sprintf("%v", value)
	}
}
