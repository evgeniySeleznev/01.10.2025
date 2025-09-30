package usecases

import (
	"context"
	"fmt"

	"file-downloader/internal/entities"
	"file-downloader/internal/interfaces"
)

// TaskUsecase реализует use case'ы управления задачами
type TaskUsecase struct {
	taskRepo       interfaces.TaskRepository
	persistentRepo interfaces.PersistentRepository
}

// NewTaskUsecase создает новый use case для задач
func NewTaskUsecase(taskRepo interfaces.TaskRepository, persistentRepo interfaces.PersistentRepository) interfaces.TaskUsecase {
	return &TaskUsecase{
		taskRepo:       taskRepo,
		persistentRepo: persistentRepo,
	}
}

// CreateTask создает новую задачу скачивания
func (u *TaskUsecase) CreateTask(ctx context.Context, urls []string) (*entities.Task, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("не предоставлены URL")
	}

	// Валидация URL
	for _, url := range urls {
		if url == "" {
			return nil, fmt.Errorf("предоставлен пустой URL")
		}
	}

	// Создание новой задачи
	task := entities.NewTask(urls)

	// Инициализация файлов с URL
	for i, url := range urls {
		task.Files[i] = entities.File{
			URL:    url,
			Status: "pending",
		}
	}

	// Сохранение в репозитории
	if err := u.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("не удалось создать задачу: %w", err)
	}

	if err := u.persistentRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("не удалось сохранить задачу: %w", err)
	}

	return task, nil
}

// GetTask получает задачу по ID
func (u *TaskUsecase) GetTask(ctx context.Context, id string) (*entities.Task, error) {
	task, err := u.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачу: %w", err)
	}

	return task, nil
}

// GetAllTasks получает все задачи
func (u *TaskUsecase) GetAllTasks(ctx context.Context) ([]*entities.Task, error) {
	tasks, err := u.taskRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить задачи: %w", err)
	}

	return tasks, nil
}

// GetTaskStatus получает статус задачи по ID
func (u *TaskUsecase) GetTaskStatus(ctx context.Context, id string) (*entities.Task, error) {
	task, err := u.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить статус задачи: %w", err)
	}

	return task, nil
}
