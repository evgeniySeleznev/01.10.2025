package repository

import (
	"context"
	"fmt"
	"sync"

	"file-downloader/internal/entities"
	"file-downloader/internal/interfaces"
)

// InMemoryTaskRepository реализует TaskRepository используя хранилище в памяти
type InMemoryTaskRepository struct {
	tasks map[string]*entities.Task
	mutex sync.RWMutex
}

// NewInMemoryTaskRepository создает новый репозиторий задач в памяти
func NewInMemoryTaskRepository() interfaces.TaskRepository {
	return &InMemoryTaskRepository{
		tasks: make(map[string]*entities.Task),
	}
}

// Create добавляет новую задачу в репозиторий
func (r *InMemoryTaskRepository) Create(ctx context.Context, task *entities.Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.tasks[task.ID.String()] = task
	return nil
}

// GetByID получает задачу по её ID
func (r *InMemoryTaskRepository) GetByID(ctx context.Context, id string) (*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("задача с id %s не найдена", id)
	}

	return task, nil
}

// GetAll получает все задачи
func (r *InMemoryTaskRepository) GetAll(ctx context.Context) ([]*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasks := make([]*entities.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Update обновляет существующую задачу
func (r *InMemoryTaskRepository) Update(ctx context.Context, task *entities.Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[task.ID.String()]; !exists {
		return fmt.Errorf("задача с id %s не найдена", task.ID.String())
	}

	r.tasks[task.ID.String()] = task
	return nil
}

// Delete удаляет задачу по её ID
func (r *InMemoryTaskRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return fmt.Errorf("задача с id %s не найдена", id)
	}

	delete(r.tasks, id)
	return nil
}

// GetPendingTasks получает все задачи со статусом "new" или "processing"
func (r *InMemoryTaskRepository) GetPendingTasks(ctx context.Context) ([]*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var pendingTasks []*entities.Task
	for _, task := range r.tasks {
		if task.Status == entities.TaskStatusNew || task.Status == entities.TaskStatusProcessing {
			pendingTasks = append(pendingTasks, task)
		}
	}

	return pendingTasks, nil
}
