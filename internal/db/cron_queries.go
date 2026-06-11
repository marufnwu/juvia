package db

import (
	"context"
	"fmt"
)

func (db *DB) CreateCronJob(ctx context.Context, userID int64, schedule, command, runAs, jobType string) (*CronJob, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO cron_jobs (user_id, schedule, command, run_as, type, enabled) VALUES (?, ?, ?, ?, ?, TRUE)`,
		userID, schedule, command, runAs, jobType)
	if err != nil {
		return nil, fmt.Errorf("create cron job: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetCronJobByID(ctx, id)
}

func (db *DB) GetCronJobByID(ctx context.Context, id int64) (*CronJob, error) {
	var j CronJob
	row := db.QueryRowContext(ctx,
		`SELECT id, user_id, schedule, command, run_as, type, enabled, last_run, last_status, last_output, created_at
		FROM cron_jobs WHERE id = ?`, id)
	if err := row.Scan(&j.ID, &j.UserID, &j.Schedule, &j.Command, &j.RunAs, &j.Type, &j.Enabled, &j.LastRun, &j.LastStatus, &j.LastOutput, &j.CreatedAt); err != nil {
		return nil, fmt.Errorf("get cron job: %w", err)
	}
	return &j, nil
}

type ListCronJobsResult struct {
	CronJobs []CronJob
	Total    int
}

func (db *DB) ListCronJobs(ctx context.Context, userID int64, page, limit int) (*ListCronJobsResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	countRow := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cron_jobs WHERE user_id = ?", userID)
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("count cron jobs: %w", err)
	}

	query := `SELECT id, user_id, schedule, command, run_as, type, enabled, last_run, last_status, last_output, created_at
		FROM cron_jobs WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list cron jobs: %w", err)
	}
	defer rows.Close()

	var jobs []CronJob
	for rows.Next() {
		var j CronJob
		if err := rows.Scan(&j.ID, &j.UserID, &j.Schedule, &j.Command, &j.RunAs, &j.Type, &j.Enabled, &j.LastRun, &j.LastStatus, &j.LastOutput, &j.CreatedAt); err != nil {
			continue
		}
		jobs = append(jobs, j)
	}
	return &ListCronJobsResult{CronJobs: jobs, Total: total}, nil
}

func (db *DB) UpdateCronJob(ctx context.Context, id int64, schedule, command, runAs string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE cron_jobs SET schedule = ?, command = ?, run_as = ? WHERE id = ?`,
		schedule, command, runAs, id)
	if err != nil {
		return fmt.Errorf("update cron job: %w", err)
	}
	return nil
}

func (db *DB) DeleteCronJob(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM cron_jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete cron job: %w", err)
	}
	return nil
}

func (db *DB) EnableCronJob(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `UPDATE cron_jobs SET enabled = TRUE WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("enable cron job: %w", err)
	}
	return nil
}

func (db *DB) DisableCronJob(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `UPDATE cron_jobs SET enabled = FALSE WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("disable cron job: %w", err)
	}
	return nil
}

func (db *DB) UpdateCronJobLastRun(ctx context.Context, id int64, status, output string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE cron_jobs SET last_run = CURRENT_TIMESTAMP, last_status = ?, last_output = ? WHERE id = ?`,
		status, output, id)
	if err != nil {
		return fmt.Errorf("update cron job last run: %w", err)
	}
	return nil
}

func (db *DB) CreateCronJobLog(ctx context.Context, cronJobID int64, durationMs int64, status, output string) (*CronJobLog, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO cron_job_logs (cron_job_id, duration_ms, status, output) VALUES (?, ?, ?, ?)`,
		cronJobID, durationMs, status, output)
	if err != nil {
		return nil, fmt.Errorf("create cron job log: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetCronJobLogByID(ctx, id)
}

func (db *DB) GetCronJobLogByID(ctx context.Context, id int64) (*CronJobLog, error) {
	var l CronJobLog
	row := db.QueryRowContext(ctx,
		`SELECT id, cron_job_id, run_at, duration_ms, status, output, created_at
		FROM cron_job_logs WHERE id = ?`, id)
	if err := row.Scan(&l.ID, &l.CronJobID, &l.RunAt, &l.DurationMs, &l.Status, &l.Output, &l.CreatedAt); err != nil {
		return nil, fmt.Errorf("get cron job log: %w", err)
	}
	return &l, nil
}

func (db *DB) ListCronJobLogs(ctx context.Context, cronJobID int64, limit int) ([]CronJobLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `SELECT id, cron_job_id, run_at, duration_ms, status, output, created_at
		FROM cron_job_logs WHERE cron_job_id = ? ORDER BY run_at DESC LIMIT ?`
	rows, err := db.QueryContext(ctx, query, cronJobID, limit)
	if err != nil {
		return nil, fmt.Errorf("list cron job logs: %w", err)
	}
	defer rows.Close()

	var logs []CronJobLog
	for rows.Next() {
		var l CronJobLog
		if err := rows.Scan(&l.ID, &l.CronJobID, &l.RunAt, &l.DurationMs, &l.Status, &l.Output, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}
	return logs, nil
}