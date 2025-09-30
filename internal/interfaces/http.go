package interfaces

import (
	"net/http"
)

// HTTPHandler определяет интерфейс для HTTP обработчиков
type HTTPHandler interface {
	CreateTask(w http.ResponseWriter, r *http.Request)
	GetTask(w http.ResponseWriter, r *http.Request)
	GetAllTasks(w http.ResponseWriter, r *http.Request)
	GetTaskStatus(w http.ResponseWriter, r *http.Request)
}
