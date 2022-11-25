/*
main package is the main entry point for the api backend.
*/
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"github.com/Walchand-Linux-Users-Group/wargames/backend/api/helpers"
	"github.com/Walchand-Linux-Users-Group/wargames/backend/api/routes"
)

func allowOriginFunc(r *http.Request, origin string) bool {
	return true
}

func initAPI() {

	helpers.InitEnv()

	helpers.InitDatabase()

	helpers.InitRedis()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  allowOriginFunc,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I am alive!"))
	})

	router.Route("/user", routes.UserRouter)
	// router.Route("/wargame", routes.WargameRoute)

	log.Fatal(http.ListenAndServe(":"+helpers.GetEnv("PORT"), router))

}

func main() {
	initAPI()
}
