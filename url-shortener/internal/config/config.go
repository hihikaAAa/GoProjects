package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct { // Конфиг будет такой же , как в .yaml файле. Без пробелов теги
	Env         string `yaml:"env" env:"ENV" env-default:"local" env-required:"true"` // Тег yaml - определяет, какое имя будет у параметра в соотв. .yaml файле. env - в env,  envDefault - local. НЕ безопасно. Обычно ставят prod
	StoragePath string `yaml:"storage_path" env-required:"true"` 
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	ReadTimeout     time.Duration `yaml:"rtimeout" env-default:"4s"`
	WriteTimeout  time.Duration `yaml:"wtimeout" env-default:"6s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User string `yaml:"user" env-required:"true"` // Храним в конфиг файле
	Password string `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"` // Храним в секретах гитхаба, Git Actions Secrets
}

func MustLoad() *Config{ //Приставка Must используется, когда функция вместо возврата ошибки будет паниковать
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("There is no CONFIG_PATH") // Не будет стандартного логера, логер потом
	}

	if _,err := os.Stat(configPath); os.IsNotExist(err){ // Проверка, существует ли файл
		log.Fatalf("config file %s does not exist",configPath)
	}

	var cfg Config
	
	if err:= cleanenv.ReadConfig(configPath, &cfg);err!=nil{
		log.Fatalf("cannot read config file: %s", err)
	}

	return &cfg
}
