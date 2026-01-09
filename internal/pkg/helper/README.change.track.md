# Change Track Helper

Helper package untuk melacak perubahan antara data lama dan data baru dengan fitur yang lebih lengkap daripada `detectChanges` yang ada.

## Fitur Utama

- **TrackChanges**: Fungsi utama untuk melacak perubahan antara dua struct
- **Functional Options Pattern**: Konfigurasi yang fleksibel dengan opsi skip fields dan custom mapper
- **BuildHistoryPayload**: Membentuk payload JSON untuk history tracking
- **Pointer Handling**: Otomatis menangani pointer fields
- **Time Comparison**: Khusus comparison untuk time.Time
- **JSON Tag Support**: Menggunakan JSON tags untuk field mapping

## Usage

### Basic Usage

```go
package service

import (
    "api-jellyfish/internal/pkg/helper"
    "api-jellyfish/internal/entity"
)

// Dalam service function
func (s *CustomerService) UpdateCustomer(ctx context.Context, id uuid.UUID, customer *entity.Customer, userID uuid.UUID) (*entity.CustomerFilterResponse, error) {
    existing, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Update customer...
    
    // Gunakan helper untuk track changes
    changes := helper.TrackChanges(existing, customer, 
        helper.WithSkipFields("UpdatedAt", "UpdatedBy", "CreatedAt", "CreatedBy"))
    
    if len(changes) > 0 {
        notes := "Customer Updated"
        payload, err := helper.BuildHistoryPayload(changes, notes)
        if err != nil {
            // Handle error
        }
        // Simpan ke history...
    }
    
    return updatedCustomer, nil
}
```

### Advanced Usage dengan Custom Mapper

```go
// Dengan custom mapper untuk lowercase comparison
changes := helper.TrackChanges(oldCustomer, newCustomer,
    helper.WithSkipFields("UpdatedAt", "UpdatedBy", "CreatedAt", "CreatedBy"),
    helper.WithCustomMapper("name", func(v interface{}) interface{} {
        if str, ok := v.(string); ok {
            return strings.ToLower(str)
        }
        return v
    }),
    helper.WithCustomMapper("code", func(v interface{}) interface{} {
        if str, ok := v.(string); ok {
            return strings.ToUpper(str)
        }
        return v
    }))
```

### Migration dari detectChanges

Contoh actual implementation di `customer_address.service.go`:

Sebelum (dengan detectChanges):

```go
// Old detectChanges function
func (s *CustomerAddressService) detectChanges(ctx context.Context, old, new *entity.CustomerAddress) map[string]interface{} {
    changes := make(map[string]interface{})
    oldVal := reflect.ValueOf(old).Elem()
    newVal := reflect.ValueOf(new).Elem()

    for i := 0; i < oldVal.NumField(); i++ {
        fieldName := oldVal.Type().Field(i).Name
        oldField := oldVal.Field(i).Interface()
        newField := newVal.Field(i).Interface()

        if fieldName == "UpdatedAt" || fieldName == "UpdatedBy" || fieldName == "CreatedAt" || fieldName == "CreatedBy" {
            continue
        }

        if !reflect.DeepEqual(oldField, newField) {
            changes[fieldName] = map[string]interface{}{
                "old": oldField,
                "new": newField,
            }
        }
    }

    return changes
}

// Usage di UpsertBulkAddresses
if operation == "upsert_updated" && existing != nil {
    change.Changes = s.detectChanges(ctx, existing, address)
}
```

Sesudah (dengan helper):

```go
// Usage di UpsertBulkAddresses (sudah diimplementasi)
if operation == "upsert_updated" && existing != nil {
    change.Old = existing
    change.New = address
    // Use helper to track changes
    trackChanges := helper.TrackChanges(existing, address,
        helper.WithSkipFields("UpdatedAt", "UpdatedBy", "CreatedAt", "CreatedBy"))
    // Convert to map[string]interface{} to maintain compatibility
    change.Changes = make(map[string]interface{})
    for field, changeData := range trackChanges {
        change.Changes[field] = map[string]interface{}{
            "old": changeData.Old,
            "new": changeData.New,
        }
    }
}
```

## Struktur Output

Output dari `TrackChanges` adalah `Changes` (map[string]ChangeTrack):

```go
type ChangeTrack struct {
    Old interface{} `json:"old"`
    New interface{} `json:"new"`
}

type Changes map[string]ChangeTrack
```

Contoh output:

```json
{
    "name": {
        "old": "Old Name",
        "new": "New Name"
    },
    "code": {
        "old": "OLD",
        "new": "NEW"
    },
    "is_active": {
        "old": true,
        "new": false
    }
}
```

## Benefits dibanding detectChanges

1. **Reusable**: Bisa digunakan di semua service tanpa duplikasi code
2. **Configurable**: Functional options pattern untuk fleksibilitas
3. **Customizable**: Custom mapper untuk logika comparison khusus
4. **Better Error Handling**: Integrated logging dan error handling
5. **Pointer Safe**: Handle pointer fields dengan aman
6. **JSON Support**: Menggunakan JSON tags untuk field mapping
7. **Tested**: Comprehensive test suite
8. **Documented**: Well documented dengan examples

## Integration

Helper ini sudah terintegrasi dengan:
- Logger v2 package untuk error logging
- Standard Go reflection package
- JSON marshaling/unmarshaling
- Time comparison handling
- Pointer dereferencing

## Testing

```bash
go test ./internal/pkg/helper/ -v -run "TestTrackChanges|TestBuildHistoryPayload"
```

Semua test sudah passing dan coverage lengkap untuk semua functionality.