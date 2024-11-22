package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/utils"

	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupTest(t *testing.T) *Runner {
	// Mock IsMaster function by setting viper config
	viper.Set("node.master", true)
	defer viper.Set("node.master", nil) // cleanup after test

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
	}
	switch runtime.GOOS {
	case "windows":
		task.Cmd = "ping -n 10 127.0.0.1"
	default: // linux and darwin (macOS)
		task.Cmd = "sleep 10"
	}
	taskId, err := service.NewModelService[models.Task]().InsertOne(*task)
	require.NoError(t, err)
	task.Id = taskId

	// Create a task handler service
	svc := newTaskHandlerService()

	// Create a task runner
	runner, _ := newTaskRunner(task.Id, svc)
	require.NotNil(t, runner)

	// Set task and spider
	runner.t = task
	runner.s = spider

	// Initialize runner
	err = runner.configureCmd()
	require.Nil(t, err)

	return runner
}

func TestRunner_HandleIPC(t *testing.T) {
	// Setup test data
	runner := setupTest(t)

	// Create a pipe for testing
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()
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
		log.Info("IPC message was handled successfully")
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for IPC message to be handled")
	}
}

func TestRunner_Cancel(t *testing.T) {
	// Setup
	runner := setupTest(t)

	// Create pipes for stdout
	pr, pw := io.Pipe()
	runner.cmd.Stdout = pw
	runner.cmd.Stderr = pw

	// Start the command
	err := runner.cmd.Start()
	assert.NoError(t, err)
	log.Infof("started process with PID: %d", runner.cmd.Process.Pid)
	runner.pid = runner.cmd.Process.Pid

	// Read and print command output
	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			log.Info(scanner.Text())
		}
	}()

	// Wait a bit longer on Windows for the process to start properly
	waitTime := 100 * time.Millisecond
	if runtime.GOOS == "windows" {
		waitTime = 1 * time.Second
	}
	time.Sleep(waitTime)

	// Verify process exists before attempting to cancel
	if !utils.ProcessIdExists(runner.pid) {
		t.Fatalf("Process with PID %d was not started successfully", runner.pid)
	}

	// Test cancel
	go func() {
		err = runner.Cancel(true)
		assert.NoError(t, err)
	}()

	// Wait for process to be killed, with shorter timeout
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !utils.ProcessIdExists(runner.pid) {
			return // Process was killed
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Errorf("Process with PID %d was not killed within timeout", runner.pid)
}
