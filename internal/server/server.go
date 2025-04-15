package server

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/wisp167/pvz/internal/data"
	"github.com/wisp167/pvz/internal/handlers"
)

const version = "1.0.0"

type config struct {
	port       int
	env        string
	numWorkers int
	db         struct {
		dsn          string
		port         string
		host         string
		name         string
		user         string
		password     string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type Application struct {
	config  config
	logger  *log.Logger
	model   *data.Models
	queue   chan struct{}
	jwtkey  []byte
	server  *echo.Echo
	handler *handlers.ServerHandler
}

func SetupApplication() (*Application, error) {
	var cfg config
	var jwtkey string

	godotenv.Load(".env")

	logger := log.New(os.Stdout, "[Application]: ", log.Ldate|log.Ltime|log.Lshortfile)

	EnvPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PORT: %v", err)
	}

	DbMaxOpenCons, err := strconv.Atoi(os.Getenv("DATABASE_MAX_OPEN_CONNS"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_MAX_OPEN_CONNS: %v", err)
	}
	DbMaxIdleCons, err := strconv.Atoi(os.Getenv("DATABASE_MAX_IDLE_CONNS"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_MAX_IDLE_CONNS: %v", err)
	}
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		return nil, fmt.Errorf("JWT_KEY environment variable is required")
	}
	flag.IntVar(&cfg.port, "port", EnvPort, "API server port")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.host, "db-host", os.Getenv("DATABASE_HOST"), "PostgreSQL host")
	flag.StringVar(&cfg.db.name, "db-name", os.Getenv("DATABASE_NAME"), "PostgreSQL database name")
	flag.StringVar(&cfg.db.user, "db-user", os.Getenv("DATABASE_USER"), "PostgreSQL user")
	flag.StringVar(&cfg.db.port, "db-port", os.Getenv("DATABASE_PORT"), "PostgreSQL port")

	flag.StringVar(&cfg.db.password, "db-password", os.Getenv("DATABASE_PASSWORD"), "PostgreSQL password")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", DbMaxOpenCons, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", DbMaxIdleCons, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", os.Getenv("DATABASE_MAX_IDLE_TIME"), "PostgreSQL max connection idle time")

	cfg.numWorkers = 50

	flag.Parse()

	logger.Printf("Config: %v", cfg)

	// Open the database connection
	db, err := OpenDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	model, err := data.NewModels(db)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	app := &Application{
		config: cfg,
		logger: logger,
		model:  &model,
		jwtkey: []byte(jwtkey),
		queue:  make(chan struct{}, cfg.numWorkers),
	}

	return app, nil
}

func (app *Application) RegisterHandler(e *echo.Echo, handler *handlers.ServerHandler) {
	JWTConfig_ := handlers.JWTConfig{
		SigningKey: app.jwtkey,
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/register" ||
				c.Path() == "/login" ||
				c.Path() == "/dummyLogin"
		},
	}

	authMiddleware := handlers.AuthWithConfig(JWTConfig_)

	//authGroup := e.Group("")
	e.Use(authMiddleware)

	handlers.RegisterHandlersMiddleware(e, handler)

}

func (app *Application) Start() error {
	e := echo.New()
	handlerLogger := log.New(os.Stdout, "[Handler]: ", log.Ldate|log.Ltime|log.Lshortfile)
	handler := &handlers.ServerHandler{
		Model: app.model,
	}

	handler.InitUnexportedVals(app.jwtkey, handlerLogger)

	e.HideBanner = true

	if app.config.env == "development" {
		e.Debug = true
		e.Use(middleware.Logger())
	}

	app.RegisterHandler(e, handler)

	app.server = e
	app.handler = handler

	app.logger.Printf("starting %s server on %d", app.config.env, app.config.port)

	address := fmt.Sprintf(":%d", app.config.port)
	serverErr := make(chan error, 1)
	go func() {
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			serverErr <- err
			app.logger.Fatalf("listen: %s\n", err)
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-time.After(100 * time.Millisecond):
		app.server = e
		return nil
	}
}

func (app *Application) Stop() error {
	if app.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			app.logger.Println("graceful shutdown timed out, forcing close")
			if closeErr := app.server.Close(); closeErr != nil {
				return fmt.Errorf("forced close error: %v (original error: %v)", closeErr, err)
			}
		}
		return fmt.Errorf("server shutdown failed: %v", err)
	}

	app.logger.Println("server stopped gracefully")
	return nil

}

func OpenDB(cfg config) (*sql.DB, error) {
	cfg.db.dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.db.user,
		cfg.db.password,
		cfg.db.host,
		cfg.db.port,
		cfg.db.name,
	)
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
