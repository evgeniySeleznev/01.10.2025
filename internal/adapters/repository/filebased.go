package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"file-downloader/internal/entities"
	"file-downloader/internal/interfaces"
)

// FileBasedTaskRepository реализует PersistentRepository используя файловое хранилище
type FileBasedTaskRepository struct {
	filePath string
	tasks    map[string]*entities.Task
	mutex    sync.RWMutex
}

// NewFileBasedTaskRepository создает новый репозиторий задач на основе файлов
func NewFileBasedTaskRepository(filePath string) interfaces.PersistentRepository {
	return &FileBasedTaskRepository{
		filePath: filePath,
		tasks:    make(map[string]*entities.Task),
	}
}

// LoadTasks загружает задачи из файла
func (r *FileBasedTaskRepository) LoadTasks() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Создание директории, если она не существует
	if err := os.MkdirAll(filepath.Dir(r.filePath), 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	// Проверка существования файла
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		// Файл не существует, создаем пустую карту
		r.tasks = make(map[string]*entities.Task)
		return nil
	}

	// Чтение файла
	data, err := ioutil.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	// Парсинг JSON
	var tasks map[string]*entities.Task
	if len(data) > 0 {
		if err := json.Unmarshal(data, &tasks); err != nil {
			return fmt.Errorf("не удалось распарсить JSON: %w", err)
		}
	} else {
		tasks = make(map[string]*entities.Task)
	}

	r.tasks = tasks
	return nil
}

// SaveTasks сохраняет задачи в файл
func (r *FileBasedTaskRepository) SaveTasks() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Создание директории, если она не существует
	if err := os.MkdirAll(filepath.Dir(r.filePath), 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	// Маршалинг в JSON
	data, err := json.MarshalIndent(r.tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось маршалить JSON: %w", err)
	}

	// Запись в файл
	if err := ioutil.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл: %w", err)
	}

	return nil
}

// Create добавляет новую задачу в репозиторий
func (r *FileBasedTaskRepository) Create(ctx context.Context, task *entities.Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.tasks[task.ID.String()] = task
	return r.saveTasksUnsafe()
}

// GetByID получает задачу по её ID
func (r *FileBasedTaskRepository) GetByID(ctx context.Context, id string) (*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("задача с id %s не найдена", id)
	}

	return task, nil
}

// GetAll получает все задачи
func (r *FileBasedTaskRepository) GetAll(ctx context.Context) ([]*entities.Task, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tasks := make([]*entities.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// Update обновляет существующую задачу
func (r *FileBasedTaskRepository) Update(ctx context.Context, task *entities.Task) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[task.ID.String()]; !exists {
		return fmt.Errorf("задача с id %s не найдена", task.ID.String())
	}

	r.tasks[task.ID.String()] = task
	return r.saveTasksUnsafe()
}

// Delete удаляет задачу по её ID
func (r *FileBasedTaskRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return fmt.Errorf("задача с id %s не найдена", id)
	}

	delete(r.tasks, id)
	return r.saveTasksUnsafe()
}

// GetPendingTasks получает все задачи со статусом "new" или "processing"
func (r *FileBasedTaskRepository) GetPendingTasks(ctx context.Context) ([]*entities.Task, error) {
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

// saveTasksUnsafe сохраняет задачи без получения блокировки (вызывающий должен держать блокировку)
func (r *FileBasedTaskRepository) saveTasksUnsafe() error {
	// Маршалинг в JSON
	data, err := json.MarshalIndent(r.tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось маршалить JSON: %w", err)
	}

	// Запись в файл
	if err := ioutil.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл: %w", err)
	}

	return nil
}
