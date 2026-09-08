// Command api is the application entry point.
//
// It wires the whole dependency graph manually (composition root):
//   - load config
//   - open the PostgreSQL pool
//   - build the user repository, transaction manager, service, and handler
//   - start the HTTP server
//
// Note: Migrations are run via CLI command (make migrate-up), not at startup.
//
// @title PengBook API
// @version 1.0
// @description Pepeng Book Journal and Notes API
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"fmt"
	"os"

	"pengbook/api/internal/config"
	"pengbook/api/internal/database"
	"pengbook/api/internal/infrastructure/postgres"
	"pengbook/api/internal/infrastructure/redis"
	"pengbook/api/internal/module/auth"
	"pengbook/api/internal/module/user"
	"pengbook/api/internal/server"
	"pengbook/api/pkg/logger"
)

func main() {
	log := logger.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	log.Info("Config loaded", "config", cfg)

	pool, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("Connected to PostgreSQL", "host", cfg.Database.Host, "port", cfg.Database.Port, "database", cfg.Database.Name)

	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	txManager := postgres.NewTxManager(pool)
	
	userRepo := postgres.NewUserRepository(pool)
	userSvc := user.NewService(userRepo, txManager)
	userHandler := user.NewHandler(userSvc)
	
	authTokenRepo := redis.NewAuthTokenRepository(redisClient)
	authSvc := auth.NewService(userRepo, authTokenRepo, txManager, cfg.JWT.Secret)
	authHandler := auth.NewHandler(authSvc)

	log.Info("Connected to Redis", "host", cfg.Redis.Host, "port", cfg.Redis.Port)


	srv := server.NewHTTP(userHandler, authHandler, cfg.HTTPPort, log)

	log.Info("Starting HTTP server", "url", fmt.Sprintf("http://localhost:%d", cfg.HTTPPort), "swagger", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.HTTPPort))

	if err := srv.Run(); err != nil {
		log.Error("http server", "error", err)
		os.Exit(1)
	}

}
