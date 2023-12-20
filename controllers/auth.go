package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AuthResource представляет ресурс для аутентификации.
type AuthResource struct{}

// Credentials содержит поля для имени пользователя и пароля.
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

var tokenAuth *jwtauth.JWTAuth

// Authenticator - промежуточное ПО для проверки аутентификации пользователя.
func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		// Проверка токена на валидность.
		if err != nil || token == nil || jwt.Validate(token) != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct {
				Error string `json:"error"`
			}{
				Error: "Unauthorized",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// init выполняется при запуске программы и инициализирует JWTAuth.
func init() {
	tokenAuth = jwtauth.New("HS256", []byte(os.Getenv("JWT_SIGN_KEY")), nil)
}

// Routes определяет маршруты для AuthResource.
func (rs AuthResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(Authenticator)

		r.Post("/logout", rs.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Post("/login", rs.Login)
		r.Post("/register", rs.Register)
	})

	return r
}

// HashPassword хеширует пароль с использованием bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// GenerateJWT генерирует JWT-токен для пользователя.
func GenerateJWT(userId string) (string, error) {
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"userId":  userId,
		"expires": time.Now().Add(time.Minute * 5).Unix(),
	})

	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Login обрабатывает запрос на вход пользователя.
func (rs AuthResource) Login(w http.ResponseWriter, r *http.Request) {
	var c Credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Поиск пользователя в базе данных по имени.
	sr := config.Users.FindOne(context.TODO(), bson.M{"name": c.Username})
	if sr.Err() != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var u models.User
	err := sr.Decode(&u)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Сравнение хеша пароля пользователя с введенным паролем.
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(c.Password))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Генерация и установка JWT-токена в куки.
	j, err := GenerateJWT(u.ID.Hex())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	cookie := &http.Cookie{
		Name:   "jwt",
		Value:  j,
		MaxAge: 60 * 60 * 24,
		Path:   "/",
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Username string `json:"username"`
		UserId   string `json:"userId"`
	}{u.Name, u.ID.Hex()})
}

// Logout обрабатывает запрос на выход пользователя.
func (rs AuthResource) Logout(w http.ResponseWriter, r *http.Request) {
	// Удаление JWT-токена из куки.
	c := &http.Cookie{
		Name:    "jwt",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	}

	http.SetCookie(w, c)
}

// Register обрабатывает запрос на регистрацию нового пользователя.
func (rs AuthResource) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var c Credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка, что имя пользователя уникально.
	sr := config.Users.FindOne(context.TODO(), bson.M{"name": c.Username})
	if sr.Err() == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(struct {
			Error string `json:"message"`
		}{"Username already in use"})
		return
	}

	// Хеширование пароля и создание нового пользователя.
	h, err := HashPassword(c.Password)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u := models.User{
		ID:       primitive.NewObjectID(),
		Name:     c.Username,
		Password: h,
	}

	err = models.AddUser(u)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

