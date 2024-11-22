package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/crawlab-team/crawlab/core/fs"
	"github.com/hashicorp/go-multierror"

	"github.com/crawlab-team/crawlab/core/models/models"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/entity"
	client2 "github.com/crawlab-team/crawlab/core/grpc/client"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/client"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/sys_exec"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
	"github.com/crawlab-team/crawlab/trace"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Runner represents a task execution handler that manages the lifecycle of a running task
type Runner struct {
	// dependencies
	svc   *Service             // task handler service
	fsSvc interfaces.FsService // task fs service

	// settings
	subscribeTimeout time.Duration // maximum time to wait for task subscription
	bufferSize       int           // buffer size for reading process output

	// internals
	cmd  *exec.Cmd                      // process command instance
	pid  int                            // process id
	tid  primitive.ObjectID             // task id
	t    *models.Task                   // task model instance
	s    *models.Spider                 // spider model instance
	ch   chan constants.TaskSignal      // channel for task status communication
	err  error                          // captures any process execution errors
	cwd  string                         // current working directory for task
	c    *client2.GrpcClient            // gRPC client for communication
	conn grpc.TaskService_ConnectClient // gRPC stream connection for task service

	// log handling
	scannerStdout *bufio.Reader // reader for process stdout
	scannerStderr *bufio.Reader // reader for process stderr
	logBatchSize  int           // number of log lines to batch before sending

	// IPC (Inter-Process Communication)
	stdinPipe  io.WriteCloser   // pipe for writing to child process
	stdoutPipe io.ReadCloser    // pipe for reading from child process
	ipcChan    chan IPCMessage  // channel for sending IPC messages
	ipcHandler func(IPCMessage) // callback for handling received IPC messages

	// goroutine management
	ctx    context.Context    // context for controlling goroutine lifecycle
	cancel context.CancelFunc // function to cancel the context
	done   chan struct{}      // channel to signal completion
	wg     sync.WaitGroup     // wait group for goroutine synchronization
}

const (
	IPCMessageData = "data" // IPCMessageData is the message type identifier for data messages
	IPCMessageLog  = "log"  // IPCMessageLog is the message type identifier for log messages
)

// IPCMessage defines the structure for messages exchanged between parent and child processes
type IPCMessage struct {
	Type    string      `json:"type"`    // message type identifier
	Payload interface{} `json:"payload"` // message content
	IPC     bool        `json:"ipc"`     // Add this field to explicitly mark IPC messages
}

// Init initializes the task runner by updating the task status and establishing gRPC connections
func (r *Runner) Init() (err error) {
	// update task
	if err := r.updateTask("", nil); err != nil {
		return err
	}

	// start grpc client
	if !r.c.IsStarted() {
		err := r.c.Start()
		if err != nil {
			return err
		}
	}

	// grpc task service stream client
	if err := r.initConnection(); err != nil {
		return err
	}

	return nil
}

// Run executes the task and manages its lifecycle, including file synchronization, process execution,
// and status monitoring. Returns an error if the task execution fails.
func (r *Runner) Run() (err error) {
	// log task started
	log.Infof("task[%s] started", r.tid.Hex())

	// configure working directory
	r.configureCwd()

	// sync files worker nodes
	if !utils.IsMaster() {
		if err := r.syncFiles(); err != nil {
			return r.updateTask(constants.TaskStatusError, err)
		}
	}

	// configure cmd
	err = r.configureCmd()
	if err != nil {
		return r.updateTask(constants.TaskStatusError, err)
	}

	// configure environment variables
	r.configureEnv()

	// start process
	if err := r.cmd.Start(); err != nil {
		return r.updateTask(constants.TaskStatusError, err)
	}

	// process id
	if r.cmd.Process == nil {
		return r.updateTask(constants.TaskStatusError, constants.ErrNotExists)
	}
	r.pid = r.cmd.Process.Pid
	r.t.Pid = r.pid

	// update task status (processing)
	if err := r.updateTask(constants.TaskStatusRunning, nil); err != nil {
		return err
	}

	// start health check
	go r.startHealthCheck()

	// Start IPC reader
	go r.startIPCReader()

	// Start IPC handler
	go r.handleIPC()

	// Ensure cleanup when Run() exits
	defer func() {
		r.cancel()       // Cancel context to stop all goroutines
		r.wg.Wait()      // Wait for all goroutines to finish
		close(r.done)    // Signal that everything is done
		close(r.ipcChan) // Close IPC channel
	}()

	// wait for process to finish
	return r.wait()
}

// Cancel terminates the running task. If force is true, the process will be killed immediately
// without waiting for graceful shutdown.
func (r *Runner) Cancel(force bool) (err error) {
	// Signal goroutines to stop
	r.cancel()

	// Kill process
	err = sys_exec.KillProcess(r.cmd, &sys_exec.KillProcessOptions{
		Force: force,
	})
	if err != nil {
		log.Errorf("kill process error: %v", err)
		return err
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), r.svc.GetCancelTimeout())
	defer cancel()

	// Wait for process to be killed with context
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for task to stop")
		case <-ticker.C:
			if !utils.ProcessIdExists(r.pid) {
				return nil
			}
		}
	}
}

func (r *Runner) SetSubscribeTimeout(timeout time.Duration) {
	r.subscribeTimeout = timeout
}

func (r *Runner) GetTaskId() (id primitive.ObjectID) {
	return r.tid
}

// configureCmd builds and configures the command to be executed, including setting up IPC pipes
// and processing command parameters
func (r *Runner) configureCmd() (err error) {
	var cmdStr string

	// customized spider
	if r.t.Cmd == "" {
		cmdStr = r.s.Cmd
	} else {
		cmdStr = r.t.Cmd
	}

	// parameters
	if r.t.Param != "" {
		cmdStr += " " + r.t.Param
	} else if r.s.Param != "" {
		cmdStr += " " + r.s.Param
	}

	// get cmd instance
	r.cmd, err = sys_exec.BuildCmd(cmdStr)
	if err != nil {
		log.Errorf("error building command: %v", err)
		return err
	}

	// set working directory
	r.cmd.Dir = r.cwd

	// Configure pipes for IPC
	r.stdinPipe, err = r.cmd.StdinPipe()
	if err != nil {
		log.Errorf("error creating stdin pipe: %v", err)
		return err
	}

	// Add stdout pipe for IPC
	r.stdoutPipe, err = r.cmd.StdoutPipe()
	if err != nil {
		log.Errorf("error creating stdout pipe: %v", err)
		return err
	}

	// Initialize IPC channel
	r.ipcChan = make(chan IPCMessage)

	return nil
}

// startHealthCheck periodically verifies that the process is still running
// If the process disappears unexpectedly, it signals a task lost condition
func (r *Runner) startHealthCheck() {
	r.wg.Add(1)
	defer r.wg.Done()

	if r.cmd.ProcessState == nil || r.cmd.ProcessState.Exited() {
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			if !utils.ProcessIdExists(r.pid) {
				// process lost
				r.ch <- constants.TaskSignalLost
				return
			}
		}
	}
}

// configureEnv sets up the environment variables for the task process, including:
// - Node.js paths
// - Crawlab-specific variables
// - Global environment variables from the system
func (r *Runner) configureEnv() {
	// By default, add Node.js's global node_modules to PATH
	envPath := os.Getenv("PATH")
	nodePath := "/usr/lib/node_modules"
	if !strings.Contains(envPath, nodePath) {
		_ = os.Setenv("PATH", nodePath+":"+envPath)
	}
	_ = os.Setenv("NODE_PATH", nodePath)

	// Default envs
	r.cmd.Env = os.Environ()
	r.cmd.Env = append(r.cmd.Env, "CRAWLAB_TASK_ID="+r.tid.Hex())
	r.cmd.Env = append(r.cmd.Env, "CRAWLAB_GRPC_ADDRESS="+utils.GetGrpcAddress())
	r.cmd.Env = append(r.cmd.Env, "CRAWLAB_GRPC_AUTH_KEY="+utils.GetAuthKey())
	r.cmd.Env = append(r.cmd.Env, "PYENV_ROOT="+utils.PyenvRoot)
	r.cmd.Env = append(r.cmd.Env, "PATH="+os.Getenv("PATH")+":"+utils.PyenvRoot+"/shims:"+utils.PyenvRoot+"/bin")

	// Global environment variables
	envs, err := client.NewModelService[models.Environment]().GetMany(nil, nil)
	if err != nil {
		trace.PrintError(err)
		return
	}
	for _, env := range envs {
		r.cmd.Env = append(r.cmd.Env, env.Key+"="+env.Value)
	}

	// Add environment variable for child processes to identify they're running under Crawlab
	r.cmd.Env = append(r.cmd.Env, "CRAWLAB_PARENT_PID="+fmt.Sprint(os.Getpid()))
}

func (r *Runner) createHttpRequest(method, path string) (*http.Response, error) {
	// Construct master URL
	var id string
	if r.s.GitId.IsZero() {
		id = r.s.Id.Hex()
	} else {
		id = r.s.GitId.Hex()
	}
	url := fmt.Sprintf("%s/sync/%s%s", utils.GetApiEndpoint(), id, path)

	// Create and execute request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for k, v := range r.getHttpRequestHeaders() {
		req.Header.Set(k, v)
	}

	return http.DefaultClient.Do(req)
}

// syncFiles synchronizes files between master and worker nodes:
// 1. Gets file list from master
// 2. Compares with local files
// 3. Downloads new/modified files
// 4. Deletes files that no longer exist on master
func (r *Runner) syncFiles() (err error) {
	workingDir := ""
	if !r.s.GitId.IsZero() {
		workingDir = r.s.GitRootPath
	}

	// get file list from master
	resp, err := r.createHttpRequest("GET", "/scan?path="+workingDir)
	if err != nil {
		log.Errorf("error getting file list from master: %v", err)
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error reading response body: %v", err)
		return err
	}
	var masterFiles map[string]entity.FsFileInfo
	err = json.Unmarshal(body, &masterFiles)
	if err != nil {
		log.Errorf("error unmarshaling JSON: %v", err)
		return err
	}

	// create a map for master files
	masterFilesMap := make(map[string]entity.FsFileInfo)
	for _, file := range masterFiles {
		masterFilesMap[file.Path] = file
	}

	// create working directory if not exists
	if _, err := os.Stat(r.cwd); os.IsNotExist(err) {
		if err := os.MkdirAll(r.cwd, os.ModePerm); err != nil {
			log.Errorf("error creating worker directory: %v", err)
			return err
		}
	}

	// get file list from worker
	workerFiles, err := utils.ScanDirectory(r.cwd)
	if err != nil {
		log.Errorf("error scanning worker directory: %v", err)
		return trace.TraceError(err)
	}

	// delete files that are deleted on master node
	for path, workerFile := range workerFiles {
		if _, exists := masterFilesMap[path]; !exists {
			log.Infof("deleting file: %s", path)
			err := os.Remove(workerFile.FullPath)
			if err != nil {
				log.Errorf("error deleting file: %v", err)
			}
		}
	}

	// set up wait group and error channel
	var wg sync.WaitGroup
	pool := make(chan struct{}, 10)

	// download files that are new or modified on master node
	for path, masterFile := range masterFilesMap {
		workerFile, exists := workerFiles[path]
		if !exists || masterFile.Hash != workerFile.Hash {
			wg.Add(1)

			// acquire token
			pool <- struct{}{}

			// start goroutine to synchronize file or directory
			go func(path string, masterFile *entity.FsFileInfo) {
				defer wg.Done()

				if masterFile.IsDir {
					log.Infof("directory needs to be synchronized: %s", path)
					_err := os.MkdirAll(filepath.Join(r.cwd, path), masterFile.Mode)
					if _err != nil {
						log.Errorf("error creating directory: %v", _err)
						err = errors.Join(err, _err)
					}
				} else {
					log.Infof("file needs to be synchronized: %s", path)
					_err := r.downloadFile(path, filepath.Join(r.cwd, path), masterFile)
					if _err != nil {
						log.Errorf("error downloading file: %v", _err)
						err = errors.Join(err, _err)
					}
				}

				// release token
				<-pool

			}(path, &masterFile)
		}
	}

	// wait for all files and directories to be synchronized
	wg.Wait()

	return err
}

// downloadFile downloads a file from the master node and saves it to the local file system
func (r *Runner) downloadFile(path string, filePath string, fileInfo *entity.FsFileInfo) error {
	resp, err := r.createHttpRequest("GET", "/download?path="+path)
	if err != nil {
		log.Errorf("error getting file response: %v", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("error downloading file: %s", resp.Status)
		return errors.New(resp.Status)
	}
	defer resp.Body.Close()

	// create directory if not exists
	dirPath := filepath.Dir(filePath)
	utils.Exists(dirPath)
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.Errorf("error creating directory: %v", err)
		return err
	}

	// create local file
	out, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode)
	if err != nil {
		log.Errorf("error creating file: %v", err)
		return err
	}
	defer out.Close()

	// copy file content to local file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Errorf("error copying file: %v", err)
		return err
	}
	return nil
}

// getHttpRequestHeaders returns the headers for HTTP requests to the master node
func (r *Runner) getHttpRequestHeaders() (headers map[string]string) {
	headers = make(map[string]string)
	headers["Authorization"] = utils.GetAuthKey()
	return headers
}

// wait monitors the process execution and sends appropriate signals based on the exit status:
// - TaskSignalFinish for successful completion
// - TaskSignalCancel for cancellation
// - TaskSignalError for execution errors
func (r *Runner) wait() (err error) {
	// start a goroutine to wait for process to finish
	go func() {
		err = r.cmd.Wait()
		if err != nil {
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) {
				r.ch <- constants.TaskSignalError
				return
			}
			exitCode := exitError.ExitCode()
			if exitCode == -1 {
				// cancel error
				r.ch <- constants.TaskSignalCancel
				return
			}

			// standard error
			r.err = err
			r.ch <- constants.TaskSignalError
			return
		}

		// success
		r.ch <- constants.TaskSignalFinish
	}()

	// declare task status
	status := ""

	// wait for signal
	signal := <-r.ch
	switch signal {
	case constants.TaskSignalFinish:
		err = nil
		status = constants.TaskStatusFinished
	case constants.TaskSignalCancel:
		err = constants.ErrTaskCancelled
		status = constants.TaskStatusCancelled
	case constants.TaskSignalError:
		err = r.err
		status = constants.TaskStatusError
	case constants.TaskSignalLost:
		err = constants.ErrTaskLost
		status = constants.TaskStatusError
	default:
		err = constants.ErrInvalidSignal
		status = constants.TaskStatusError
	}

	// update task status
	if err := r.updateTask(status, err); err != nil {
		log.Errorf("error updating task status: %v", err)
		return err
	}

	return nil
}

// updateTask updates the task status and related statistics in the database
// If running on a worker node, updates are sent to the master
func (r *Runner) updateTask(status string, e error) (err error) {
	if r.t != nil && status != "" {
		// update task status
		r.t.Status = status
		if e != nil {
			r.t.Error = e.Error()
		}
		if r.svc.GetNodeConfigService().IsMaster() {
			err = service.NewModelService[models.Task]().ReplaceById(r.t.Id, *r.t)
			if err != nil {
				return err
			}
		} else {
			err = client.NewModelService[models.Task]().ReplaceById(r.t.Id, *r.t)
			if err != nil {
				return err
			}
		}

		// update stats
		r._updateTaskStat(status)
		r._updateSpiderStat(status)

		// send notification
		go r.sendNotification()
	}

	// get task
	r.t, err = r.svc.GetTaskById(r.tid)
	if err != nil {
		return err
	}

	return nil
}

// initConnection establishes a gRPC connection to the task service
func (r *Runner) initConnection() (err error) {
	r.conn, err = r.c.TaskClient.Connect(context.Background())
	if err != nil {
		log.Errorf("error connecting to task service: %v", err)
		return err
	}
	return nil
}

// writeLogLines marshals log lines to JSON and sends them to the task service
func (r *Runner) writeLogLines(lines []string) {
	linesBytes, err := json.Marshal(lines)
	if err != nil {
		log.Errorf("error marshaling log lines: %v", err)
		return
	}
	msg := &grpc.TaskServiceConnectRequest{
		Code:   grpc.TaskServiceConnectCode_INSERT_LOGS,
		TaskId: r.tid.Hex(),
		Data:   linesBytes,
	}
	if err := r.conn.Send(msg); err != nil {
		log.Errorf("error sending log lines: %v", err)
		return
	}
}

// _updateTaskStat updates task statistics based on the current status:
// - For running tasks: sets start time and wait duration
// - For completed tasks: sets end time and calculates durations
func (r *Runner) _updateTaskStat(status string) {
	ts, err := client.NewModelService[models.TaskStat]().GetById(r.tid)
	if err != nil {
		trace.PrintError(err)
		return
	}
	switch status {
	case constants.TaskStatusPending:
		// do nothing
	case constants.TaskStatusRunning:
		ts.StartTs = time.Now()
		ts.WaitDuration = ts.StartTs.Sub(ts.CreatedAt).Milliseconds()
	case constants.TaskStatusFinished, constants.TaskStatusError, constants.TaskStatusCancelled:
		if ts.StartTs.IsZero() {
			ts.StartTs = time.Now()
			ts.WaitDuration = ts.StartTs.Sub(ts.CreatedAt).Milliseconds()
		}
		ts.EndTs = time.Now()
		ts.RuntimeDuration = ts.EndTs.Sub(ts.StartTs).Milliseconds()
		ts.TotalDuration = ts.EndTs.Sub(ts.CreatedAt).Milliseconds()
	}
	if r.svc.GetNodeConfigService().IsMaster() {
		err = service.NewModelService[models.TaskStat]().ReplaceById(ts.Id, *ts)
		if err != nil {
			trace.PrintError(err)
			return
		}
	} else {
		err = client.NewModelService[models.TaskStat]().ReplaceById(ts.Id, *ts)
		if err != nil {
			trace.PrintError(err)
			return
		}
	}
}

// sendNotification sends a notification to the task service
func (r *Runner) sendNotification() {
	req := &grpc.TaskServiceSendNotificationRequest{
		NodeKey: r.svc.GetNodeConfigService().GetNodeKey(),
		TaskId:  r.tid.Hex(),
	}
	_, err := r.c.TaskClient.SendNotification(context.Background(), req)
	if err != nil {
		log.Errorf("error sending notification: %v", err)
		trace.PrintError(err)
		return
	}
}

// _updateSpiderStat updates spider statistics based on task completion:
// - Updates last task ID
// - Increments task counts
// - Updates duration metrics
func (r *Runner) _updateSpiderStat(status string) {
	// task stat
	ts, err := client.NewModelService[models.TaskStat]().GetById(r.tid)
	if err != nil {
		trace.PrintError(err)
		return
	}

	// update
	var update bson.M
	switch status {
	case constants.TaskStatusPending, constants.TaskStatusRunning:
		update = bson.M{
			"$set": bson.M{
				"last_task_id": r.tid, // last task id
			},
			"$inc": bson.M{
				"tasks":         1,               // task count
				"wait_duration": ts.WaitDuration, // wait duration
			},
		}
	case constants.TaskStatusFinished, constants.TaskStatusError, constants.TaskStatusCancelled:
		update = bson.M{
			"$set": bson.M{
				"last_task_id": r.tid, // last task id
			},
			"$inc": bson.M{
				"results":          ts.ResultCount,            // results
				"runtime_duration": ts.RuntimeDuration / 1000, // runtime duration
				"total_duration":   ts.TotalDuration / 1000,   // total duration
			},
		}
	default:
		log.Errorf("Invalid task status: %s", status)
		trace.PrintError(errors.New("invalid task status"))
		return
	}

	// perform update
	if r.svc.GetNodeConfigService().IsMaster() {
		err = service.NewModelService[models.SpiderStat]().UpdateById(r.s.Id, update)
		if err != nil {
			trace.PrintError(err)
			return
		}
	} else {
		err = client.NewModelService[models.SpiderStat]().UpdateById(r.s.Id, update)
		if err != nil {
			trace.PrintError(err)
			return
		}
	}
}

// configureCwd sets the working directory for the task based on the spider's configuration
func (r *Runner) configureCwd() {
	workspacePath := utils.GetWorkspace()
	if r.s.GitId.IsZero() {
		// not git
		r.cwd = filepath.Join(workspacePath, r.s.Id.Hex())
	} else {
		// git
		r.cwd = filepath.Join(workspacePath, r.s.GitId.Hex(), r.s.GitRootPath)
	}
}

// handleIPC processes incoming IPC messages from the child process
// Messages are converted to JSON and written to the child process's stdin
func (r *Runner) handleIPC() {
	for msg := range r.ipcChan {
		// Convert message to JSON
		jsonData, err := json.Marshal(msg)
		if err != nil {
			log.Errorf("error marshaling IPC message: %v", err)
			continue
		}

		// Write to child process's stdin
		_, err = fmt.Fprintln(r.stdinPipe, string(jsonData))
		if err != nil {
			log.Errorf("error writing to child process: %v", err)
		}
	}
}

// SendToChild sends a message to the child process through the IPC channel
// msgType: type of message being sent
// payload: data to be sent to the child process
func (r *Runner) SendToChild(msgType string, payload interface{}) {
	r.ipcChan <- IPCMessage{
		Type:    msgType,
		Payload: payload,
		IPC:     true, // Explicitly mark as IPC message
	}
}

// SetIPCHandler sets the handler for incoming IPC messages
func (r *Runner) SetIPCHandler(handler func(IPCMessage)) {
	r.ipcHandler = handler
}

// startIPCReader continuously reads IPC messages from the child process's stdout
// Messages are parsed and either handled by the IPC handler or written to logs
func (r *Runner) startIPCReader() {
	r.wg.Add(1)
	defer r.wg.Done()

	scanner := bufio.NewScanner(r.stdoutPipe)
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
			if !scanner.Scan() {
				return
			}
			line := scanner.Text()

			var ipcMsg IPCMessage
			err := json.Unmarshal([]byte(line), &ipcMsg)
			if err == nil && ipcMsg.IPC {
				// Only handle as IPC if it's valid JSON AND has IPC flag set
				if r.ipcHandler != nil {
					r.ipcHandler(ipcMsg)
				} else {
					// Default handler (insert data)
					if ipcMsg.Type == "" || ipcMsg.Type == IPCMessageData {
						r.handleIPCInsertDataMessage(ipcMsg)
					} else {
						log.Warnf("no IPC handler set for message: %v", ipcMsg)
					}
				}
			} else {
				// Everything else is treated as logs
				r.writeLogLines([]string{line})
			}
		}
	}
}

// handleIPCInsertDataMessage converts the IPC message payload to JSON and sends it to the master node
func (r *Runner) handleIPCInsertDataMessage(ipcMsg IPCMessage) {
	// Validate message
	if ipcMsg.Payload == nil {
		log.Errorf("empty payload in IPC message")
		return
	}

	// Convert payload to data to be inserted
	var records []map[string]interface{}

	switch payload := ipcMsg.Payload.(type) {
	case []interface{}: // Handle array of objects
		records = make([]map[string]interface{}, 0, len(payload))
		for i, item := range payload {
			if itemMap, ok := item.(map[string]interface{}); ok {
				records = append(records, itemMap)
			} else {
				log.Errorf("invalid record at index %d: %v", i, item)
				continue
			}
		}
	case []map[string]interface{}: // Handle direct array of maps
		records = payload
	case map[string]interface{}: // Handle single object
		records = []map[string]interface{}{payload}
	case interface{}: // Handle generic interface
		if itemMap, ok := payload.(map[string]interface{}); ok {
			records = []map[string]interface{}{itemMap}
		} else {
			log.Errorf("invalid payload type: %T", payload)
			return
		}
	default:
		log.Errorf("unsupported payload type: %T, value: %v", payload, ipcMsg.Payload)
		return
	}

	// Validate records
	if len(records) == 0 {
		log.Warnf("no valid records to insert")
		return
	}

	// Marshal data with error handling
	data, err := json.Marshal(records)
	if err != nil {
		log.Errorf("error marshaling records: %v", err)
		return
	}

	// Validate connection
	if r.conn == nil {
		log.Errorf("gRPC connection not initialized")
		return
	}

	// Send IPC message to master with context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create gRPC message
	grpcMsg := &grpc.TaskServiceConnectRequest{
		Code:   grpc.TaskServiceConnectCode_INSERT_DATA,
		TaskId: r.tid.Hex(),
		Data:   data,
	}

	// Use context for sending
	select {
	case <-ctx.Done():
		log.Errorf("timeout sending IPC message")
		return
	default:
		if err := r.conn.Send(grpcMsg); err != nil {
			log.Errorf("error sending IPC message: %v", err)
			return
		}
	}
}

// newTaskRunner creates a new task runner instance with the specified task ID
// It initializes all necessary components and establishes required connections
func newTaskRunner(id primitive.ObjectID, svc *Service) (r *Runner, err error) {
	// validate options
	if id.IsZero() {
		err = fmt.Errorf("invalid task id: %s", id.Hex())
		log.Errorf("error creating task runner: %v", err)
		return nil, err
	}

	// runner
	r = &Runner{
		subscribeTimeout: 30 * time.Second,
		bufferSize:       1024 * 1024,
		svc:              svc,
		tid:              id,
		ch:               make(chan constants.TaskSignal),
		logBatchSize:     20,
	}

	// multi error
	var errs multierror.Error

	// task
	r.t, err = svc.GetTaskById(id)
	if err != nil {
		errs.Errors = append(errs.Errors, err)
	} else {
		// spider
		r.s, err = svc.GetSpiderById(r.t.SpiderId)
		if err != nil {
			errs.Errors = append(errs.Errors, err)
		} else {
			// task fs service
			r.fsSvc = fs.NewFsService(filepath.Join(utils.GetWorkspace(), r.s.Id.Hex()))
		}
	}

	// grpc client
	r.c = client2.GetGrpcClient()

	// Initialize context and done channel
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.done = make(chan struct{})

	// initialize task runner
	if err := r.Init(); err != nil {
		log.Errorf("error initializing task runner: %v", err)
		errs.Errors = append(errs.Errors, err)
	}

	return r, errs.ErrorOrNil()
}
