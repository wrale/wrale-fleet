package coordinator

import (
	"context"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/types"
)

func TestScheduler(t *testing.T) {
	ctx := context.Background()
	scheduler := NewScheduler()

	// Test scheduling a task
	task := types.Task{
		ID:        "task-1",
		DeviceIDs: []types.DeviceID{"device-1", "device-2"},
		Operation: "test_operation",
		Priority:  1,
		Resources: map[types.ResourceType]float64{
			types.ResourceCPU:    50.0,
			types.ResourceMemory: 60.0,
		},
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	t.Run("Schedule Task", func(t *testing.T) {
		err := scheduler.Schedule(ctx, task)
		if err != nil {
			t.Errorf("Failed to schedule task: %v", err)
		}

		// Verify task was scheduled
		entry, err := scheduler.GetTask(ctx, task.ID)
		if err != nil {
			t.Errorf("Failed to get task: %v", err)
		}
		if entry == nil {
			t.Error("Task not found after scheduling")
		}
		if entry.State != TaskStatePending {
			t.Errorf("Expected task state %s, got %s", TaskStatePending, entry.State)
		}
	})

	t.Run("Start Task", func(t *testing.T) {
		err := scheduler.StartTask(ctx, task.ID)
		if err != nil {
			t.Errorf("Failed to start task: %v", err)
		}

		entry, err := scheduler.GetTask(ctx, task.ID)
		if err != nil {
			t.Errorf("Failed to get task: %v", err)
		}
		if entry.State != TaskStateRunning {
			t.Errorf("Expected task state %s, got %s", TaskStateRunning, entry.State)
		}
		if entry.StartedAt == nil {
			t.Error("StartedAt not set after starting task")
		}
	})

	t.Run("Complete Task", func(t *testing.T) {
		err := scheduler.CompleteTask(ctx, task.ID, nil)
		if err != nil {
			t.Errorf("Failed to complete task: %v", err)
		}

		entry, err := scheduler.GetTask(ctx, task.ID)
		if err != nil {
			t.Errorf("Failed to get task: %v", err)
		}
		if entry.State != TaskStateCompleted {
			t.Errorf("Expected task state %s, got %s", TaskStateCompleted, entry.State)
		}
		if entry.EndedAt == nil {
			t.Error("EndedAt not set after completing task")
		}
	})

	t.Run("List Tasks", func(t *testing.T) {
		tasks, err := scheduler.ListTasks(ctx)
		if err != nil {
			t.Errorf("Failed to list tasks: %v", err)
		}
		if len(tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(tasks))
		}
	})

	t.Run("Cancel Task", func(t *testing.T) {
		// Schedule a new task to cancel
		task2 := task
		task2.ID = "task-2"
		
		err := scheduler.Schedule(ctx, task2)
		if err != nil {
			t.Errorf("Failed to schedule task: %v", err)
		}

		err = scheduler.Cancel(ctx, task2.ID)
		if err != nil {
			t.Errorf("Failed to cancel task: %v", err)
		}

		entry, err := scheduler.GetTask(ctx, task2.ID)
		if err != nil {
			t.Errorf("Failed to get task: %v", err)
		}
		if entry.State != TaskStateCanceled {
			t.Errorf("Expected task state %s, got %s", TaskStateCanceled, entry.State)
		}
	})

	t.Run("Priority Ordering", func(t *testing.T) {
		scheduler := NewScheduler()

		// Schedule tasks with different priorities
		task1 := task
		task1.ID = "task-1"
		task1.Priority = 1

		task2 := task
		task2.ID = "task-2"
		task2.Priority = 3

		task3 := task
		task3.ID = "task-3"
		task3.Priority = 2

		// Schedule in reverse priority order
		if err := scheduler.Schedule(ctx, task1); err != nil {
			t.Errorf("Failed to schedule task1: %v", err)
		}
		if err := scheduler.Schedule(ctx, task2); err != nil {
			t.Errorf("Failed to schedule task2: %v", err)
		}
		if err := scheduler.Schedule(ctx, task3); err != nil {
			t.Errorf("Failed to schedule task3: %v", err)
		}

        // Verify priority ordering by checking the order of pending tasks
        tasks, err := scheduler.ListTasks(ctx)
        if err != nil {
            t.Errorf("Failed to list tasks: %v", err)
        }

        var pendingTasks []TaskEntry
        for _, entry := range tasks {
            if entry.State == TaskStatePending {
                pendingTasks = append(pendingTasks, entry)
            }
        }

        if len(pendingTasks) != 3 {
            t.Errorf("Expected 3 pending tasks, got %d", len(pendingTasks))
        }

        // Check that tasks are ordered by priority (highest first)
        for i := 1; i < len(pendingTasks); i++ {
            if pendingTasks[i-1].Task.Priority < pendingTasks[i].Task.Priority {
                t.Errorf("Tasks not properly ordered by priority: %d before %d",
                    pendingTasks[i-1].Task.Priority, pendingTasks[i].Task.Priority)
            }
        }
    })

    t.Run("Failed Task", func(t *testing.T) {
        task := types.Task{
            ID:        "failed-task",
            DeviceIDs: []types.DeviceID{"device-1"},
            Operation: "test_operation",
            Priority:  1,
            CreatedAt: time.Now(),
        }

        if err := scheduler.Schedule(ctx, task); err != nil {
            t.Errorf("Failed to schedule task: %v", err)
        }

        if err := scheduler.StartTask(ctx, task.ID); err != nil {
            t.Errorf("Failed to start task: %v", err)
        }

        testError := fmt.Errorf("test error")
        if err := scheduler.CompleteTask(ctx, task.ID, testError); err != nil {
            t.Errorf("Failed to complete task: %v", err)
        }

        entry, err := scheduler.GetTask(ctx, task.ID)
        if err != nil {
            t.Errorf("Failed to get task: %v", err)
        }

        if entry.State != TaskStateFailed {
            t.Errorf("Expected task state %s, got %s", TaskStateFailed, entry.State)
        }

        if entry.Error == nil || entry.Error.Error() != testError.Error() {
            t.Errorf("Task error not properly recorded")
        }
    })
}