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

	"github.com/crawlab-team/crawlab/core/dependency"
	"github.com/crawlab-team/crawlab/core/fs"
	"github.com/hashicorp/go-multierror"

	"github.com/crawlab-team/crawlab/core/models/models"

	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/entity"
	client2 "github.com/crawlab-team/crawlab/core/grpc/client"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/client"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// newTaskRunner creates a new task runner instance with the specified task ID
// It initializes all necessary components and establishes required connections
func newTaskRunner(id primitive.ObjectID, svc *Service) (r *Runner, err error) {
	// validate options
	if id.IsZero() {
		err = fmt.Errorf("invalid task id: %s", id.Hex())
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
		Logger:           utils.NewLogger("TaskRunner"),
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

	// Initialize context and done channel
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.done = make(chan struct{})

	// initialize task runner
	if err := r.Init(); err != nil {
		r.Errorf("error initializing task runner: %v", err)
		errs.Errors = append(errs.Errors, err)
	}

	return r, errs.ErrorOrNil()
}

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
	conn grpc.TaskService_ConnectClient // gRPC stream connection for task service
	interfaces.Logger

	// log handling
	readerStdout *bufio.Reader // reader for process stdout
	readerStderr *bufio.Reader // reader for process stderr
	logBatchSize int           // number of log lines to batch before sending

	// IPC (Inter-Process Communication)
	stdinPipe  io.WriteCloser          // pipe for writing to child process
	stdoutPipe io.ReadCloser           // pipe for reading from child process
	ipcChan    chan entity.IPCMessage  // channel for sending IPC messages
	ipcHandler func(entity.IPCMessage) // callback for handling received IPC messages

	// goroutine management
	ctx    context.Context    // context for controlling goroutine lifecycle
	cancel context.CancelFunc // function to cancel the context
	done   chan struct{}      // channel to signal completion
	wg     sync.WaitGroup     // wait group for goroutine synchronization
}

// Init initializes the task runner by updating the task status and establishing gRPC connections
func (r *Runner) Init() (err error) {
	// wait for grpc client ready
	client2.GetGrpcClient().WaitForReady()

	// update task
	if err := r.updateTask("", nil); err != nil {
		return err
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
	r.Infof("task[%s] started", r.tid.Hex())

	// update task status (processing)
	if err := r.updateTask(constants.TaskStatusRunning, nil); err != nil {
		return err
	}

	// configure working directory
	r.configureCwd()

	// sync files worker nodes
	if !utils.IsMaster() {
		if err := r.syncFiles(); err != nil {
			return r.updateTask(constants.TaskStatusError, err)
		}
	}

	// install dependencies
	if err := r.installDependenciesIfAvailable(); err != nil {
		r.Warnf("error installing dependencies: %v", err)
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

	// start health check
	go r.startHealthCheck()

	// Start IPC reader
	go r.startIPCReader()

	// Start IPC handler
	go r.handleIPC()

	// Ensure cleanup when Run() exits
	defer func() {
		_ = r.conn.CloseSend() // Close gRPC connection
		r.cancel()             // Cancel context to stop all goroutines
		r.wg.Wait()            // Wait for all goroutines to finish
		close(r.done)          // Signal that everything is done
		close(r.ipcChan)       // Close IPC channel
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
	err = utils.KillProcess(r.cmd, force)
	if err != nil {
		r.Errorf("kill process error: %v", err)
		return err
	}
	r.Debugf("attempt to kill process[%d]", r.pid)

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

	// command
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
	r.cmd, err = utils.BuildCmd(cmdStr)
	if err != nil {
		r.Errorf("error building command: %v", err)
		return err
	}

	// set working directory
	r.cmd.Dir = r.cwd

	// Configure pipes for IPC and logs
	r.stdinPipe, err = r.cmd.StdinPipe()
	if err != nil {
		r.Errorf("error creating stdin pipe: %v", err)
		return err
	}

	// Add stdout pipe for IPC and logs
	r.stdoutPipe, err = r.cmd.StdoutPipe()
	if err != nil {
		r.Errorf("error creating stdout pipe: %v", err)
		return err
	}

	// Add stderr pipe for error logs
	stderrPipe, err := r.cmd.StderrPipe()
	if err != nil {
		r.Errorf("error creating stderr pipe: %v", err)
		return err
	}

	// Create buffered readers
	r.readerStdout = bufio.NewReader(r.stdoutPipe)
	r.readerStderr = bufio.NewReader(stderrPipe)

	// Initialize IPC channel
	r.ipcChan = make(chan entity.IPCMessage)

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
// - Python paths
// - Node.js paths
// - Go paths
// - Crawlab-specific variables
// - Global environment variables from the system
func (r *Runner) configureEnv() {
	// Start with the current environment
	env := os.Environ()

	// Create a map for easier manipulation and to avoid duplicates
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Handle PATH non-existence
	if _, exists := envMap["PATH"]; !exists {
		envMap["PATH"] = ""
	}

	// Configure Python path
	pyenvRoot := utils.GetPyenvPath()
	pyenvShimsPath := pyenvRoot + "/shims"
	pyenvBinPath := pyenvRoot + "/bin"
	envMap["PYENV_ROOT"] = pyenvRoot
	if !strings.Contains(envMap["PATH"], pyenvShimsPath) {
		envMap["PATH"] = pyenvShimsPath + ":" + envMap["PATH"]
		r.Debugf("added pyenv shims path to PATH: %s", pyenvShimsPath)
	}
	if !strings.Contains(envMap["PATH"], pyenvBinPath) {
		envMap["PATH"] = pyenvBinPath + ":" + envMap["PATH"]
		r.Debugf("added pyenv bin path to PATH: %s", pyenvBinPath)
	}

	// Configure Node.js path
	nodePath := utils.GetNodeModulesPath()
	nodeBinPath := utils.GetNodeBinPath()
	envMap["NODE_PATH"] = nodePath
	if !strings.Contains(envMap["PATH"], nodePath) {
		envMap["PATH"] = nodePath + ":" + envMap["PATH"]
		r.Debugf("added node modules path to PATH: %s", nodePath)
	}
	if !strings.Contains(envMap["PATH"], nodeBinPath) {
		envMap["PATH"] = nodeBinPath + ":" + envMap["PATH"]
		r.Debugf("added node bin path to PATH: %s", nodeBinPath)
	}

	// Configure Go path
	goPath := utils.GetGoPath()
	if goPath != "" {
		envMap["GOPATH"] = goPath
		r.Debugf("set GOPATH: %s", goPath)
	}

	// Crawlab-specific variables
	envMap["CRAWLAB_TASK_ID"] = r.tid.Hex()
	envMap["CRAWLAB_PARENT_PID"] = fmt.Sprint(os.Getpid())

	// Global environment variables
	envs, err := client.NewModelService[models.Environment]().GetMany(nil, nil)
	if err != nil {
		r.Errorf("error getting environment variables: %v", err)
		return
	}
	for _, env := range envs {
		envMap[env.Key] = env.Value
		r.Debugf("set environment variable: %s", env.Key)
	}

	// Convert the map back to the []string format for r.cmd.Env
	r.cmd.Env = make([]string, 0, len(envMap))
	for key, value := range envMap {
		r.cmd.Env = append(r.cmd.Env, key+"="+value)
	}

	r.Debugf("environment configuration completed with %d variables", len(r.cmd.Env))
}

func (r *Runner) createHttpRequest(method, path string) (*http.Response, error) {
	// Normalize path
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	// Construct master URL
	var id string
	if r.s.GitId.IsZero() {
		id = r.s.Id.Hex()
	} else {
		id = r.s.GitId.Hex()
	}
	url := fmt.Sprintf("%s/sync/%s/%s", utils.GetApiEndpoint(), id, path)

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
	r.Infof("starting file synchronization for spider: %s", r.s.Id.Hex())

	workingDir := ""
	if !r.s.GitId.IsZero() {
		workingDir = r.s.GitRootPath
		r.Debugf("using git root path: %s", workingDir)
	}

	// get file list from master
	r.Infof("fetching file list from master node")
	resp, err := r.createHttpRequest("GET", "/scan?path="+workingDir)
	if err != nil {
		r.Errorf("error getting file list from master: %v", err)
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Errorf("error reading response body: %v", err)
		return err
	}
	var masterFiles map[string]entity.FsFileInfo
	err = json.Unmarshal(body, &masterFiles)
	if err != nil {
		r.Errorf("error unmarshaling JSON for URL: %s", resp.Request.URL.String())
		r.Errorf("error details: %v", err)
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
			r.Errorf("error creating worker directory: %v", err)
			return err
		}
	}

	// get file list from worker
	workerFiles, err := utils.ScanDirectory(r.cwd)
	if err != nil {
		r.Errorf("error scanning worker directory: %v", err)
		return err
	}

	// delete files that are deleted on master node
	for path, workerFile := range workerFiles {
		if _, exists := masterFilesMap[path]; !exists {
			r.Infof("deleting file: %s", path)
			err := os.Remove(workerFile.FullPath)
			if err != nil {
				r.Errorf("error deleting file: %v", err)
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
					r.Infof("directory needs to be synchronized: %s", path)
					_err := os.MkdirAll(filepath.Join(r.cwd, path), masterFile.Mode)
					if _err != nil {
						r.Errorf("error creating directory: %v", _err)
						err = errors.Join(err, _err)
					}
				} else {
					r.Infof("file needs to be synchronized: %s", path)
					_err := r.downloadFile(path, filepath.Join(r.cwd, path), masterFile)
					if _err != nil {
						r.Errorf("error downloading file: %v", _err)
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

	r.Infof("file synchronization completed successfully")
	return err
}

// downloadFile downloads a file from the master node and saves it to the local file system
func (r *Runner) downloadFile(path string, filePath string, fileInfo *entity.FsFileInfo) error {
	r.Debugf("downloading file: %s -> %s", path, filePath)

	resp, err := r.createHttpRequest("GET", "/download?path="+path)
	if err != nil {
		r.Errorf("error getting file response: %v", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		r.Errorf("error downloading file: %s", resp.Status)
		return errors.New(resp.Status)
	}
	defer resp.Body.Close()

	// create directory if not exists
	dirPath := filepath.Dir(filePath)
	utils.Exists(dirPath)
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		r.Errorf("error creating directory: %v", err)
		return err
	}

	// create local file
	out, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode)
	if err != nil {
		r.Errorf("error creating file: %v", err)
		return err
	}
	defer out.Close()

	// copy file content to local file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		r.Errorf("error copying file: %v", err)
		return err
	}

	r.Debugf("successfully downloaded file: %s (size: %d bytes)", path, fileInfo.FileSize)
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
		r.Debugf("waiting for process[%d] to finish", r.pid)
		err = r.cmd.Wait()
		if err != nil {
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) {
				r.ch <- constants.TaskSignalError
				r.Debugf("process[%d] exited with error: %v", r.pid, err)
				return
			}
			exitCode := exitError.ExitCode()
			if exitCode == -1 {
				// cancel error
				r.ch <- constants.TaskSignalCancel
				r.Debugf("process[%d] cancelled", r.pid)
				return
			}

			// standard error
			r.err = err
			r.ch <- constants.TaskSignalError
			r.Debugf("process[%d] exited with error: %v", r.pid, err)
			return
		}

		// success
		r.ch <- constants.TaskSignalFinish
		r.Debugf("process[%d] exited successfully", r.pid)
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
		r.Errorf("error updating task status: %v", err)
		return err
	}

	// log according to status
	switch status {
	case constants.TaskStatusFinished:
		r.Infof("task[%s] finished", r.tid.Hex())
	case constants.TaskStatusCancelled:
		r.Infof("task[%s] cancelled", r.tid.Hex())
	case constants.TaskStatusError:
		r.Errorf("task[%s] error: %v", r.tid.Hex(), err)
	default:
		r.Errorf("invalid task status: %s", status)
	}

	return nil
}

// updateTask updates the task status and related statistics in the database
// If running on a worker node, updates are sent to the master
func (r *Runner) updateTask(status string, e error) (err error) {
	if status != "" {
		r.Debugf("updating task status to: %s", status)
	}

	if r.t != nil && status != "" {
		// update task status
		r.t.Status = status
		if e != nil {
			r.t.Error = e.Error()
		}
		if utils.IsMaster() {
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
	r.Debugf("fetching updated task from database")
	r.t, err = r.svc.GetTaskById(r.tid)
	if err != nil {
		r.Errorf("failed to get updated task: %v", err)
		return err
	}

	return nil
}

// initConnection establishes a gRPC connection to the task service
func (r *Runner) initConnection() (err error) {
	r.conn, err = client2.GetGrpcClient().TaskClient.Connect(context.Background())
	if err != nil {
		r.Errorf("error connecting to task service: %v", err)
		return err
	}
	return nil
}

// writeLogLines marshals log lines to JSON and sends them to the task service
func (r *Runner) writeLogLines(lines []string) {
	linesBytes, err := json.Marshal(lines)
	if err != nil {
		r.Errorf("error marshaling log lines: %v", err)
		return
	}
	msg := &grpc.TaskServiceConnectRequest{
		Code:   grpc.TaskServiceConnectCode_INSERT_LOGS,
		TaskId: r.tid.Hex(),
		Data:   linesBytes,
	}
	if err := r.conn.Send(msg); err != nil {
		r.Errorf("error sending log lines: %v", err)
		return
	}
}

// _updateTaskStat updates task statistics based on the current status:
// - For running tasks: sets start time and wait duration
// - For completed tasks: sets end time and calculates durations
func (r *Runner) _updateTaskStat(status string) {
	if status != "" {
		r.Debugf("updating task statistics for status: %s", status)
	}

	ts, err := client.NewModelService[models.TaskStat]().GetById(r.tid)
	if err != nil {
		r.Errorf("error getting task stat: %v", err)
		return
	}

	r.Debugf("current task statistics - wait_duration: %dms, runtime_duration: %dms", ts.WaitDuration, ts.RuntimeDuration)

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
	if utils.IsMaster() {
		err = service.NewModelService[models.TaskStat]().ReplaceById(ts.Id, *ts)
		if err != nil {
			r.Errorf("error updating task stat: %v", err)
			return
		}
	} else {
		err = client.NewModelService[models.TaskStat]().ReplaceById(ts.Id, *ts)
		if err != nil {
			r.Errorf("error updating task stat: %v", err)
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
	_, err := client2.GetGrpcClient().TaskClient.SendNotification(context.Background(), req)
	if err != nil {
		r.Errorf("error sending notification: %v", err)
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
		r.Errorf("error getting task stat: %v", err)
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
		r.Errorf("Invalid task status: %s", status)
		return
	}

	// perform update
	if utils.IsMaster() {
		err = service.NewModelService[models.SpiderStat]().UpdateById(r.s.Id, update)
		if err != nil {
			r.Errorf("error updating spider stat: %v", err)
			return
		}
	} else {
		err = client.NewModelService[models.SpiderStat]().UpdateById(r.s.Id, update)
		if err != nil {
			r.Errorf("error updating spider stat: %v", err)
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
			r.Errorf("error marshaling IPC message: %v", err)
			continue
		}

		// Write to child process's stdin
		_, err = fmt.Fprintln(r.stdinPipe, string(jsonData))
		if err != nil {
			r.Errorf("error writing to child process: %v", err)
		}
	}
}

// SetIPCHandler sets the handler for incoming IPC messages
func (r *Runner) SetIPCHandler(handler func(entity.IPCMessage)) {
	r.ipcHandler = handler
}

// startIPCReader continuously reads IPC messages from the child process's stdout
// Messages are parsed and either handled by the IPC handler or written to logs
func (r *Runner) startIPCReader() {
	r.wg.Add(2) // Add 2 to wait group for both stdout and stderr readers

	// Start stdout reader
	go func() {
		defer r.wg.Done()
		r.readOutput(r.readerStdout, true) // true for stdout
	}()

	// Start stderr reader
	go func() {
		defer r.wg.Done()
		r.readOutput(r.readerStderr, false) // false for stderr
	}()
}

func (r *Runner) readOutput(reader *bufio.Reader, isStdout bool) {
	scanner := bufio.NewScanner(reader)
	for {
		select {
		case <-r.ctx.Done():
			// Context cancelled, stop reading
			return
		default:
			// Scan the next line
			if !scanner.Scan() {
				return
			}

			// Get the line
			line := scanner.Text()

			// Trim the line
			line = strings.TrimRight(line, "\n\r")

			// For stdout, try to parse as IPC message first
			if isStdout {
				var ipcMsg entity.IPCMessage
				if err := json.Unmarshal([]byte(line), &ipcMsg); err == nil && ipcMsg.IPC {
					if r.ipcHandler != nil {
						r.ipcHandler(ipcMsg)
					} else {
						// Default handler (insert data)
						if ipcMsg.Type == "" || ipcMsg.Type == constants.IPCMessageData {
							r.handleIPCInsertDataMessage(ipcMsg)
						} else {
							r.Warnf("no IPC handler set for message: %v", ipcMsg)
						}
					}
					continue
				}
			}

			// If not an IPC message or from stderr, treat as log
			r.writeLogLines([]string{line})
		}
	}
}

// handleIPCInsertDataMessage converts the IPC message payload to JSON and sends it to the master node
func (r *Runner) handleIPCInsertDataMessage(ipcMsg entity.IPCMessage) {
	if ipcMsg.Payload == nil {
		r.Errorf("empty payload in IPC message")
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
				r.Errorf("invalid record at index %d: %v", i, item)
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
			r.Errorf("invalid payload type: %T", payload)
			return
		}
	default:
		r.Errorf("unsupported payload type: %T, value: %v", payload, ipcMsg.Payload)
		return
	}

	// Validate records
	if len(records) == 0 {
		r.Warnf("no valid records to insert")
		return
	}

	// Marshal data with error handling
	data, err := json.Marshal(records)
	if err != nil {
		r.Errorf("error marshaling records: %v", err)
		return
	}

	// Validate connection
	if r.conn == nil {
		r.Errorf("gRPC connection not initialized")
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
		r.Errorf("timeout sending IPC message")
		return
	default:
		if err := r.conn.Send(grpcMsg); err != nil {
			r.Errorf("error sending IPC message: %v", err)
			return
		}
	}
}

func (r *Runner) installDependenciesIfAvailable() (err error) {
	if !utils.IsPro() {
		return nil
	}

	// Get dependency installer service
	depSvc := dependency.GetDependencyInstallerRegistryService()
	if depSvc == nil {
		r.Warnf("dependency installer service not available")
		return nil
	}

	// Check if auto install is enabled
	if !depSvc.IsAutoInstallEnabled() {
		r.Debug("auto dependency installation is disabled")
		return nil
	}

	// Get install command
	cmd, err := depSvc.GetInstallDependencyRequirementsCmdBySpiderId(r.s.Id)
	if err != nil {
		return err
	}
	if cmd == nil {
		return nil
	}

	// Set up pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.Errorf("error creating stdout pipe for dependency installation: %v", err)
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		r.Errorf("error creating stderr pipe for dependency installation: %v", err)
		return err
	}

	// Start the command
	r.Infof("installing dependencies for spider: %s", r.s.Id.Hex())
	r.Infof("command for dependencies installation: %s", cmd.String())
	if err := cmd.Start(); err != nil {
		r.Errorf("error starting dependency installation command: %v", err)
		return err
	}

	// Create wait group for log readers
	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			r.Info(line)
		}
	}()

	// Read stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			r.Error(line)
		}
	}()

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		r.Errorf("dependency installation failed: %v", err)
		return err
	}

	// Wait for log readers to finish
	wg.Wait()

	return nil
}

// logInternally sends internal runner logs to the same logging system as the task
func (r *Runner) logInternally(level string, message string) {
	// Format the internal log with a prefix
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")

	// Pad level
	level = fmt.Sprintf("%-5s", level)

	// Format the log message
	internalLog := fmt.Sprintf("%s [%s] [%s] %s", level, timestamp, "Crawlab", message)

	// Send to the same log system as task logs
	if r.conn != nil {
		r.writeLogLines([]string{internalLog})
	}

	// Also log through the standard logger
	switch level {
	case "ERROR":
		r.Logger.Error(message)
	case "WARN":
		r.Logger.Warn(message)
	case "INFO":
		r.Logger.Info(message)
	case "DEBUG":
		r.Logger.Debug(message)
	}
}

func (r *Runner) Error(message string) {
	msg := fmt.Sprintf(message)
	r.logInternally("ERROR", msg)
}

func (r *Runner) Warn(message string) {
	msg := fmt.Sprintf(message)
	r.logInternally("WARN", msg)
}

func (r *Runner) Info(message string) {
	msg := fmt.Sprintf(message)
	r.logInternally("INFO", msg)
}

func (r *Runner) Debug(message string) {
	msg := fmt.Sprintf(message)
	r.logInternally("DEBUG", msg)
}

func (r *Runner) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.logInternally("ERROR", msg)
}

func (r *Runner) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.logInternally("WARN", msg)
}

func (r *Runner) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.logInternally("INFO", msg)
}

func (r *Runner) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.logInternally("DEBUG", msg)
}
