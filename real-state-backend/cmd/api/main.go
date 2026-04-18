package main

import (
	"database/sql"
	_ "fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	// Drivers
	_ "github.com/lib/pq" // O github.com/go-sql-driver/mysql

	// Internal
	"real-state-backend/config"
	"real-state-backend/internal/core/ports"
	"real-state-backend/internal/handlers"
	"real-state-backend/internal/repository"
	"real-state-backend/internal/services"
	"real-state-backend/pkg/cache"
	"real-state-backend/pkg/middleware"
)

func main() {
	// 1. Cargar Configuración
	cfg := config.LoadConfig()

	// 2. Logger Estructurado (JSON para producción)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 3. Conexión a Base de Datos
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		slog.Error("Failed to connect to DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Configurar pool de conexiones (Clave para escalabilidad)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Cargar configuración de seguridad desde BD
	cfg = config.LoadConfigFromDB(cfg, db)

	// 3.5 Inicializar Redis Cache (opcional, fallback a sin cache si falla)
	var redisCache cache.Cache
	redisCache, err = cache.NewRedisCache(cfg.RedisAddr)
	if err != nil {
		slog.Warn("Redis cache unavailable, running without cache", "error", err)
		redisCache = nil
	}
	defer func() {
		if rc, ok := redisCache.(*cache.RedisCache); ok {
			if err := rc.Close(); err != nil {
				slog.Warn("Failed to close Redis connection", "error", err)
			}
		}
	}()

	// 4. Inyección de Dependencias
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	authServiceBase := services.NewAuthService(userRepo, sessionRepo, auditRepo, cfg.JWTSecret, cfg.JWTPepper, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.MaxFailedAttempts, cfg.LockoutDuration)

	// Usar authServiceBase como authService
	var authService ports.AuthService = authServiceBase
	if redisCache != nil {
		authService = services.NewCachedAuthService(authServiceBase, redisCache, cfg.CacheTTL)
		slog.Info("AuthService initialized with Redis caching")
	} else {
		slog.Info("AuthService initialized without caching")
	}

	authHandler := handlers.NewAuthHandler(authService)

	propRepo := repository.NewPropertyRepository(db)
	propService := services.NewPropertyService(propRepo)
	propHandler := handlers.NewPropertyHandler(propService)

	configRepo := repository.NewSecurityConfigRepository(db)
	configHandler := handlers.NewConfigHandler(configRepo, auditRepo, authService)

	// 5. Router y Rutas
	mux := http.NewServeMux()
	// Endpoints públicos (v1)
	mux.HandleFunc("POST /v1/login", authHandler.Login)
	mux.HandleFunc("POST /v1/refresh", authHandler.RefreshToken)
	// Backward compatibility (sin versión)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /refresh", authHandler.RefreshToken)

	// Subrouter para rutas protegidas (v1)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("POST /v1/verify-mfa", authHandler.VerifyMFA)
	protectedMux.HandleFunc("POST /v1/logout", authHandler.Logout)
	protectedMux.HandleFunc("GET /v1/properties", propHandler.GetAll)
	protectedMux.HandleFunc("GET /v1/properties/{id}", propHandler.GetByID)
	protectedMux.HandleFunc("POST /v1/properties", propHandler.CreateProperty)
	protectedMux.HandleFunc("GET /v1/config", configHandler.GetSecurityConfig)
	protectedMux.HandleFunc("PUT /v1/config", configHandler.UpdateSecurityConfig)
	// Backward compatibility (sin versión)
	protectedMux.HandleFunc("POST /verify-mfa", authHandler.VerifyMFA)
	protectedMux.HandleFunc("POST /logout", authHandler.Logout)
	protectedMux.HandleFunc("GET /properties", propHandler.GetAll)
	protectedMux.HandleFunc("GET /properties/{id}", propHandler.GetByID)
	protectedMux.HandleFunc("POST /properties", propHandler.CreateProperty)
	protectedMux.HandleFunc("GET /config", configHandler.GetSecurityConfig)
	protectedMux.HandleFunc("PUT /config", configHandler.UpdateSecurityConfig)

	// Aplicar middlewares
	jwtMiddleware := middleware.JWTMiddleware(authService, cfg.JWTSecret)

	protectedHandler := jwtMiddleware(protectedMux)

	// Combinar routers (v1)
	mux.Handle("/v1/verify-mfa", protectedHandler)
	mux.Handle("/v1/logout", protectedHandler)
	mux.Handle("/v1/properties", protectedHandler)
	mux.Handle("/v1/properties/", protectedHandler)
	mux.Handle("POST /v1/properties", protectedHandler)

	// Rutas para /v1/config (protegidas, RBAC validado en el handler)
	mux.Handle("/v1/config", protectedHandler)
	mux.Handle("/v1/config/", protectedHandler)

	// Backward compatibility (sin versión)
	mux.Handle("/verify-mfa", protectedHandler)
	mux.Handle("/logout", protectedHandler)
	mux.Handle("/properties", protectedHandler)
	mux.Handle("/properties/", protectedHandler)
	mux.Handle("POST /properties", protectedHandler)
	mux.Handle("/config", protectedHandler)
	mux.Handle("/config/", protectedHandler)

	// 6. Aplicar Middleware
	handlerWithMiddleware := middleware.SecurityMiddleware(cfg.AllowedOrigins)(
		middleware.ClientIdentificationMiddleware(
			middleware.RequestIDMiddleware(mux),
		),
	)

	// 7. Iniciar Servidor
	slog.Info("Server starting", "port", cfg.ServerPort)
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handlerWithMiddleware,
		ReadTimeout:  10 * time.Second, // Protección contra Slowloris attacks
		WriteTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("Server failed", "error", err)
	}
}
