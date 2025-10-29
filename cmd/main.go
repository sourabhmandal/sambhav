package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sambhav/internal/general"
	"sambhav/internal/repository"
	"sambhav/internal/user"
	"sambhav/pkg/database"
	"sambhav/pkg/env"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := env.EnvVars()
	if err != nil {
		log.Fatalf("Error parsing environment variables: %v", err)
	}

	serverPort := cfg.ServerPort

	// Declare a flag to run migrations only
	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	dbInst := database.NewDatabaseMongo(
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseHost,
		cfg.DatabaseName,
		cfg.DatabaseAppName)

	newServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: registerRoutes(dbInst),
	}
	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(done, newServer, dbInst)

	// start the server
	if err := newServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}

func registerRoutes(dbInst database.Database) *gin.Engine {

	// declare generic handlers
	generalHandlers := general.NewGeneralHandler(dbInst)
	// declare user handlers
	userRepository := repository.NewUserRepository(dbInst)
	userService := user.NewUserService(userRepository)
	userHandlers := user.NewUserHandler(userService)

	router := gin.Default()
	// generic routes
	router.GET("/health", generalHandlers.HealthCheck)
	// user routes
	userRouter := router.Group("/user")
	userRouter.POST("/", userHandlers.RegisterUser)
	userRouter.GET("/", userHandlers.GetAllUsers)

	return router
}

func gracefulShutdown(done chan bool, server *http.Server, db database.Database) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// shut down any database connections
	db.Close()

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}
