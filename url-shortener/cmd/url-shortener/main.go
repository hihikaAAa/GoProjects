package main 

import(
	"log/slog"
	"os"

	"github.com/hihikaAAa/GoProjects/url-shortener/internal/config"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/storage/sqlite"
	"github.com/hihikaAAa/GoProjects/url-shortener/internal/lib/logger/sl"

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
}

func setupLogger(env string) *slog.Logger{
	var log *slog.Logger

	switch env{
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout,&slog.HandlerOptions{Level: slog.LevelDebug}),
	)
	case envDev:
	log = slog.New(slog.NewJSONHandler(os.Stdout,&slog.HandlerOptions{Level: slog.LevelDebug}),)
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout,&slog.HandlerOptions{Level: slog.LevelInfo}),)
	}
	return log
}