// @title           TMS API
// @version         1.0
// @description     Tenant Management System — REST API
// @termsOfService  http://swagger.io/terms/
//
// @contact.name   TMS Support
// @contact.email  support@auditrakkr.com
//
// @license.name  MIT
//
// @host      localhost:8080
// @BasePath  /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/auditrakkr/tms-fullstack/tms-backend/docs"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
	"github.com/auditrakkr/tms-fullstack/tms-backend/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// ── Structured logging ──────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── Configuration ───────────────────────────────────────────
	config.LoadConfig()
	cfg := config.AppConfig

	// ── App settings (upload paths, protocol, file size limits) ─
	global.InitAppSettings(
		cfg.App.Protocol,
		cfg.Uploads.Directory,
		cfg.Uploads.PhotoFileSizeLimit,
		cfg.Uploads.LogoFileSizeLimit,
		cfg.Uploads.GeneralFileSizeLimit,
	)

	// ── Crypto (dedicated key, separate from JWT) ────────────────
	if err := utils.InitCrypto(cfg.CryptoKey); err != nil {
		slog.Warn("crypto key not configured — region password encryption disabled", "error", err)
	}

	// ── Mail (SMTP + configurable templates) ─────────────────────
	global.InitMailConfig(
		cfg.SMTP.Host,
		fmt.Sprintf("%d", cfg.SMTP.Port),
		cfg.SMTP.User,
		cfg.SMTP.Password,
	)
	global.InitMailTemplates(
		cfg.Mail.FromAddress,
		cfg.Mail.ConfirmEmailSubject,
		cfg.Mail.ConfirmEmailTemplate,
		cfg.Mail.ResetPasswordSubject,
		cfg.Mail.ResetPasswordTemplate,
	)

	// ── Database ────────────────────────────────────────────────
	database.ConnectDB()
	defer database.CloseDB()

	// ── Router ─────────────────────────────────────────────────
	r := gin.Default()

	// Session store for PKCE code verifier (Google OAuth2).
	sessionSecret := cfg.JWT.Secret
	if len(sessionSecret) > 32 {
		sessionSecret = sessionSecret[:32]
	}
	store := cookie.NewStore([]byte(sessionSecret))
	r.Use(sessions.Sessions("tms-session", store))

	r.Static("/assets", "./assets/dist")
	r.Static("/js", "./assets/js")
	r.LoadHTMLGlob("./views/**/*.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": global.APP_NAME})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	SetupRoutes(r, database.DB, database.Cache)

	// ── HTTP server with graceful shutdown ──────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("TMS server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received — draining requests...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
	slog.Info("server stopped")
}
