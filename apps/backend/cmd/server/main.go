package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cifo-monitoring/backend/internal/config"
	"github.com/cifo-monitoring/backend/internal/handler"
	"github.com/cifo-monitoring/backend/internal/integration"
	"github.com/cifo-monitoring/backend/internal/middleware"
	"github.com/cifo-monitoring/backend/internal/repository"
	"github.com/cifo-monitoring/backend/internal/service"
	"github.com/cifo-monitoring/backend/internal/ws"
	"github.com/cifo-monitoring/backend/pkg/logger"
	"github.com/cifo-monitoring/backend/pkg/validator"
	"github.com/labstack/echo/v4"
)

func main() {
	// load env config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// init structured logger
	appLogger := logger.New(cfg.LogLevel)
	appLogger.Info("starting cifo backend server",
		slog.String("env", cfg.Environment),
		slog.Int("port", cfg.Port),
	)

	// create base context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// init database pool
	dbPool, err := repository.NewPostgresPool(ctx, cfg.DatabaseDSN, appLogger)
	if err != nil {
		appLogger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	// init redis client
	redisClient, err := repository.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPass, appLogger)
	if err != nil {
		appLogger.Error("failed to connect redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	// run database migrations
	migrationsDir := findMigrationsDir()
	migrator := repository.NewMigrator(dbPool, migrationsDir, appLogger)
	if err := migrator.Up(ctx); err != nil {
		appLogger.Error("database migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	appLogger.Info("database migrations verified")

	// init echo router
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = validator.New()

	// init repositories
	userRepo := repository.NewUserRepository(dbPool)
	auditRepo := repository.NewAuditRepository(dbPool)

	// init auth service & jwks cache
	jwksCache := service.NewJWKSCache(cfg.KeycloakJWKSURL, 1*time.Hour)
	authService := service.NewAuthService(cfg, userRepo, auditRepo, redisClient, jwksCache, appLogger)

	// init docker client & services
	dockerClient, err := integration.NewDockerClient(cfg.DockerHost)
	if err != nil {
		appLogger.Warn("failed to initialize docker client", slog.String("error", err.Error()))
	} else {
		defer dockerClient.Close()
	}

	dockerService := service.NewDockerService(dockerClient, auditRepo, redisClient, appLogger)

	// init kubernetes & argocd clients
	var k8sService service.KubernetesService
	var argoService service.ArgoCDService

	k8sClient, err := integration.NewKubernetesClient("")
	if err != nil {
		appLogger.Warn("failed to initialize kubernetes client", slog.String("error", err.Error()))
	} else {
		appLogger.Info("kubernetes client initialized successfully")
		k8sService = service.NewKubernetesService(k8sClient, auditRepo, appLogger)

		if restCfg := k8sClient.GetRESTConfig(); restCfg != nil {
			argoClient, aErr := integration.NewArgoCDClient(restCfg)
			if aErr != nil {
				appLogger.Warn("failed to initialize argocd client", slog.String("error", aErr.Error()))
			} else {
				appLogger.Info("argocd dynamic client initialized successfully")
				argoService = service.NewArgoCDService(argoClient, auditRepo, appLogger)
			}
		}
	}

	monitoringService := service.NewMonitoringService(dockerService, k8sService, appLogger)

	// init telegram & incident services
	telegramClient := integration.NewTelegramClient(cfg.TelegramToken, cfg.TelegramChatID, appLogger)
	telegramService := service.NewTelegramService(telegramClient, redisClient, appLogger)
	incidentRepo := repository.NewIncidentRepository(dbPool)

	// init websocket hub & streamer
	wsHub := ws.NewHub(appLogger)
	go wsHub.Run(ctx)
	wsStreamer := ws.NewStreamer(wsHub, dockerClient, k8sClient, appLogger)
	wsStreamer.Start(ctx)
	wsHandler := handler.NewWebSocketHandler(wsHub, authService, appLogger)

	notificationService := service.NewNotificationService(telegramService, wsHub, incidentRepo, appLogger)
	incidentService := service.NewIncidentService(incidentRepo, auditRepo, notificationService, appLogger)

	// background escalation & retry ticker
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = incidentService.CheckEscalations(ctx)
				_ = telegramService.ProcessRetryQueue(ctx)
			}
		}
	}()

	// init handlers
	healthHandler := handler.NewHealthHandler(dbPool, redisClient, cfg.DockerHost)
	metrics := handler.NewMetrics(dbPool)
	rateLimiter := middleware.NewRateLimiter(redisClient, appLogger)
	authHandler := handler.NewAuthHandler(authService, userRepo, auditRepo)
	dockerHandler := handler.NewDockerHandler(dockerService)
	monitoringHandler := handler.NewMonitoringHandler(monitoringService)
	k8sHandler := handler.NewKubernetesHandler(k8sService)
	argoHandler := handler.NewArgoCDHandler(argoService)
	incidentHandler := handler.NewIncidentHandler(incidentService)

	// init ai engine
	aiRepo := repository.NewPostgresAIRepository(dbPool)
	aiClient := integration.NewHTTPAIClient(cfg.AIServiceURL)
	aiService := service.NewDefaultAIService(aiRepo, incidentRepo, aiClient, dockerService, k8sService, argoService)
	aiHandler := handler.NewAIHandler(aiService)

	// init settings engine
	settingsRepo := repository.NewSettingsRepository(dbPool)
	settingsService := service.NewSettingsService(settingsRepo, userRepo, auditRepo, telegramService, appLogger)
	settingsHandler := handler.NewSettingsHandler(settingsService)

	// register global middleware
	e.Use(middleware.RequestLogger(appLogger))
	e.Use(middleware.Recover(appLogger))
	e.Use(middleware.CORS(cfg.AllowedOrigins))
	e.Use(metrics.Middleware())
	e.Use(rateLimiter.LimitIP(100, time.Minute))

	// register probe handlers
	e.GET("/healthz", healthHandler.Liveness)
	e.GET("/readyz", healthHandler.Readiness)
	e.GET("/metrics", metrics.Handler())

	// register websocket route
	e.GET("/ws", wsHandler.HandleWebSocket)

	// register auth and admin routes
	api := e.Group("/api/v1")
	api.POST("/auth/login", authHandler.Login)
	api.POST("/webhooks/alertmanager", incidentHandler.HandleAlertmanagerWebhook)

	protectedAuth := api.Group("/auth", middleware.RequireAuth(authService))
	protectedAuth.GET("/me", authHandler.Me)
	protectedAuth.POST("/logout", authHandler.Logout)

	admin := api.Group("/admin", middleware.RequireAuth(authService), middleware.RequireRole(authService, "admin"))
	admin.GET("/users", authHandler.ListUsers)
	admin.GET("/audit-logs", authHandler.ListAuditLogs)

	// register docker routes
	dockerGroup := api.Group("/docker", middleware.RequireAuth(authService))
	dockerGroup.GET("/containers", dockerHandler.ListContainers)
	dockerGroup.GET("/containers/:id", dockerHandler.GetContainer)
	dockerGroup.GET("/containers/:id/stats", dockerHandler.GetContainerStats)
	dockerGroup.GET("/containers/:id/logs", dockerHandler.GetContainerLogs)
	dockerGroup.POST("/containers/:id/restart", dockerHandler.RestartContainer, middleware.RequireRole(authService, "devops"))
	dockerGroup.POST("/containers/:id/stop", dockerHandler.StopContainer, middleware.RequireRole(authService, "admin"))
	dockerGroup.GET("/images", dockerHandler.ListImages)
	dockerGroup.GET("/volumes", dockerHandler.ListVolumes)
	dockerGroup.GET("/networks", dockerHandler.ListNetworks)
	dockerGroup.GET("/system", dockerHandler.GetSystemInfo)

	// register kubernetes routes
	k8sGroup := api.Group("/kubernetes", middleware.RequireAuth(authService))
	k8sGroup.GET("/pods", k8sHandler.ListPods)
	k8sGroup.GET("/pods/:namespace/:name", k8sHandler.GetPod)
	k8sGroup.GET("/pods/:namespace/:name/logs", k8sHandler.GetPodLogs)
	k8sGroup.GET("/deployments", k8sHandler.ListDeployments)
	k8sGroup.GET("/deployments/:namespace/:name", k8sHandler.GetDeployment)
	k8sGroup.POST("/deployments/:namespace/:name/restart", k8sHandler.RestartDeployment, middleware.RequireRole(authService, "devops"))
	k8sGroup.POST("/deployments/:namespace/:name/scale", k8sHandler.ScaleDeployment, middleware.RequireRole(authService, "devops"))
	k8sGroup.GET("/nodes", k8sHandler.ListNodes)
	k8sGroup.GET("/services", k8sHandler.ListServices)
	k8sGroup.GET("/overview", k8sHandler.GetClusterOverview)

	// register argocd routes
	argoGroup := api.Group("/argocd", middleware.RequireAuth(authService))
	argoGroup.GET("/applications", argoHandler.ListApplications)
	argoGroup.GET("/applications/:name", argoHandler.GetApplication)
	argoGroup.POST("/applications/:name/sync", argoHandler.SyncApplication, middleware.RequireRole(authService, "devops"))
	argoGroup.GET("/overview", argoHandler.GetOverview)

	// register monitoring routes
	monitoringGroup := api.Group("/monitoring", middleware.RequireAuth(authService))
	monitoringGroup.GET("/stats", monitoringHandler.GetStats)
	monitoringGroup.GET("/metrics/cpu", monitoringHandler.GetCPUMetrics)
	monitoringGroup.GET("/metrics/memory", monitoringHandler.GetMemoryMetrics)
	monitoringGroup.GET("/metrics/network", monitoringHandler.GetNetworkMetrics)

	// register incident routes
	incidentGroup := api.Group("/incidents", middleware.RequireAuth(authService))
	incidentGroup.GET("", incidentHandler.ListIncidents)
	incidentGroup.GET("/stats", incidentHandler.GetIncidentStats)
	incidentGroup.GET("/:id", incidentHandler.GetIncident)
	incidentGroup.POST("/:id/acknowledge", incidentHandler.AcknowledgeIncident, middleware.RequireRole(authService, "devops"))
	incidentGroup.POST("/:id/resolve", incidentHandler.ResolveIncident, middleware.RequireRole(authService, "devops"))
	incidentGroup.POST("/:id/close", incidentHandler.CloseIncident, middleware.RequireRole(authService, "admin"))
	incidentGroup.POST("/:id/rca", aiHandler.HandleGenerateRCA)

	// register ai routes
	aiGroup := api.Group("/ai", middleware.RequireAuth(authService))
	aiGroup.GET("/models", aiHandler.HandleListModels)
	aiGroup.POST("/chat", aiHandler.HandleChat)
	aiGroup.GET("/sessions", aiHandler.HandleListSessions)
	aiGroup.GET("/sessions/:id/messages", aiHandler.HandleGetSessionMessages)
	aiGroup.POST("/tools/:id/approve", aiHandler.HandleApproveTool)
	aiGroup.POST("/tools/:id/reject", aiHandler.HandleRejectTool)
	aiGroup.GET("/usage", aiHandler.HandleGetUsage)

	// register settings routes (admin only)
	settingsGroup := api.Group("/settings", middleware.RequireAuth(authService), middleware.RequireRole(authService, "admin"))
	settingsGroup.GET("", settingsHandler.GetSettings)
	settingsGroup.PUT("", settingsHandler.UpdateSettings)
	settingsGroup.POST("/notifications/test", settingsHandler.TestNotification)
	settingsGroup.POST("/test-notification", settingsHandler.TestNotification)
	settingsGroup.GET("/users", settingsHandler.ListUsers)
	settingsGroup.PUT("/users/:id/role", settingsHandler.UpdateUserRole)
	settingsGroup.DELETE("/users/:id", settingsHandler.DeactivateUser)
	settingsGroup.POST("/users/:id/deactivate", settingsHandler.DeactivateUser)
	settingsGroup.PUT("/users/:id/activate", settingsHandler.ReactivateUser)
	settingsGroup.POST("/users/:id/reactivate", settingsHandler.ReactivateUser)

	// start http server
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	go func() {
		if err := e.Start(serverAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("http server failed", slog.String("error", err.Error()))
			stop()
		}
	}()
	appLogger.Info("http server listening", slog.String("addr", serverAddr))

	// wait for termination signal
	<-ctx.Done()
	appLogger.Info("shutting down server gracefully")

	// timeout for active connections
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("error during server shutdown", slog.String("error", err.Error()))
	}

	appLogger.Info("server shutdown complete")
}

// findMigrationsDir locates migrations
func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"apps/backend/migrations",
		"../migrations",
		"../../migrations",
	}

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return "migrations"
}
