package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql"
	gqlHandlers "github.com/99designs/gqlgen/graphql/handler"
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/HelloSundayMorning/apputils/tracing"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gofrs/uuid"
	gHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/net/context"
)

type (
	Initialize func(srv *AppServer) (err error)
	CleanUp    func(srv *AppServer) (err error)

	// AppServer
	// Application Server object that controls the application state and life cycle.
	// It's based on the http Server from net/http package, and offers the ability to register HTTP routes.
	// The router is Handler uses the gorilla/mux implementation
	// Initialization, Shutdown and Cleanup is managed by the AppServer. Custom functions for initialization and
	// cleanup are provided so the life cycle of other objects can be added to it.
	AppServer struct {
		*http.Server
		gqlServer      *gqlHandlers.Server
		AppID          app.ApplicationID // Unique identifier for the application
		initializeFunc Initialize        // Custom initialization function
		cleanupFunc    CleanUp           // Custom cleanup function
		corsOrigins    []string          // Enable CORS and set origins
		environment    string            // The environment name
	}
)

const (
	component = "server"
)

var (
	getAppEnv = func() string {
		return os.Getenv(app.AppEnvironmentEnv)
	}
)

// NewServer
// Create a new Application Server instance.
func NewServer(appID app.ApplicationID, port int) *AppServer {

	env := getAppEnv()

	if env != app.LocalEnvironment && env != app.StagingEnvironment && env != app.ProductionEnvironment {
		log.FatalfNoContext(appID, component, "Cannot create a new server, env variable %s invalid. Expected %s | %s | %s", app.AppEnvironmentEnv, app.LocalEnvironment, app.StagingEnvironment, app.ProductionEnvironment)
	}

	router := mux.NewRouter()

	router.Use()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	server := &AppServer{
		Server:      httpServer,
		AppID:       appID,
		environment: env,
	}

	return server
}

// NewServerWithInitialization
// Create a new Application Server instance, with Custom initialization and cleanup functions
func NewServerWithInitialization(appID app.ApplicationID, port int, initializeFunc Initialize, cleanupFunc CleanUp, corsOrigins ...string) *AppServer {

	server := NewServer(appID, port)

	server.initializeFunc = initializeFunc
	server.cleanupFunc = cleanupFunc

	if len(corsOrigins) > 0 {
		server.corsOrigins = corsOrigins
	}

	return server
}

// EnableAWSXrayTracing
// Set env variable AWS_XRAY_SDK_DISABLED to false to enable AWS XRay Tracing
// Must be called before server.Start()
func (srv *AppServer) EnableAWSXrayTracing() {
	err := os.Setenv(app.AwsXrayDisableEnv, "FALSE")

	if err != nil {
		log.FatalfNoContext(srv.AppID, component, "Error setting env variable %s, %s", app.AwsXrayDisableEnv, err)
	}
}

// DisableAWSXrayTracing
// Set env variable AWS_XRAY_SDK_DISABLED to true to disable AWS XRay Tracing
// Must be called before server.Start()
func (srv *AppServer) DisableAWSXrayTracing() {
	err := os.Setenv(app.AwsXrayDisableEnv, "TRUE")

	if err != nil {
		log.FatalfNoContext(srv.AppID, component, "Error setting env variable %s, %s", app.AwsXrayDisableEnv, err)
	}
}

// AddRoute
// Add http route to the server with a method.
// The path is added after the server appID ie. /appID/path
// path - the HTTP path route
// method - the HTTP verb
// handler - next HTTP handler called if user is authorized
func (srv *AppServer) AddRoute(path, method string, handler http.HandlerFunc) error {

	path = fmt.Sprintf("/%s%s", srv.AppID, path)

	srv.router().HandleFunc(path, srv.requestInterceptor(handler)).Methods(method)

	log.PrintfNoContext(srv.AppID, component, "Added route %s %s for app %s", method, path, srv.AppID)

	return nil
}

// AddRouteNoAppID
// Add http route to the server with a method.
// The path is added without appID ie. /path
// path - the HTTP path route
// method - the HTTP verb
// handler - next HTTP handler called if user is authorized
func (srv *AppServer) AddRouteNoAppID(path, method string, handler http.HandlerFunc) error {

	srv.router().HandleFunc(path, srv.requestInterceptor(handler)).Methods(method)

	log.PrintfNoContext(srv.AppID, component, "Added route %s %s for app %s", method, path, srv.AppID)

	return nil
}

// AddAuthorizedRoute
// Add http route to the server with a method. It enforces the existence of an authorized user in the context
// and validates roles for the authorized user. Return 401 if no authorized user is present or 403
// if the authorized user doesn't have any of the required roles.
// The path is added after the server appID ie. /appID/path
// path - the HTTP path route
// method - the HTTP verb
// authorizedRoles - list of roles the user must have one to be authorized
// handler - next HTTP handler called if user is authorized
func (srv *AppServer) AddAuthorizedRoute(path, method string, authorizedRoles []string, handler http.HandlerFunc) error {

	path = fmt.Sprintf("/%s%s", srv.AppID, path)

	srv.router().HandleFunc(path, srv.requestInterceptor(srv.AuthorizeInterceptor(handler, authorizedRoles))).Methods(method)

	if len(authorizedRoles) > 0 {
		log.PrintfNoContext(srv.AppID, component, "Added authorized route %s %s with roles %s for app %s", method, path, authorizedRoles, srv.AppID)
	} else {
		log.PrintfNoContext(srv.AppID, component, "Added authorized route %s %s for app %s", method, path, srv.AppID)
	}

	return nil
}

func (srv *AppServer) AddRoutePrefix(path string, handler http.HandlerFunc) error {

	path = fmt.Sprintf("/%s%s", srv.AppID, path)

	srv.router().PathPrefix(path).HandlerFunc(srv.requestInterceptor(handler))

	log.PrintfNoContext(srv.AppID, component, "Added route prefix %s for app %s", path, srv.AppID)

	return nil
}

// AddGraphQLHandler
// Adds the http handler for the graphQL schema on POST /query endpoint
// Supports the ExecutableSchema from gqlgen lib (https://gqlgen.com/)
//
// path - url path that will respond to graphQL queries. It's always a POST
// gqlSchema - The executable Schema from gqlgen. It's a generated code
//
// Example call:
//
// 			srv.AddGraphQLHandler("/query", generated.NewExecutableSchema(generated.Config{
//					Resolvers: &resolver.Resolver{},
//			}))
//
func (srv *AppServer) AddGraphQLHandler(path string, gqlSchema graphql.ExecutableSchema) (err error) {

	path = fmt.Sprintf("/%s%s", srv.AppID, path)

	gqlServer := gqlHandlers.NewDefaultServer(gqlSchema)

	gqlServer.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		tracing.AddTracingGraphQLInfo(ctx)
		return next(ctx)
	})

	srv.router().HandleFunc(path, srv.requestInterceptor(func(writer http.ResponseWriter, request *http.Request) {

		gqlServer.ServeHTTP(writer, request)

	})).Methods("POST")

	srv.gqlServer = gqlServer

	log.PrintfNoContext(srv.AppID, component, "Added GraphQL route %s %s for app %s", "POST", path, srv.AppID)

	return nil
}

// SetGQLErrorPresenter
// Sets the ErrorPresenter for the GQL handler
// ref: https://gqlgen.com/reference/errors/#hooks
//
// ctx - the error context
// presenterFunc - a function that takes a context.Context and the original error, and returns the error to be presented
func (srv *AppServer) SetGQLErrorPresenter(presenterFunc func(ctx context.Context, originalErr error) (presentedErr error)) {

	srv.gqlServer.SetErrorPresenter(func(ctx context.Context, originalErr error) *gqlerror.Error {
		presentedErr := graphql.DefaultErrorPresenter(ctx, originalErr)

		if presenterFunc != nil {
			err := presenterFunc(ctx, originalErr)
			presentedErr.Message = err.Error()
		}

		return presentedErr
	})

}

func (srv *AppServer) Vars(r *http.Request) map[string]string {

	return mux.Vars(r)
}

func (srv *AppServer) NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {

	newR, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	valueUserID := ctx.Value(appctx.AuthorizedUserIDHeader)
	userID := ""

	if valueUserID != nil {
		userID = valueUserID.(string)
	}

	newR.Header.Set(appctx.AppIdHeader, ctx.Value(appctx.AppIdHeader).(string))
	newR.Header.Set(appctx.CorrelationIdHeader, ctx.Value(appctx.CorrelationIdHeader).(string))
	newR.Header.Set(appctx.AuthorizedUserIDHeader, userID)

	return newR, nil
}

func (srv *AppServer) configureAwsXray() {

	switch srv.environment {
	case app.ProductionEnvironment:
		if os.Getenv(app.AwsXrayHostEnv) == "" {
			srv.DisableAWSXrayTracing()
			log.PrintfNoContext(srv.AppID, component, "Env var %s not found. Disabling AWS XRay", app.AwsXrayHostEnv)

			return
		}

		// Configuring AWS Xray
		err := xray.Configure(xray.Config{
			DaemonAddr:     os.Getenv(app.AwsXrayHostEnv),
			ServiceVersion: os.Getenv(app.AppVersionEnv),
		})

		if err != nil {
			log.PrintfNoContext(srv.AppID, component, "Failed to configure AWS X-Ray configuration, %s. Proceeding with AWS XRay disabled", err)
			srv.DisableAWSXrayTracing()
			return
		}

		log.PrintfNoContext(srv.AppID, component, "AWS XRay successfully configured.")

		srv.EnableAWSXrayTracing()

	default:
		srv.DisableAWSXrayTracing()

	}

}

func (srv *AppServer) Start() {

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT)  // Handling Ctrl + C
	signal.Notify(sigChan, syscall.SIGTERM) // Handling Docker stop

	log.PrintfNoContext(srv.AppID, component, "Initializing resources for %s environment", srv.environment)

	srv.configureAwsXray()

	srv.addVersionHandler()
	srv.addHealthHandler()

	if srv.initializeFunc != nil {

		err := srv.initializeFunc(srv)

		if err != nil {
			log.FatalfNoContext(srv.AppID, component, "Failed to initialize resources, %s", err)
		}

		if len(srv.corsOrigins) > 0 {

			log.PrintfNoContext(srv.AppID, component, "Setting CORS origins: %s", srv.corsOrigins)

			headersOk := gHandlers.AllowedHeaders([]string{"Authorization", "Content-Type", "Accept,Origin", "User-Agent", "DNT,Cache-Control", "X-Mx-ReqToken", "Keep-Alive", "X-Requested-With", "If-Modified-Since", "Origin"})
			originsOk := gHandlers.AllowedOrigins(srv.corsOrigins)
			methodsOk := gHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE", "PATCH"})
			credOK := gHandlers.AllowCredentials()
			ageOK := gHandlers.MaxAge(600)

			srv.Handler = gHandlers.CORS(headersOk, originsOk, methodsOk, credOK, ageOK)(srv.Handler)
		}

	}

	log.PrintfNoContext(srv.AppID, component, "Starting app server %s", srv.AppID)

	go func() {
		log.PrintfNoContext(srv.AppID, component, "Listening on port %s. Ctrl+C to stop", srv.Addr)

		err := srv.ListenAndServe()

		if err != http.ErrServerClosed {
			log.FatalfNoContext(srv.AppID, component, "Failed to start server, %s", err)
		}
	}()

	<-sigChan

	srv.prepareShutdown()

}

func (srv *AppServer) prepareShutdown() {

	log.PrintfNoContext(srv.AppID, component, "Cleaning up resources")

	if srv.cleanupFunc != nil {
		err := srv.cleanupFunc(srv)

		if err != nil {
			log.ErrorfNoContext(srv.AppID, component, "Error cleaning up resources ,%s", err)
		}
	}

	log.PrintfNoContext(srv.AppID, component, "Shutting down app server %s", srv.AppID)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	err := srv.Shutdown(ctx)

	if err != nil {
		log.FatalfNoContext(srv.AppID, component, "Error shutting down server, %s", err)
	} else {
		log.PrintfNoContext(srv.AppID, component, "App server %s gracefully stopped", srv.AppID)
	}
}

func (srv *AppServer) router() *mux.Router {

	return srv.Handler.(*mux.Router)
}

func (srv *AppServer) addHealthHandler() {

	path := "/healthz"
	pathAppID := fmt.Sprintf("/%s/healthz", srv.AppID)
	method := "GET"

	f := func(writer http.ResponseWriter, request *http.Request) {}

	srv.router().HandleFunc(path, f).Methods(method)
	srv.router().HandleFunc(pathAppID, f).Methods(method)

	log.PrintfNoContext(srv.AppID, component, "Added route %s %s for app %s", method, path, srv.AppID)
	log.PrintfNoContext(srv.AppID, component, "Added route %s %s for app %s", method, pathAppID, srv.AppID)
}

func (srv *AppServer) addVersionHandler() {

	path := fmt.Sprintf("/%s/version", srv.AppID)
	method := "GET"

	srv.router().HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {

		ctx := appctx.NewContext(request)

		type v struct {
			AppID   app.ApplicationID
			Version string
		}

		vJSON, _ := json.Marshal(&v{
			AppID:   srv.AppID,
			Version: os.Getenv(app.AppVersionEnv),
		})

		writer.Header().Set("Content-Type", "application/json")

		_, err := writer.Write(vJSON)

		if err != nil {
			log.Errorf(ctx, component, "Error serializing JSON for /version call, %s", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}

		return

	}).Methods(method)

	log.PrintfNoContext(srv.AppID, component, "Added route %s %s for app %s", method, path, srv.AppID)
}

func (srv *AppServer) AuthorizeInterceptor(next http.HandlerFunc, authorizedRoles []string) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {

		ctx := appctx.NewContext(request)

		authUserID := appctx.GetAuthorizedUserID(ctx)

		if authUserID == "" {
			log.Errorf(ctx, component, "Unauthorized request")
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		if len(authorizedRoles) > 0 {

			isForbidden := true

			for _, role := range authorizedRoles {

				if appctx.HasRole(ctx, role) {
					isForbidden = false
					break
				}
			}

			if isForbidden {
				log.Errorf(ctx, component, "User %s forbidden", authUserID)
				writer.WriteHeader(http.StatusForbidden)
				return
			}
		}

		log.Printf(ctx, component, "User %s authorized", authUserID)

		next.ServeHTTP(writer, request)
	}
}

func (srv *AppServer) requestInterceptor(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get(appctx.CorrelationIdHeader) == "" {
			id, _ := uuid.NewV4()
			r.Header.Set(appctx.CorrelationIdHeader, id.String())
		}

		previousAppID := r.Header.Get(appctx.AppIdHeader)

		r.Header.Set(appctx.AppIdHeader, string(srv.AppID))

		ctx := appctx.NewContext(r)

		if previousAppID != "" {
			log.Printf(ctx, component, "Request %s %s from app %s", r.Method, r.RequestURI, previousAppID)
		} else {
			log.Printf(ctx, component, "Request %s %s", r.Method, r.RequestURI)
		}

		// Here wrapping request in a AWS XRay segment handler to trace the Request
		xRayHandler := srv.newXraySegmentHandler(next)

		xRayHandler.ServeHTTP(w, r.WithContext(ctx))

	}
}

func (srv *AppServer) newXraySegmentHandler(next http.HandlerFunc) http.Handler {
	xRayHandler := xray.Handler(xray.NewFixedSegmentNamer(string(srv.AppID)), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.NewContext(r)

		tracing.AddCustomTracingWorkloadType(ctx, tracing.WorkloadTypeHTTPCall)
		tracing.AddTracingAnnotationFromCtx(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	}))
	return xRayHandler
}
