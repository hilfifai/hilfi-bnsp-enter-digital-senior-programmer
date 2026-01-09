package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	types "api-jellyfish/internal/common/type"
	"api-jellyfish/internal/pkg/logger/v2"
	"api-jellyfish/internal/pkg/pagination"
	"api-jellyfish/internal/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// HTTP Status Code Constants
const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusNoContent           = http.StatusNoContent
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusConflict            = http.StatusConflict
	StatusUnprocessableEntity = http.StatusUnprocessableEntity
	StatusTooManyRequests     = http.StatusTooManyRequests
	StatusInternalServerError = http.StatusInternalServerError
	StatusServiceUnavailable  = http.StatusServiceUnavailable
	StatusTemporaryRedirect   = http.StatusTemporaryRedirect
)

// Default Response Messages
const (
	MsgSuccess               = "Success"
	MsgCreated               = "Created successfully"
	MsgBadRequest            = "Bad request"
	MsgUnauthorized          = "Unauthorized"
	MsgForbidden             = "Forbidden"
	MsgNotFound              = "Not found"
	MsgConflict              = "Conflict"
	MsgUnprocessableEntity   = "Unprocessable entity"
	MsgInternalServerError   = "Internal server error"
	MsgServiceUnavailable    = "Service unavailable"
	MsgValidationFailed      = "Validation failed"
	MsgInvalidRequest        = "Invalid request"
	MsgValidationSystemError = "Validation system error"
	MsgTooManyRequests       = "Too many requests, please try again later"
	MsgMaintenanceMode       = "Service is currently under maintenance"
	MsgTokenExpired          = "Token has expired"
	MsgInvalidToken          = "Invalid or malformed token"
	MsgInsufficientPerms     = "Insufficient permissions to perform this action"
	MsgFileUploaded          = "File uploaded successfully"
	MsgFileNotFound          = "File not found or cannot be accessed"
)

// Context Keys
const (
	ContextKeyDebug     = "debug"
	ContextKeyStartTime = "start-time"
	ContextKeyRequestID = "request_id"
	ContextKeyVersion   = "version"
)

// Depricated dont use in handler directly APIResponse sends a standardized API response with optional debug information
func APIResponse(c *gin.Context, statusCode int, message string, data interface{}, err error) {
	if c == nil {
		logger.Log.Error().Msg("gin.Context is nil in APIResponse")
		return
	}
	// Get logger from context if available, fallback to default logger
	var log zerolog.Logger
	if c.Request != nil {
		log = logger.FromContext(c.Request.Context())
	} else {
		log = logger.Log
	}

	shouldDebug := gin.Mode() == gin.DebugMode

	response := types.ResponseAPI{
		Data:    data,
		Message: message,
	}

	// Add debug information in debug mode
	if shouldDebug {
		debug := buildDebugInfo(c, err)
		response.Debug = debug
	}

	// Log error if present
	if err != nil {
		log.Error().
			Err(err).
			Int("status_code", statusCode).
			Str("message", message).
			Msg("API error response")
	}

	c.AbortWithStatusJSON(statusCode, response)
}

// buildDebugInfo constructs debug information for API responses
func buildDebugInfo(c *gin.Context, err error) *types.ResponseAPIDebug {
	var debug *types.ResponseAPIDebug

	if val, exists := c.Get(ContextKeyDebug); exists && val != nil {
		debug = val.(*types.ResponseAPIDebug)
		debug.EndTime = time.Now()
		debug.RuntimeMs = debug.EndTime.Sub(debug.StartTime).Milliseconds()
	} else {
		startTime := getStartTime(c)
		endTime := time.Now()
		debug = &types.ResponseAPIDebug{
			RequestID: c.GetString(ContextKeyRequestID),
			Version:   c.GetString(ContextKeyVersion),
			StartTime: startTime,
			EndTime:   endTime,
			RuntimeMs: endTime.Sub(startTime).Milliseconds(),
		}
	}

	// Set error message if exists
	if err != nil {
		errMsg := err.Error()
		debug.Error = &errMsg
	}

	return debug
}

// getStartTime retrieves start time from context or returns current time
func getStartTime(c *gin.Context) time.Time {
	if value, exists := c.Get(ContextKeyStartTime); exists && value != nil {
		if t, ok := value.(time.Time); ok {
			return t
		}
	}
	return time.Now()
}

// Success Response Handlers

// APISuccess returns a successful response with 200 status code
func APISuccess(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = MsgSuccess
	}
	APIResponse(c, StatusOK, message, data, nil)
}

// APICreated returns a created response with 201 status code
func APICreated(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = MsgCreated
	}
	APIResponse(c, StatusCreated, message, data, nil)
}

// APINoContent returns 204 No Content response
func APINoContent(c *gin.Context) {
	if c == nil {
		logger.Log.Error().Msg("gin.Context is nil in APINoContent")
		return
	}
	c.AbortWithStatus(StatusNoContent)
}

// Error Response Handlers

// APIBadRequest returns a bad request error with 400 status code
func APIBadRequest(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgBadRequest
	}
	APIResponse(c, StatusBadRequest, message, nil, err)
}

// APIUnauthorized returns an unauthorized error with 401 status code
func APIUnauthorized(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgUnauthorized
	}
	APIResponse(c, StatusUnauthorized, message, nil, err)
}

// APIForbidden returns a forbidden error with 403 status code
func APIForbidden(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgForbidden
	}
	APIResponse(c, StatusForbidden, message, nil, err)
}

// APINotFound returns a not found error with 404 status code
func APINotFound(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgNotFound
	}
	APIResponse(c, StatusNotFound, message, nil, err)
}

// APIConflict returns a conflict error with 409 status code
func APIConflict(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgConflict
	}
	APIResponse(c, StatusConflict, message, nil, err)
}

// APIUnprocessableEntity returns an unprocessable entity error with 422 status code
func APIUnprocessableEntity(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgUnprocessableEntity
	}
	APIResponse(c, StatusUnprocessableEntity, message, nil, err)
}

// APIInternalServerError returns an internal server error with 500 status code
func APIInternalServerError(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgInternalServerError
	}
	APIResponse(c, StatusInternalServerError, message, nil, err)
}

// APIServiceUnavailable returns a service unavailable error with 503 status code
func APIServiceUnavailable(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgServiceUnavailable
	}
	APIResponse(c, StatusServiceUnavailable, message, nil, err)
}

// APIRateLimitExceeded returns rate limit error with 429 status code
func APIRateLimitExceeded(c *gin.Context) {
	APIResponse(c, StatusTooManyRequests, MsgTooManyRequests, nil, nil)
}

// Validation Error Handlers

// APIValidationError returns a validation error response with structured error data
func APIValidationError(c *gin.Context, message string, validationErr error) {
	if message == "" {
		message = MsgValidationFailed
	}

	if validationErr == nil {
		APIBadRequest(c, message, errors.New("validation error is nil"))
		return
	}

	validationErrors := validation.ValidationErrorResponse(validationErr)
	APIResponse(c, StatusBadRequest, message, validationErrors, validationErr)
}

// APIBindingError handles both JSON binding and validation errors consistently
func APIBindingError(c *gin.Context, err error) {
	if err == nil {
		APIBadRequest(c, MsgInvalidRequest, errors.New("binding error is nil"))
		return
	}

	validationErrors := validation.ValidationErrorResponse(err)
	APIResponse(c, StatusBadRequest, MsgInvalidRequest, validationErrors, err)
}

// APISystemValidationError handles system-level validation errors (500)
func APISystemValidationError(c *gin.Context, message string, err error) {
	if message == "" {
		message = MsgValidationSystemError
	}
	APIInternalServerError(c, message, err)
}

// Advanced Response Handlers

// APIErrorWithData returns an error response with custom data
func APIErrorWithData(c *gin.Context, statusCode int, message string, data interface{}, err error) {
	APIResponse(c, statusCode, message, data, err)
}

// APICustomError returns custom error with specific status and message
func APICustomError(c *gin.Context, statusCode int, message string, err error) {
	APIResponse(c, statusCode, message, nil, err)
}

// Pagination Response Handlers

// APIPaginated returns a paginated response using standard pagination structure
func APIPaginated(c *gin.Context, message string, data interface{}, page, pageSize, total int) {
	if message == "" {
		message = "Data retrieved successfully"
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if total < 0 {
		total = 0
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	paginatedData := map[string]interface{}{
		"data":        data,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}

	APIResponse(c, StatusOK, message, paginatedData, nil)
}

// APIPaginatedFromStruct returns a paginated response using PaginatedResponse struct directly
func APIPaginatedFromStruct[T any](c *gin.Context, message string, paginatedResponse pagination.PaginatedResponse[T]) {
	if message == "" {
		message = "Data retrieved successfully"
	}

	APIResponse(c, StatusOK, message, paginatedResponse, nil)
}

// APIPaginatedWithCustomMeta returns a paginated response with additional metadata
func APIPaginatedWithCustomMeta(c *gin.Context, message string, data interface{}, page, pageSize, total int, meta map[string]interface{}) {
	if message == "" {
		message = "Data retrieved successfully"
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if total < 0 {
		total = 0
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	response := map[string]interface{}{
		"data":        data,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"meta":        meta,
	}

	APIResponse(c, StatusOK, message, response, nil)
}

// APISuccessWithMeta returns success response with additional metadata
func APISuccessWithMeta(c *gin.Context, message string, data interface{}, meta map[string]interface{}) {
	if message == "" {
		message = MsgSuccess
	}

	response := map[string]interface{}{
		"data": data,
		"meta": meta,
	}

	APIResponse(c, StatusOK, message, response, nil)
}

// Service Error Handler

// APIErrorFromService handles common service layer errors with intelligent error mapping
func APIErrorFromService(c *gin.Context, err error) {
	if err == nil {
		APISuccess(c, "", nil)
		return
	}

	errMsg := err.Error()
	switch errMsg {
	case "not_found", "record_not_found", "user_not_found":
		APINotFound(c, "", err)
	case "unauthorized", "invalid_token", "token_expired":
		APIUnauthorized(c, "", err)
	case "forbidden", "access_denied", "insufficient_permissions":
		APIForbidden(c, "", err)
	case "conflict", "already_exists", "duplicate_entry":
		APIConflict(c, "", err)
	case "validation_failed", "invalid_input":
		APIBadRequest(c, "", err)
	default:
		APIInternalServerError(c, "", err)
	}
}

// CRUD Operation Helpers

// APICreateSuccess returns standardized create response
func APICreateSuccess(c *gin.Context, resourceName string, data interface{}) {
	if resourceName == "" {
		resourceName = "Resource"
	}
	message := resourceName + " created successfully"
	APICreated(c, message, data)
}

// APIUpdateSuccess returns standardized update response
func APIUpdateSuccess(c *gin.Context, resourceName string, data interface{}) {
	if resourceName == "" {
		resourceName = "Resource"
	}
	message := resourceName + " updated successfully"
	APISuccess(c, message, data)
}

// APIDeleteSuccess returns standardized delete response
func APIDeleteSuccess(c *gin.Context, resourceName string) {
	if resourceName == "" {
		resourceName = "Resource"
	}
	message := resourceName + " deleted successfully"
	APISuccess(c, message, nil)
}

// APIGetSuccess returns standardized get response
func APIGetSuccess(c *gin.Context, resourceName string, data interface{}) {
	if resourceName == "" {
		resourceName = "Resource"
	}
	message := resourceName + " retrieved successfully"
	APISuccess(c, message, data)
}

// Field Validation Helpers

// APIRequiredField returns error for missing required field
func APIRequiredField(c *gin.Context, fieldName string) {
	if fieldName == "" {
		fieldName = "unknown"
	}
	message := "Field '" + fieldName + "' is required"
	APIBadRequest(c, message, errors.New("required field missing: "+fieldName))
}

// APIInvalidField returns error for invalid field value
func APIInvalidField(c *gin.Context, fieldName string, reason string) {
	if fieldName == "" {
		fieldName = "unknown"
	}
	message := "Field '" + fieldName + "' is invalid"
	if reason != "" {
		message += ": " + reason
	}
	APIBadRequest(c, message, errors.New("invalid field: "+fieldName))
}

// Resource Helpers

// APIResourceNotFound returns standardized not found error
func APIResourceNotFound(c *gin.Context, resourceName string) {
	if resourceName == "" {
		resourceName = "Resource"
	}
	message := resourceName + " not found"
	APINotFound(c, message, errors.New("resource not found: "+resourceName))
}

// APIResourceAlreadyExists returns standardized conflict error
func APIResourceAlreadyExists(c *gin.Context, resourceName string) {
	if resourceName == "" {
		resourceName = "Resource"
	}
	message := resourceName + " already exists"
	APIConflict(c, message, errors.New("resource already exists: "+resourceName))
}

// File Operation Helpers

// APIFileUploadSuccess returns file upload success response
func APIFileUploadSuccess(c *gin.Context, fileInfo interface{}) {
	APICreated(c, MsgFileUploaded, fileInfo)
}

// APIFileDownloadError returns file download error
func APIFileDownloadError(c *gin.Context, err error) {
	APINotFound(c, MsgFileNotFound, err)
}

// Authentication and Authorization Helpers

// APITokenExpired returns token expired error
func APITokenExpired(c *gin.Context) {
	APIUnauthorized(c, MsgTokenExpired, errors.New("token expired"))
}

// APIInvalidToken returns invalid token error
func APIInvalidToken(c *gin.Context) {
	APIUnauthorized(c, MsgInvalidToken, errors.New("invalid token"))
}

// APIInsufficientPermissions returns insufficient permissions error
func APIInsufficientPermissions(c *gin.Context) {
	APIForbidden(c, MsgInsufficientPerms, errors.New("insufficient permissions"))
}

// Utility Response Helpers

// APIMaintenanceMode returns maintenance mode error
func APIMaintenanceMode(c *gin.Context) {
	APIServiceUnavailable(c, MsgMaintenanceMode, errors.New("maintenance mode"))
}

// JSON Utility Functions

// JSONToString converts any payload to JSON string with proper error handling
func JSONToString(payload interface{}) (string, error) {
	if payload == nil {
		return "", errors.New("payload is nil")
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal payload to JSON")
		return "", err
	}

	return string(jsonBytes), nil
}

// JSONToStruct converts any payload to specified struct type with proper error handling
func JSONToStruct[T any](payload interface{}) (*T, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal payload to JSON")
		return nil, err
	}

	var result T
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to unmarshal JSON to struct")
		return nil, err
	}

	return &result, nil
}

// JSONToByte converts any payload to JSON byte array with proper error handling
func JSONToByte(payload interface{}) ([]byte, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal payload to JSON")
		return nil, err
	}

	return jsonBytes, nil
}

// JSONStringToStruct converts JSON string to specified struct type
func JSONStringToStruct[T any](jsonStr string) (*T, error) {
	if jsonStr == "" {
		return nil, errors.New("JSON string is empty")
	}

	var result T
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		logger.Log.Error().Err(err).Str("json", jsonStr).Msg("Failed to unmarshal JSON string to struct")
		return nil, err
	}

	return &result, nil
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(jsonStr string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// Helper function to safely get string from interface{}
func SafeStringFromInterface(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// Helper function to safely get int from interface{}
func SafeIntFromInterface(value interface{}, defaultValue int) int {
	if value == nil {
		return defaultValue
	}
	if i, ok := value.(int); ok {
		return i
	}
	if f, ok := value.(float64); ok {
		return int(f)
	}
	return defaultValue
}
