package update

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	resp "github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/api/response"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/logger/sl"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/storage"
)

type Request struct{
	URL string `json:"url" validate:"required,url"` 
}

type URLUpdater interface{
	UpdateURL(alias, newURL string)(error)
}

func New(log *slog.Logger, urlUpdater URLUpdater) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "handlers.url.update.New"

		log = log.With(slog.String("op",op), slog.String("request_id", middleware.GetReqID(r.Context())),)

		var req Request

		err := render.DecodeJSON(r.Body,&req)

		if err != nil{
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return 
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil{
			validateErr := err.(validator.ValidationErrors) 

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := chi.URLParam(r,"alias")
		if alias == ""{
			log.Info("alias not found")

			render.JSON(w,r,resp.Error("invalid request"))

			return
		}

		err = urlUpdater.UpdateURL(alias, req.URL)
		if errors.Is(err, storage.ErrURLNotFound){
			log.Info("url not found")

			render.JSON(w,r,resp.Error("not found"))

			return
		}
		if err != nil{
			log.Error("failed to update url", sl.Err(err))

			render.JSON(w,r, resp.Error("failed to update url"))

			return
		}

		log.Info("url updated", slog.String("alias", alias))

		w.WriteHeader(http.StatusNoContent)
	}
}