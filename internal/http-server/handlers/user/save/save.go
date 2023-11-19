package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"testRest/internal/lib/api/response"
	"testRest/internal/lib/logger/sl"
	"testRest/internal/storage"
)

type Request struct {
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
	PressTime    []int  `json:"presstimes" validate:"required"`
	IntervalTime []int  `json:"intervaltimes" validate:"required"`
}

type Response struct {
	response.Response
	Username string `json:"username"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserSaver
type UserSaver interface {
	SaveUser(username string, password string, presstimes []int, intervaltimes []int) (string, error)
}

func New(log *slog.Logger, userSaver UserSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("failed to validate request", sl.Err(err))
			render.JSON(w, r, response.Error("failed to validate request"))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		username, err := userSaver.SaveUser(req.Username, req.Password, req.PressTime, req.IntervalTime)
		if errors.Is(err, storage.ErrUserExist) {
			log.Error("user already exist", sl.Err(err))
			render.JSON(w, r, response.Error("user already exist"))
			return
		}

		log.Info("user saved", slog.String("username", username))

		render.JSON(w, r, response.OK())
	}
}
