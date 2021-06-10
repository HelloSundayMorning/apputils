package gql

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// NewUnauthorizedError creates a new instance of Unauthorized error with the given context and message
func NewUnauthorizedError(ctx context.Context, message *string) *gqlerror.Error {

	return newError(ctx, message, http.StatusUnauthorized)
}

// NewForbiddenError creates a new instance of Forbidden error with the given context and message
func NewForbiddenError(ctx context.Context, message *string) *gqlerror.Error {

	return newError(ctx, message, http.StatusForbidden)
}

// NewResourceNotFoundError creates a new instance of ResourceNotFound error with the given context and message
// StatusResourceNotFound (4004) is used instead of http.StatusNotFound
func NewResourceNotFoundError(ctx context.Context, message *string) *gqlerror.Error {

	return newError(ctx, message, StatusResourceNotFound)
}

func newError(ctx context.Context, message *string, errorCode int) *gqlerror.Error {
	path := graphql.GetPath(ctx)

	return &gqlerror.Error{
		Path:    path,
		Message: *message,
		Extensions: map[string]interface{}{
			"code": errorCode,
		},
	}
}
