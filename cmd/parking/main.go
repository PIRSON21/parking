package main

import (
	"errors"
	"flag"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/parking"
	"github.com/PIRSON21/parking/internal/http-server/handler/user"
	authMiddleware "github.com/PIRSON21/parking/internal/lib/api/auth/middleware"
	"github.com/PIRSON21/parking/internal/storage/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"log/slog"
	"net/http"
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

	// установка роутера chi
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.RedirectSlashes)

	router.Group(func(public chi.Router) {
		public.Post("/login", user.LoginHandler(log, db, cfg))
	})

	router.Group(func(user chi.Router) {
		user.Use(authMiddleware.AuthMiddleware(db))
		user.Get("/parking", parking.AllParkingsHandler(log, db, cfg))
	})

	router.Group(func(manager chi.Router) {
		manager.Use(authMiddleware.AuthMiddleware(db))
		manager.Use(authMiddleware.ManagerMiddleware)
		manager.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusSeeOther)
		})
	})

	router.Group(func(admin chi.Router) {
		admin.Use(authMiddleware.AuthMiddleware(db))
		admin.Use(authMiddleware.AdminMiddleware)
		admin.Route("/parking", func(r chi.Router) {
			r.Get("/{id}", parking.GetParkingHandler(log, db, cfg))
			r.Post("/add", parking.AddParkingHandler(log, db, cfg))
		})
		admin.Post("/create_manager", user.CreateManagerHandler(log, db, cfg))
	})

	// задание настроек сервера
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  40 * time.Second,
		WriteTimeout: 40 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("server started", slog.String("addr", srv.Addr))

	// запуск сервера
	if err := srv.ListenAndServe(); err != nil {
		log.Error("error while serving: ", slog.String("err", err.Error()))
		return
	}
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
