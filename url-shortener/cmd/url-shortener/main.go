package main

import (
	"log/slog"
	"os"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/config"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/http-server/handlers/url/save"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/http-server/middleware/logger"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/logger/sl"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/storage/sqlite"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/http-server/handlers/url/redirect"
)
const(
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)
func main(){
	cfg := config.MustLoad()
	
	log := setupLogger(cfg.Env)

	storage, err := sqlite.New(cfg.StoragePath)
	if err!= nil{
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log)) 
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat) // для красивых логов. Использовать осторожно, если можем привязаться к chi

	router.Post("/url", save.New(log,storage)) // Подключили хендлер на save
	router.Get("/{alias}", redirect.New(log,storage)) // Подключение redirect. Ищет в бд сохраненный URL и делает на него redirect. Благодаря URLFormat и {alias} мы получим alias

	log.Info("starting server", slog.String("address", cfg.Address))
	
	srv := &http.Server{
		Addr : cfg.Address,
		Handler: router,
		ReadTimeout: cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	} 
	
	if err := srv.ListenAndServe(); err != nil{ // Запуск сервера. Блокирующая функция. При вызове не позволяет коду двигаться дальше
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger{
	var log *slog.Logger

	switch env{
	case envLocal:
		log = setupPrettySlog()
	case envDev:
	log = slog.New(slog.NewJSONHandler(os.Stdout,&slog.HandlerOptions{Level: slog.LevelDebug}),)
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout,&slog.HandlerOptions{Level: slog.LevelInfo}),)
	}
	return log
}

func setupPrettySlog() *slog.Logger{
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}