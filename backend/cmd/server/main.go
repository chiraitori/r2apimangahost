package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"r2apimangahost/backend/internal/auth"
	"r2apimangahost/backend/internal/config"
	"r2apimangahost/backend/internal/db"
	"r2apimangahost/backend/internal/handlers"
	"r2apimangahost/backend/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.LoadConfig()

	log.Println("==================================================")
	log.Println("🚀 Starting Cloudflare R2 Manga Host API (Go)...")
	log.Println("==================================================")

	// Connect to MongoDB
	mongoDB, err := db.ConnectMongoDB(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoDB.Client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize R2 Storage
	r2Store, err := storage.NewR2Storage(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Cloudflare R2 client: %v", err)
	}

	// Initialize Services & Handlers
	authService := auth.NewAuthService(cfg)
	authHandler := handlers.NewAuthHandler(mongoDB, authService, cfg)
	mangaHandler := handlers.NewMangaHandler(mongoDB, r2Store)
	chapterHandler := handlers.NewChapterHandler(mongoDB, r2Store)
	genreHandler := handlers.NewGenreHandler(mongoDB)

	// Setup Router
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(120 * time.Second)) // generous timeout for large zip uploads

	// CORS Configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","time":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth
		r.Post("/auth/login", authHandler.Login)

		// Public Content
		r.Get("/home", genreHandler.GetHome)
		r.Get("/genres", genreHandler.GetGenres)
		r.Get("/stats", genreHandler.GetStats)

		r.Get("/manga", mangaHandler.GetMangas)
		r.Get("/manga/{id}", mangaHandler.GetManga)
		r.Get("/chapter/{id}", chapterHandler.GetChapter)

		// Admin Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(authService.RequireAdminMiddleware)

			r.Get("/auth/me", authHandler.Me)

			// Manga Management
			r.Post("/manga", mangaHandler.CreateManga)
			r.Put("/manga/{id}", mangaHandler.UpdateManga)
			r.Delete("/manga/{id}", mangaHandler.DeleteManga)

			// Chapter Management
			r.Post("/manga/{mangaId}/chapters", chapterHandler.UploadChapter)
			r.Delete("/chapter/{id}", chapterHandler.DeleteChapter)
		})
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server run context for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		log.Println("🛑 Shutdown signal received, shutting down gracefully...")
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
		serverStopCtx()
	}()

	log.Printf("⚡ Manga Host API server running on port :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Server ListenAndServe error: %v", err)
	}

	<-serverCtx.Done()
	log.Println("👋 Server stopped cleanly.")
}
