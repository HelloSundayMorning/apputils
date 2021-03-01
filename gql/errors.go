package gql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

// UnauthorizedErrorCode 401 indicates an unauthorized request
const UnauthorizedErrorCode = 401

// NewUnaothorizedError creates a new instance of Unauthorized error with the given context and message
func NewUnaothorizedError(ctx context.Context, message *string) *gqlerror.Error {
	paths := graphql.GetPath(ctx)
	iPaths := make([]interface{}, len(paths))
	for i, path := range paths {
		iPaths[i] = path
	}

	return &gqlerror.Error{
		Path:    iPaths,
		Message: *message,
		Extensions: map[string]interface{}{
			"code": UnauthorizedErrorCode,
		},
	}
}
