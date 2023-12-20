package models

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetUserIdByName возвращает идентификатор пользователя по его имени.
func GetUserIdByName(userName string) (primitive.ObjectID, error) {
	// Ищем пользователя в базе данных по имени.
	sr := config.Users.FindOne(context.TODO(), bson.M{"name": userName})
	var u User
	err := sr.Decode(&u)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return u.ID, nil
}

// AllShareInviteShoppingLists возвращает все списки покупок, к которым пользователь приглашен.
func AllShareInviteShoppingLists(userId primitive.ObjectID) (*[]ShoppingList, error) {
	// Формируем конвейер для агрегации данных в MongoDB.
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.M{"sharingInviteIds": userId}},
		},
		{
			{Key: "$lookup", Value: bson.M{
				"from":         "users",
				"localField":   "ownerId",
				"foreignField": "_id",
				"as":           "owner",
			}},
		},
		{
			{Key: "$addFields", Value: bson.M{
				"ownerName": bson.M{"$arrayElemAt": []interface{}{"$owner.name", 0}},
			}},
		},
		{
			{Key: "$lookup", Value: bson.M{
				"from":         "users",
				"localField":   "sharingIds",
				"foreignField": "_id",
				"as":           "sharings",
			}},
		},
		{
			{Key: "$addFields", Value: bson.M{
				"sharingNames": "$sharings.name",
			}},
		},
	}

	// Выполняем агрегацию данных в MongoDB.
	var result []ShoppingList
	cursor, err := config.ShoppingLists.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	if err = cursor.All(context.Background(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ShareListWithUser делится списком покупок с пользователем.
func ShareListWithUser(listId, userId primitive.ObjectID) (bool, error) {
	// Формируем фильтр и обновление для добавления пользователя в список покупок и удаления приглашения.
	filter := bson.M{"_id": listId, "sharingInviteIds": userId}
	update := bson.M{
		"$addToSet": bson.M{
			"sharingIds": userId,
		},
		"$pull": bson.M{
			"sharingInviteIds": userId,
		},
	}
	// Выполняем обновление в MongoDB.
	result, err := config.ShoppingLists.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	if result.ModifiedCount == 0 {
		return false, nil
	}
	return true, nil
}

// DeclineShareListWithUser отклоняет приглашение пользователя к списку покупок.
func DeclineShareListWithUser(listId, userId primitive.ObjectID) (bool, error) {
	// Формируем фильтр и обновление для удаления приглашения пользователя к списку покупок.
	filter := bson.M{"_id": listId, "sharingInviteIds": userId}
	update := bson.M{
		"$pull": bson.M{
			"sharingInviteIds": userId,
		},
	}
	// Выполняем обновление в MongoDB.
	result, err := config.ShoppingLists.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return false, err
	}
	if result.ModifiedCount == 0 {
		return false, nil
	}
	return true, nil
}

// AddShareInvite добавляет приглашение пользователя к списку покупок.
func AddShareInvite(ownerId, listId, userId primitive.ObjectID) (bool, error) {
	// Формируем обновление для добавления пользователя в список покупок.
	update := bson.M{
		"$addToSet": bson.M{
			"sharingInviteIds": userId,
		},
	}
	// Выполняем обновление в MongoDB.
	result, err := config.ShoppingLists.UpdateOne(context.TODO(), bson.M{"ownerId": ownerId, "_id": listId}, update)

	if err != nil || result.ModifiedCount != 1 {
		return false, err
	}
	return true, nil
}

