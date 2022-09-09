package routes

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"fmt"
	"errors"
)

func UserRouter(router chi.Router){
	router.Route("/{ID[0-9]+}", func(r chi.Router) {
		r.use(userCtx)
		r.Get("/", getUser)
	  })
}

func userCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user *User
		var err error

		if userID := chi.URLParam(r, "ID"); userID != "" {
			user, err = dbGetUser(userID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}

		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "article", article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("ID").(*User)

	if err := render.Render(w, r, NewArticleResponse(user)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}