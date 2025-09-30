package interfaces

import (
	"context"

	"file-downloader/internal/entities"
)

// TaskUsecase определяет интерфейс для операций управления задачами
type TaskUsecase interface {
	CreateTask(ctx context.Context, urls []string) (*entities.Task, error)
	GetTask(ctx context.Context, id string) (*entities.Task, error)
	GetAllTasks(ctx context.Context) ([]*entities.Task, error)
	GetTaskStatus(ctx context.Context, id string) (*entities.Task, error)
}

// DownloadUsecase определяет интерфейс для операций скачивания файлов
type DownloadUsecase interface {
	ProcessTask(ctx context.Context, task *entities.Task) error
	DownloadFile(ctx context.Context, url string, taskID string, fileIndex int) error
	GetPendingTasks(ctx context.Context) ([]*entities.Task, error)
}
