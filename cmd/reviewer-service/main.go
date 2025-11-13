package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	prHandler "github.com/Skorpsrgvch/reviewer-service/internal/handler/pullrequest"
	statsHandler "github.com/Skorpsrgvch/reviewer-service/internal/handler/stats"
	teamHandler "github.com/Skorpsrgvch/reviewer-service/internal/handler/team"
	userHandler "github.com/Skorpsrgvch/reviewer-service/internal/handler/user"

	prSvc "github.com/Skorpsrgvch/reviewer-service/internal/service/pullrequest"
	statsSvc "github.com/Skorpsrgvch/reviewer-service/internal/service/stats"
	teamSvc "github.com/Skorpsrgvch/reviewer-service/internal/service/team"
	userSvc "github.com/Skorpsrgvch/reviewer-service/internal/service/user"

	"github.com/Skorpsrgvch/reviewer-service/internal/repository/postgres"
	"github.com/Skorpsrgvch/reviewer-service/pkg/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dbURL string) error {
	log.Println("Running migrations...")

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@db:5432/avito?sslmode=disable"
	}

	// Ждем готовности БД
	if err := waitForDB(dbURL); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Запускаем миграции
	if err := runMigrations(dbURL); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	// Подключаемся к БД
	dbConn, err := db.NewPostgresDB(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbConn.Close()

	// Репозиторий (реализует все интерфейсы)
	repo := postgres.NewDBRepo(dbConn)

	// Сервисы
	teamSvc := teamSvc.NewService(repo)
	userSvc := userSvc.NewService(repo, repo)
	prSvc := prSvc.NewService(repo, repo)
	statsSvc := statsSvc.NewService(repo)

	// Хендлеры
	teamHandler := teamHandler.NewHandler(teamSvc)
	userHandler := userHandler.NewHandler(userSvc)
	prHandler := prHandler.NewHandler(prSvc)
	statsHandler := statsHandler.NewHandler(statsSvc)

	// Роутер
	r := gin.New()
	r.Use(gin.Recovery())

	// Роуты
	r.POST("/team/add", teamHandler.AddTeam)
	r.GET("/team/get", teamHandler.GetTeam)

	r.POST("/users/setIsActive", userHandler.SetIsActive)
	r.GET("/users/getReview", userHandler.GetReview)

	r.POST("/pullRequest/create", prHandler.Create)
	r.POST("/pullRequest/merge", prHandler.Merge)
	r.POST("/pullRequest/reassign", prHandler.Reassign)

	r.GET("/stats", statsHandler.GetStats)

	// Сервер
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}

func waitForDB(dbURL string) error {
	for i := 0; i < 30; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		db, err := db.NewPostgresDB(ctx, dbURL)
		if err != nil {
			cancel()
			time.Sleep(2 * time.Second)
			continue
		}
		db.Close()
		cancel()

		return nil
	}
	return fmt.Errorf("timeout waiting for database")
}
