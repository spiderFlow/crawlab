package apps

import (
	"fmt"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/node/service"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/spf13/viper"
	"net/http"
	_ "net/http/pprof"
	"sync"
)

type Server struct {
	// modules
	nodeSvc interfaces.NodeService
	api     *Api

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
}

func (app *Server) initPprof() {
	if viper.GetBool("pprof") {
		go func() {
			fmt.Println(http.ListenAndServe("0.0.0.0:6060", nil))
		}()
	}
}

func newServer() App {
	// server
	svr := &Server{
		quit: make(chan int, 1),
	}

	// master modules
	if utils.IsMaster() {
		// api
		svr.api = GetApi()
	}

	// node service
	if utils.IsMaster() {
		svr.nodeSvc = service.GetMasterService()
	} else {
		svr.nodeSvc = service.GetWorkerService()
	}

	return svr
}

var server App
var serverOnce sync.Once

func GetServer() App {
	serverOnce.Do(func() {
		server = newServer()
	})
	return server
}
