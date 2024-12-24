package apps

import (
	"github.com/crawlab-team/crawlab/core/utils"
)

var utilsLogger = utils.NewLogger("AppsUtils")

func Start(app App) {
	start(app)
}

func start(app App) {
	app.Init()
	go app.Start()
	app.Wait()
	app.Stop()
}

func initModule(name string, fn func() error) (err error) {
	if err := fn(); err != nil {
		utilsLogger.Errorf("init %s error: %v", name, err)
		panic(err)
	}
	utilsLogger.Infof("initialized %s successfully", name)
	return nil
}
