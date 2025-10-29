package repository

import (
	"context"
	"sambhav/pkg/database"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserService interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, userID string) (*repository.User, error)
	// Other methods...
}

type userRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewUserRepository(dbInstance database.Database) *userRepository {
	db := dbInstance.Connection()
	collection := db.Collection("users")

	return &userRepository{db: db, collection: collection}
}
