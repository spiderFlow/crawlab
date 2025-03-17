package scheduler

import (
	errors2 "errors"
	"fmt"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/grpc/server"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/task/handler"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

type Service struct {
	// dependencies
	svr        *server.GrpcServer
	handlerSvc *handler.Service

	// settings
	interval time.Duration

	// internals
	interfaces.Logger
}

func (svc *Service) Start() {
	go svc.initTaskStatus()
	go svc.cleanupTasks()
	utils.DefaultWait()
}

func (svc *Service) Enqueue(t *models.Task, by primitive.ObjectID) (t2 *models.Task, err error) {
	// set task status
	t.Status = constants.TaskStatusPending
	t.SetCreated(by)
	t.SetUpdated(by)

	// add task
	taskModelSvc := service.NewModelService[models.Task]()
	t.Id, err = taskModelSvc.InsertOne(*t)
	if err != nil {
		return nil, err
	}

	// task stat
	ts := models.TaskStat{}
	ts.SetId(t.Id)
	ts.SetCreated(by)
	ts.SetUpdated(by)

	// add task stat
	_, err = service.NewModelService[models.TaskStat]().InsertOne(ts)
	if err != nil {
		svc.Errorf("failed to add task stat: %s", t.Id.Hex())
		return nil, err
	}

	// success
	return t, nil
}

func (svc *Service) Cancel(id, by primitive.ObjectID, force bool) (err error) {
	// task
	t, err := service.NewModelService[models.Task]().GetById(id)
	if err != nil {
		svc.Errorf("task not found: %s", id.Hex())
		return err
	}

	// initial status
	initialStatus := t.Status

	// set status of pending tasks as "cancelled"
	if initialStatus == constants.TaskStatusPending {
		t.Status = constants.TaskStatusCancelled
		return svc.SaveTask(t, by)
	}

	// whether task is running on master node
	isMasterTask, err := svc.isMasterNode(t)
	if err != nil {
		err := fmt.Errorf("failed to check if task is running on master node: %s", t.Id.Hex())
		t.Status = constants.TaskStatusAbnormal
		t.Error = err.Error()
		return svc.SaveTask(t, by)
	}

	if isMasterTask {
		// cancel task on master
		return svc.cancelOnMaster(t, by, force)
	} else {
		// send to cancel task on worker nodes
		return svc.cancelOnWorker(t, by, force)
	}
}

func (svc *Service) cancelOnMaster(t *models.Task, by primitive.ObjectID, force bool) (err error) {
	if err := svc.handlerSvc.Cancel(t.Id, force); err != nil {
		svc.Errorf("failed to cancel task on master: %s", t.Id.Hex())
		return err
	}

	// set task status as "cancelled"
	t.Status = constants.TaskStatusCancelled
	return svc.SaveTask(t, by)
}

func (svc *Service) cancelOnWorker(t *models.Task, by primitive.ObjectID, force bool) (err error) {
	// get subscribe stream
	stream, ok := svc.svr.TaskSvr.GetSubscribeStream(t.Id)
	if !ok {
		err := fmt.Errorf("stream not found for task: %s", t.Id.Hex())
		svc.Errorf(err.Error())
		t.Status = constants.TaskStatusAbnormal
		t.Error = err.Error()
		return svc.SaveTask(t, by)
	}

	// send cancel request
	err = stream.Send(&grpc.TaskServiceSubscribeResponse{
		Code:   grpc.TaskServiceSubscribeCode_CANCEL,
		TaskId: t.Id.Hex(),
		Force:  force,
	})
	if err != nil {
		svc.Errorf("failed to send cancel request to worker: %s", t.Id.Hex())
		return err
	}

	return nil
}

func (svc *Service) SetInterval(interval time.Duration) {
	svc.interval = interval
}

func (svc *Service) SaveTask(t *models.Task, by primitive.ObjectID) (err error) {
	if t.Id.IsZero() {
		t.SetCreated(by)
		t.SetUpdated(by)
		_, err = service.NewModelService[models.Task]().InsertOne(*t)
		return err
	} else {
		t.SetUpdated(by)
		return service.NewModelService[models.Task]().ReplaceById(t.Id, *t)
	}
}

// initTaskStatus initialize task status of existing tasks
func (svc *Service) initTaskStatus() {
	// set status of running tasks as TaskStatusAbnormal
	runningTasks, err := service.NewModelService[models.Task]().GetMany(bson.M{
		"status": bson.M{
			"$in": []string{
				constants.TaskStatusPending,
				constants.TaskStatusAssigned,
				constants.TaskStatusRunning,
			},
		},
	}, nil)
	if err != nil {
		if errors2.Is(err, mongo2.ErrNoDocuments) {
			return
		}
		svc.Errorf("failed to get running tasks: %v", err)
		return
	}
	for _, t := range runningTasks {
		go func(t *models.Task) {
			t.Status = constants.TaskStatusAbnormal
			if err := svc.SaveTask(t, primitive.NilObjectID); err != nil {
				svc.Errorf("failed to set task status as TaskStatusAbnormal: %s", t.Id.Hex())
				return
			}
		}(&t)
	}
}

func (svc *Service) isMasterNode(t *models.Task) (ok bool, err error) {
	if t.NodeId.IsZero() {
		err = fmt.Errorf("task %s has no node id", t.Id.Hex())
		svc.Errorf("%v", err)
		return false, err
	}
	n, err := service.NewModelService[models.Node]().GetById(t.NodeId)
	if err != nil {
		if errors2.Is(err, mongo2.ErrNoDocuments) {
			svc.Errorf("node not found: %s", t.NodeId.Hex())
			return false, err
		}
		svc.Errorf("failed to get node: %s", t.NodeId.Hex())
		return false, err
	}
	return n.IsMaster, nil
}

func (svc *Service) cleanupTasks() {
	for {
		// task stats over 30 days ago
		taskStats, err := service.NewModelService[models.TaskStat]().GetMany(bson.M{
			"created_ts": bson.M{
				"$lt": time.Now().Add(-30 * 24 * time.Hour),
			},
		}, nil)
		if err != nil {
			time.Sleep(30 * time.Minute)
			continue
		}

		// task ids
		var ids []primitive.ObjectID
		for _, ts := range taskStats {
			ids = append(ids, ts.Id)
		}

		if len(ids) > 0 {
			// remove tasks
			if err := service.NewModelService[models.Task]().DeleteMany(bson.M{
				"_id": bson.M{"$in": ids},
			}); err != nil {
				svc.Warnf("failed to remove tasks: %v", err)
			}

			// remove task stats
			if err := service.NewModelService[models.TaskStat]().DeleteMany(bson.M{
				"_id": bson.M{"$in": ids},
			}); err != nil {
				svc.Warnf("failed to remove task stats: %v", err)
			}
		}

		time.Sleep(30 * time.Minute)
	}
}

func newTaskSchedulerService() *Service {
	return &Service{
		interval:   5 * time.Second,
		svr:        server.GetGrpcServer(),
		handlerSvc: handler.GetTaskHandlerService(),
		Logger:     utils.NewLogger("TaskScheduler"),
	}
}

var _service *Service
var _serviceOnce sync.Once

func GetTaskSchedulerService() *Service {
	_serviceOnce.Do(func() {
		_service = newTaskSchedulerService()
	})
	return _service
}
