package client

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/cenkalti/backoff/v4"
	"github.com/crawlab-team/crawlab/core/grpc/middlewares"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/utils"
	grpc2 "github.com/crawlab-team/crawlab/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	// dependencies
	nodeCfgSvc interfaces.NodeConfigService

	// settings
	address string
	timeout time.Duration

	// internals
	conn    *grpc.ClientConn
	err     error
	once    sync.Once
	stopped bool
	stop    chan struct{}

	// clients
	NodeClient             grpc2.NodeServiceClient
	TaskClient             grpc2.TaskServiceClient
	ModelBaseServiceClient grpc2.ModelBaseServiceClient
	DependencyClient       grpc2.DependencyServiceClient
	MetricClient           grpc2.MetricServiceClient
}

func (c *GrpcClient) Start() (err error) {
	c.once.Do(func() {
		// connect
		err = c.connect()
		if err != nil {
			return
		}

		// register rpc services
		c.register()
	})

	return err
}

func (c *GrpcClient) Stop() (err error) {
	// set stopped flag
	c.stopped = true
	c.stop <- struct{}{}
	log.Infof("[GrpcClient] stopped")

	// skip if connection is nil
	if c.conn == nil {
		return nil
	}

	// close connection
	if err := c.conn.Close(); err != nil {
		return err
	}
	log.Infof("grpc client disconnected from %s", c.address)

	return nil
}

func (c *GrpcClient) register() {
	c.NodeClient = grpc2.NewNodeServiceClient(c.conn)
	c.ModelBaseServiceClient = grpc2.NewModelBaseServiceClient(c.conn)
	c.TaskClient = grpc2.NewTaskServiceClient(c.conn)
	c.DependencyClient = grpc2.NewDependencyServiceClient(c.conn)
	c.MetricClient = grpc2.NewMetricServiceClient(c.conn)
}

func (c *GrpcClient) Context() (ctx context.Context, cancel context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

func (c *GrpcClient) IsStarted() (res bool) {
	return c.conn != nil
}

func (c *GrpcClient) IsClosed() (res bool) {
	if c.conn != nil {
		return c.conn.GetState() == connectivity.Shutdown
	}
	return false
}

func (c *GrpcClient) getRequestData(d interface{}) (data []byte) {
	if d == nil {
		return data
	}
	switch d.(type) {
	case []byte:
		data = d.([]byte)
	default:
		var err error
		data, err = json.Marshal(d)
		if err != nil {
			panic(err)
		}
	}
	return data
}

func (c *GrpcClient) connect() (err error) {
	op := func() error {
		// connection options
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithChainUnaryInterceptor(middlewares.GetAuthTokenUnaryChainInterceptor()),
			grpc.WithChainStreamInterceptor(middlewares.GetAuthTokenStreamChainInterceptor()),
		}

		// create new client connection
		c.conn, err = grpc.NewClient(c.address, opts...)
		if err != nil {
			log.Errorf("[GrpcClient] grpc client failed to connect to %s: %v", c.address, err)
			return err
		}

		// connect
		c.conn.Connect()
		log.Infof("[GrpcClient] grpc client connected to %s", c.address)

		return nil
	}
	return backoff.RetryNotify(op, backoff.NewExponentialBackOff(), utils.BackoffErrorNotify("grpc client connect"))
}

func newGrpcClient() (c *GrpcClient) {
	return &GrpcClient{
		address: utils.GetGrpcAddress(),
		timeout: 10 * time.Second,
		stop:    make(chan struct{}),
	}
}

var _client *GrpcClient
var _clientOnce sync.Once

func GetGrpcClient() *GrpcClient {
	_clientOnce.Do(func() {
		_client = newGrpcClient()
	})
	return _client
}
