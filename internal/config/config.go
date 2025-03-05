package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
)

type Config struct {
	Environment string `env:"ENV" 	  envDefault:"prod"`
	Address     string `env:"ADDRESS" envDefault:"localhost:8000"`
	ConfigDB
}

type ConfigDB struct {
	Name     string `env:"DB_NAME,required"`
	Username string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
}

// MustCreateConfig создает структуру конфига из файла, путь которого
// передан в path. Если возникла ошибка, приложение падает.
func MustCreateConfig(path string) *Config {
	var cfg Config
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}
