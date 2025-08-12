package save

import (
	"log/slog"
	"net/http"
	"errors"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	resp "github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/api/response"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/logger/sl"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/random"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/storage"
)

type Request struct{
	URL string `json:"url" validate:"required,url"` // validate - дает информацию пакету валидартор. Это обязательное поле, если нет - получаем ошибку. И это обязательно должен быть url
	Alias string `json:"alias,omitempty"`
}

type Response struct{
	resp.Response
	Alias string `json:"alias,omitempty"`
}
type URLSaver interface{
	SaveURL(urlToSave string, alias string) (int64,error)
}

const aliasLength = 6
const maxAliasTries = 15

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		const op = "handlers.url.save.New"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())),)

		var req Request

		err := render.DecodeJSON(r.Body, &req) // Помогает распарсить запрос

		if err != nil{
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return // Не ЗАБЫВАТЬ return , тк не останавливает функцию
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil{ // Валидирование данных. Информация должна быть читаемой
			validateErr := err.(validator.ValidationErrors) // Приводим ошибку к нужному типу

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}
		
		alias := req.Alias // alias - не обязательный парамерт. Если не указан - генерируем свой
		var id int64
		if alias == ""{
			alias = random.NewUniqueRandomString(aliasLength)
			id, err = urlSaver.SaveURL(req.URL,alias)
			if err!= nil && errors.Is(err,storage.ErrUrlExists){
				for i:=0; i< maxAliasTries; i++{
					alias = random.NewUniqueRandomString(aliasLength)
					id, err = urlSaver.SaveURL(req.URL,alias);
					if err == nil{
						break
					}
					if errors.Is(err,storage.ErrUrlExists){
						continue
					}
					log.Error("failed to add url",sl.Err(err))

					render.JSON(w, r, resp.Error("failed to add url"))
					
					return
				}
				if err != nil{
					log.Error("failed to generate alias", sl.Err(err))

					render.JSON(w, r, resp.Error("failed to generate alias"))
					
					return
				}
			}else if err != nil {
				log.Error("failed to add url", sl.Err(err))

				render.JSON(w, r, resp.Error("failed to add url"))

				return
    }
		}else{
			id, err = urlSaver.SaveURL(req.URL,alias)
			if errors.Is(err, storage.ErrUrlExists){
				log.Info("url already exists", slog.String("url", req.URL)) // Логируем именно с уровнем инфо, тк это нормальная ситуация

				render.JSON(w, r, resp.Error("url already exitst"))

				return
			}
			if err != nil{
				log.Error("failed to add url", sl.Err(err))

				render.JSON(w, r, resp.Error("failed to add url"))

				return
			}
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, 
			Response{
				Response: resp.OK(),
				Alias: alias,
		})
	}
}