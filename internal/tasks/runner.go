package tasks

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"juvia/internal/db"
)

// Runner manages async tasks.
type Runner struct {
	db  *sql.DB
	log *slog.Logger
}

// NewRunner creates a new task runner.
func NewRunner(db *sql.DB, log *slog.Logger) *Runner {
	return &Runner{db: db, log: log}
}

// CreateTask inserts a new pending task.
func (r *Runner) CreateTask(ctx context.Context, taskType string, createdBy int64) (*db.Task, error) {
	taskID := fmt.Sprintf("task_%d%d", time.Now().UnixNano(), rand.Intn(10000))
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (task_id, type, status, progress, created_by) VALUES (?, ?, 'pending', 0, ?)`,
		taskID, taskType, createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return r.GetTaskByTaskID(ctx, taskID)
}

// UpdateTaskProgress updates task progress and steps.
func (r *Runner) UpdateTaskProgress(ctx context.Context, taskID string, progress int, steps string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET progress = ?, steps = ?, updated_at = CURRENT_TIMESTAMP WHERE task_id = ?`,
		progress, steps, taskID,
	)
	if err != nil {
		return fmt.Errorf("update task progress: %w", err)
	}
	return nil
}

// CompleteTask marks a task as done.
func (r *Runner) CompleteTask(ctx context.Context, taskID string, result string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET status = 'done', progress = 100, result = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE task_id = ?`,
		result, taskID,
	)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}
	return nil
}

// FailTask marks a task as failed.
func (r *Runner) FailTask(ctx context.Context, taskID string, result string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET status = 'failed', result = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE task_id = ?`,
		result, taskID,
	)
	if err != nil {
		return fmt.Errorf("fail task: %w", err)
	}
	return nil
}

// GetTaskByTaskID retrieves a task by its task_id.
func (r *Runner) GetTaskByTaskID(ctx context.Context, taskID string) (*db.Task, error) {
	var t db.Task
	row := r.db.QueryRowContext(ctx, `SELECT id, task_id, type, status, progress, steps, result, created_by, created_at, updated_at, completed_at FROM tasks WHERE task_id = ?`, taskID)
	if err := row.Scan(&t.ID, &t.TaskID, &t.Type, &t.Status, &t.Progress, &t.Steps, &t.Result, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return &t, nil
}

// ListTasks returns tasks ordered by creation date.
func (r *Runner) ListTasks(ctx context.Context, limit int) ([]db.Task, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, task_id, type, status, progress, steps, result, created_by, created_at, updated_at, completed_at FROM tasks ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []db.Task
	for rows.Next() {
		var t db.Task
		if err := rows.Scan(&t.ID, &t.TaskID, &t.Type, &t.Status, &t.Progress, &t.Steps, &t.Result, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
