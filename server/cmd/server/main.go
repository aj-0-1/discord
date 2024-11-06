package main

import (
	"context"
	"discord/internal/auth"
	"discord/internal/chat"
	"discord/internal/config"
	"discord/internal/database"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "discord/docs"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatal("failed to connect to database")
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to redis")
	}

	authService := auth.NewService(
		db,
		[]byte(cfg.JWT.Secret),
		&logger,
	)
	authHandler := auth.NewHandler(authService, &logger)

	chatService := chat.NewService(db, redisClient, &logger)
	chatHandler := chat.NewHandler(chatService, &logger)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", authHandler.Routes())

		r.Group(func(r chi.Router) {
			r.Use(authService.Middleware)
			r.Mount("/chat", chatHandler.Routes())
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal().Msg("graceful shutdown timed out.. forcing exit")
			}
		}()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			logger.Fatal().Err(err).Msg("server shutdown failed")
		}
		serverStopCtx()
	}()

	logger.Info().Msgf("server starting on port %d", cfg.Server.Port)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("server failed to start")
	}

	<-serverCtx.Done()
}
