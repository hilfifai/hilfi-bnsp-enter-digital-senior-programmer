package query_catalog

import (
	"api-jellyfish/internal/pkg/helper"
	"api-jellyfish/internal/pkg/jwt"
	"api-jellyfish/internal/service"

	"github.com/gin-gonic/gin"
)

type IHandler interface {
	NewRoutes(e *gin.RouterGroup)
	ExecuteQuery(c *gin.Context)
}

type Handler struct {
	auth    jwt.IJWTAuth
	service service.QueryService
}

func NewQueryHandler(auth jwt.IJWTAuth, service service.QueryService) IHandler {
	return &Handler{auth: auth, service: service}
}

// GetAll
// @Summary Query Catalog Execution
// @Description Execute a query from the catalog
// @Tags Query Catalog
// @Produce json
// @Param code path string true "Catalog Code"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} types.ResponseAPI{data=any} "Successful operation"
// @Failure 500 {object} types.ResponseAPI "Internal server error"
// @Router /api/v1/query/{code} [get]
// @Security BearerAuth
func (h *Handler) ExecuteQuery(c *gin.Context) {
	code := c.Param("code")
	ctx := c.Request.Context()

	resp, err := h.service.Execute(ctx, code, c.Request.URL.Query())
	if err != nil {
		helper.APIBadRequest(c, "", err)
		return
	}
	helper.APISuccess(c, "", resp)
}
