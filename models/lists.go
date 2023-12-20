package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ListItem представляет элемент списка покупок.
type ListItem struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name"`
	IsCompleted bool               `json:"isCompleted" bson:"isCompleted"`
}

// ShoppingList представляет список покупок.
type ShoppingList struct {
	ID               primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	OwnerId          primitive.ObjectID   `json:"ownerId" bson:"ownerId"`
	OwnerName        string               `json:"ownerName" bson:"ownerName"`
	Name             string               `json:"name" bson:"name"`
	Items            []ListItem           `json:"items" bson:"items"`
	SharingIds       []primitive.ObjectID `json:"sharingIds" bson:"sharingIds"`
	SharingInviteIds []primitive.ObjectID `json:"sharingInviteIds" bson:"sharingInviteIds"`
	SharingNames     []string             `json:"sharingNames" bson:"sharingNames"`
}

// AllShoppingLists возвращает все списки покупок для заданного пользователя.
func AllShoppingLists(userId primitive.ObjectID) (*[]ShoppingList, error) {
	// Формируем конвейер для агрегации данных в MongoDB.
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.M{
				"$or": []interface{}{
					bson.M{"ownerId": userId},
					bson.M{"sharingIds": userId},
				},
			}},
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

// AddListItem добавляет новый элемент в список покупок.
func AddListItem(name string, userId, listId primitive.ObjectID) error {
	// Создаем новый элемент списка покупок.
	li := ListItem{
		ID:          primitive.NewObjectID(),
		Name:        name,
		IsCompleted: false,
	}

	// Формируем фильтр для определения списка, к которому добавляется элемент.
	filter := bson.M{
		"_id": listId,
		"$or": bson.A{
			bson.M{"ownerId": userId},
			bson.M{"sharingIds": userId},
		},
	}

	// Формируем обновление для добавления элемента в список.
	update := bson.M{
		"$push": bson.M{
			"items": li,
		},
	}

	// Выполняем обновление в MongoDB.
	_, err := config.ShoppingLists.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

// ModifyListItem обновляет информацию об элементе списка покупок.
func ModifyListItem(userId primitive.ObjectID, li ListItem) error {
	// Формируем фильтр для поиска элемента списка.
	filter := bson.M{
		"items._id": li.ID,
		"$or": bson.A{
			bson.M{"ownerId": userId},
			bson.M{"sharingIds": userId},
		},
	}

	// Формируем обновление для изменения информации об элементе.
	update := bson.M{
		"$set": bson.M{
			"items.$": li,
		},
	}

	// Выполняем обновление в MongoDB.
	_, err := config.ShoppingLists.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

// AddNewShoppingList добавляет новый список покупок для пользователя.
func AddNewShoppingList(name string, ownerId primitive.ObjectID) (string, error) {
	// Создаем новый список покупок.
	t := ShoppingList{
		ID:               primitive.NewObjectID(),
		OwnerId:          ownerId,
		Name:             name,
		Items:            make([]ListItem, 0),
		SharingIds:       make([]primitive.ObjectID, 0),
		SharingInviteIds: make([]primitive.ObjectID, 0),
	}
	// Вставляем новый список в MongoDB.
	_, err := config.ShoppingLists.InsertOne(context.TODO(), t)
	if err != nil {
		return "", err
	}
	return t.ID.Hex(), nil
}

// AddShoppingLists добавляет несколько списков покупок для пользователя.
func AddShoppingLists(sl []ShoppingList, ownerId primitive.ObjectID) error {
	// Преобразуем структуры данных в формат, подходящий для вставки в MongoDB.
	slInterface := make([]interface{}, len(sl))
	for i := range sl {
		slInterface[i] = sl[i]
	}

	// Вставляем несколько списков покупок в MongoDB.
	_, err := config.ShoppingLists.InsertMany(context.TODO(), slInterface)
	if err != nil {
		return err
	}
	return nil
}

// CheckoutList удаляет завершенные элементы из списка покупок.
func CheckoutList(listId primitive.ObjectID) error {
	// Формируем фильтр для поиска списка по ID.
	filter := bson.M{"_id": listId}

	// Формируем обновление для удаления завершенных элементов из списка.
	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{
				"isCompleted": true,
			},
		},
	}

	// Выполняем обновление в MongoDB.
	updatedResult, err := config.ShoppingLists.UpdateOne(context.TODO(), filter, update)
	fmt.Println("Update result", updatedResult.ModifiedCount)
	if err != nil {
		return err
	}

	return nil
}

// RemoveListItem удаляет элемент списка покупок.
func RemoveListItem(itemId string, userId primitive.ObjectID) error {
	// Преобразуем строковый идентификатор элемента в ObjectID.
	objId, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return err
	}

	// Формируем фильтр для поиска элемента по ID пользователя.
	filter := bson.M{
		"$or": bson.A{
			bson.M{"ownerId": userId},
			bson.M{"sharingIds": userId},
		},
	}

	// Формируем обновление для удаления элемента из списка.
	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{
				"_id": objId,
			},
		},
	}

	// Выполняем обновление в MongoDB.
	_, err = config.ShoppingLists.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

// RemoveList удаляет список покупок пользователя.
func RemoveList(listId string, ownerId primitive.ObjectID) (success bool, err error) {
	// Преобразуем строковый идентификатор списка в ObjectID.
	listObjId, err := primitive.ObjectIDFromHex(listId)
	if err != nil {
		return false, err
	}

	// Формируем запрос на удаление списка из MongoDB.
	dr, err := config.ShoppingLists.DeleteOne(context.TODO(), bson.M{
		"_id":     listObjId,
		"ownerId": ownerId,
	})

	if err != nil {
		return false, err
	} else if dr.DeletedCount == 0 {
		return false, nil
	}

	return true, nil
}

