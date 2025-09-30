package entities

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	urls := []string{"https://example.com/file1.jpg", "https://example.com/file2.pdf"}

	task := NewTask(urls)

	if task.ID.String() == "" {
		t.Error("Expected task ID to be generated")
	}

	if len(task.URLs) != len(urls) {
		t.Errorf("Expected %d URLs, got %d", len(urls), len(task.URLs))
	}

	if task.Status != TaskStatusNew {
		t.Errorf("Expected status %s, got %s", TaskStatusNew, task.Status)
	}

	if task.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if task.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	if len(task.Files) != len(urls) {
		t.Errorf("Expected %d files, got %d", len(urls), len(task.Files))
	}
}

func TestUpdateStatus(t *testing.T) {
	task := NewTask([]string{"https://example.com/file1.jpg"})
	originalTime := task.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	task.UpdateStatus(TaskStatusProcessing)

	if task.Status != TaskStatusProcessing {
		t.Errorf("Expected status %s, got %s", TaskStatusProcessing, task.Status)
	}

	if !task.UpdatedAt.After(originalTime) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestSetError(t *testing.T) {
	task := NewTask([]string{"https://example.com/file1.jpg"})
	errorMsg := "Download failed"

	task.SetError(errorMsg)

	if task.Error != errorMsg {
		t.Errorf("Expected error %s, got %s", errorMsg, task.Error)
	}

	if task.Status != TaskStatusFailed {
		t.Errorf("Expected status %s, got %s", TaskStatusFailed, task.Status)
	}
}

func TestIsCompleted(t *testing.T) {
	task := NewTask([]string{"https://example.com/file1.jpg", "https://example.com/file2.pdf"})

	// Initially not completed
	if task.IsCompleted() {
		t.Error("Expected task to not be completed initially")
	}

	// Mark all files as completed
	for i := range task.Files {
		task.Files[i].Status = "completed"
	}

	if !task.IsCompleted() {
		t.Error("Expected task to be completed")
	}

	// Mark one file as failed
	task.Files[0].Status = "failed"

	if task.IsCompleted() {
		t.Error("Expected task to not be completed when one file failed")
	}
}

func TestIsFailed(t *testing.T) {
	task := NewTask([]string{"https://example.com/file1.jpg", "https://example.com/file2.pdf"})

	// Initially not failed
	if task.IsFailed() {
		t.Error("Expected task to not be failed initially")
	}

	// Mark one file as failed
	task.Files[0].Status = "failed"

	if !task.IsFailed() {
		t.Error("Expected task to be failed")
	}
}

func TestGetProgress(t *testing.T) {
	task := NewTask([]string{"https://example.com/file1.jpg", "https://example.com/file2.pdf"})

	// Initially 0% progress
	if progress := task.GetProgress(); progress != 0 {
		t.Errorf("Expected 0%% progress, got %d%%", progress)
	}

	// Mark one file as completed - 50% progress
	task.Files[0].Status = "completed"
	if progress := task.GetProgress(); progress != 50 {
		t.Errorf("Expected 50%% progress, got %d%%", progress)
	}

	// Mark both files as completed - 100% progress
	task.Files[1].Status = "completed"
	if progress := task.GetProgress(); progress != 100 {
		t.Errorf("Expected 100%% progress, got %d%%", progress)
	}

	// Empty files should return 0%
	emptyTask := NewTask([]string{})
	if progress := emptyTask.GetProgress(); progress != 0 {
		t.Errorf("Expected 0%% progress for empty task, got %d%%", progress)
	}
}

