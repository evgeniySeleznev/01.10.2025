package usecases

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"file-downloader/internal/entities"
	"file-downloader/internal/interfaces"
)

// DownloadUsecase реализует use case'ы скачивания файлов
type DownloadUsecase struct {
	taskRepo       interfaces.TaskRepository
	persistentRepo interfaces.PersistentRepository
	downloadDir    string
}

// NewDownloadUsecase создает новый use case для скачивания
func NewDownloadUsecase(taskRepo interfaces.TaskRepository, persistentRepo interfaces.PersistentRepository) interfaces.DownloadUsecase {
	return &DownloadUsecase{
		taskRepo:       taskRepo,
		persistentRepo: persistentRepo,
		downloadDir:    "./downloads",
	}
}

// ProcessTask обрабатывает задачу, скачивая все её файлы
func (u *DownloadUsecase) ProcessTask(ctx context.Context, task *entities.Task) error {
	// Обновление статуса задачи на processing
	task.UpdateStatus(entities.TaskStatusProcessing)
	if err := u.updateTask(task); err != nil {
		return fmt.Errorf("не удалось обновить статус задачи: %w", err)
	}

	// Создание директории для скачивания этой задачи
	taskDir := filepath.Join(u.downloadDir, task.ID.String())
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		task.SetError(fmt.Sprintf("не удалось создать директорию для скачивания: %v", err))
		u.updateTask(task)
		return fmt.Errorf("не удалось создать директорию для скачивания: %w", err)
	}

	// Скачивание каждого файла
	for i := range task.Files {
		if err := u.DownloadFile(ctx, task.Files[i].URL, task.ID.String(), i); err != nil {
			task.Files[i].Status = "failed"
			task.Files[i].Error = err.Error()
		}

		// Обновление задачи после каждого файла
		if err := u.updateTask(task); err != nil {
			return fmt.Errorf("не удалось обновить задачу: %w", err)
		}
	}

	// Проверка финального статуса
	if task.IsCompleted() {
		task.UpdateStatus(entities.TaskStatusCompleted)
	} else if task.IsFailed() {
		task.UpdateStatus(entities.TaskStatusFailed)
	}

	return u.updateTask(task)
}

// DownloadFile скачивает один файл
func (u *DownloadUsecase) DownloadFile(ctx context.Context, url string, taskID string, fileIndex int) error {
	// Получение задачи
	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("не удалось получить задачу: %w", err)
	}

	if fileIndex >= len(task.Files) {
		return fmt.Errorf("неверный индекс файла: %d", fileIndex)
	}

	file := &task.Files[fileIndex]
	file.Status = "downloading"

	// Создание HTTP клиента с таймаутом
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Выполнение запроса
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		file.Status = "failed"
		file.Error = fmt.Sprintf("не удалось создать запрос: %v", err)
		return err
	}

	// Получение информации о файле
	resp, err := client.Do(req)
	if err != nil {
		file.Status = "failed"
		file.Error = fmt.Sprintf("не удалось скачать: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		file.Status = "failed"
		file.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Получение имени файла из URL или заголовка Content-Disposition
	fileName := u.getFileName(url, resp.Header.Get("Content-Disposition"))
	filePath := filepath.Join(u.downloadDir, taskID, fileName)
	file.Path = filePath

	// Создание файла
	destFile, err := os.Create(filePath)
	if err != nil {
		file.Status = "failed"
		file.Error = fmt.Sprintf("не удалось создать файл: %v", err)
		return err
	}
	defer destFile.Close()

	// Копирование данных
	written, err := io.Copy(destFile, resp.Body)
	if err != nil {
		file.Status = "failed"
		file.Error = fmt.Sprintf("не удалось записать файл: %v", err)
		return err
	}

	// Обновление информации о файле
	file.Size = written
	file.Status = "completed"

	return nil
}

// GetPendingTasks получает все ожидающие задачи
func (u *DownloadUsecase) GetPendingTasks(ctx context.Context) ([]*entities.Task, error) {
	return u.taskRepo.GetPendingTasks(ctx)
}

// updateTask обновляет задачу в обоих репозиториях
func (u *DownloadUsecase) updateTask(task *entities.Task) error {
	if err := u.taskRepo.Update(context.Background(), task); err != nil {
		return err
	}
	return u.persistentRepo.Update(context.Background(), task)
}

// getFileName извлекает имя файла из URL или заголовка Content-Disposition
func (u *DownloadUsecase) getFileName(url, contentDisposition string) string {
	// Попытка получить имя файла из заголовка Content-Disposition
	if contentDisposition != "" {
		parts := strings.Split(contentDisposition, "filename=")
		if len(parts) > 1 {
			filename := strings.Trim(parts[1], `"`)
			if filename != "" {
				return filename
			}
		}
	}

	// Извлечение имени файла из URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if filename != "" && !strings.Contains(filename, "?") {
			return filename
		}
	}

	// Генерация имени файла по умолчанию
	return fmt.Sprintf("file_%d", time.Now().Unix())
}
