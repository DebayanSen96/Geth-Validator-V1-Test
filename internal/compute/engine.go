package compute

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Task represents a computation task
type Task struct {
	ID       string
	Data     []byte
	Status   string
	Result   []byte
	Created  time.Time
	Finished time.Time
}

// Engine represents a computation engine for off-chain tasks
type Engine struct {
	tasks map[string]*Task
	mutex sync.Mutex
}

// NewEngine creates a new compute engine
func NewEngine() *Engine {
	return &Engine{
		tasks: make(map[string]*Task),
		mutex: sync.Mutex{},
	}
}

// SubmitTask submits a new computation task
func (e *Engine) SubmitTask(taskID string, data []byte) string {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Create a new task
	task := &Task{
		ID:      taskID,
		Data:    data,
		Status:  "pending",
		Created: time.Now(),
	}

	// Store the task
	e.tasks[taskID] = task

	// Start processing the task in a goroutine
	go e.processTask(taskID)

	return taskID
}

// GetTaskStatus gets the status of a task
func (e *Engine) GetTaskStatus(taskID string) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	task, ok := e.tasks[taskID]
	if !ok {
		return "", errors.New("task not found")
	}

	return task.Status, nil
}

// GetTaskResult gets the result of a completed task
func (e *Engine) GetTaskResult(taskID string) ([]byte, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	task, ok := e.tasks[taskID]
	if !ok {
		return nil, errors.New("task not found")
	}

	if task.Status != "completed" {
		return nil, fmt.Errorf("task is not completed, current status: %s", task.Status)
	}

	return task.Result, nil
}

// WaitForResult waits for a task to complete and returns the result
func (e *Engine) WaitForResult(taskID string, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := e.GetTaskStatus(taskID)
		if err != nil {
			return nil, err
		}

		if status == "completed" {
			return e.GetTaskResult(taskID)
		} else if status == "failed" {
			return nil, errors.New("task failed")
		}

		// Wait a bit before checking again
		time.Sleep(100 * time.Millisecond)
	}

	return nil, errors.New("timeout waiting for task completion")
}

// processTask processes a computation task
func (e *Engine) processTask(taskID string) {
	// Simulate computation time
	time.Sleep(1 * time.Second)

	e.mutex.Lock()
	defer e.mutex.Unlock()

	task, ok := e.tasks[taskID]
	if !ok {
		return
	}

	// Perform the computation
	// For this example, we'll just compute a hash of the data
	hash := sha256.Sum256(task.Data)
	result := []byte(hex.EncodeToString(hash[:]))

	// Update the task
	task.Status = "completed"
	task.Result = result
	task.Finished = time.Now()
}
