package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Environment string `env:"ENV" 	  envDefault:"prod"`
	Address     string `env:"ADDRESS" envDefault:"localhost:8000"`
	ConfigDB
}

type ConfigDB struct {
	DBHost     string `env:"DB_HOST,required"`
	DBName     string `env:"DB_NAME,required"`
	DBUsername string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
}

// MustCreateConfig создает структуру конфига из файла, путь которого
// передан в path. Если возникла ошибка, приложение падает.
func MustCreateConfig(path string) *Config {
	var cfg Config
	log.Println("reading config from file: ", path)
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}
