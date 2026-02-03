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
	"real-state-backend/internal/handlers"
	"real-state-backend/internal/repository"
	"real-state-backend/internal/services"
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

	// 4. Inyección de Dependencias
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	authService := services.NewAuthService(userRepo, sessionRepo, auditRepo, cfg.JWTSecret, cfg.JWTPepper, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.MaxFailedAttempts, cfg.LockoutDuration)
	authHandler := handlers.NewAuthHandler(authService)

	propRepo := repository.NewPropertyRepository(db)
	propService := services.NewPropertyService(propRepo)
	propHandler := handlers.NewPropertyHandler(propService)

	configRepo := repository.NewSecurityConfigRepository(db)
	configHandler := handlers.NewConfigHandler(configRepo, auditRepo)

	// 5. Router y Rutas
	mux := http.NewServeMux()
	// Endpoints públicos
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /refresh", authHandler.RefreshToken)

	// Subrouter para rutas protegidas
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("POST /verify-mfa", authHandler.VerifyMFA)
	protectedMux.HandleFunc("POST /logout", authHandler.Logout)
	protectedMux.HandleFunc("GET /properties", propHandler.GetAll)
	protectedMux.HandleFunc("GET /properties/{id}", propHandler.GetByID)
	protectedMux.HandleFunc("POST /properties", propHandler.CreateProperty)
	// Manejar /config por método. PUT requiere permiso 'manage_security_config'
	protectedMux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			configHandler.GetSecurityConfig(w, r)
		case http.MethodPut:
			// Aplica RBAC sólo para la acción de actualizar configuración
			rbacManage := middleware.RBACMiddleware(authService, "manage_security_config")
			rbacManage(http.HandlerFunc(configHandler.UpdateSecurityConfig)).ServeHTTP(w, r)
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// Aplicar middlewares
	jwtMiddleware := middleware.JWTMiddleware(authService, cfg.JWTSecret)
	rbacMiddleware := middleware.RBACMiddleware(authService, "create_property")

	protectedHandler := jwtMiddleware(protectedMux)
	// Para rutas específicas con RBAC, aplicar adicionalmente
	createPropertyHandler := jwtMiddleware(rbacMiddleware(http.HandlerFunc(propHandler.CreateProperty)))

	// Combinar routers
	mux.Handle("/verify-mfa", protectedHandler)
	mux.Handle("/logout", protectedHandler)
	mux.Handle("/properties", protectedHandler)
	mux.Handle("/properties/", protectedHandler)
	mux.Handle("POST /properties", createPropertyHandler) // Sobrescribir con RBAC

	// Rutas para /config (protegidas). GET/PUT se despachan dentro de protectedMux
	mux.Handle("/config", protectedHandler)
	mux.Handle("/config/", protectedHandler)

	// 6. Aplicar Middleware
	handlerWithMiddleware := middleware.SecurityMiddleware(mux)

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
