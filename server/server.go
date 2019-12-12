package server

import (
	"encoding/json"
	"fmt"
	"github.com/HelloSundayMorning/apputils/app"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/HelloSundayMorning/apputils/log"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"io"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	Initialize func(srv *AppServer) (err error)
	CleanUp func(srv *AppServer) (err error)

	// Application Server object that controls the application state and life cycle.
	// It's based on the http Server from net/http package, and offers the ability to register HTTP routes.
	// The router is Handler uses the gorilla/mux implementation
	// Initialization, Shutdown and Cleanup is managed by the AppServer. Custom functions for initialization and
	// cleanup are provided so the life cycle of other objects can be added to it.
	AppServer struct {
		*http.Server
		AppID          app.ApplicationID  // Unique identifier for the application
		initializeFunc Initialize         // Custom initialization function
		cleanupFunc    CleanUp            // Custom cleanup function
	}
)

const (
	component = "server"
)

// Create a new Application Server instance.
func NewServer(appID app.ApplicationID, port int) *AppServer {

	router := mux.NewRouter()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	server := &AppServer{
		Server: httpServer,
		AppID:  appID,
	}

	return server
}

// Create a new Application Server instance with a custom http handler.
// the AddRoute method does not work with custom http handler, so routes must be added
// outside the AppServer in the http Handler it self
func NewServerWithHandler(appID app.ApplicationID, port int, handler http.Handler) *AppServer {

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	server := &AppServer{
		Server: httpServer,
		AppID:  appID,
	}

	return server
}

// Create a new Application Server instance, with Custom initialization and cleanup functions
func NewServerWithInitialization(appID app.ApplicationID, port int, initializeFunc Initialize, cleanupFunc CleanUp) *AppServer {

	server := NewServer(appID, port)

	server.initializeFunc = initializeFunc
	server.cleanupFunc = cleanupFunc

	return server
}

func (srv *AppServer) AddRoute(path, method string, handler http.HandlerFunc) error {

	path = fmt.Sprintf("/%s%s", srv.AppID, path)

	srv.router().HandleFunc(path, srv.requestInterceptor(handler)).Methods(method)

	log.PrintfNoContext(srv.AppID, component, "Added route %s %s for app %s", method, path, srv.AppID)

	return nil
}

func (srv *AppServer) AddRoutePrefix(path string, handler http.HandlerFunc) error {

	path = fmt.Sprintf("/%s%s", srv.AppID, path)

	srv.router().PathPrefix(path).HandlerFunc(srv.requestInterceptor(handler))

	log.PrintfNoContext(srv.AppID, component, "Added route prefix %s for app %s", path, srv.AppID)

	return nil
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

func (srv *AppServer) Start() {

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT)  // Handling Ctrl + C
	signal.Notify(sigChan, syscall.SIGTERM) // Handling Docker stop

	log.PrintfNoContext(srv.AppID, component, "Initializing resources")

	srv.addVersionHandler()

	if srv.initializeFunc != nil {

		err := srv.initializeFunc(srv)

		if err != nil {
			log.FatalfNoContext(srv.AppID, component, "Failed to initialize resources, %s", err)
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
			Version: os.Getenv("APP_VERSION"),
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

func (srv *AppServer) requestInterceptor(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get(appctx.CorrelationIdHeader) == "" {
			id := uuid.NewV4()
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

		next.ServeHTTP(w, r)

	}
}
