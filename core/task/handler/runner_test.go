package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/grpc/client"
	"github.com/crawlab-team/crawlab/core/grpc/server"
	"github.com/crawlab-team/crawlab/core/utils"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/grpc"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupGrpc(t *testing.T) {
	// Mock IsMaster function by setting viper config
	viper.Set("node.master", true)
	defer viper.Set("node.master", nil) // cleanup after test

	// Start a gRPC server
	svr := server.GetGrpcServer()
	err := svr.Start()
	require.Nil(t, err)

	// Start a gRPC client
	client.GetGrpcClient().Start()
	require.Nil(t, err)

	// Cleanup
	t.Cleanup(func() {
		err = svr.Stop()
		if err != nil {
			log.Warnf("failed to stop gRPC server: %v", err)
		}
	})
}

func setupRunner(t *testing.T) *Runner {
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

func TestRunner(t *testing.T) {
	// Setup test data
	setupGrpc(t)

	t.Run("HandleIPC", func(t *testing.T) {
		// Create a runner
		runner := setupRunner(t)

		// Create a pipe for testing
		pr, pw := io.Pipe()
		defer func() {
			_ = pr.Close()
			log.Infof("closed reader pipe")
		}()
		defer func() {
			_ = pw.Close()
			log.Infof("closed writer pipe")
		}()
		runner.stdoutPipe = pr

		// Initialize context and other required fields
		runner.ctx, runner.cancel = context.WithCancel(context.Background())
		runner.wg = sync.WaitGroup{}
		runner.done = make(chan struct{})
		runner.ipcChan = make(chan entity.IPCMessage)

		// Create a channel to signal that the reader is ready
		readerReady := make(chan struct{})

		// Start IPC reader with ready signal
		go func() {
			defer runner.wg.Done()
			runner.wg.Add(1)
			close(readerReady) // Signal that reader is ready

			// Read directly from the pipe for debugging
			scanner := bufio.NewScanner(pr)
			for scanner.Scan() {
				line := scanner.Text()
				log.Infof("Read from pipe: %s", line)

				// Try to parse as IPC message
				var ipcMsg entity.IPCMessage
				if err := json.Unmarshal([]byte(line), &ipcMsg); err != nil {
					log.Errorf("Failed to unmarshal IPC message: %v", err)
					continue
				}

				if ipcMsg.IPC {
					log.Infof("Valid IPC message received: %+v", ipcMsg)
					if runner.ipcHandler != nil {
						runner.ipcHandler(ipcMsg)
					}
				}
			}

			if err := scanner.Err(); err != nil {
				log.Errorf("Scanner error: %v", err)
			}
		}()

		// Wait for reader to be ready
		<-readerReady

		// Create test message
		testMsg := entity.IPCMessage{
			Type:    "test_type",
			Payload: map[string]interface{}{"key": "value"},
			IPC:     true,
		}

		// Create channels for synchronization
		handled := make(chan bool)
		messageError := make(chan error, 1)

		// Set up message handler
		runner.SetIPCHandler(func(msg entity.IPCMessage) {
			log.Infof("Handler received IPC message: %+v", msg)
			if msg.Type != testMsg.Type {
				messageError <- fmt.Errorf("expected message type %s, got %s", testMsg.Type, msg.Type)
				return
			}
			if !reflect.DeepEqual(msg.Payload, testMsg.Payload) {
				messageError <- fmt.Errorf("expected payload %v, got %v", testMsg.Payload, msg.Payload)
				return
			}
			handled <- true
		})

		// Convert message to JSON
		jsonData, err := json.Marshal(testMsg)
		if err != nil {
			t.Fatalf("failed to marshal test message: %v", err)
		}

		// Write message to pipe
		log.Infof("Writing message to pipe: %s", string(jsonData))
		_, err = fmt.Fprintln(pw, string(jsonData))
		if err != nil {
			t.Fatalf("failed to write to pipe: %v", err)
		}
		log.Info("Message written to pipe")

		// Wait for message handling with timeout
		select {
		case <-handled:
			log.Info("IPC message was handled successfully")
		case err := <-messageError:
			t.Fatalf("error handling message: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for IPC message to be handled")
		}

		// Clean up
		runner.cancel() // Cancel context to stop readers
	})

	t.Run("Cancel", func(t *testing.T) {
		// Create a runner
		runner := setupRunner(t)

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
			require.Fail(t, fmt.Sprintf("Process with PID %d was not started successfully", runner.pid))
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
				require.Fail(t, fmt.Sprintf("Process with PID %d was not killed within timeout", runner.pid))
			case <-ticker.C:
				exists := utils.ProcessIdExists(runner.pid)
				if !exists {
					return // Exit the test when process is killed
				}
			}
		}
	})

	t.Run("HandleIPCData", func(t *testing.T) {
		// Create a runner
		runner := setupRunner(t)

		// Create pipes for testing
		pr, pw := io.Pipe()
		defer pr.Close()
		defer pw.Close()
		runner.stdoutPipe = pr

		// Create a channel to signal that the reader is ready
		readerReady := make(chan struct{})

		// Start IPC reader with ready signal
		go func() {
			close(readerReady) // Signal that reader is ready
			runner.startIPCReader()
		}()

		// Wait for reader to be ready
		<-readerReady

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
					jsonData, _ := json.Marshal(testMsg)
					_, _ = fmt.Fprintln(pw, string(jsonData))
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
	})

	t.Run("HandleIPCInvalidData", func(t *testing.T) {
		// Create a runner
		runner := setupRunner(t)

		// Create pipes for testing
		pr, pw := io.Pipe()
		defer pr.Close()
		defer pw.Close()
		runner.stdoutPipe = pr

		// Create a channel to signal that the reader is ready
		readerReady := make(chan struct{})

		// Start IPC reader with ready signal
		go func() {
			close(readerReady) // Signal that reader is ready
			runner.startIPCReader()
		}()

		// Wait for reader to be ready
		<-readerReady

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
					if err != nil {
						log.Errorf("failed to write to pipe: %v", err)
					}
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
	})
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
