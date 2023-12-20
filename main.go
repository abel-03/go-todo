package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed static-ui
var staticFS embed.FS

func main() {
	// Создаем новый роутер Chi.
	r := chi.NewRouter()

	// Используем промежуточные обработчики для общих операций, таких как логирование и восстановление после паники.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Монтируем статические файлы из встроенного файла системы.
	FileServer(r, "/", getFileSystem(staticFS))

	// Монтируем роутеры для API функционала.
	r.Mount("/api/auth", controllers.AuthResource{}.Routes())
	r.Mount("/api/lists", controllers.ShoppingListsResource{}.Routes())
	r.Mount("/api/share-lists", controllers.ShareListsResource{}.Routes())

	// Получаем порт из переменной окружения.
	port := os.Getenv("PORT")

	// Запускаем веб-сервер на указанном порту.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

// FileServer создает обработчик файлов для заданного пути и корневой файловой системы.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	// Проверяем, содержит ли путь символы, которые могут представлять URL-параметры.
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	// Если путь не "/", а последний символ не "/", то редиректим запрос на версию с "/"
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	// Обработчик, который удаляет префикс пути и обслуживает файлы из корневой файловой системы.
	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

// getFileSystem возвращает файловую систему для встроенных статических файлов.
func getFileSystem(embedFS embed.FS) http.FileSystem {
	// Получаем поддерево файловой системы для папки "static-ui".
	fsys, err := fs.Sub(embedFS, "static-ui")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}

