package interfaces

import (
	"context"

	"file-downloader/internal/entities"
)

// TaskRepository определяет интерфейс для операций хранения задач
type TaskRepository interface {
	Create(ctx context.Context, task *entities.Task) error
	GetByID(ctx context.Context, id string) (*entities.Task, error)
	GetAll(ctx context.Context) ([]*entities.Task, error)
	Update(ctx context.Context, task *entities.Task) error
	Delete(ctx context.Context, id string) error
	GetPendingTasks(ctx context.Context) ([]*entities.Task, error)
}

// PersistentRepository определяет интерфейс для постоянного хранилища
type PersistentRepository interface {
	TaskRepository
	LoadTasks() error
	SaveTasks() error
}
