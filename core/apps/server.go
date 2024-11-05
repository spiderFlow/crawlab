package apps

import (
	"fmt"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/config"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/node/service"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/spf13/viper"
	"net/http"
	_ "net/http/pprof"
	"sync"
)

type Server struct {
	// settings
	grpcAddress interfaces.Address

	// dependencies
	interfaces.WithConfigPath

	// modules
	nodeSvc interfaces.NodeService
	api     *Api
	dck     *Docker

	// internals
	quit chan int
}

func (app *Server) Init() {
	// log node info
	app.logNodeInfo()

	// pprof
	app.initPprof()
}

func (app *Server) Start() {
	if utils.IsMaster() {
		// start docker app
		if utils.IsDocker() {
			go app.dck.Start()
		}

		// start api
		go app.api.Start()
	}

	// start node service
	go app.nodeSvc.Start()
}

func (app *Server) Wait() {
	<-app.quit
}

func (app *Server) Stop() {
	app.api.Stop()
	app.quit <- 1
}

func (app *Server) GetApi() ApiApp {
	return app.api
}

func (app *Server) GetNodeService() interfaces.NodeService {
	return app.nodeSvc
}

func (app *Server) logNodeInfo() {
	log.Infof("current node type: %s", utils.GetNodeType())
	if utils.IsDocker() {
		log.Infof("running in docker container")
	}
}

func (app *Server) initPprof() {
	if viper.GetBool("pprof") {
		go func() {
			fmt.Println(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
	}
}

func newServer() NodeApp {
	// server
	svr := &Server{
		WithConfigPath: config.NewConfigPathService(),
		quit:           make(chan int, 1),
	}

	// master modules
	if utils.IsMaster() {
		// api
		svr.api = GetApi()

		// docker
		if utils.IsDocker() {
			svr.dck = GetDocker(svr)
		}
	}

	// node service
	if utils.IsMaster() {
		svr.nodeSvc = service.GetMasterService()
	} else {
		svr.nodeSvc = service.GetWorkerService()
	}

	return svr
}

var server NodeApp
var serverOnce sync.Once

func GetServer() NodeApp {
	serverOnce.Do(func() {
		server = newServer()
	})
	return server
}
