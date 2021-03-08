package gql

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/gqlerror"
)

// NewUnauthorizedError creates a new instance of Unauthorized error with the given context and message
func NewUnauthorizedError(ctx context.Context, message *string) *gqlerror.Error {
	iPaths := getGQLPath(ctx)

	return newError(iPaths, message, http.StatusUnauthorized)
}

// NewForbiddenError creates a new instance of Forbidden error with the given context and message
func NewForbiddenError(ctx context.Context, message *string) *gqlerror.Error {
	iPaths := getGQLPath(ctx)

	return newError(iPaths, message, http.StatusForbidden)
}

func getGQLPath(ctx context.Context) []interface{} {
	paths := graphql.GetPath(ctx)
	iPaths := make([]interface{}, len(paths))
	for i, path := range paths {
		iPaths[i] = path
	}
	return iPaths
}

func newError(path []interface{}, message *string, errorCode int) *gqlerror.Error {
	return &gqlerror.Error{
		Path:    path,
		Message: *message,
		Extensions: map[string]interface{}{
			"code": errorCode,
		},
	}
}
