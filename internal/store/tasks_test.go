package store

import (
	"testing"
)

func TestCreateTask(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	id, err := CreateTask(db, "Buy groceries")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestListTasks(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	CreateTask(db, "Task A")
	CreateTask(db, "Task B")
	CreateTask(db, "Task C")

	tasks, err := ListTasks(db)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
	// Should be ordered by created_at
	if tasks[0].Title != "Task A" {
		t.Errorf("expected first task 'Task A', got %s", tasks[0].Title)
	}
}

func TestToggleTask(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	id, _ := CreateTask(db, "Toggle me")

	tasks, _ := ListTasks(db)
	if tasks[0].Status != "todo" {
		t.Fatalf("expected initial status 'todo', got %s", tasks[0].Status)
	}

	// Toggle to done
	if err := ToggleTask(db, id); err != nil {
		t.Fatalf("toggle to done: %v", err)
	}
	tasks, _ = ListTasks(db)
	if tasks[0].Status != "done" {
		t.Errorf("expected 'done', got %s", tasks[0].Status)
	}
	if tasks[0].CompletedAt == nil {
		t.Error("completed_at should be set after done toggle")
	}

	// Toggle back to todo
	if err := ToggleTask(db, id); err != nil {
		t.Fatalf("toggle to todo: %v", err)
	}
	tasks, _ = ListTasks(db)
	if tasks[0].Status != "todo" {
		t.Errorf("expected 'todo', got %s", tasks[0].Status)
	}
	if tasks[0].CompletedAt != nil {
		t.Error("completed_at should be nil after todo toggle")
	}
}

func TestDeleteTask(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	id, _ := CreateTask(db, "Delete me")
	if err := DeleteTask(db, id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	tasks, _ := ListTasks(db)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestTaskStats(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	id1, _ := CreateTask(db, "T1")
	CreateTask(db, "T2")
	CreateTask(db, "T3")

	ToggleTask(db, id1) // mark T1 as done

	completed, total, err := TaskStats(db)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if completed != 1 {
		t.Errorf("expected 1 completed, got %d", completed)
	}
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
}
