package application

import (
	"net/http"

	"github.com/averek7/order-api/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func loadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", loadOrderRoutes)
	return router
}

func loadOrderRoutes(router chi.Router) {
	orderHander := &handler.Order{}
	
	router.Post("/", orderHander.Create)
	router.Get("/", orderHander.List)
	router.Get("/{id}", orderHander.GetById)
	router.Put("/{id}", orderHander.UpdateById)
	router.Delete("/{id}", orderHander.DeleteById)
}
