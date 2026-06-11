package tasks

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"juvia/internal/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := db.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := d.ApplyMigrations("../db/migrations"); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return d.DB
}

func TestCreateTask(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	r := NewRunner(sqlDB, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()

	task, err := r.CreateTask(ctx, "test_task", 1)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.TaskID == "" {
		t.Fatal("expected task_id")
	}
	if task.Status != "pending" {
		t.Errorf("expected status pending, got %s", task.Status)
	}
	if task.Type != "test_task" {
		t.Errorf("expected type test_task, got %s", task.Type)
	}
}

func TestUpdateAndCompleteTask(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	r := NewRunner(sqlDB, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()

	task, _ := r.CreateTask(ctx, "test_task", 1)

	if err := r.UpdateTaskProgress(ctx, task.TaskID, 50, `[{"step":"running"}]`); err != nil {
		t.Fatalf("UpdateTaskProgress failed: %v", err)
	}

	if err := r.CompleteTask(ctx, task.TaskID, `{"ok":true}`); err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	task2, err := r.GetTaskByTaskID(ctx, task.TaskID)
	if err != nil {
		t.Fatalf("GetTaskByTaskID failed: %v", err)
	}
	if task2.Status != "done" {
		t.Errorf("expected status done, got %s", task2.Status)
	}
	if task2.Progress != 100 {
		t.Errorf("expected progress 100, got %d", task2.Progress)
	}
}

func TestFailTask(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	r := NewRunner(sqlDB, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()

	task, _ := r.CreateTask(ctx, "test_task", 1)
	if err := r.FailTask(ctx, task.TaskID, `{"error":"failed"}`); err != nil {
		t.Fatalf("FailTask failed: %v", err)
	}

	task2, err := r.GetTaskByTaskID(ctx, task.TaskID)
	if err != nil {
		t.Fatalf("GetTaskByTaskID failed: %v", err)
	}
	if task2.Status != "failed" {
		t.Errorf("expected status failed, got %s", task2.Status)
	}
}

func TestListTasks(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()

	r := NewRunner(sqlDB, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()

	r.CreateTask(ctx, "task_a", 1)
	r.CreateTask(ctx, "task_b", 1)

	tasks, err := r.ListTasks(ctx, 10)
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}
