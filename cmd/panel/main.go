package main

import (
	"context"

	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"juvia/internal/api"
	"juvia/internal/auth"
	"juvia/internal/db"
	"juvia/internal/events"
	"juvia/internal/socket"
	"juvia/internal/web"
	"juvia/internal/ws"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbPath := os.Getenv("JUVIA_DB")
	if dbPath == "" {
		dbPath = "./juvia.db"
	}

	database, err := db.OpenDB(dbPath)
	if err != nil {
		log.Error("open db", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.ApplyMigrations("/usr/share/juvia/migrations"); err != nil {
		log.Error("apply migrations", "error", err)
		os.Exit(1)
	}

	jwtSecret := os.Getenv("JUVIA_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}
	jwtManager := auth.NewJWTManager(jwtSecret)
	sessionStore := auth.NewSessionStore(database.DB)

	socketPath := os.Getenv("JUVIA_SOCKET")
	if socketPath == "" {
		socketPath = "/var/run/juvia/agent.sock"
	}
	agentClient := socket.NewClient(socketPath, log)
	defer agentClient.Close()

	eventBus := events.NewBus()
	wsServer := ws.NewServer(log, eventBus)

	cfg := api.RouterConfig{
		DB:          database,
		JWT:         jwtManager,
		Sessions:    sessionStore,
		AgentClient: agentClient,
		Log:         log,
		StaticFS:    web.StaticFS,
	}

	router := api.NewRouter(cfg)

	router.GET("/ws/v1/metrics", wsServer.HandleMetrics)
	router.GET("/ws/v1/tasks/:taskId", wsServer.HandleTasks)

	ctx := context.Background()

	bindAddr := os.Getenv("JUVIA_BIND")
	if bindAddr == "" {
		bindAddr, _ = database.GetSetting(ctx, "panel_bind_address")
		if bindAddr == "" {
			bindAddr = "0.0.0.0"
		}
	}

	port := os.Getenv("JUVIA_PORT")
	if port == "" {
		port, _ = database.GetSetting(ctx, "panel_port")
		if port == "" {
			hostname, _ := database.GetSetting(ctx, "panel_hostname")
			if hostname == "" {
				port = "18473"
			} else {
				port = "8080"
			}
		}
	}

	server := &http.Server{
		Addr:    bindAddr + ":" + port,
		Handler: router,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info("panel started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
		}
	}()

	<-sigCh
	log.Info("panel shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}
}
