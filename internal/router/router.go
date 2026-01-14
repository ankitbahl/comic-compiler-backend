package router

import (
	"github.com/ankitbahl/comic-compiler-backend/internal/controller"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func New() *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/comics", controller.Comics)
		r.Get("/comic-info", controller.ComicInfo)
		r.Post("/compile-comic", controller.CompileComic)
		r.Get("/comic-progress", controller.GetJobProgress)
		r.Post("/download-comic", controller.DownloadComic)
		r.Get("/downloadable-comics", controller.GetDownloadableComics)
	})

	return r
}
