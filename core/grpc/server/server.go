package server

import (
	"fmt"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/grpc/middlewares"
	"github.com/crawlab-team/crawlab/core/utils"
	grpc2 "github.com/crawlab-team/crawlab/grpc"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	errors2 "github.com/pkg/errors"
	"google.golang.org/grpc"
	"net"
	"sync"
)

type GrpcServer struct {
	// settings
	address string

	// internals
	svr     *grpc.Server
	l       net.Listener
	stopped bool

	// servers
	NodeSvr             *NodeServiceServer
	TaskSvr             *TaskServiceServer
	ModelBaseServiceSvr *ModelBaseServiceServer
	DependencySvr       *DependencyServiceServer
	MetricSvr           *MetricServiceServer
}

func (svr *GrpcServer) Init() {
	svr.register()
}

func (svr *GrpcServer) Start() (err error) {
	// listener
	svr.l, err = net.Listen("tcp", svr.address)
	if err != nil {
		log.Errorf("[GrpcServer] failed to listen: %v", err)
		return err
	}
	log.Infof("[GrpcServer] grpc server listens to %s", svr.address)

	// start grpc server
	go func() {
		if err := svr.svr.Serve(svr.l); err != nil {
			if errors2.Is(err, grpc.ErrServerStopped) {
				return
			}
			log.Errorf("[GrpcServer] failed to serve: %v", err)
		}
	}()

	return nil
}

func (svr *GrpcServer) Stop() (err error) {
	// skip if listener is nil
	if svr.l == nil {
		return nil
	}

	// graceful stop
	log.Infof("[GrpcServer] grpc server stopping...")
	svr.svr.Stop()

	// close listener
	log.Infof("[GrpcServer] grpc server closing listener...")
	_ = svr.l.Close()

	// mark as stopped
	svr.stopped = true

	// log
	log.Infof("[GrpcServer] grpc server stopped")

	return nil
}

func (svr *GrpcServer) register() {
	grpc2.RegisterNodeServiceServer(svr.svr, *svr.NodeSvr)
	grpc2.RegisterModelBaseServiceServer(svr.svr, *svr.ModelBaseServiceSvr)
	grpc2.RegisterTaskServiceServer(svr.svr, *svr.TaskSvr)
	grpc2.RegisterDependencyServiceServer(svr.svr, *svr.DependencySvr)
	grpc2.RegisterMetricServiceServer(svr.svr, *svr.MetricSvr)
}

func (svr *GrpcServer) recoveryHandlerFunc(p interface{}) (err error) {
	log.Errorf("[GrpcServer] recovered from panic: %v", p)
	return fmt.Errorf("recovered from panic: %v", p)
}

func newGrpcServer() *GrpcServer {
	// server
	svr := &GrpcServer{
		address: utils.GetGrpcServerAddress(),
	}

	// services servers
	svr.NodeSvr = GetNodeServiceServer()
	svr.ModelBaseServiceSvr = GetModelBaseServiceServer()
	svr.TaskSvr = GetTaskServiceServer()
	svr.DependencySvr = GetDependencyServer()
	svr.MetricSvr = GetMetricsServer()

	// recovery options
	recoveryOpts := []grpcrecovery.Option{
		grpcrecovery.WithRecoveryHandler(svr.recoveryHandlerFunc),
	}

	// grpc server
	svr.svr = grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcrecovery.UnaryServerInterceptor(recoveryOpts...),
			grpcauth.UnaryServerInterceptor(middlewares.GetAuthTokenFunc()),
		),
		grpcmiddleware.WithStreamServerChain(
			grpcrecovery.StreamServerInterceptor(recoveryOpts...),
			grpcauth.StreamServerInterceptor(middlewares.GetAuthTokenFunc()),
		),
	)

	// initialize
	svr.Init()

	return svr
}

var _server *GrpcServer
var _serverOnce sync.Once

func GetGrpcServer() *GrpcServer {
	_serverOnce.Do(func() {
		_server = newGrpcServer()
	})
	return _server
}

func NewGrpcServer() *GrpcServer {
	return newGrpcServer()
}
