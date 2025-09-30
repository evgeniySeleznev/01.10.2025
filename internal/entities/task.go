package entities

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus представляет статус задачи
type TaskStatus string

const (
	TaskStatusNew        TaskStatus = "new"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// Task представляет задачу скачивания
type Task struct {
	ID        uuid.UUID  `json:"id"`
	URLs      []string   `json:"urls"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Files     []File     `json:"files"`
	Error     string     `json:"error,omitempty"`
}

// File представляет файл в рамках задачи
type File struct {
	URL    string `json:"url"`
	Path   string `json:"path,omitempty"`
	Size   int64  `json:"size,omitempty"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// NewTask создает новую задачу с указанными URL
func NewTask(urls []string) *Task {
	return &Task{
		ID:        uuid.New(),
		URLs:      urls,
		Status:    TaskStatusNew,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Files:     make([]File, len(urls)),
	}
}

// UpdateStatus обновляет статус задачи и временную метку
func (t *Task) UpdateStatus(status TaskStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
}

// SetError устанавливает сообщение об ошибке и обновляет статус на failed
func (t *Task) SetError(err string) {
	t.Error = err
	t.Status = TaskStatusFailed
	t.UpdatedAt = time.Now()
}

// IsCompleted возвращает true, если все файлы успешно скачаны
func (t *Task) IsCompleted() bool {
	if len(t.Files) == 0 {
		return false
	}

	for _, file := range t.Files {
		if file.Status != "completed" {
			return false
		}
	}
	return true
}

// IsFailed возвращает true, если любой файл не удалось скачать
func (t *Task) IsFailed() bool {
	for _, file := range t.Files {
		if file.Status == "failed" {
			return true
		}
	}
	return false
}

// GetProgress возвращает процент выполнения
func (t *Task) GetProgress() int {
	if len(t.Files) == 0 {
		return 0
	}

	completed := 0
	for _, file := range t.Files {
		if file.Status == "completed" {
			completed++
		}
	}

	return (completed * 100) / len(t.Files)
}
