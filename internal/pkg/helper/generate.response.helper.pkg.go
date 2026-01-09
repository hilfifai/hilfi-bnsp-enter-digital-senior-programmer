package helper

import (
	_type "api-jellyfish/internal/common/type"
	"net/http"
)

func ParseResponse(r *_type.Response) *_type.Response {
	// Handle nil input
	if r == nil {
		return nil
	}

	if r.Code < 200 || r.Code >= 599 {
		r.Code = http.StatusInternalServerError
	}
	if r.Message == "" {
		generateMessage(r)
	}
	return r
}

func generateMessage(r *_type.Response) {
	switch r.Code {
	case http.StatusOK:
		r.Message = "Success"
	case http.StatusCreated:
		r.Message = "Created"
	case http.StatusBadRequest:
		r.Message = "Bad Request"
	case http.StatusUnauthorized:
		r.Message = "Unauthorized"
	case http.StatusForbidden:
		r.Message = "Forbidden"
	case http.StatusNotFound:
		r.Message = "Not Found"
	case http.StatusMethodNotAllowed:
		r.Message = "Method Not Allowed"
	case http.StatusInternalServerError:
		r.Message = "Internal Server Error"
	case http.StatusServiceUnavailable:
		r.Message = "Service Unavailable"
	default:
		r.Message = http.StatusText(r.Code)
	}
}
