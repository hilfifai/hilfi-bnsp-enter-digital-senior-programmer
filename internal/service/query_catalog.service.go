package service

import (
	"api-jellyfish/internal/entity"
	"api-jellyfish/internal/pkg/logger/v2"
	repository "api-jellyfish/internal/repository"
	"context"
	"encoding/json"
	"fmt"
)

type QueryService interface {
	Execute(ctx context.Context, code string, queryParams map[string][]string) (any, error)
}

type queryService struct {
	repo repository.QueryRepository
}

func NewQueryService(repo repository.QueryRepository) QueryService {
	return &queryService{repo}
}

func (s *queryService) Execute(ctx context.Context, code string, queryParams map[string][]string) (any, error) {
	log := logger.FromContext(ctx)
	catalog, err := s.repo.GetCatalogByCode(ctx, code)
	if err != nil {
		log.Error().Err(err).Msg("Error getting catalog by code")
		return nil, err
	}

	// Unmarshal allowed params
	var allowed []string
	err = json.Unmarshal(catalog.AllowParams, &allowed)
	if err != nil {
		log.Error().Err(err).Msg("Error unmarshaling allowed params")
		return nil, err
	}

	params := map[string]interface{}{}
	for _, key := range allowed {
		vals := queryParams[key]
		if len(vals) > 0 {
			params[key] = vals[0]
		} else {
			params[key] = nil
		}
	}

	// Pagination
	pag := entity.Pagination{Page: 1, PageSize: 10}
	if v, ok := queryParams["page"]; ok && len(v) > 0 {
		fmt.Sscanf(v[0], "%d", &pag.Page)
	}
	if v, ok := queryParams["page_size"]; ok && len(v) > 0 {
		fmt.Sscanf(v[0], "%d", &pag.PageSize)
	}
	if pag.Page < 1 {
		pag.Page = 1
	}
	if pag.PageSize < 1 {
		pag.PageSize = 10
	}

	results, total, err := s.repo.ExecuteQuery(ctx, catalog, params, pag)
	if err != nil {
		log.Error().Err(err).Msg("Error executing query catalog")
		return nil, err
	}

	if catalog.AllowPagination {
		pagination := map[string]interface{}{
			"data":        results,
			"total":       total,
			"page":        pag.Page,
			"page_size":   pag.PageSize,
			"total_pages": (total + pag.PageSize - 1) / pag.PageSize,
		}
		return pagination, nil
	}

	return results, nil
}
