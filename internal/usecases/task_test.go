package usecases

import (
	"context"
	"testing"

	"file-downloader/internal/entities"
)

// MockTaskRepository is a mock implementation of TaskRepository
type MockTaskRepository struct {
	tasks map[string]*entities.Task
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks: make(map[string]*entities.Task),
	}
}

func (m *MockTaskRepository) Create(ctx context.Context, task *entities.Task) error {
	m.tasks[task.ID.String()] = task
	return nil
}

func (m *MockTaskRepository) GetByID(ctx context.Context, id string) (*entities.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, &TaskNotFoundError{id}
	}
	return task, nil
}

func (m *MockTaskRepository) GetAll(ctx context.Context) ([]*entities.Task, error) {
	tasks := make([]*entities.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *MockTaskRepository) Update(ctx context.Context, task *entities.Task) error {
	if _, exists := m.tasks[task.ID.String()]; !exists {
		return &TaskNotFoundError{task.ID.String()}
	}
	m.tasks[task.ID.String()] = task
	return nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.tasks[id]; !exists {
		return &TaskNotFoundError{id}
	}
	delete(m.tasks, id)
	return nil
}

func (m *MockTaskRepository) GetPendingTasks(ctx context.Context) ([]*entities.Task, error) {
	var pendingTasks []*entities.Task
	for _, task := range m.tasks {
		if task.Status == entities.TaskStatusNew || task.Status == entities.TaskStatusProcessing {
			pendingTasks = append(pendingTasks, task)
		}
	}
	return pendingTasks, nil
}

func (m *MockTaskRepository) LoadTasks() error {
	return nil
}

func (m *MockTaskRepository) SaveTasks() error {
	return nil
}

type TaskNotFoundError struct {
	id string
}

func (e *TaskNotFoundError) Error() string {
	return "task with id " + e.id + " not found"
}

func TestCreateTask(t *testing.T) {
	// Setup
	mockRepo := NewMockTaskRepository()
	usecase := NewTaskUsecase(mockRepo, mockRepo)
	ctx := context.Background()

	// Test data
	urls := []string{
		"https://example.com/file1.jpg",
		"https://example.com/file2.pdf",
	}

	// Execute
	task, err := usecase.CreateTask(ctx, urls)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if task == nil {
		t.Fatal("Expected task to be created")
	}

	if len(task.URLs) != len(urls) {
		t.Errorf("Expected %d URLs, got %d", len(urls), len(task.URLs))
	}

	if task.Status != entities.TaskStatusNew {
		t.Errorf("Expected status %s, got %s", entities.TaskStatusNew, task.Status)
	}

	if len(task.Files) != len(urls) {
		t.Errorf("Expected %d files, got %d", len(urls), len(task.Files))
	}

	// Verify URLs match
	for i, url := range urls {
		if task.URLs[i] != url {
			t.Errorf("Expected URL %s, got %s", url, task.URLs[i])
		}
		if task.Files[i].URL != url {
			t.Errorf("Expected file URL %s, got %s", url, task.Files[i].URL)
		}
	}
}

func TestCreateTaskEmptyURLs(t *testing.T) {
	// Setup
	mockRepo := NewMockTaskRepository()
	usecase := NewTaskUsecase(mockRepo, mockRepo)
	ctx := context.Background()

	// Execute
	task, err := usecase.CreateTask(ctx, []string{})

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty URLs")
	}

	if task != nil {
		t.Fatal("Expected task to be nil")
	}
}

func TestCreateTaskEmptyURL(t *testing.T) {
	// Setup
	mockRepo := NewMockTaskRepository()
	usecase := NewTaskUsecase(mockRepo, mockRepo)
	ctx := context.Background()

	// Execute
	task, err := usecase.CreateTask(ctx, []string{""})

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty URL")
	}

	if task != nil {
		t.Fatal("Expected task to be nil")
	}
}

func TestGetTask(t *testing.T) {
	// Setup
	mockRepo := NewMockTaskRepository()
	usecase := NewTaskUsecase(mockRepo, mockRepo)
	ctx := context.Background()

	// Create a task first
	urls := []string{"https://example.com/file1.jpg"}
	task, err := usecase.CreateTask(ctx, urls)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Execute
	retrievedTask, err := usecase.GetTask(ctx, task.ID.String())

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrievedTask == nil {
		t.Fatal("Expected task to be retrieved")
	}

	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	// Setup
	mockRepo := NewMockTaskRepository()
	usecase := NewTaskUsecase(mockRepo, mockRepo)
	ctx := context.Background()

	// Execute
	task, err := usecase.GetTask(ctx, "non-existent-id")

	// Assert
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}

	if task != nil {
		t.Fatal("Expected task to be nil")
	}
}

func TestGetAllTasks(t *testing.T) {
	// Setup
	mockRepo := NewMockTaskRepository()
	usecase := NewTaskUsecase(mockRepo, mockRepo)
	ctx := context.Background()

	// Create multiple tasks
	urls1 := []string{"https://example.com/file1.jpg"}
	urls2 := []string{"https://example.com/file2.pdf"}

	task1, err := usecase.CreateTask(ctx, urls1)
	if err != nil {
		t.Fatalf("Failed to create task 1: %v", err)
	}

	task2, err := usecase.CreateTask(ctx, urls2)
	if err != nil {
		t.Fatalf("Failed to create task 2: %v", err)
	}

	// Execute
	tasks, err := usecase.GetAllTasks(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Verify both tasks are present
	found1, found2 := false, false
	for _, task := range tasks {
		if task.ID == task1.ID {
			found1 = true
		}
		if task.ID == task2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Task 1 not found in results")
	}
	if !found2 {
		t.Error("Task 2 not found in results")
	}
}

func TestGetTaskStatus(t *testing.T) {
	// Setup
	mockRepo := NewMockTaskRepository()
	usecase := NewTaskUsecase(mockRepo, mockRepo)
	ctx := context.Background()

	// Create a task
	urls := []string{"https://example.com/file1.jpg"}
	task, err := usecase.CreateTask(ctx, urls)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Execute
	statusTask, err := usecase.GetTaskStatus(ctx, task.ID.String())

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if statusTask == nil {
		t.Fatal("Expected task to be retrieved")
	}

	if statusTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, statusTask.ID)
	}

	if statusTask.Status != entities.TaskStatusNew {
		t.Errorf("Expected status %s, got %s", entities.TaskStatusNew, statusTask.Status)
	}
}
