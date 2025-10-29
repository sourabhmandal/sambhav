package database

import (
	"fmt"
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"golang.org/x/net/context"
)

// Service represents a service that interacts with a database.
type Database interface {

	// Connection returns the underlying database connection.
	Connection() *mongo.Database
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close()
}

type mongoDatabase struct {
	client *mongo.Client
	db     *mongo.Database
}

var (
	dbInstance *mongoDatabase
)

func NewDatabaseMongo(username, password, host, name, appName string) Database {
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("mongodb+srv://%s:%s@%s/?appName=%s", username, password, host, appName)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connStr).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}
	dbInstance = &mongoDatabase{
		client: client,
		db:     client.Database(name),
	}
	return dbInstance
}

func (s *mongoDatabase) Connection() *mongo.Database {
	return s.db
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *mongoDatabase) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	if err := s.client.Ping(ctx, readpref.Primary()); err != nil {
		stats["status"] = "down"
		stats["message"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	return stats
}

// Close closes the database connection.
func (s *mongoDatabase) Close() {
	defer func() {
		if err := s.client.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()
	log.Printf("Disconnected from database")
}
