package get

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"testRest/internal/lib/api/response"
	"testRest/internal/lib/logger/sl"
	"testRest/internal/storage"
)

type UserGetter interface {
	GetUser(username string) (string, []int, []int, error)
}

func New(log *slog.Logger, userGetter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		username := chi.URLParam(r, "username")
		if username == "" {
			log.Error("username not found", username)
			render.JSON(w, r, response.Error("invalid username"))
			return
		}
		_, _, _, err := userGetter.GetUser(username)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found", slog.String("username", username))
			render.JSON(w, r, response.Error("user not found"))
			return
		}
		if err != nil {
			log.Error("failed to get user", sl.Err(err))
			render.JSON(w, r, response.Error("failed to get user"))
			return
		}
		log.Info("user found", slog.String("username", username))
		render.JSON(w, r, response.OK)
		// return password, presstimes, intervaltimes to client
		http.Redirect(w, r, "/"+username, http.StatusFound)
	}
}
