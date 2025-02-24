package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/controller"
)

func NewRouter(productController controller.ProductsController) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Route("/", func(r chi.Router) {
		r.Post("/products", productController.CreateProduct)
		r.Get("/products", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World!"))

		})
	})

	return router

}
