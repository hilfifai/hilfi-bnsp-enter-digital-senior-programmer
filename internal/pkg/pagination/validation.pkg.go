package pagination

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// validateFilterValue memvalidation nilai filter berdasarkan tipe data
func validateFilterValue(val string, config FieldConfig) bool {
	// Validasi panjang string
	if len(val) > MaxStringLength {
		return false
	}

	switch config.DataType {
	case "boolean":
		return val == "true" || val == "false"
	case "number":
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return false
		}
		// Validasi range angka
		return num >= -1e9 && num <= 1e9
	case "date":
		_, err := time.Parse(time.RFC3339, val)
		return err == nil
	case "string":
		return IsValidString(val)
	case "in_year":
		return IsValidString(val)
	case "uuid":
		_, err := uuid.Parse(val)
		if err == nil {
			return true
		}
		return false
	default:
		return false
	}
}

// getDefaultValueForType mengembalikan nilai default untuk tipe data
func getDefaultValueForType(config FieldConfig) string {
	switch config.DataType {
	case "boolean":
		return "false"
	case "number":
		return "0"
	case "date":
		return time.Now().Format(time.RFC3339)
	default:
		return ""
	}
}

// IsValidString memvalidation string untuk mencegah SQL injection
func IsValidString(s string) bool {
	// Pattern berbahaya yang tidak boleh ada dalam query
	patterns := []string{"--", "/*", "*/", ";;", ";", "@@"}
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return false
		}
	}

	words := strings.Fields(strings.ToUpper(s))
	keywords := []string{
		"DROP", "DELETE", "UPDATE", "TRUNCATE",
		"ALTER", "GRANT", "REVOKE", "EXECUTE",
	}

	for _, word := range words {
		cleanWord := strings.Trim(word, `"'()[];,`)
		for _, keyword := range keywords {
			if cleanWord == keyword {
				return false
			}
		}
	}
	return true
}

// isValidFieldName memvalidation nama field
func isValidFieldName(name string) bool {
	if name == "" {
		return false
	}

	// Hanya izinkan alfanumerik dan underscore
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}
	return true
}

// isValidTransformFunction memvalidation fungsi transform
func isValidTransformFunction(name string) bool {
	// Whitelist fungsi SQL yang diperbolehkan
	allowedFunctions := map[string]bool{
		"LOWER":    true,
		"UPPER":    true,
		"TRIM":     true,
		"COALESCE": true,
	}
	return allowedFunctions[strings.ToUpper(name)]
}

// isValidQueryString memvalidation string query
func isValidQueryString(query string) bool {
	if query == "" {
		return false
	}

	// Pattern berbahaya yang tidak boleh ada dalam query
	patterns := []string{"--", "/*", "*/", ";;"}
	for _, p := range patterns {
		if strings.Contains(query, p) {
			return false
		}
	}

	// Split query menjadi tokens
	words := strings.Fields(strings.ToUpper(query))

	// Keywords SQL berbahaya yang tidak boleh muncul sebagai kata yang berdiri sendiri
	keywords := []string{
		"DROP", "DELETE", "UPDATE", "TRUNCATE",
		"ALTER", "GRANT", "REVOKE", "EXECUTE",
	}

	for _, word := range words {
		// Skip kata dalam tanda kutip (nama kolom)
		if strings.HasPrefix(word, `"`) && strings.HasSuffix(word, `"`) {
			continue
		}
		if strings.HasPrefix(word, "'") && strings.HasSuffix(word, "'") {
			continue
		}

		// Bersihkan kata dari karakter khusus
		cleanWord := strings.Trim(word, `"'()[];,`)

		// Skip jika kata adalah bagian dari qualified name atau kosong
		if strings.Contains(cleanWord, ".") || cleanWord == "" {
			continue
		}

		// Skip jika kata adalah bagian dari nama kolom (updated_at, updated_by, dll)
		if strings.Contains(cleanWord, "_") {
			continue
		}

		// Cek apakah kata adalah keyword berbahaya yang berdiri sendiri
		for _, keyword := range keywords {
			if cleanWord == keyword {
				return false
			}
		}
	}

	return true
}
