package redirect

import (
	"log/slog"
	"net/http"
	"errors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	resp "github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/api/response"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/storage"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/logger/sl"
)

type URLGetter interface{
	GetURL(alias string)(string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "handlers.url.redirect.New"

		log = log.With(slog.String("op",op), slog.String("request_id", middleware.GetReqID(r.Context())),)

		alias := chi.URLParam(r, "alias") // Получаем параметр alias из нашего роутера, благодаря {alias}. Здесь мы привязались к chi

		if alias == ""{
			log.Info("alias is empty")

			render.JSON(w,r, resp.Error("invalid request"))

			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound){
			log.Info("url not fount", "alias", alias)

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil{
			log.Error("failed to add url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))

			return
		}
		log.Info("url found", slog.String("alias", resURL))

		http.Redirect(w, r, resURL, http.StatusFound) // Редиректим на URL
	}
}
