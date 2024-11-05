package service

import (
	"errors"
	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	config2 "github.com/crawlab-team/crawlab/core/config"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/grpc/server"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/common"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/node/config"
	"github.com/crawlab-team/crawlab/core/notification"
	"github.com/crawlab-team/crawlab/core/schedule"
	"github.com/crawlab-team/crawlab/core/system"
	"github.com/crawlab-team/crawlab/core/task/handler"
	"github.com/crawlab-team/crawlab/core/task/scheduler"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
	"github.com/crawlab-team/crawlab/trace"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

type MasterService struct {
	// dependencies
	cfgSvc           interfaces.NodeConfigService
	server           *server.GrpcServer
	taskSchedulerSvc *scheduler.Service
	taskHandlerSvc   *handler.Service
	scheduleSvc      *schedule.Service
	systemSvc        *system.Service

	// settings
	cfgPath         string
	address         interfaces.Address
	monitorInterval time.Duration
	stopOnError     bool
}

func (svc *MasterService) Start() {
	// create indexes
	common.InitIndexes()

	// start grpc server
	if err := svc.server.Start(); err != nil {
		panic(err)
	}

	// register to db
	if err := svc.Register(); err != nil {
		panic(err)
	}

	// start monitoring worker nodes
	go svc.Monitor()

	// start task handler
	go svc.taskHandlerSvc.Start()

	// start task scheduler
	go svc.taskSchedulerSvc.Start()

	// start schedule service
	go svc.scheduleSvc.Start()

	// wait for quit signal
	svc.Wait()

	// stop
	svc.Stop()
}

func (svc *MasterService) Wait() {
	utils.DefaultWait()
}

func (svc *MasterService) Stop() {
	_ = svc.server.Stop()
	svc.taskHandlerSvc.Stop()
	log.Infof("master[%s] service has stopped", svc.GetConfigService().GetNodeKey())
}

func (svc *MasterService) Monitor() {
	log.Infof("master[%s] monitoring started", svc.GetConfigService().GetNodeKey())

	// ticker
	ticker := time.NewTicker(svc.monitorInterval)

	for {
		// monitor
		err := svc.monitor()
		if err != nil {
			trace.PrintError(err)
			if svc.stopOnError {
				log.Errorf("master[%s] monitor error, now stopping...", svc.GetConfigService().GetNodeKey())
				svc.Stop()
				return
			}
		}

		// wait
		<-ticker.C
	}
}

func (svc *MasterService) GetConfigService() (cfgSvc interfaces.NodeConfigService) {
	return svc.cfgSvc
}

func (svc *MasterService) GetConfigPath() (path string) {
	return svc.cfgPath
}

func (svc *MasterService) SetConfigPath(path string) {
	svc.cfgPath = path
}

func (svc *MasterService) SetMonitorInterval(duration time.Duration) {
	svc.monitorInterval = duration
}

func (svc *MasterService) Register() (err error) {
	nodeKey := svc.GetConfigService().GetNodeKey()
	nodeName := svc.GetConfigService().GetNodeName()
	node, err := service.NewModelService[models.Node]().GetOne(bson.M{"key": nodeKey}, nil)
	if err != nil && err.Error() == mongo2.ErrNoDocuments.Error() {
		// not exists
		log.Infof("master[%s] does not exist in db", nodeKey)
		node := models.Node{
			Key:        nodeKey,
			Name:       nodeName,
			MaxRunners: config.DefaultConfigOptions.MaxRunners,
			IsMaster:   true,
			Status:     constants.NodeStatusOnline,
			Enabled:    true,
			Active:     true,
			ActiveAt:   time.Now(),
		}
		node.SetCreated(primitive.NilObjectID)
		node.SetUpdated(primitive.NilObjectID)
		id, err := service.NewModelService[models.Node]().InsertOne(node)
		if err != nil {
			return err
		}
		log.Infof("added master[%s] in db. id: %s", nodeKey, id.Hex())
		return nil
	} else if err == nil {
		// exists
		log.Infof("master[%s] exists in db", nodeKey)
		node.Status = constants.NodeStatusOnline
		node.Active = true
		node.ActiveAt = time.Now()
		err = service.NewModelService[models.Node]().ReplaceById(node.Id, *node)
		if err != nil {
			return err
		}
		log.Infof("updated master[%s] in db. id: %s", nodeKey, node.Id.Hex())
		return nil
	} else {
		// error
		return err
	}
}

func (svc *MasterService) monitor() (err error) {
	// update master node status in db
	if err := svc.updateMasterNodeStatus(); err != nil {
		if err.Error() == mongo2.ErrNoDocuments.Error() {
			return nil
		}
		return err
	}

	// all worker nodes
	workerNodes, err := svc.getAllWorkerNodes()
	if err != nil {
		return err
	}

	// iterate all worker nodes
	wg := sync.WaitGroup{}
	wg.Add(len(workerNodes))
	for _, n := range workerNodes {
		go func(n *models.Node) {
			defer wg.Done()

			// subscribe
			ok := svc.subscribeNode(n)
			if !ok {
				go svc.setWorkerNodeOffline(n)
				return
			}

			// ping client
			ok = svc.pingNodeClient(n)
			if !ok {
				go svc.setWorkerNodeOffline(n)
				return
			}

			// update node available runners
			if err := svc.updateNodeAvailableRunners(n); err != nil {
				trace.PrintError(err)
				return
			}
		}(&n)
	}

	wg.Wait()

	return nil
}

func (svc *MasterService) getAllWorkerNodes() (nodes []models.Node, err error) {
	query := bson.M{
		"key":    bson.M{"$ne": svc.cfgSvc.GetNodeKey()}, // not self
		"active": true,                                   // active
	}
	nodes, err = service.NewModelService[models.Node]().GetMany(query, nil)
	if err != nil {
		if errors.Is(err, mongo2.ErrNoDocuments) {
			return nil, nil
		}
		return nil, trace.TraceError(err)
	}
	return nodes, nil
}

func (svc *MasterService) updateMasterNodeStatus() (err error) {
	nodeKey := svc.GetConfigService().GetNodeKey()
	node, err := service.NewModelService[models.Node]().GetOne(bson.M{"key": nodeKey}, nil)
	if err != nil {
		return err
	}
	oldStatus := node.Status

	node.Status = constants.NodeStatusOnline
	node.Active = true
	node.ActiveAt = time.Now()
	newStatus := node.Status

	err = service.NewModelService[models.Node]().ReplaceById(node.Id, *node)
	if err != nil {
		return err
	}

	if utils.IsPro() {
		if oldStatus != newStatus {
			go svc.sendNotification(node)
		}
	}

	return nil
}

func (svc *MasterService) setWorkerNodeOffline(node *models.Node) {
	node.Status = constants.NodeStatusOffline
	node.Active = false
	err := backoff.Retry(func() error {
		return service.NewModelService[models.Node]().ReplaceById(node.Id, *node)
	}, backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 3))
	if err != nil {
		trace.PrintError(err)
	}
	svc.sendNotification(node)
}

func (svc *MasterService) subscribeNode(n *models.Node) (ok bool) {
	_, ok = svc.server.NodeSvr.GetSubscribeStream(n.Id)
	return ok
}

func (svc *MasterService) pingNodeClient(n *models.Node) (ok bool) {
	stream, ok := svc.server.NodeSvr.GetSubscribeStream(n.Id)
	if !ok {
		log.Errorf("cannot get worker node client[%s]", n.Key)
		return false
	}
	err := stream.Send(&grpc.NodeServiceSubscribeResponse{
		Code: grpc.NodeServiceSubscribeCode_PING,
	})
	if err != nil {
		log.Errorf("failed to ping worker node client[%s]: %v", n.Key, err)
		return false
	}
	return true
}

func (svc *MasterService) updateNodeAvailableRunners(node *models.Node) (err error) {
	query := bson.M{
		"node_id": node.Id,
		"status":  constants.TaskStatusRunning,
	}
	runningTasksCount, err := service.NewModelService[models.Task]().Count(query)
	if err != nil {
		return trace.TraceError(err)
	}
	node.AvailableRunners = node.MaxRunners - runningTasksCount
	err = service.NewModelService[models.Node]().ReplaceById(node.Id, *node)
	if err != nil {
		return err
	}
	return nil
}

func (svc *MasterService) sendNotification(node *models.Node) {
	if !utils.IsPro() {
		return
	}
	go notification.GetNotificationService().SendNodeNotification(node)
}

func newMasterService() *MasterService {
	return &MasterService{
		cfgPath:          config2.GetConfigPath(),
		cfgSvc:           config.GetNodeConfigService(),
		monitorInterval:  15 * time.Second,
		stopOnError:      false,
		server:           server.GetGrpcServer(),
		taskSchedulerSvc: scheduler.GetTaskSchedulerService(),
		taskHandlerSvc:   handler.GetTaskHandlerService(),
		scheduleSvc:      schedule.GetScheduleService(),
		systemSvc:        system.GetSystemService(),
	}
}

var masterService *MasterService
var masterServiceOnce sync.Once

func GetMasterService() *MasterService {
	masterServiceOnce.Do(func() {
		masterService = newMasterService()
	})
	return masterService
}
