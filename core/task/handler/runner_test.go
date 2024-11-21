package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupTest(t *testing.T) *Runner {
	// Create a test spider
	spider := &models.Spider{
		Name: "Test Spider",
	}
	spiderId, err := service.NewModelService[models.Spider]().InsertOne(*spider)
	require.NoError(t, err)
	spider.Id = spiderId

	// Create a test task
	task := &models.Task{
		SpiderId: spiderId,
		Status:   constants.TaskStatusPending,
		Type:     "test",
		Mode:     "test",
		NodeId:   primitive.NewObjectID(),
		Cmd:      "python script.py",
	}
	taskId, err := service.NewModelService[models.Task]().InsertOne(*task)
	require.NoError(t, err)
	task.Id = taskId

	// Create a test runner
	svc := newTaskHandlerService()
	runner, _ := newTaskRunner(task.Id, svc)
	err = runner.updateTask("", nil)
	require.Nil(t, err)
	_ = runner.Init()
	err = runner.configureCmd()
	require.Nil(t, err)

	return runner
}

func TestRunner_HandleIPC(t *testing.T) {
	// Setup test data
	runner := setupTest(t)

	// Create a pipe for testing
	pr, pw := io.Pipe()
	runner.stdoutPipe = pr

	// Start IPC reader
	go runner.startIPCReader()

	// Create test message
	testMsg := IPCMessage{
		Type:    "test_type",
		Payload: map[string]interface{}{"key": "value"},
		IPC:     true,
	}

	// Create a channel to signal that the message was handled
	handled := make(chan bool)
	runner.SetIPCHandler(func(msg IPCMessage) {
		assert.Equal(t, testMsg.Type, msg.Type)
		assert.Equal(t, testMsg.Payload, msg.Payload)
		handled <- true
	})

	// Convert message to JSON and write to pipe
	go func() {
		jsonData, err := json.Marshal(testMsg)
		if err != nil {
			t.Errorf("failed to marshal test message: %v", err)
			return
		}

		// Write message followed by newline
		_, err = fmt.Fprintln(pw, string(jsonData))
		if err != nil {
			t.Errorf("failed to write to pipe: %v", err)
			return
		}
	}()

	select {
	case <-handled:
		// Message was handled successfully
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for IPC message to be handled")
	}

	// Clean up
	pw.Close()
	pr.Close()
}

func TestRunner_Cancel(t *testing.T) {
	// Setup
	runner := setupTest(t)

	// Start a long-running command
	runner.t.Cmd = "sleep 10"
	err := runner.cmd.Start()
	assert.NoError(t, err)
	runner.pid = runner.cmd.Process.Pid

	// Test cancel
	err = runner.Cancel(true)
	assert.NoError(t, err)

	// Verify process was killed
	// Wait a short time for the process to be killed
	time.Sleep(100 * time.Millisecond)

	process, err := os.FindProcess(runner.pid)
	require.NoError(t, err)
	err = process.Signal(syscall.Signal(0))
	assert.Error(t, err) // Process should not exist
}

// Helper function to create a temporary workspace for testing
func createTestWorkspace(t *testing.T) string {
	dir, err := os.MkdirTemp("", "crawlab-test-*")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}
