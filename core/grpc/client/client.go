package client

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	interfaces.Logger

	// clients
	NodeClient             grpc2.NodeServiceClient
	TaskClient             grpc2.TaskServiceClient
	ModelBaseServiceClient grpc2.ModelBaseServiceClient
	DependencyClient       grpc2.DependencyServiceClient
	MetricClient           grpc2.MetricServiceClient

	// Add new fields for state management
	state     connectivity.State
	stateMux  sync.RWMutex
	reconnect chan struct{}
}

func (c *GrpcClient) Start() {
	c.once.Do(func() {
		// initialize reconnect channel
		c.reconnect = make(chan struct{})

		// start state monitor
		go c.monitorState()

		// connect
		err := c.connect()
		if err != nil {
			c.Fatalf("failed to connect: %v", err)
			return
		}

		// register rpc services
		c.register()
	})
}

func (c *GrpcClient) Stop() (err error) {
	// set stopped flag
	c.stopped = true
	c.stop <- struct{}{}
	c.Infof("stopped")

	// skip if connection is nil
	if c.conn == nil {
		return nil
	}

	// close connection
	if err := c.conn.Close(); err != nil {
		c.Errorf("failed to close connection: %v", err)
		return err
	}
	c.Infof("disconnected from %s", c.address)

	return nil
}

func (c *GrpcClient) WaitForReady() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if c.IsReady() {
				c.Debugf("client is now ready")
				return
			}
		case <-c.stop:
			c.Errorf("client has stopped")
		}
	}
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

func (c *GrpcClient) IsReady() (res bool) {
	state := c.conn.GetState()
	return c.conn != nil && state == connectivity.Ready
}

func (c *GrpcClient) IsClosed() (res bool) {
	if c.conn != nil {
		return c.conn.GetState() == connectivity.Shutdown
	}
	return false
}

func (c *GrpcClient) monitorState() {
	for {
		select {
		case <-c.stop:
			return
		default:
			if c.conn == nil {
				time.Sleep(time.Second)
				continue
			}

			previous := c.getState()
			current := c.conn.GetState()

			if previous != current {
				c.setState(current)
				c.Infof("state changed from %s to %s", previous, current)

				// Trigger reconnect if connection is lost or becomes idle from ready state
				if current == connectivity.TransientFailure ||
					current == connectivity.Shutdown ||
					(previous == connectivity.Ready && current == connectivity.Idle) {
					select {
					case c.reconnect <- struct{}{}:
						c.Infof("triggering reconnection due to state change to %s", current)
					default:
					}
				}
			}

			time.Sleep(time.Second)
		}
	}
}

func (c *GrpcClient) setState(state connectivity.State) {
	c.stateMux.Lock()
	defer c.stateMux.Unlock()
	c.state = state
}

func (c *GrpcClient) getState() connectivity.State {
	c.stateMux.RLock()
	defer c.stateMux.RUnlock()
	return c.state
}

func (c *GrpcClient) connect() (err error) {
	// Start reconnection loop
	go func() {
		for {
			select {
			case <-c.stop:
				return
			case <-c.reconnect:
				if !c.stopped {
					c.Infof("attempting to reconnect to %s", c.address)
					if err := c.doConnect(); err != nil {
						c.Errorf("reconnection failed: %v", err)
					}
				}
			}
		}
	}()

	return c.doConnect()
}

func (c *GrpcClient) doConnect() (err error) {
	op := func() error {
		// connection options
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithChainUnaryInterceptor(middlewares.GetGrpcClientAuthTokenUnaryChainInterceptor()),
			grpc.WithChainStreamInterceptor(middlewares.GetGrpcClientAuthTokenStreamChainInterceptor()),
		}

		// create new client connection
		c.conn, err = grpc.NewClient(c.address, opts...)
		if err != nil {
			c.Errorf("failed to connect to %s: %v", c.address, err)
			return err
		}

		// connect
		c.Infof("connecting to %s", c.address)
		c.conn.Connect()

		// wait for connection to be ready
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		ok := c.conn.WaitForStateChange(ctx, connectivity.Ready)
		if !ok {
			return fmt.Errorf("failed to connect to %s: timed out", c.address)
		}

		// success
		c.Infof("connected to %s", c.address)

		return nil
	}
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 5 * time.Second
	b.MaxElapsedTime = 10 * time.Minute
	n := func(err error, duration time.Duration) {
		c.Errorf("failed to connect to %s: %v, retrying in %s", c.address, err, duration)
	}
	return backoff.RetryNotify(op, b, n)
}

func newGrpcClient() (c *GrpcClient) {
	return &GrpcClient{
		address: utils.GetGrpcAddress(),
		timeout: 10 * time.Second,
		stop:    make(chan struct{}),
		Logger:  utils.NewLogger("GrpcClient"),
		state:   connectivity.Idle,
	}
}

var _client *GrpcClient
var _clientOnce sync.Once

func GetGrpcClient() *GrpcClient {
	_clientOnce.Do(func() {
		_client = newGrpcClient()
		go _client.Start()
	})
	return _client
}
