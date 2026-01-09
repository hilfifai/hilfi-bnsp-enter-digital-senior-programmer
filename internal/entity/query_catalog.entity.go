package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type QueryCatalog struct {
	ID              uuid.UUID       `db:"id"`
	Code            string          `db:"code"`
	Description     string          `db:"description"`
	SQLText         string          `db:"sql_text"`
	AllowParams     json.RawMessage `db:"allow_params"`
	AllowPagination bool            `db:"allow_pagination"`
	CreatedAt       *time.Time      `db:"created_at"`
	IsActive        bool            `db:"is_active"`
}

type Pagination struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
