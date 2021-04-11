package gue

import (
	"time"

	"github.com/vgarvardt/gue/v2/adapter"
)

// WorkerOption defines a type that allows to set worker properties during the build-time.
type WorkerOption func(*Worker)

// WorkerPoolOption defines a type that allows to set worker pool properties during the build-time.
type WorkerPoolOption func(pool *WorkerPool)

// WithWorkerPollInterval overrides default poll interval with the given value.
// Poll interval is the "sleep" duration if there were no jobs found in the DB.
func WithWorkerPollInterval(d time.Duration) WorkerOption {
	return func(w *Worker) {
		w.interval = d
	}
}

// WithWorkerQueue overrides default worker queue name with the given value.
func WithWorkerQueue(queue string) WorkerOption {
	return func(w *Worker) {
		w.queue = queue
	}
}

// WithWorkerID sets worker ID for easier identification in logs
func WithWorkerID(id string) WorkerOption {
	return func(w *Worker) {
		w.id = id
	}
}

// WithWorkerLogger sets Logger implementation to worker
func WithWorkerLogger(logger adapter.Logger) WorkerOption {
	return func(w *Worker) {
		w.logger = logger
	}
}

// WithWorkerAllJobs sets the worker to process every job type
func WithWorkerAllJobs(all bool) WorkerOption {
	return func(w *Worker) {
		w.allJobs = all
	}
}

// WithWorkerMaxErrorCount sets the maximum amount of errors a job can have before it is ignored
func WithWorkerMaxErrorCount(count int) WorkerOption {
	return func(c *Worker) {
		c.maxErrorCount = count
	}
}

// WithWorkerMinErrorCount sets the minimum amount of can a job can have before it is selected
func WithWorkerMinErrorCount(count int) WorkerOption {
	return func(c *Worker) {
		c.minErrorCount = count
	}
}

// WithPoolPollInterval overrides default poll interval with the given value.
// Poll interval is the "sleep" duration if there were no jobs found in the DB.
func WithPoolPollInterval(d time.Duration) WorkerPoolOption {
	return func(w *WorkerPool) {
		w.interval = d
	}
}

// WithPoolQueue overrides default worker queue name with the given value.
func WithPoolQueue(queue string) WorkerPoolOption {
	return func(w *WorkerPool) {
		w.queue = queue
	}
}

// WithPoolID sets worker pool ID for easier identification in logs
func WithPoolID(id string) WorkerPoolOption {
	return func(w *WorkerPool) {
		w.id = id
	}
}

// WithPoolLogger sets Logger implementation to worker pool
func WithPoolLogger(logger adapter.Logger) WorkerPoolOption {
	return func(w *WorkerPool) {
		w.logger = logger
	}
}

// WithWorkerPoolAllJobs sets the worker to process every job type
func WithWorkerPoolAllJobs(all bool) WorkerPoolOption {
	return func(w *WorkerPool) {
		w.allJobs = all
	}
}

// WithWorkerPoolMaxErrorCount sets the maximum amount of errors a job can have before it is ignored
func WithWorkerPoolMaxErrorCount(count int) WorkerPoolOption {
	return func(c *WorkerPool) {
		c.maxErrorCount = count
	}
}

// WithWorkerPoolMinErrorCount sets the minimum amount of can a job can have before it is selected
func WithWorkerPoolMinErrorCount(count int) WorkerPoolOption {
	return func(c *WorkerPool) {
		c.minErrorCount = count
	}
}
