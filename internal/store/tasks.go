package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Task struct {
	ID          int64
	Title       string
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func CreateTask(db *sql.DB, title string) (int64, error) {
	res, err := db.Exec("INSERT INTO tasks (title, status) VALUES (?, 'todo')", title)
	if err != nil {
		return 0, fmt.Errorf("create task: %w", err)
	}
	return res.LastInsertId()
}

func ListTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query(
		"SELECT id, title, status, created_at, completed_at FROM tasks ORDER BY created_at",
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ToggleTask flips a task between todo and done.
// Setting to done records completed_at; setting back to todo clears it.
func ToggleTask(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		UPDATE tasks SET
			status = CASE WHEN status = 'todo' THEN 'done' ELSE 'todo' END,
			completed_at = CASE WHEN status = 'todo' THEN CURRENT_TIMESTAMP ELSE NULL END
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("toggle task: %w", err)
	}
	return nil
}

func DeleteTask(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func TaskStats(db *sql.DB) (completed, total int, err error) {
	err = db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END), 0),
			COUNT(*)
		FROM tasks
	`).Scan(&completed, &total)
	if err != nil {
		return 0, 0, fmt.Errorf("task stats: %w", err)
	}
	return completed, total, nil
}
