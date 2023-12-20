package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User представляет собой модель пользователя.
type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name     string             `json:"name"`
	Password string             `json:"password"`
}

// AddUser добавляет нового пользователя в базу данных.
func AddUser(u User) error {
	// Вставляем пользователя в коллекцию "users".
	_, err := config.Users.InsertOne(context.TODO(), u)
	if err != nil {
		return err
	}

	return nil
}

