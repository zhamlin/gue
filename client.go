package gue

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/vgarvardt/gue/v2/adapter"
	"github.com/vgarvardt/gue/v2/adapter/exponential"
)

// ErrMissingType is returned when you attempt to enqueue a job with no Type
// specified.
var ErrMissingType = errors.New("job type must be specified")

// Client is a Gue client that can add jobs to the queue and remove jobs from
// the queue.
type Client struct {
	pool    adapter.ConnPool
	logger  adapter.Logger
	id      string
	backoff Backoff
}

// NewClient creates a new Client that uses the pgx pool.
func NewClient(pool adapter.ConnPool, options ...ClientOption) *Client {
	instance := Client{
		pool:    pool,
		logger:  adapter.NoOpLogger{},
		backoff: exponential.Default,
	}

	for _, option := range options {
		option(&instance)
	}

	if instance.id == "" {
		instance.id = newID()
	}

	instance.logger = instance.logger.With(adapter.F("client-id", instance.id))

	return &instance
}

// Enqueue adds a job to the queue.
func (c *Client) Enqueue(ctx context.Context, j *Job) error {
	return c.execEnqueue(ctx, j, c.pool)
}

// EnqueueTx adds a job to the queue within the scope of the transaction.
// This allows you to guarantee that an enqueued job will either be committed or
// rolled back atomically with other changes in the course of this transaction.
//
// It is the caller's responsibility to Commit or Rollback the transaction after
// this function is called.
func (c *Client) EnqueueTx(ctx context.Context, j *Job, tx adapter.Tx) error {
	return c.execEnqueue(ctx, j, tx)
}

func (c *Client) execEnqueue(ctx context.Context, j *Job, q adapter.Queryable) error {
	if j.Type == "" {
		return ErrMissingType
	}

	now := time.Now()

	runAt := j.RunAt
	if runAt.IsZero() {
		j.RunAt = now
	}

	if len(j.Args) == 0 {
		j.Args = []byte(`[]`)
	}

	err := q.QueryRow(ctx, `INSERT INTO gue_jobs
(queue, priority, run_at, job_type, args, created_at, updated_at)
VALUES
($1, $2, $3, $4, $5, $6, $6) RETURNING job_id
`, j.Queue, j.Priority, j.RunAt, j.Type, j.Args, now).Scan(&j.ID)

	c.logger.Debug(
		"Tried to enqueue a job",
		adapter.Err(err),
		adapter.F("queue", j.Queue),
		adapter.F("id", j.ID),
	)

	return err
}

// LockJob attempts to retrieve a Job from the database in the specified queue.
// If a job is found, it will be locked on the transactional level, so other workers
// will be skipping it. If no job is found, nil will be returned instead of an error.
//
// Because Gue uses transaction-level locks, we have to hold the
// same transaction throughout the process of getting a job, working it,
// deleting it, and releasing the lock.
//
// After the Job has been worked, you must call either Done() or Error() on it
// in order to commit transaction to persist Job changes (remove or update it).
func (c *Client) LockJob(ctx context.Context, queue string, maxErrorCount int) (*Job, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	j := Job{pool: c.pool, tx: tx, backoff: c.backoff}

	query := `
SELECT job_id, queue, priority, run_at, job_type, args, error_count
FROM gue_jobs
WHERE run_at <= $1 AND queue = $2
%s
ORDER BY priority ASC
LIMIT 1 FOR UPDATE SKIP LOCKED`

	args := []interface{}{time.Now(), queue}
	var whereCond string
	if maxErrorCount > 0 {
		args = append(args, maxErrorCount)
		whereCond = fmt.Sprintf("AND error_count < $%d", len(args))
	}

	q := fmt.Sprintf(query, whereCond)
	err = tx.QueryRow(ctx, q, args...).Scan(
		&j.ID,
		&j.Queue,
		&j.Priority,
		&j.RunAt,
		&j.Type,
		&j.Args,
		&j.ErrorCount,
	)
	if err == nil {
		return &j, nil
	}

	rbErr := tx.Rollback(ctx)
	if err == adapter.ErrNoRows {
		return nil, rbErr
	}

	return nil, fmt.Errorf("could not lock a job (rollback result: %v): %w", rbErr, err)
}

func (c *Client) LockJobMinError(ctx context.Context, queue string, minErrorCount int) (*Job, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	j := Job{pool: c.pool, tx: tx, backoff: c.backoff}

	query := `
SELECT job_id, queue, priority, run_at, job_type, args, error_count
FROM gue_jobs
WHERE error_count >= $1
%s
ORDER BY priority ASC
LIMIT 1 FOR UPDATE SKIP LOCKED`

	args := []interface{}{minErrorCount}
	var whereCond string
	if queue != "" {
		args = append(args, queue)
		whereCond = fmt.Sprintf("AND queue = $%d", len(args))
	}

	q := fmt.Sprintf(query, whereCond)
	err = tx.QueryRow(ctx, q, args...).Scan(
		&j.ID,
		&j.Queue,
		&j.Priority,
		&j.RunAt,
		&j.Type,
		&j.Args,
		&j.ErrorCount,
	)
	if err == nil {
		return &j, nil
	}

	rbErr := tx.Rollback(ctx)
	if err == adapter.ErrNoRows {
		return nil, rbErr
	}

	return nil, fmt.Errorf("could not lock a job (rollback result: %v): %w", rbErr, err)
}

func newID() string {
	hasher := md5.New()
	// nolint:errcheck
	hasher.Write([]byte(time.Now().Format(time.RFC3339Nano)))
	return hex.EncodeToString(hasher.Sum(nil))[:6]
}
