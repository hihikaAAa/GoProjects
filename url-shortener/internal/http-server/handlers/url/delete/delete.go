package delete

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	resp "github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/api/response"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/logger/sl"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/storage"
)

type URLDeleter interface{
	DeleteURL(alias string)(error)
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "handlers.url.delete.New"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())),)

		alias := chi.URLParam(r, "alias")

		if alias == ""{
			log.Info("alias is empty")

			render.JSON(w,r,resp.Error("invalid request"))

			return
		}

		err := urlDeleter.DeleteURL(alias)
		if errors.Is(err,storage.ErrURLNotFound){
			log.Info("url not found")

			render.JSON(w,r,resp.Error("not found"))

			return
		}
		if err != nil{
			log.Error("failed to delete url", sl.Err(err))

			render.JSON(w,r,resp.Error("failed to delete url"))

			return
		}

		log.Info("url deleted", slog.String("alias", alias))

		w.WriteHeader(http.StatusNoContent)
	}
}