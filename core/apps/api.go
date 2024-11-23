package apps

import (
	"context"
	"errors"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/controllers"
	"github.com/crawlab-team/crawlab/core/middlewares"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"sync"
	"time"
)

func init() {
	// set gin mode
	gin.SetMode(utils.GetGinMode())
}

type Api struct {
	// internals
	app *gin.Engine
	ln  net.Listener
	srv *http.Server
}

func (app *Api) Init() {
	// initialize middlewares
	_ = app.initModuleWithApp("middlewares", middlewares.InitMiddlewares)

	// initialize routes
	_ = app.initModuleWithApp("routes", controllers.InitRoutes)
}

func (app *Api) Start() {
	// address
	address := utils.GetServerAddress()

	// http server
	app.srv = &http.Server{
		Handler: app.app,
		Addr:    address,
	}

	// listen
	var err error
	app.ln, err = net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	log.Infof("api server listening on %s", address)

	// serve
	if err := http.Serve(app.ln, app.app); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("run api server error: %v", err)
		} else {
			log.Info("api server graceful down")
		}
	}
}

func (app *Api) Wait() {
	utils.DefaultWait()
}

func (app *Api) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.srv.Shutdown(ctx); err != nil {
		log.Errorf("shutdown api server error: %v", err)
	}
}

func (app *Api) GetGinEngine() *gin.Engine {
	return app.app
}

func (app *Api) GetHttpServer() *http.Server {
	return app.srv
}

func (app *Api) initModuleWithApp(name string, fn func(app *gin.Engine) error) (err error) {
	return initModule(name, func() error {
		return fn(app.app)
	})
}

func newApi() *Api {
	return &Api{
		app: gin.New(),
	}
}

var api *Api
var apiOnce sync.Once

func GetApi() *Api {
	apiOnce.Do(func() {
		api = newApi()
	})
	return api
}
