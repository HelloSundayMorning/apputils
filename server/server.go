package server

import (
	"encoding/json"
	"fmt"
	"github.com/HelloSundayMorning/apputils/appctx"
	"github.com/HelloSundayMorning/apputils/db"
	"github.com/HelloSundayMorning/apputils/eventpubsub"
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
	CleanUp func(srv *AppServer)
	RequestHandler func(srv *AppServer, dependencies ...interface{}) (handler http.HandlerFunc, err error)
	EventHandler func(srv *AppServer, dependencies ...interface{}) (handler eventpubsub.ProcessEvent, err error)

	AppServer struct {
		*http.Server
		AppID          string
		initializeFunc Initialize
		cleanupFunc    CleanUp
		pubSub         eventpubsub.EventPubSub
		db.AppSqlDb
	}
)

func NewServer(appID string, port int, eventPubSub eventpubsub.EventPubSub, appDb db.AppSqlDb) *AppServer {

	router := mux.NewRouter()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	server := &AppServer{
		Server:   httpServer,
		AppID:    appID,
		pubSub:   eventPubSub,
		AppSqlDb: appDb,
	}

	return server
}

func NewServerWithInitialization(appID string, port int, eventPubSub eventpubsub.EventPubSub, appDb db.AppSqlDb, initializeFunc Initialize, cleanupFunc CleanUp) *AppServer {

	server := NewServer(appID, port, eventPubSub, appDb)

	server.initializeFunc = initializeFunc
	server.cleanupFunc = cleanupFunc

	return server
}

func (srv *AppServer) AddRoute(path, method string, handler RequestHandler, dependencies ...interface{}) error {

	h, err := handler(srv)

	if err != nil {
		return err
	}

	srv.router().HandleFunc(path, srv.requestInterceptor(h)).Methods(method)

	log.PrintfNoContext(srv.AppID, "server", "Added route %s %s for app %s", method, path, srv.AppID)

	return nil
}

func (srv *AppServer) RegisterTopic(topic string) (err error) {

	err = srv.pubSub.RegisterTopic(srv.AppID, topic)

	if err != nil {
		return err
	}

	log.PrintfNoContext(srv.AppID, "server", "Registered topic %s for app %s", topic, srv.AppID)

	return nil
}

func (srv *AppServer) InitializeQueue(topic string) (err error) {

	for attempts := 1; attempts < 4; attempts++ {

		err = srv.pubSub.InitializeQueue(srv.AppID, topic)
		if err == nil {
			break
		}

		log.PrintfNoContext(srv.AppID, "server", "Failed to initialize queue for topic %s. Waiting (30 sec) for next attempt. Total Attempts = %d", topic, attempts)

		time.Sleep(time.Second * 30)

	}

	if err != nil {
		log.PrintfNoContext(srv.AppID, "server", "Failed to initialize queue for topic %s. %s", topic, err)
	}

	return nil
}

func (srv *AppServer) SubscribeToTopicWithMaxMsg(topic string, eventHandler EventHandler, maxMessages int, dependencies ...interface{}) (err error) {

	h, err := eventHandler(srv, dependencies...)

	if err != nil {
		return err
	}

	err = srv.pubSub.SubscribeWithMaxMsg(srv.AppID, topic, h, maxMessages)

	if err != nil {
		return err
	}

	log.PrintfNoContext(srv.AppID, "server", "App %s Subscribed to topic %s", srv.AppID, topic)

	return nil
}

func (srv *AppServer) SubscribeToTopic(topic string, eventHandler EventHandler, dependencies ...interface{}) (err error) {

	h, err := eventHandler(srv, dependencies...)

	if err != nil {
		return err
	}

	err = srv.pubSub.Subscribe(srv.AppID, topic, h)

	if err != nil {
		return err
	}

	log.PrintfNoContext(srv.AppID, "server", "App %s Subscribed to topic %s", srv.AppID, topic)

	return nil
}

func (srv *AppServer) PublishToTopic(ctx context.Context, topic, contentType string, event []byte) (err error) {

	err = srv.pubSub.Publish(ctx, topic, event, contentType)

	if err != nil {
		return err
	}

	return nil
}

func (srv *AppServer) NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {

	newR, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	newR.Header.Set(appctx.AppIdHeader, ctx.Value(appctx.AppIdHeader).(string))
	newR.Header.Set(appctx.CorrelationIdHeader, ctx.Value(appctx.CorrelationIdHeader).(string))

	return newR, nil
}

func (srv *AppServer) Start() {

	sigc := make(chan os.Signal, 1)

	signal.Notify(sigc, syscall.SIGINT)  // Handling Ctrl + C
	signal.Notify(sigc, syscall.SIGTERM) // Handling Docker stop

	log.PrintfNoContext(srv.AppID, "server", "Initializing resources")

	srv.addVersionHandler()

	if srv.initializeFunc != nil {

		err := srv.initializeFunc(srv)

		if err != nil {
			log.FatalfNoContext(srv.AppID, "server", "Failed to initialize resources, %s", err)
		}
	}

	log.PrintfNoContext(srv.AppID, "server", "Starting app server %s", srv.AppID)

	go func() {
		log.PrintfNoContext(srv.AppID, "server", "Listening on port %s. Ctrl+C to stop", srv.Addr)

		err := srv.ListenAndServe()

		if err != http.ErrServerClosed {
			log.FatalfNoContext(srv.AppID, "server", "Failed to start server, %s", err)
		}
	}()

	<-sigc

	srv.prepareShutdown()

}

func (srv *AppServer) prepareShutdown() {

	log.PrintfNoContext(srv.AppID, "server", "Cleaning up resources")

	srv.pubSub.CleanUp()

	if srv.cleanupFunc != nil {
		srv.cleanupFunc(srv)
	}

	log.PrintfNoContext(srv.AppID, "server", "Shutting down app server %s", srv.AppID)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	srv.Shutdown(ctx)

	log.PrintfNoContext(srv.AppID, "server", "App server %s gracefully stopped", srv.AppID)
}

func (srv *AppServer) router() *mux.Router {
	return srv.Handler.(*mux.Router)
}

func (srv *AppServer) addVersionHandler() {

	path := "/version"
	method := "GET"

	srv.router().HandleFunc(path, srv.requestInterceptor(func(writer http.ResponseWriter, request *http.Request) {

		type v struct {
			AppID   string
			Version string
		}

		vJSON, _ := json.Marshal(&v{
			AppID:   srv.AppID,
			Version: os.Getenv("APP_VERSION"),
		})

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(vJSON)

	})).Methods(method)

	log.PrintfNoContext(srv.AppID, "server", "Added route %s %s for app %s", method, path, srv.AppID)
}

func (srv *AppServer) requestInterceptor(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get(appctx.CorrelationIdHeader) == "" {
			id, _ := uuid.NewV4()
			r.Header.Set(appctx.CorrelationIdHeader, id.String())
		}

		previousAppID := r.Header.Get(appctx.AppIdHeader)

		r.Header.Set(appctx.AppIdHeader, srv.AppID)

		ctx := appctx.NewContext(r)

		if previousAppID != "" {
			log.Printf(ctx, "server", "Request %s %s from app %s", r.Method, r.RequestURI, previousAppID)
		} else {
			log.Printf(ctx, "server", "Request %s %s", r.Method, r.RequestURI)
		}

		next.ServeHTTP(w, r)

	}
}
