package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpHandlers "file-downloader/internal/adapters/http"
	"file-downloader/internal/adapters/repository"
	"file-downloader/internal/entities"
	"file-downloader/internal/infrastructure"
	"file-downloader/internal/interfaces"
	"file-downloader/internal/usecases"
)

// syncRepositories синхронизирует данные между in-memory и file-based репозиториями
func syncRepositories(taskRepo interfaces.TaskRepository, fileRepo interfaces.PersistentRepository) error {
	// Получаем все задачи из file-based репозитория
	tasks, err := fileRepo.GetAll(context.Background())
	if err != nil {
		return err
	}

	// Добавляем их в in-memory репозиторий
	for _, task := range tasks {
		if err := taskRepo.Create(context.Background(), task); err != nil {
			log.Printf("Предупреждение: не удалось добавить задачу %s в in-memory репозиторий: %v", task.ID.String(), err)
		}
	}

	log.Printf("Синхронизировано %d задач между репозиториями", len(tasks))
	return nil
}

func main() {
	// Инициализация зависимостей
	taskRepo := repository.NewInMemoryTaskRepository()
	fileRepo := repository.NewFileBasedTaskRepository("./data/tasks.json")

	// Загрузка существующих задач из файла
	if err := fileRepo.LoadTasks(); err != nil {
		log.Printf("Предупреждение: не удалось загрузить задачи из файла: %v", err)
	}

	// Синхронизация данных между репозиториями
	if err := syncRepositories(taskRepo, fileRepo); err != nil {
		log.Printf("Предупреждение: не удалось синхронизировать репозитории: %v", err)
	}

	// Инициализация use case'ов
	taskUsecase := usecases.NewTaskUsecase(taskRepo, fileRepo)
	downloadUsecase := usecases.NewDownloadUsecase(taskRepo, fileRepo)

	// Инициализация HTTP обработчиков
	taskHandler := httpHandlers.NewTaskHandler(taskUsecase, downloadUsecase)

	// Инициализация сервера
	server := &http.Server{
		Addr:    ":8080",
		Handler: httpHandlers.SetupRoutes(taskHandler),
	}

	// Инициализация пула воркеров для скачивания
	workerPool := infrastructure.NewWorkerPool(3, downloadUsecase) // 3 параллельных скачивания
	workerPool.Start()

	// Настройка graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск процессора задач для обработки новых задач
	go func() {
		log.Println("Процессор задач запущен")
		for {
			select {
			case <-ctx.Done():
				log.Println("Процессор задач остановлен")
				return
			default:
				// Получение ожидающих задач и добавление их в пул воркеров
				pendingTasks, err := downloadUsecase.GetPendingTasks(ctx)
				if err != nil {
					log.Printf("Ошибка получения ожидающих задач: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				log.Printf("Найдено %d ожидающих задач", len(pendingTasks))
				for _, task := range pendingTasks {
					log.Printf("Задача %s имеет статус: %s", task.ID.String(), task.Status)
					if task.Status == entities.TaskStatusNew {
						log.Printf("Добавляем задачу %s в пул воркеров", task.ID.String())
						if err := workerPool.AddTask(task.ID.String()); err != nil {
							log.Printf("Ошибка добавления задачи в пул воркеров: %v", err)
						} else {
							log.Printf("Задача %s успешно добавлена в пул воркеров", task.ID.String())
						}
					}
				}

				time.Sleep(2 * time.Second)
			}
		}
	}()

	// Запуск сервера в горутине
	go func() {
		log.Println("Запуск сервера на :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Не удалось запустить сервер: %v", err)
		}
	}()

	// Обработка graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Остановка сервера...")

	// Отмена контекста для прекращения приема новых задач
	cancel()

	// Graceful остановка пула воркеров
	workerPool.Stop()

	// Сохранение текущего состояния в файл
	if err := fileRepo.SaveTasks(); err != nil {
		log.Printf("Ошибка сохранения задач: %v", err)
	}

	// Остановка HTTP сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Принудительная остановка сервера: %v", err)
	}

	log.Println("Сервер остановлен")
}
