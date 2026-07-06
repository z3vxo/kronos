package teamserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/z3vxo/kronos/internal/auth"
	"github.com/z3vxo/kronos/internal/broker"
	"github.com/z3vxo/kronos/internal/config"
	"github.com/z3vxo/kronos/internal/database"
	"github.com/z3vxo/kronos/internal/files"
	"github.com/z3vxo/kronos/internal/logger"
)

func NewTeamServer() (*TeamServer, error) {
	a := auth.NewAuth(config.Cfg.TS.Auth.Username, config.Cfg.TS.Auth.Password,
		config.Cfg.TS.Auth.JwtSecret, config.Cfg.TS.Auth.TokenHours, config.Cfg.TS.Auth.TokenRefreshHours)
	d, err := database.NewDB()
	if err != nil {
		return nil, err
	}

	l, err := logger.New()
	if err != nil {
		return nil, err
	}

	return &TeamServer{
		httpServer: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", config.Cfg.TS.ListenInterface, config.Cfg.TS.Port),
			ReadHeaderTimeout: 15 * time.Second,
			WriteTimeout:      0,
			IdleTimeout:       0,
		},
		SSE:       broker.NewBroker(),
		Auth:      a,
		db:        d,
		Listeners: &Listeners{ListenerMap: make(map[string]Listener), GetEndpoint: config.Cfg.Server.GetEndpoint, PostEndpoint: config.Cfg.Server.PostEndpoint},
		Logger:    l,
		FileMgr:   files.NewFileManager(d),
	}, nil
}

func (ts *TeamServer) Start() error {

	r := chi.NewRouter()
	ts.httpServer.Handler = r
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(notFound))

	})

	r.Route("/ts", func(r chi.Router) {
		r.Post("/rest/login", ts.loginHandler)

		r.Group(func(r chi.Router) {
			r.Use(ts.Auth.AuthMiddleWare)
			r.Get("/events", ts.SSE.EventHandler)
			r.Get("/rest/agents/list", ts.AgentListHandler)
			r.Get("/rest/agents/resolve/{codename}", ts.AgentResolveHandler)
			r.Get("/rest/agents/info/{codename}", ts.AgentInfoHandler)
			r.Delete("/rest/agents/delete/all", ts.AgentDeleteAllHandler)
			r.Delete("/rest/agents/delete/{codename}", ts.AgentDeleteHandler)

			r.Post("/rest/tasks/new", ts.CommandNewHandler)
			r.Delete("/rest/tasks/delete/{guid}/{taskID}", ts.CommandDeleteHandler)
			r.Get("/rest/tasks/list/{guid}", ts.ListTasksHandler)
			r.Get("/rest/tasks/history/{guid}", ts.ListHistoryHandler)
			r.Get("/rest/tasks/history/{guid}/{taskID}", ts.GetHistoryOutputHandler)

			r.Get("/rest/files/list", ts.FilesListHandler)
			r.Get("/rest/files/sync/{code}", ts.FilesSyncHandler)
			r.Post("/rest/file/download/task", ts.DownloadTaskHandler) 
			r.Post("/rest/file/upload/start", ts.UploadStartHandler)
			r.Put("/rest/file/upload/{id}", ts.HandleUpload)

			r.Get("/rest/listeners/list", ts.ListListenerHandler)
			r.Post("/rest/listeners/new", ts.NewListenerHandler)
			r.Post("/rest/listeners/start/{name}", ts.StartListenerHandler)
			r.Post("/rest/listeners/stop/{name}", ts.StopListenerHandler)
			r.Delete("/rest/listeners/delete/{name}", ts.DeleteListnerHandler)

			r.Post("/rest/agent/profiles/new", ts.HandleNewProfile)
			r.Get("/rest/agent/profiles/list", ts.HandleListProfiles)
			r.Post("/rest/agent/generate/{profile}", ts.HandleProfileGenerate)
			r.Post("/rest/agent/profiles/delete/{profile}", ts.HandleDeleteProfile)

		})
	})
	fmt.Println("Server Started!")
	if err := ts.StartListenersFromDB(); err != nil {
		return err
	}

	return ts.httpServer.ListenAndServe()
}

func (ts *TeamServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ts.StopAllListeners()
	ts.httpServer.Shutdown(ctx)
}
