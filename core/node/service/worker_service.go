package service

import (
	"context"
	"errors"
	"github.com/crawlab-team/crawlab/core/controllers"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/crawlab-team/crawlab/core/models/models"

	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/crawlab-team/crawlab/core/grpc/client"
	"github.com/crawlab-team/crawlab/core/interfaces"
	client2 "github.com/crawlab-team/crawlab/core/models/client"
	nodeconfig "github.com/crawlab-team/crawlab/core/node/config"
	"github.com/crawlab-team/crawlab/core/task/handler"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
	"github.com/crawlab-team/crawlab/trace"
	"go.mongodb.org/mongo-driver/bson"
)

type WorkerService struct {
	// dependencies
	cfgSvc     interfaces.NodeConfigService
	client     *client.GrpcClient
	handlerSvc *handler.Service

	// settings
	address           interfaces.Address
	heartbeatInterval time.Duration

	// internals
	stopped bool
	n       *models.Node
	s       grpc.NodeService_SubscribeClient
	isReady bool
}

func (svc *WorkerService) Start() {
	// start grpc client (retry if failed)
	err := backoff.RetryNotify(
		func() error {
			return svc.client.Start()
		},
		backoff.NewExponentialBackOff(
			backoff.WithInitialInterval(1*time.Second),
			backoff.WithMaxInterval(1*time.Minute),
			backoff.WithMaxElapsedTime(10*time.Minute),
		),
		func(err error, duration time.Duration) {
			log.Errorf("failed to start grpc client: %v", err)
			log.Infof("retrying in %s", duration)
		},
	)
	if err != nil {
		log.Fatalf("failed to start grpc client: %v", err)
		panic(err)
	}

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
	_ = svc.client.Stop()
	svc.handlerSvc.Stop()
	log.Infof("worker[%s] service has stopped", svc.cfgSvc.GetNodeKey())
}

func (svc *WorkerService) register() {
	ctx, cancel := svc.client.Context()
	defer cancel()
	_, err := svc.client.NodeClient.Register(ctx, &grpc.NodeServiceRegisterRequest{
		NodeKey:    svc.cfgSvc.GetNodeKey(),
		NodeName:   svc.cfgSvc.GetNodeName(),
		MaxRunners: int32(svc.cfgSvc.GetMaxRunners()),
	})
	if err != nil {
		log.Fatalf("failed to register worker[%s] to master: %v", svc.cfgSvc.GetNodeKey(), err)
		panic(err)
	}
	svc.n, err = client2.NewModelService[models.Node]().GetOne(bson.M{"key": svc.GetConfigService().GetNodeKey()}, nil)
	if err != nil {
		log.Fatalf("failed to get node: %v", err)
		panic(err)
	}
	log.Infof("worker[%s] registered to master. id: %s", svc.GetConfigService().GetNodeKey(), svc.n.Id.Hex())
	return
}

func (svc *WorkerService) reportStatus() {
	ticker := time.NewTicker(svc.heartbeatInterval)
	for {
		// return if client is closed
		if svc.client.IsClosed() {
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
			return
		}

		// Use backoff for connection attempts
		operation := func() error {
			stream, err := svc.client.NodeClient.Subscribe(context.Background(), &grpc.NodeServiceSubscribeRequest{
				NodeKey: svc.cfgSvc.GetNodeKey(),
			})
			if err != nil {
				log.Errorf("failed to subscribe to master: %v", err)
				return err
			}

			// Handle messages
			for {
				if svc.stopped {
					return nil
				}

				msg, err := stream.Recv()
				if err != nil {
					if svc.client.IsClosed() {
						log.Errorf("connection to master is closed: %v", err)
						return err
					}
					log.Errorf("failed to receive message from master: %v", err)
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
			log.Errorf("subscription failed after max retries: %v", err)
			return
		}

		// Wait before attempting to reconnect
		time.Sleep(time.Second)
	}
}

func (svc *WorkerService) sendHeartbeat() {
	ctx, cancel := context.WithTimeout(context.Background(), svc.heartbeatInterval)
	defer cancel()
	_, err := svc.client.NodeClient.SendHeartbeat(ctx, &grpc.NodeServiceSendHeartbeatRequest{
		NodeKey: svc.cfgSvc.GetNodeKey(),
	})
	if err != nil {
		trace.PrintError(err)
	}
}

func (svc *WorkerService) startHealthServer() {
	// handlers
	app := gin.New()
	app.GET("/health", controllers.GetHealthFn(func() bool {
		return svc.isReady && !svc.stopped && svc.client != nil && !svc.client.IsClosed()
	}))

	// listen
	ln, err := net.Listen("tcp", utils.GetServerAddress())
	if err != nil {
		panic(err)
	}

	// serve
	if err := http.Serve(ln, app); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("run server error:" + err.Error())
		} else {
			log.Info("server graceful down")
		}
	}
}

func newWorkerService() *WorkerService {
	return &WorkerService{
		heartbeatInterval: 15 * time.Second,
		cfgSvc:            nodeconfig.GetNodeConfigService(),
		client:            client.GetGrpcClient(),
		handlerSvc:        handler.GetTaskHandlerService(),
		isReady:           false,
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
