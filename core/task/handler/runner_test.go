package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/crawlab-team/crawlab/core/entity"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
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
		task.Cmd = "ping -c 10 127.0.0.1"
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
	testMsg := entity.IPCMessage{
		Type:    "test_type",
		Payload: map[string]interface{}{"key": "value"},
		IPC:     true,
	}

	// Create a channel to signal that the message was handled
	handled := make(chan bool)
	runner.SetIPCHandler(func(msg entity.IPCMessage) {
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

	// Wait for process to finish
	go func() {
		err = runner.cmd.Wait()
		if err != nil {
			log.Errorf("process[%d] exited with error: %v", runner.pid, err)
			return
		}
		log.Infof("process[%d] exited successfully", runner.pid)
	}()

	// Wait for a certain period for the process to start properly
	time.Sleep(1 * time.Second)

	// Verify process exists before attempting to cancel
	if !utils.ProcessIdExists(runner.pid) {
		t.Fatalf("Process with PID %d was not started successfully", runner.pid)
	}

	// Test cancel
	go func() {
		err = runner.Cancel(true)
		assert.NoError(t, err)
		log.Infof("process[%d] cancelled", runner.pid)
	}()

	// Wait for process to be killed, with shorter timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Process with PID %d was not killed within timeout", runner.pid)
		case <-ticker.C:
			exists := utils.ProcessIdExists(runner.pid)
			if !exists {
				return // Exit the test when process is killed
			}
		}
	}
}

func TestRunner_HandleIPCData(t *testing.T) {
	// Setup test data
	runner := setupTest(t)

	// Create pipes for testing
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()
	runner.stdoutPipe = pr

	// Start IPC reader
	go runner.startIPCReader()

	// Test cases
	testCases := []struct {
		name     string
		payload  interface{}
		expected int // expected number of records
	}{
		{
			name: "single object",
			payload: map[string]interface{}{
				"field1": "value1",
				"field2": 123,
			},
			expected: 1,
		},
		{
			name: "array of objects",
			payload: []map[string]interface{}{
				{
					"field1": "value1",
					"field2": 123,
				},
				{
					"field1": "value2",
					"field2": 456,
				},
			},
			expected: 2,
		},
		{
			name:     "empty payload",
			payload:  nil,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a channel to track processed messages
			processed := make(chan int)

			// Mock the gRPC connection
			runner.conn = &mockConnectClient{
				sendFunc: func(req *grpc.TaskServiceConnectRequest) error {
					// Verify the request
					assert.Equal(t, grpc.TaskServiceConnectCode_INSERT_DATA, req.Code)
					assert.Equal(t, runner.tid.Hex(), req.TaskId)

					// If payload was nil, we expect no data
					if tc.payload == nil {
						processed <- 0
						return nil
					}

					// Unmarshal the data to verify record count
					var records []map[string]interface{}
					err := json.Unmarshal(req.Data, &records)
					assert.NoError(t, err)
					processed <- len(records)
					return nil
				},
			}

			// Create test message
			testMsg := entity.IPCMessage{
				Type:    constants.IPCMessageData,
				Payload: tc.payload,
				IPC:     true,
			}

			// Convert message to JSON and write to pipe
			go func() {
				jsonData, err := json.Marshal(testMsg)
				assert.NoError(t, err)
				_, err = fmt.Fprintln(pw, string(jsonData))
				assert.NoError(t, err)
			}()

			// Wait for processing with timeout
			select {
			case recordCount := <-processed:
				assert.Equal(t, tc.expected, recordCount)
			case <-time.After(1 * time.Second):
				if tc.expected > 0 {
					t.Fatal("timeout waiting for IPC message to be processed")
				}
			}
		})
	}
}

// mockConnectClient is a mock implementation of the gRPC Connect client
type mockConnectClient struct {
	grpc.TaskService_ConnectClient
	sendFunc func(*grpc.TaskServiceConnectRequest) error
}

func (m *mockConnectClient) Send(req *grpc.TaskServiceConnectRequest) error {
	if m.sendFunc != nil {
		return m.sendFunc(req)
	}
	return nil
}

func TestRunner_HandleIPCInvalidData(t *testing.T) {
	// Setup test data
	runner := setupTest(t)

	// Create pipes for testing
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()
	runner.stdoutPipe = pr

	// Start IPC reader
	go runner.startIPCReader()

	// Test cases for invalid data
	testCases := []struct {
		name    string
		message string // Raw message to send
	}{
		{
			name:    "invalid json",
			message: "{ invalid json",
		},
		{
			name:    "non-ipc json",
			message: `{"type": "data", "payload": {"field": "value"}}`, // Missing IPC flag
		},
		{
			name:    "invalid payload type",
			message: `{"type": "data", "payload": "invalid", "ipc": true}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a channel to ensure no data is processed
			processed := make(chan struct{})

			// Mock the gRPC connection
			runner.conn = &mockConnectClient{
				sendFunc: func(req *grpc.TaskServiceConnectRequest) error {
					if req.Code == grpc.TaskServiceConnectCode_INSERT_DATA {
						// This should not be called for invalid data
						processed <- struct{}{}
					}
					return nil
				},
			}

			// Write test message to pipe
			go func() {
				_, err := fmt.Fprintln(pw, tc.message)
				assert.NoError(t, err)
			}()

			// Wait briefly to ensure no processing occurs
			select {
			case <-processed:
				t.Error("invalid message was processed")
			case <-time.After(1 * time.Second):
				// Success - no processing occurred
			}
		})
	}
}
