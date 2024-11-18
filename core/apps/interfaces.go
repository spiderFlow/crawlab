package apps

import (
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/gin-gonic/gin"
	"net/http"
)

type App interface {
	Init()
	Start()
	Wait()
	Stop()
}

type ApiApp interface {
	App
	GetGinEngine() (engine *gin.Engine)
	GetHttpServer() (svr *http.Server)
}

type ServerApp interface {
	App
	GetApi() (api ApiApp)
	GetNodeService() (masterSvc interfaces.NodeService)
}
