package database

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"` // MongoDB's _id field
	Name  string             `bson:"name" json:"name"`
	Email string             `bson:"email" json:"email"`
	Bio   *string            `bson:"bio" json:"bio"`
}
