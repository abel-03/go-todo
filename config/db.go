package config

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// DB - переменная для представления объекта MongoDB базы данных.
var DB *mongo.Database

// ShoppingLists и Users - переменные для представления коллекций MongoDB.
var ShoppingLists, Users *mongo.Collection

// init - функция, вызываемая автоматически при запуске программы.
func init() {
	// Подключение к MongoDB с использованием URI, который хранится в переменной окружения "MONGO_DB_URI".
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_DB_URI")))
	if err != nil {
		// Если произошла ошибка при подключении, программа завершает выполнение с выводом ошибки.
		panic(err)
	}

	// Проверка, что установленное соединение с MongoDB работает (посылается ping запрос к Primary серверу).
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		// Если ping неудачен, программа завершает выполнение с выводом ошибки.
		panic(err)
	}
	fmt.Println("Successfully connected and pinged.")

	// Установка глобальных переменных для представления базы данных и коллекций MongoDB.
	DB = client.Database("planpulse")
	ShoppingLists = client.Database("planpulse").Collection("shoppingLists")
	Users = client.Database("planpulse").Collection("users")
}

