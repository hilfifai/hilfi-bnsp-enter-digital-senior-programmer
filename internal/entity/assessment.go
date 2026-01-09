// internal/entity/assessment.go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type StringArray []string

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("StringArray: unsupported type %T", value)
	}

	str = strings.TrimSpace(str)
	if str == "" {
		*a = []string{}
		return nil
	}

	// Case 1: JSON array
	if strings.HasPrefix(str, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(str), &arr); err != nil {
			return fmt.Errorf("StringArray: invalid JSON array: %w", err)
		}
		*a = arr
		return nil
	}

	// Case 2: Postgres array literal, e.g. {a,b,c}
	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		// Remove { and }
		inner := str[1 : len(str)-1]
		if inner == "" {
			*a = []string{}
			return nil
		}
		// Split by comma
		parts := strings.Split(inner, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			parts[i] = strings.Trim(parts[i], `"`) // remove quotes if any
		}
		*a = parts
		return nil
	}

	// Case 3: Unexpected formats
	return fmt.Errorf("StringArray: invalid format %s", str)
}
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	if len(a) == 0 {
		return "{}", nil // kosong tapi valid untuk Postgres array
	}

	// buat format array literal PostgreSQL: {a,b,c}
	var b strings.Builder
	b.WriteByte('{')
	for i, v := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		// escape tanda kutip jika ada
		if strings.ContainsAny(v, `," {}`) {
			b.WriteString(`"`)
			b.WriteString(strings.ReplaceAll(v, `"`, `\"`))
			b.WriteString(`"`)
		} else {
			b.WriteString(v)
		}
	}
	b.WriteByte('}')
	return b.String(), nil
}

type Assessment struct {
	ID              uuid.UUID   `db:"id" json:"id"`
	ModuleID        uuid.UUID   `db:"module_id" json:"module_id"`
	Name            string      `db:"name" json:"name"`
	Ord             int         `db:"ord" json:"ord"`
	IsActive        bool        `db:"is_active" json:"is_active"`
	Code            string      `db:"code" json:"code"`
	CreatedBy       uuid.UUID   `db:"created_by" json:"created_by"`
	UpdatedBy       uuid.UUID   `db:"updated_by" json:"updated_by"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
	AssignedRoleIDs StringArray `db:"assigned_role_ids" json:"assigned_role_ids,omitempty"`
}

type AssessmentItem struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	AssessmentID uuid.UUID  `db:"assessment_id" json:"assessment_id"`
	Code         string     `db:"code" json:"code"`
	Name         string     `db:"name" json:"name"`
	FileID       *uuid.UUID `db:"file_id" json:"file_id,omitempty"`
	Filename     *string    `db:"filename" json:"filename,omitempty"`
	Ord          int        `db:"ord" json:"ord"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedBy    uuid.UUID  `db:"created_by" json:"-"`
	UpdatedBy    uuid.UUID  `db:"updated_by" json:"-"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

type AssessmentWithItems struct {
	Assessment
	Items []AssessmentItem `json:"items,omitempty"`
}
