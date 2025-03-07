package main

import (
	"errors"
	"flag"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/storage/postgresql"
	"log"
	"log/slog"
	"os"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// cfg - конфиг сервера.
var cfg *config.Config

func main() {
	var configPath string
	flag.StringVar(&configPath, "path", "", "положение файла конфигурации")

	// чтение параметров
	flag.Parse()

	if configPath == "" {
		log.Fatal("не указано место файла конфигурации")
	}

	// получаем файл конфига
	cfg = config.MustCreateConfig(configPath)

	// logFile - буфер файла, в котором буду храниться логи.
	// Для каждого запуска свои логи.
	var logFile *os.File

	if cfg.Environment != envLocal {
		// создаю файл логирования, если нужен
		logFile = mustCreateLogFile()
		defer logFile.Close()
	}

	// создаю и задаю логер
	log := mustCreateLogger(cfg.Environment, logFile)

	log.Info("logger started successfully", slog.String("env", cfg.Environment))
	// подключение БД
	db := postgresql.MustConnectDB(cfg)

	_ = db // TODO: убрать

	// TODO: установить роутер + выбрать пакет для websocket

	// TODO: запустить сервер
}

// mustCreateLogger создает логер исходя из текущего окружения.
//
// Если логер не создастся, случится паника.
func mustCreateLogger(env string, logFile *os.File) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		logger = slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log.Fatal("неправильное окружение")
	}

	return logger
}

// mustCreateLogFile создает файл для хранения логов в формате "DD.MM.YYYY hh.mm.ss".
//
// Если файл не создастся, случится паника.
func mustCreateLogFile() *os.File {
	err := os.Mkdir("logs/", os.ModeDir)
	if errors.Is(err, os.ErrExist) {
		log.Println("directory \"logs/\" already exists")
	} else if err != nil {
		log.Fatal("error while creating \"logs/\" directory: ", err)
	}

	fileName := time.Now().Format("02.01.2006 15.04.05")

	logFile, err := os.Create("./logs/" + fileName + ".json")
	if err != nil {
		log.Fatal("error while create log file "+fileName+": ", err)
	}

	return logFile
}
