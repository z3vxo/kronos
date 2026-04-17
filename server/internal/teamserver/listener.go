package teamserver

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/z3vxo/kronos/internal/config"
	"github.com/z3vxo/kronos/internal/server"
)

type Listener struct {
	httpServer *http.Server
}

type Listeners struct {
	Mu           sync.RWMutex
	ListenerMap  map[string]Listener
	GetEndpoint  string
	PostEndpoint string
}

func (ts *TeamServer) NewListener(port int) (string, error) {
	id := "1111"

	r := chi.NewRouter()
	r.Get(config.Cfg.Server.GetEndpoint, server.AgentCheckInHandler)
	r.Post(config.Cfg.Server.PostEndpoint, server.AgentUploadHandler)
	ts.Listeners.Mu.Lock()
	ts.Listeners.ListenerMap[id] = Listener{
		httpServer: &http.Server{Addr: fmt.Sprintf(":%d", port),
			Handler: r,
		},
	}
	ts.Listeners.Mu.Unlock()

	err := ts.db.InsertListener(port, id)
	if err != nil {
		return "", err
	}

	return id, nil

}

func (ts *TeamServer) StartListener(id string) error {
	go func() {
		if err := ts.Listeners.ListenerMap[id].httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("listener %s error: %v\n", id, err)
		}
	}()
	return nil
}

func (ts *TeamServer) StopListener(id string) error {
	if err := ts.db.DeleteListener(id); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ts.Listeners.ListenerMap[id].httpServer.Shutdown(ctx)
	return nil

}
