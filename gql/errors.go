package gql

import (
	"github.com/vektah/gqlparser/gqlerror"
)

// UnauthorizedErrorCode 401 indicates an unauthorized request
const UnauthorizedErrorCode = 401

// GetUnauthorizedError creates a new instance of Unauthorized error with the given path and message
func GetUnauthorizedError(path *[]interface{}, message *string) *gqlerror.Error {
	return &gqlerror.Error{
		Path:    *path,
		Message: *message,
		Extensions: map[string]interface{}{
			"code": UnauthorizedErrorCode,
		},
	}
}
