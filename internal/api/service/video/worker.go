package video

import (
	"context"
	"fmt"
	"go-fitness/external/logger/sl"
	"log/slog"
)

type Transcoder interface {
	ProcessTranscode(ctx context.Context, task TranscodeTask) error
}

type WorkerPool struct {
	log              *slog.Logger
	taskQueue        chan TranscodeTask
	transcodeService Transcoder
}

// TranscodeTask is a struct for video transcode task
type TranscodeTask struct {
	UploadPath string
	VideoID    int64
	DstPath    string
	ChunkHash  string
}

func NewWorkerPool(
	log *slog.Logger,
	transcodeService Transcoder,
) *WorkerPool {
	pool := &WorkerPool{
		log:              log,
		taskQueue:        make(chan TranscodeTask, 50),
		transcodeService: transcodeService,
	}

	workerCount := 1

	for i := 0; i < workerCount; i++ {
		go pool.worker(i)
	}

	return pool
}

func (p *WorkerPool) worker(id int) {
	const op string = "WorkerPool.worker"

	log := p.log.With(
		sl.String("op", op),
		sl.Int("worker_id", id),
	)

	for task := range p.taskQueue {
		log.Info("Processing task", sl.String("task", fmt.Sprintf("%+v", task)))
		if err := p.transcodeService.ProcessTranscode(context.Background(), task); err != nil {
			log.Error("Failed to process transcode", sl.Err(err))
		}
	}
}

// AddTask adds a task to the task queue
func (p *WorkerPool) AddTask(task TranscodeTask) {
	p.taskQueue <- task
}
