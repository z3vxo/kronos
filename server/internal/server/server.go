package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Setup() error {
	//err := database.CheckAndSetupDB()

	r := chi.NewRouter()

	r.Route("/rest", func(r chi.Router) {
		r.Post("/login", loginHandler)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleWare)
			r.Get("/agents", nyx_AgentHandler)
			r.Post("/agents/resolve/{codeName}", nyx_AgentResolveHandler)
		})
	})
	fmt.Println("Server Started!")
	http.ListenAndServe(":3000", r)

	return nil
}
