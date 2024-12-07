package coordinator

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// TaskState tracks the state of a task
type TaskState string

const (
	TaskStatePending   TaskState = "pending"
	TaskStateRunning   TaskState = "running"
	TaskStateCompleted TaskState = "completed"
	TaskStateFailed    TaskState = "failed"
	TaskStateCanceled  TaskState = "canceled"
)

// TaskEntry tracks a task and its current state
type TaskEntry struct {
	Task      types.Task
	State     TaskState
	StartedAt *time.Time
	EndedAt   *time.Time
	Error     error
}

// Scheduler implements task scheduling and tracking
type Scheduler struct {
	pending   []TaskEntry
	running   map[types.TaskID]TaskEntry
	completed []TaskEntry
	mu        sync.RWMutex
}

// NewScheduler creates a new scheduler instance
func NewScheduler() *Scheduler {
	return &Scheduler{
		pending:   make([]TaskEntry, 0),
		running:   make(map[types.TaskID]TaskEntry),
		completed: make([]TaskEntry, 0),
	}
}

// Schedule adds a new task to be scheduled
func (s *Scheduler) Schedule(ctx context.Context, task types.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create task entry
	entry := TaskEntry{
		Task:  task,
		State: TaskStatePending,
	}

	// Add to pending queue
	s.pending = append(s.pending, entry)

	// Sort by priority (higher priority first)
	sort.Slice(s.pending, func(i, j int) bool {
		return s.pending[i].Task.Priority > s.pending[j].Task.Priority
	})

	return nil
}

// StartTask moves a task from pending to running
func (s *Scheduler) StartTask(ctx context.Context, taskID types.TaskID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find task in pending queue
	for i, entry := range s.pending {
		if entry.Task.ID == taskID {
			// Update state
			now := time.Now()
			entry.State = TaskStateRunning
			entry.StartedAt = &now

			// Move to running map
			s.running[taskID] = entry

			// Remove from pending
			s.pending = append(s.pending[:i], s.pending[i+1:]...)
			return nil
		}
	}

	return nil
}

// CompleteTask marks a task as completed
func (s *Scheduler) CompleteTask(ctx context.Context, taskID types.TaskID, err error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, exists := s.running[taskID]; exists {
		// Update state
		now := time.Now()
		entry.EndedAt = &now
		entry.Error = err
		if err == nil {
			entry.State = TaskStateCompleted
		} else {
			entry.State = TaskStateFailed
		}

		// Move to completed list
		s.completed = append(s.completed, entry)

		// Remove from running
		delete(s.running, taskID)
		return nil
	}

	return nil
}

// Cancel cancels a pending or running task
func (s *Scheduler) Cancel(ctx context.Context, taskID types.TaskID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check pending tasks
	for i, entry := range s.pending {
		if entry.Task.ID == taskID {
			now := time.Now()
			entry.State = TaskStateCanceled
			entry.EndedAt = &now
			s.completed = append(s.completed, entry)
			s.pending = append(s.pending[:i], s.pending[i+1:]...)
			return nil
		}
	}

	// Check running tasks
	if entry, exists := s.running[taskID]; exists {
		now := time.Now()
		entry.State = TaskStateCanceled
		entry.EndedAt = &now
		s.completed = append(s.completed, entry)
		delete(s.running, taskID)
	}

	return nil
}

// GetTask returns information about a specific task
func (s *Scheduler) GetTask(ctx context.Context, taskID types.TaskID) (*TaskEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check pending tasks
	for _, entry := range s.pending {
		if entry.Task.ID == taskID {
			return &entry, nil
		}
	}

	// Check running tasks
	if entry, exists := s.running[taskID]; exists {
		return &entry, nil
	}

	// Check completed tasks
	for _, entry := range s.completed {
		if entry.Task.ID == taskID {
			return &entry, nil
		}
	}

	return nil, nil
}

// ListTasks returns all tasks
func (s *Scheduler) ListTasks(ctx context.Context) ([]TaskEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.pending) + len(s.running) + len(s.completed)
	tasks := make([]TaskEntry, 0, total)

	// Add all tasks
	tasks = append(tasks, s.pending...)
	for _, entry := range s.running {
		tasks = append(tasks, entry)
	}
	tasks = append(tasks, s.completed...)

	return tasks, nil
}