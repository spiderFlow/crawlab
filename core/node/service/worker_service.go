package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/crawlab-team/crawlab/core/controllers"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/crawlab-team/crawlab/core/models/models"

	"github.com/cenkalti/backoff/v4"
	"github.com/crawlab-team/crawlab/core/grpc/client"
	"github.com/crawlab-team/crawlab/core/interfaces"
	client2 "github.com/crawlab-team/crawlab/core/models/client"
	nodeconfig "github.com/crawlab-team/crawlab/core/node/config"
	"github.com/crawlab-team/crawlab/core/task/handler"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
	"go.mongodb.org/mongo-driver/bson"
)

type WorkerService struct {
	// dependencies
	cfgSvc     interfaces.NodeConfigService
	handlerSvc *handler.Service

	// settings
	address           interfaces.Address
	heartbeatInterval time.Duration

	// internals
	stopped bool
	n       *models.Node
	s       grpc.NodeService_SubscribeClient
	isReady bool
	interfaces.Logger
}

func (svc *WorkerService) Start() {
	// wait for grpc client ready
	client.GetGrpcClient().WaitForReady()

	// register to master
	svc.register()

	// mark as ready after registration
	svc.isReady = true

	// start health check server
	go svc.startHealthServer()

	// subscribe
	go svc.subscribe()

	// start sending heartbeat to master
	go svc.reportStatus()

	// start task handler
	go svc.handlerSvc.Start()

	// wait for quit signal
	svc.Wait()

	// stop
	svc.Stop()
}

func (svc *WorkerService) Wait() {
	utils.DefaultWait()
}

func (svc *WorkerService) Stop() {
	svc.stopped = true
	_ = client.GetGrpcClient().Stop()
	svc.handlerSvc.Stop()
	svc.Infof("worker[%s] service has stopped", svc.cfgSvc.GetNodeKey())
}

func (svc *WorkerService) register() {
	op := func() (err error) {
		ctx, cancel := client.GetGrpcClient().Context()
		defer cancel()
		_, err = client.GetGrpcClient().NodeClient.Register(ctx, &grpc.NodeServiceRegisterRequest{
			NodeKey:    svc.cfgSvc.GetNodeKey(),
			NodeName:   svc.cfgSvc.GetNodeName(),
			MaxRunners: int32(svc.cfgSvc.GetMaxRunners()),
		})
		if err != nil {
			err = fmt.Errorf("failed to register worker[%s]: %v", svc.cfgSvc.GetNodeKey(), err)
			return err
		}
		svc.n, err = client2.NewModelService[models.Node]().GetOne(bson.M{"key": svc.GetConfigService().GetNodeKey()}, nil)
		if err != nil {
			err = fmt.Errorf("failed to get node: %v", err)
			return err
		}
		svc.Infof("worker[%s] registered to master. id: %s", svc.GetConfigService().GetNodeKey(), svc.n.Id.Hex())
		return nil
	}
	b := backoff.NewExponentialBackOff()
	n := func(err error, duration time.Duration) {
		svc.Errorf("register worker[%s] error: %v", svc.cfgSvc.GetNodeKey(), err)
		svc.Infof("retry in %.1f seconds", duration.Seconds())
	}
	err := backoff.RetryNotify(op, b, n)
	if err != nil {
		svc.Fatalf("failed to register worker[%s]: %v", svc.cfgSvc.GetNodeKey(), err)
		panic(err)
	}
}

func (svc *WorkerService) reportStatus() {
	ticker := time.NewTicker(svc.heartbeatInterval)
	for {
		// return if client is closed
		if client.GetGrpcClient().IsClosed() {
			ticker.Stop()
			return
		}

		// send heartbeat
		svc.sendHeartbeat()

		// sleep
		<-ticker.C
	}
}

func (svc *WorkerService) GetConfigService() (cfgSvc interfaces.NodeConfigService) {
	return svc.cfgSvc
}

func (svc *WorkerService) subscribe() {
	// Configure exponential backoff
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Second
	b.MaxInterval = 1 * time.Minute
	b.MaxElapsedTime = 10 * time.Minute
	b.Multiplier = 2.0

	for {
		if svc.stopped {
			svc.Infof("subscription stopped. exiting...")
			return
		}

		// Use backoff for connection attempts
		operation := func() error {
			svc.Debugf("attempting to subscribe to master")
			stream, err := client.GetGrpcClient().NodeClient.Subscribe(context.Background(), &grpc.NodeServiceSubscribeRequest{
				NodeKey: svc.cfgSvc.GetNodeKey(),
			})
			if err != nil {
				svc.Errorf("failed to subscribe to master: %v", err)
				return err
			}
			svc.Debugf("subscribed to master")

			// Handle messages
			for {
				if svc.stopped {
					return nil
				}

				msg, err := stream.Recv()
				if err != nil {
					if client.GetGrpcClient().IsClosed() {
						svc.Errorf("connection to master is closed: %v", err)
						return err
					}
					svc.Errorf("failed to receive message from master: %v", err)
					return err
				}

				switch msg.Code {
				case grpc.NodeServiceSubscribeCode_PING:
					// do nothing
				}
			}
		}

		// Execute with backoff
		err := backoff.Retry(operation, b)
		if err != nil {
			svc.Errorf("subscription failed after max retries: %v", err)
			return
		}

		// Wait before attempting to reconnect
		time.Sleep(time.Second)
	}
}

func (svc *WorkerService) sendHeartbeat() {
	ctx, cancel := context.WithTimeout(context.Background(), svc.heartbeatInterval)
	defer cancel()
	_, err := client.GetGrpcClient().NodeClient.SendHeartbeat(ctx, &grpc.NodeServiceSendHeartbeatRequest{
		NodeKey: svc.cfgSvc.GetNodeKey(),
	})
	if err != nil {
		svc.Errorf("failed to send heartbeat to master: %v", err)
	}
}

func (svc *WorkerService) startHealthServer() {
	// handlers
	app := gin.New()
	app.GET("/health", controllers.GetHealthFn(func() bool {
		return svc.isReady && !svc.stopped && client.GetGrpcClient() != nil && !client.GetGrpcClient().IsClosed()
	}))

	// listen
	ln, err := net.Listen("tcp", utils.GetServerAddress())
	if err != nil {
		panic(err)
	}

	// serve
	if err := http.Serve(ln, app); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			svc.Errorf("run server error: %v", err)
		} else {
			svc.Info("server graceful down")
		}
	}
}

func newWorkerService() *WorkerService {
	return &WorkerService{
		heartbeatInterval: 15 * time.Second,
		cfgSvc:            nodeconfig.GetNodeConfigService(),
		handlerSvc:        handler.GetTaskHandlerService(),
		isReady:           false,
		Logger:            utils.NewLogger("WorkerService"),
	}
}

var workerService *WorkerService
var workerServiceOnce sync.Once

func GetWorkerService() *WorkerService {
	workerServiceOnce.Do(func() {
		workerService = newWorkerService()
	})
	return workerService
}
