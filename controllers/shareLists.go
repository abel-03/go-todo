package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShareListsResource представляет ресурс для управления запросами на обмен списками.
type ShareListsResource struct{}

// ShareListReq содержит поля для запроса обмена списками.
type ShareListReq struct {
	ListId   string `json:"listId"`
	UserName string `json:"userName"`
}

// HandleShareReq содержит поля для обработки запроса на обмен списками.
type HandleShareReq struct {
	ListId      string `json:"listId"`
	IsAccepting bool   `json:"isAccepting"`
}

// Routes определяет маршруты для ShareListsResource.
func (rs ShareListsResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Use(Authenticator)

	r.Get("/", rs.GetShareInviteLists)
	r.Post("/create", rs.CreateShareRequest)
	r.Post("/respond", rs.RespondToShareRequest)

	return r
}

// GetShareInviteLists возвращает списки приглашений на обмен пользователя.
func (rs ShareListsResource) GetShareInviteLists(w http.ResponseWriter, r *http.Request) {
	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Получение списков приглашений на обмен для пользователя.
	items, err := models.AllShareInviteShoppingLists(ownerId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Отправка списков в формате JSON.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// CreateShareRequest обрабатывает запрос на создание запроса на обмен списками.
func (rs ShareListsResource) CreateShareRequest(w http.ResponseWriter, r *http.Request) {
	var s ShareListReq
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	ownerId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Преобразование строковых идентификаторов в ObjectID.
	listId, err := primitive.ObjectIDFromHex(s.ListId)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Получение идентификатора пользователя по имени.
	userId, err := models.GetUserIdByName(s.UserName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Проверка на самообмен и добавление запроса на обмен в базу данных.
	if userId == ownerId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	success, err := models.AddShareInvite(ownerId, listId, userId)
	if !success || err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RespondToShareRequest обрабатывает запрос на обработку запроса на обмен списками.
func (rs ShareListsResource) RespondToShareRequest(w http.ResponseWriter, r *http.Request) {
	var h HandleShareReq
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Извлечение идентификатора пользователя из токена.
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Преобразование строковых идентификаторов в ObjectID.
	listId, err := primitive.ObjectIDFromHex(h.ListId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var success bool
	// Обработка запроса на обмен (принятие или отклонение).
	if h.IsAccepting {
		success, err = models.ShareListWithUser(listId, userId)
	} else {
		success, err = models.DeclineShareListWithUser(listId, userId)
	}

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Возврат статуса в зависимости от результата операции.
	if success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

