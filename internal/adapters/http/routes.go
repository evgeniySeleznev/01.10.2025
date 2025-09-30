package http

import (
	"net/http"
	"strings"

	"file-downloader/internal/interfaces"
)

// SetupRoutes настраивает HTTP маршруты
func SetupRoutes(handler interfaces.HTTPHandler) http.Handler {
	mux := http.NewServeMux()

	// Маршруты задач
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateTask(w, r)
		case http.MethodGet:
			handler.GetAllTasks(w, r)
		default:
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		}
	})

	// Маршрут для конкретных задач и их статуса
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}

		// Проверяем, является ли это запросом статуса
		if strings.HasSuffix(r.URL.Path, "/status") {
			handler.GetTaskStatus(w, r)
			return
		}

		// Иначе это запрос конкретной задачи
		handler.GetTask(w, r)
	})

	// Проверка здоровья
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}
