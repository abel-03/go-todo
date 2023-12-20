package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewListItemReq содержит данные для создания нового элемента списка.
type NewListItemReq struct {
	Name        string `json:"name"`
	ListId      string `json:"listId"`
	IsCompleted bool   `json:"isCompleted"`
}

// ShoppingListReq представляет структуру для запроса списка покупок.
type ShoppingListReq struct {
	ID    string        `json:"id"`
	Name  string        `json:"name" bson:"name"`
	Items []ListItemReq `json:"items"`
}

// ListItemReq представляет структуру для запроса элемента списка.
type ListItemReq struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsCompleted bool   `json:"isCompleted"`
}

// ShoppingListsResource представляет ресурс для управления списками покупок.
type ShoppingListsResource struct{}

// Routes определяет маршруты для ShoppingListsResource.
func (rs ShoppingListsResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Use(Authenticator)

	r.Post("/", rs.CreateList)
	r.Get("/", rs.GetLists)
	r.Delete("/{id}", rs.DeleteList)
	r.Post("/checkout/{id}", rs.CheckoutList)

	r.Route("/bulk", func(r chi.Router) {
		r.Post("/", rs.AddLists)
	})

	r.Route("/items", func(r chi.Router) {
		r.Post("/", rs.CreateListItem)
		r.Delete("/{id}", rs.DeleteListItem)
		r.Put("/{id}", rs.UpdateListItem)
	})

	return r
}

// GetLists возвращает списки покупок пользователя.
func (rs ShoppingListsResource) GetLists(w http.ResponseWriter, r *http.Request) {
	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	objId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Получение списков покупок для пользователя из базы данных.
	items, err := models.AllShoppingLists(objId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// CreateList создает новый список покупок для пользователя.
func (rs ShoppingListsResource) CreateList(w http.ResponseWriter, r *http.Request) {
	l := struct{ Name string }{}
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Добавление нового списка покупок в базу данных.
	id, err := models.AddNewShoppingList(l.Name, ownerId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}{l.Name, id})
}

// DeleteList удаляет список покупок пользователя.
func (rs ShoppingListsResource) DeleteList(w http.ResponseWriter, r *http.Request) {
	listId := chi.URLParam(r, "id")

	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Удаление списка покупок из базы данных.
	success, err := models.RemoveList(listId, ownerId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if !success {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddLists добавляет несколько списков покупок для пользователя.
func (rs ShoppingListsResource) AddLists(w http.ResponseWriter, r *http.Request) {
	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Декодирование запроса с несколькими списками покупок.
	var lists []ShoppingListReq
	if err := json.NewDecoder(r.Body).Decode(&lists); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Преобразование структур запроса в модели данных и добавление списков в базу данных.
	var dbLists []models.ShoppingList
	for _, l := range lists {
		newList := models.ShoppingList{
			ID:               primitive.NewObjectID(),
			OwnerId:          ownerId,
			Name:             l.Name,
			Items:            make([]models.ListItem, 0),
			SharingIds:       make([]primitive.ObjectID, 0),
			SharingInviteIds: make([]primitive.ObjectID, 0),
			SharingNames:     make([]string, 0),
		}
		for _, item := range l.Items {
			newItem := models.ListItem{
				ID:          primitive.NewObjectID(),
				Name:        item.Name,
				IsCompleted: item.IsCompleted,
			}
			newList.Items = append(newList.Items, newItem)
		}
		dbLists = append(dbLists, newList)
	}

	// Добавление списков покупок в базу данных.
	err = models.AddShoppingLists(dbLists, ownerId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// CheckoutList отмечает список покупок как завершенный.
func (rs ShoppingListsResource) CheckoutList(w http.ResponseWriter, r *http.Request) {
	listId := chi.URLParam(r, "id")
	listObjId, err := primitive.ObjectIDFromHex(listId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Отмечение списка покупок как завершенного в базе данных.
	err = models.CheckoutList(listObjId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// CreateListItem создает новый элемент списка покупок.
func (rs ShoppingListsResource) CreateListItem(w http.ResponseWriter, r *http.Request) {
	var itemData NewListItemReq
	if err := json.NewDecoder(r.Body).Decode(&itemData); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Преобразование строкового идентификатора в ObjectID.
	listId, err := primitive.ObjectIDFromHex(itemData.ListId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Добавление нового элемента в список покупок в базе данных.
	err = models.AddListItem(itemData.Name, ownerId, listId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// DeleteListItem удаляет элемент списка покупок.
func (rs ShoppingListsResource) DeleteListItem(w http.ResponseWriter, r *http.Request) {
	itemId := chi.URLParam(r, "id")

	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Удаление элемента списка покупок из базы данных.
	err = models.RemoveListItem(itemId, ownerId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateListItem обновляет информацию об элементе списка покупок.
func (rs ShoppingListsResource) UpdateListItem(w http.ResponseWriter, r *http.Request) {
	listItemId := chi.URLParam(r, "id")

	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Декодирование запроса с новыми данными об элементе списка.
	var itemData NewListItemReq
	if err := json.NewDecoder(r.Body).Decode(&itemData); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Преобразование строкового идентификатора в ObjectID.
	liId, err := primitive.ObjectIDFromHex(listItemId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Создание структуры для обновленных данных об элементе списка.
	li := models.ListItem{
		ID:          liId,
		Name:        itemData.Name,
		IsCompleted: itemData.IsCompleted,
	}

	// Обновление данных об элементе списка в базе данных.
	err = models.ModifyListItem(ownerId, li)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

