package repository

import (
	entity "api-jellyfish/internal/entity"
	database "api-jellyfish/internal/pkg/db"
	"api-jellyfish/internal/pkg/logger/v2"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type QueryRepository interface {
	GetCatalogByCode(ctx context.Context, code string) (*entity.QueryCatalog, error)
	ExecuteQuery(ctx context.Context, catalog *entity.QueryCatalog, params map[string]interface{}, pag entity.Pagination) ([]map[string]interface{}, int, error)
}

type queryRepository struct {
	db *database.Database
}

func NewQueryRepository(db *database.Database) QueryRepository {
	return &queryRepository{db}
}

func (r *queryRepository) GetCatalogByCode(ctx context.Context, code string) (*entity.QueryCatalog, error) {
	log := logger.FromContext(ctx)
	var catalog entity.QueryCatalog
	if err := r.db.DB.GetContext(ctx, &catalog, "SELECT * FROM m_query_catalog WHERE code = $1 and is_active is true", code); err != nil {
		log.Error().Err(err).Msg("Error executing query")
		return nil, err
	}
	return &catalog, nil
}

func (r *queryRepository) ExecuteQuery(ctx context.Context, catalog *entity.QueryCatalog, params map[string]interface{}, pag entity.Pagination) ([]map[string]interface{}, int, error) {
	log := logger.FromContext(ctx)

	finalQuery, countQuery := r.buildQueries(catalog, pag)
	total, err := r.getTotal(ctx, catalog, countQuery, params)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.executeMainQuery(ctx, finalQuery, params)
	if err != nil {
		log.Error().Err(err).Msg("Error executing query")
		return nil, 0, err
	}
	defer rows.Close()

	results, err := r.scanResults(ctx, rows)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *queryRepository) buildQueries(catalog *entity.QueryCatalog, pag entity.Pagination) (string, string) {
	finalQuery := catalog.SQLText
	countQuery := ""

	if catalog.AllowPagination {
		finalQuery = fmt.Sprintf("%s LIMIT %d OFFSET %d",
			strings.TrimSpace(catalog.SQLText),
			pag.PageSize,
			(pag.Page-1)*pag.PageSize,
		)
		countQuery = fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS sub", catalog.SQLText)
	}

	return finalQuery, countQuery
}

func (r *queryRepository) getTotal(ctx context.Context, catalog *entity.QueryCatalog, countQuery string, params map[string]interface{}) (int, error) {
	if !catalog.AllowPagination {
		return 0, nil
	}

	log := logger.FromContext(ctx)
	var total int

	query, args, err := sqlx.Named(countQuery, params)
	if err != nil {
		log.Error().Err(err).Msg("Error preparing count query")
		return 0, err
	}

	query = r.db.DB.Rebind(query)
	if err := r.db.DB.GetContext(ctx, &total, query, args...); err != nil && err != sql.ErrNoRows {
		log.Error().Err(err).Msg("Error executing count query")
		return 0, err
	}

	return total, nil
}

func (r *queryRepository) executeMainQuery(ctx context.Context, finalQuery string, params map[string]interface{}) (*sqlx.Rows, error) {
	if len(params) == 0 {
		return r.db.DB.QueryxContext(ctx, finalQuery)
	}
	return r.db.DB.NamedQueryContext(ctx, finalQuery, params)
}

func (r *queryRepository) scanResults(ctx context.Context, rows *sqlx.Rows) ([]map[string]interface{}, error) {
	log := logger.FromContext(ctx)
	results := []map[string]interface{}{}

	for rows.Next() {
		row := make(map[string]interface{})
		if err := rows.MapScan(row); err != nil {
			log.Error().Err(err).Msg("Error scanning row")
			return nil, err
		}

		// Handle JSON data types - convert byte arrays to proper JSON
		r.processJSONFields(row)
		results = append(results, row)
	}

	return results, nil
}

func (r *queryRepository) processJSONFields(row map[string]interface{}) {
	for key, value := range row {
		if value == nil {
			continue
		}

		if newValue := r.processValueType(value); newValue != nil {
			row[key] = newValue
		}
	}
}

func (r *queryRepository) processValueType(value interface{}) interface{} {
	// Handle byte arrays (common for JSON/JSONB columns)
	if bytes, ok := value.([]byte); ok {
		return r.processByteArray(bytes)
	}

	// Handle driver.Valuer interface (like pq.NullTime, etc.)
	if valuer, ok := value.(interface{ Value() (interface{}, error) }); ok {
		if val, err := valuer.Value(); err == nil {
			return val
		}
	}

	// Handle SQL null types
	return r.processSQLNullTypes(value)
}

func (r *queryRepository) processByteArray(bytes []byte) interface{} {
	// Try to parse as JSON first
	var jsonValue interface{}
	if err := json.Unmarshal(bytes, &jsonValue); err == nil {
		return jsonValue
	}
	// If not valid JSON, keep as string
	return string(bytes)
}

func (r *queryRepository) processSQLNullTypes(value interface{}) interface{} {
	switch v := value.(type) {
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return nil
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return nil
	case sql.NullFloat64:
		if v.Valid {
			return v.Float64
		}
		return nil
	case sql.NullBool:
		if v.Valid {
			return v.Bool
		}
		return nil
	case sql.NullTime:
		if v.Valid {
			return v.Time
		}
		return nil
	default:
		return value
	}
}
