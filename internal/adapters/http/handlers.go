package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"file-downloader/internal/interfaces"
)

// TaskHandler обрабатывает HTTP запросы для операций с задачами
type TaskHandler struct {
	taskUsecase     interfaces.TaskUsecase
	downloadUsecase interfaces.DownloadUsecase
}

// NewTaskHandler создает новый обработчик задач
func NewTaskHandler(taskUsecase interfaces.TaskUsecase, downloadUsecase interfaces.DownloadUsecase) interfaces.HTTPHandler {
	return &TaskHandler{
		taskUsecase:     taskUsecase,
		downloadUsecase: downloadUsecase,
	}
}

// CreateTaskRequest представляет тело запроса для создания задачи
type CreateTaskRequest struct {
	URLs []string `json:"urls"`
}

// CreateTask обрабатывает POST /tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "URL обязательны", http.StatusBadRequest)
		return
	}

	task, err := h.taskUsecase.CreateTask(r.Context(), req.URLs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Не удалось создать задачу: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTask обрабатывает GET /tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	id := h.extractTaskID(r.URL.Path)
	if id == "" {
		http.Error(w, "ID задачи обязателен", http.StatusBadRequest)
		return
	}

	task, err := h.taskUsecase.GetTask(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Задача не найдена", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Не удалось получить задачу: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// GetAllTasks обрабатывает GET /tasks
func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := h.taskUsecase.GetAllTasks(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Не удалось получить задачи: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTaskStatus обрабатывает GET /tasks/{id}/status
func (h *TaskHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	id := h.extractTaskID(r.URL.Path)
	if id == "" {
		http.Error(w, "ID задачи обязателен", http.StatusBadRequest)
		return
	}

	task, err := h.taskUsecase.GetTaskStatus(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Задача не найдена", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Не удалось получить статус задачи: %v", err), http.StatusInternalServerError)
		return
	}

	// Возврат только информации о статусе
	statusResponse := map[string]interface{}{
		"id":         task.ID,
		"status":     task.Status,
		"progress":   task.GetProgress(),
		"created_at": task.CreatedAt,
		"updated_at": task.UpdatedAt,
		"files":      task.Files,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusResponse)
}

// extractTaskID извлекает ID задачи из пути URL
func (h *TaskHandler) extractTaskID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "tasks" {
		return parts[1]
	}
	return ""
}
