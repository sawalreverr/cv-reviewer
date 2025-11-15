package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sawalreverr/cv-reviewer/config"
	"github.com/sawalreverr/cv-reviewer/pkg/errors"
)

type Job struct {
	ID uuid.UUID
	JobTitle string
	CVID uuid.UUID
	ProjectID uuid.UUID
}

type JobQueue interface {
	Enqueue(job Job) error
	Start(ctx context.Context)
	Stop()
} 

type jobQueue struct {
	queue chan Job
	workerCount int
	processor JobProcessor
	wg sync.WaitGroup
	ctx context.Context
	cancel context.CancelFunc
}

type JobProcessor interface {
	Process(ctx context.Context, job Job) error
}

func NewJobQueue(cfg *config.QueueConfig, processor JobProcessor) JobQueue {
	ctx, cancel := context.WithCancel(context.Background())

	return &jobQueue{
		queue: make(chan Job, cfg.QueueSize),
		workerCount: cfg.WorkerCount,
		processor: processor,
		ctx: ctx,
		cancel: cancel,
	}
}

func (q *jobQueue) Enqueue(job Job) error {
	select{
	case q.queue <- job:
		log.Printf("job %s enqueued successfully", job.ID)
		return nil
	case <-time.After(5 *time.Second):
		return errors.ErrQueueFull
	}
}

func (q *jobQueue) Start(ctx context.Context) {
	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

func (q *jobQueue) Stop() {
	log.Println("stopping job queue...")
	q.cancel()
	close(q.queue)
	q.wg.Wait()
}

func (q *jobQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case job, ok := <- q.queue:
			if !ok {return}

			log.Printf("worker %d: processing job %s", id, job.ID)
			if err := q.processor.Process(q.ctx, job); err != nil {
				log.Printf("worker %d: failed to process job %s: %v", id, job.ID, err)
			} else {
				log.Printf("worker %d: success to process job %s", id, job.ID)
			}
		case <-q.ctx.Done(): return
		}
	}
}
