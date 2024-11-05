package system

import (
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
)

type Service struct {
}

func (svc *Service) Init() (err error) {
	// initialize data
	if err := svc.initData(); err != nil {
		return err
	}

	return nil
}

func (svc *Service) initData() (err error) {
	total, err := service.NewModelService[models.Setting]().Count(bson.M{
		"key": "site_title",
	})
	if err != nil {
		return err
	}
	if total > 0 {
		return nil
	}

	// data to initialize
	settings := []models.Setting{
		{
			Key: "site_title",
			Value: bson.M{
				"customize_site_title": false,
				"site_title":           "",
			},
		},
	}
	_, err = service.NewModelService[models.Setting]().InsertMany(settings)
	if err != nil {
		return err
	}
	return nil
}

func newSystemService() *Service {
	// service
	svc := &Service{}

	if err := svc.Init(); err != nil {
		panic(err)
	}

	return svc
}

var _service *Service
var _serviceOnce sync.Once

func GetSystemService() *Service {
	if _service == nil {
		_service = newSystemService()
	}
	_serviceOnce.Do(func() {
		_service = newSystemService()
	})
	return _service
}
