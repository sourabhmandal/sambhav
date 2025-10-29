package user

import (
	"database/sql"
	"errors"
	"fmt"
	"sambhav/internal/repository"
	"sambhav/pkg/database"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
)

type UserService interface {
	RegisterUser(ctx context.Context, name, email string) error
	GetAllUsers(ctx context.Context) ([]*database.User, error)
	GetUserByID(ctx context.Context, userID string) (*database.User, error)
}

type userService struct {
	userRepository repository.UserService
}

func NewUserService(userRepo repository.UserService) UserService {
	return &userService{userRepository: userRepo}
}

// RegisterUser registers a new user in the system.
func (u *userService) RegisterUser(ctx context.Context, name, email string) error {
	// Check if the user already exists based on email
	existingUser, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// Create a new user in the schema
		_, err = u.userRepository.CreateUser(ctx, &database.User{
			ID:    primitive.NewObjectID(),
			Email: email,
			Name:  name,
			Bio:   nil, // Assuming empty bio for new user
		})
		if err != nil {
			return err
		}
	} else if existingUser != nil {
		return errors.New("user already exists")
	} else {
		return err
	}

	return nil
}

// GetUserByID retrieves a user by their ID.
func (u *userService) GetAllUsers(ctx context.Context) ([]*database.User, error) {

	// Get the user by ID from the schema
	users, err := u.userRepository.ListAllUsers(ctx)
	fmt.Println(users, err)

	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserByID retrieves a user by their ID.
func (u *userService) GetUserByID(ctx context.Context, userID string) (*database.User, error) {
	// Get the user by ID from the schema
	user, err := u.userRepository.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
