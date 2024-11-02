package main

import (
	"discord/internal/database"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"discord/internal/auth"
	"discord/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := auth.NewUserRepository(db)
	tokenService := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.Duration)
	authService := auth.NewAuthService(userRepo, tokenService)
	authHandler := auth.NewHandler(authService, tokenService)

	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", authHandler.Routes())
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
