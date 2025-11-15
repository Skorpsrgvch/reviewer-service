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

	// Адаптеры
	"github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/middleware"
	"github.com/Skorpsrgvch/reviewer-service/internal/adapter/postgres"

	// Юзкейсы
	statsUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/stats/get"
	teamCreateUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/team/create"
	teamGetUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/team/get"

	userGetReviewUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/user/getReview"
	userSetActiveUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/user/setActive"

	prCreateUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/pullrequest/create"
	prMergeUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/pullrequest/merge"
	prReassignUC "github.com/Skorpsrgvch/reviewer-service/internal/usecase/pullrequest/reassign"

	// Хендлеры
	prHttp "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/pullrequest"
	statsHttp "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/stats"
	teamHttp "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/team"
	userHttp "github.com/Skorpsrgvch/reviewer-service/internal/adapter/http/user"

	// База
	"github.com/Skorpsrgvch/reviewer-service/pkg/db"

	// Миграции
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/gin-gonic/gin"
)

func runMigrations(dbURL string) error {
	log.Println("Running migrations...")

	m, err := migrate.New("file://migrations", dbURL)
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

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@db:5432/avito?sslmode=disable"
	}

	// Ждём готовности БД
	if err := waitForDB(dbURL); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Запускаем миграции
	if err := runMigrations(dbURL); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	// Подключаемся к БД
	dbConn, err := db.NewPostgresDB(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbConn.Close()

	// === Адаптеры (репозитории) ===
	teamRepo := postgres.NewTeamRepo(dbConn)
	userRepo := postgres.NewUserRepo(dbConn)
	prRepo := postgres.NewPullRequestRepo(dbConn)

	statsRepo := postgres.NewPullRequestRepo(dbConn)

	getStatsUC, _ := statsUC.NewUsecase(statsRepo)
	getStatsHandler := statsHttp.NewGetHandler(getStatsUC)

	// === Юзкейсы ===
	createTeamUC, err := teamCreateUC.NewUsecase(teamRepo)
	if err != nil {
		log.Fatalf("Failed to init createTeamUC: %v", err)
	}

	getTeamUC, err := teamGetUC.NewUsecase(teamRepo)
	if err != nil {
		log.Fatalf("Failed to init getTeamUC: %v", err)
	}

	setActiveUC, err := userSetActiveUC.NewUsecase(userRepo, userRepo)
	if err != nil {
		log.Fatalf("Failed to init setActiveUC: %v", err)
	}

	getReviewUC, err := userGetReviewUC.NewUsecase(prRepo, userRepo)
	if err != nil {
		log.Fatalf("Failed to init getReviewUC: %v", err)
	}

	createPRUC, err := prCreateUC.NewUsecase(prRepo, userRepo, teamRepo)
	if err != nil {
		log.Fatalf("Failed to init createPRUC: %v", err)
	}

	mergePRUC, err := prMergeUC.NewUsecase(prRepo, prRepo)
	if err != nil {
		log.Fatalf("Failed to init mergePRUC: %v", err)
	}

	reassignPRUC, err := prReassignUC.NewUsecase(prRepo, userRepo, teamRepo)
	if err != nil {
		log.Fatalf("Failed to init reassignPRUC: %v", err)
	}

	// === Хендлеры ===
	createTeamHandler := teamHttp.NewCreateHandler(createTeamUC)
	getTeamHandler := teamHttp.NewGetHandler(getTeamUC)

	setActiveHandler := userHttp.NewSetActiveHandler(setActiveUC)
	getReviewHandler := userHttp.NewGetReviewHandler(getReviewUC)

	createPRHandler := prHttp.NewCreateHandler(createPRUC)
	mergePRHandler := prHttp.NewMergeHandler(mergePRUC)
	reassignPRHandler := prHttp.NewReassignHandler(reassignPRUC)

	// === Роутер ===
	r := gin.New()
	r.Use(gin.Recovery())

	adminGroup := r.Group("/")
	adminGroup.Use(middleware.AuthMiddleware())
	{
		adminGroup.POST("/team/add", createTeamHandler.Handle)

		adminGroup.POST("/users/setIsActive", setActiveHandler.Handle)

		adminGroup.POST("/pullRequest/create", createPRHandler.Handle)
		adminGroup.POST("/pullRequest/merge", mergePRHandler.Handle)
		adminGroup.POST("/pullRequest/reassign", reassignPRHandler.Handle)
	}
	r.GET("/stats", getStatsHandler.Handle)
	r.GET("/team/get", getTeamHandler.Handle)
	r.GET("/users/getReview", getReviewHandler.Handle)

	// Запуск сервера
	srv := &http.Server{Addr: ":8080", Handler: r}

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}

// waitForDB — ожидание подключения к БД
func waitForDB(dbURL string) error {
	for i := 0; i < 30; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		dbConn, err := db.NewPostgresDB(ctx, dbURL)
		cancel()
		if err != nil {
			log.Printf("Waiting for DB... attempt %d", i+1)
			time.Sleep(2 * time.Second)
			continue
		}
		dbConn.Close()
		return nil
	}
	return fmt.Errorf("timeout waiting for database")
}
