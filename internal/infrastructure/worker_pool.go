package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"file-downloader/internal/interfaces"
)

// WorkerPool управляет параллельными скачиваниями файлов
type WorkerPool struct {
	workerCount     int
	downloadUsecase interfaces.DownloadUsecase
	taskQueue       chan *TaskJob
	workers         []*Worker
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	running         bool
}

// TaskJob представляет задачу для пула воркеров
type TaskJob struct {
	TaskID string
}

// Worker представляет одного воркера в пуле
type Worker struct {
	id       int
	pool     *WorkerPool
	jobQueue chan *TaskJob
	quit     chan bool
	busy     bool
	mu       sync.Mutex
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(workerCount int, downloadUsecase interfaces.DownloadUsecase) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workerCount:     workerCount,
		downloadUsecase: downloadUsecase,
		taskQueue:       make(chan *TaskJob, 100), // Буфер для 100 задач
		ctx:             ctx,
		cancel:          cancel,
		running:         false,
	}
}

// Start запускает пул воркеров
func (wp *WorkerPool) Start() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		return
	}

	wp.running = true

	// Создание воркеров
	wp.workers = make([]*Worker, wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		worker := &Worker{
			id:       i,
			pool:     wp,
			jobQueue: make(chan *TaskJob, 1),
			quit:     make(chan bool),
		}
		wp.workers[i] = worker

		wp.wg.Add(1)
		go worker.start()
	}

	// Запуск диспетчера задач
	go wp.dispatchTasks()

	log.Printf("Пул воркеров запущен с %d воркерами", wp.workerCount)
}

// Stop останавливает пул воркеров gracefully
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.running {
		return
	}

	log.Println("Остановка пула воркеров...")

	// Отмена контекста для прекращения приема новых задач
	wp.cancel()

	// Остановка всех воркеров
	for _, worker := range wp.workers {
		worker.stop()
	}

	// Ожидание завершения всех воркеров
	wp.wg.Wait()

	wp.running = false
	log.Println("Пул воркеров остановлен")
}

// AddTask добавляет задачу в пул воркеров
func (wp *WorkerPool) AddTask(taskID string) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if !wp.running {
		return fmt.Errorf("пул воркеров не запущен")
	}

	select {
	case wp.taskQueue <- &TaskJob{TaskID: taskID}:
		log.Printf("Задача %s добавлена в пул воркеров", taskID)
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("пул воркеров завершает работу")
	default:
		return fmt.Errorf("очередь задач переполнена")
	}
}

// dispatchTasks распределяет задачи между доступными воркерами
func (wp *WorkerPool) dispatchTasks() {
	log.Println("Диспетчер задач запущен")
	for {
		select {
		case job := <-wp.taskQueue:
			log.Printf("Диспетчер получил задачу %s", job.TaskID)
			// Поиск доступного воркера
			worker := wp.findAvailableWorker()
			if worker != nil {
				log.Printf("Найден доступный воркер %d для задачи %s", worker.id, job.TaskID)
				select {
				case worker.jobQueue <- job:
					log.Printf("Задача %s передана воркеру %d", job.TaskID, worker.id)
				default:
					log.Printf("Воркер %d занят, возвращаем задачу %s в очередь", worker.id, job.TaskID)
					// Воркер занят, возвращаем задачу в очередь
					go func() {
						time.Sleep(100 * time.Millisecond)
						select {
						case wp.taskQueue <- job:
						case <-wp.ctx.Done():
						}
					}()
				}
			} else {
				log.Printf("Нет доступных воркеров для задачи %s, возвращаем в очередь", job.TaskID)
				// Нет доступных воркеров, возвращаем задачу в очередь
				go func() {
					time.Sleep(100 * time.Millisecond)
					select {
					case wp.taskQueue <- job:
					case <-wp.ctx.Done():
					}
				}()
			}
		case <-wp.ctx.Done():
			log.Println("Диспетчер задач остановлен")
			return
		}
	}
}

// findAvailableWorker находит доступного воркера
func (wp *WorkerPool) findAvailableWorker() *Worker {
	for _, worker := range wp.workers {
		worker.mu.Lock()
		if !worker.busy {
			worker.busy = true
			worker.mu.Unlock()
			return worker
		}
		worker.mu.Unlock()
	}
	return nil
}

// start запускает воркера
func (w *Worker) start() {
	defer w.pool.wg.Done()

	log.Printf("Воркер %d запущен", w.id)

	for {
		select {
		case job := <-w.jobQueue:
			if job != nil {
				w.processJob(job)
				// Освобождаем воркера после обработки задачи
				w.mu.Lock()
				w.busy = false
				w.mu.Unlock()
			}
		case <-w.quit:
			log.Printf("Воркер %d остановлен", w.id)
			return
		}
	}
}

// stop останавливает воркера
func (w *Worker) stop() {
	w.quit <- true
}

// processJob обрабатывает задачу
func (w *Worker) processJob(job *TaskJob) {
	log.Printf("Воркер %d обрабатывает задачу %s", w.id, job.TaskID)

	// Получение ожидающих задач и обработка той, которая соответствует ID
	tasks, err := w.pool.downloadUsecase.GetPendingTasks(w.pool.ctx)
	if err != nil {
		log.Printf("Воркер %d не смог получить ожидающие задачи: %v", w.id, err)
		return
	}

	for _, task := range tasks {
		if task.ID.String() == job.TaskID {
			if err := w.pool.downloadUsecase.ProcessTask(w.pool.ctx, task); err != nil {
				log.Printf("Воркер %d не смог обработать задачу %s: %v", w.id, job.TaskID, err)
			} else {
				log.Printf("Воркер %d завершил задачу %s", w.id, job.TaskID)
			}
			return
		}
	}

	log.Printf("Воркер %d не смог найти задачу %s", w.id, job.TaskID)
}
