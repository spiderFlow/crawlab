package schedule

import (
	"github.com/crawlab-team/crawlab/core/interfaces"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/spider/admin"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
	"time"
)

type Service struct {
	// dependencies
	modelSvc *service.ModelService[models.Schedule]
	adminSvc *admin.Service

	// settings variables
	loc            *time.Location
	delay          bool
	skip           bool
	updateInterval time.Duration

	// internals
	cron      *cron.Cron
	logger    cron.Logger
	schedules []models.Schedule
	stopped   bool
	mu        sync.Mutex
	interfaces.Logger
}

func (svc *Service) GetLocation() (loc *time.Location) {
	return svc.loc
}

func (svc *Service) SetLocation(loc *time.Location) {
	svc.loc = loc
}

func (svc *Service) GetDelay() (delay bool) {
	return svc.delay
}

func (svc *Service) SetDelay(delay bool) {
	svc.delay = delay
}

func (svc *Service) GetSkip() (skip bool) {
	return svc.skip
}

func (svc *Service) SetSkip(skip bool) {
	svc.skip = skip
}

func (svc *Service) GetUpdateInterval() (interval time.Duration) {
	return svc.updateInterval
}

func (svc *Service) SetUpdateInterval(interval time.Duration) {
	svc.updateInterval = interval
}

func (svc *Service) Init() (err error) {
	err = svc.fetch()
	if err != nil {
		svc.Fatalf("failed to initialize schedule service: %v", err)
		return err
	}
	return nil
}

func (svc *Service) Start() {
	svc.cron.Start()
	go svc.Update()
}

func (svc *Service) Wait() {
	utils.DefaultWait()
	svc.Stop()
}

func (svc *Service) Stop() {
	svc.stopped = true
	svc.cron.Stop()
}

func (svc *Service) Enable(s models.Schedule, by primitive.ObjectID) (err error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	id, err := svc.cron.AddFunc(s.Cron, svc.schedule(s.Id))
	if err != nil {
		svc.Errorf("failed to add cron job: %v", err)
		return err
	}
	s.Enabled = true
	s.EntryId = id
	s.SetUpdated(by)
	return svc.modelSvc.ReplaceById(s.Id, s)
}

func (svc *Service) Disable(s models.Schedule, by primitive.ObjectID) (err error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	svc.cron.Remove(s.EntryId)
	s.Enabled = false
	s.EntryId = -1
	s.SetUpdated(by)
	return svc.modelSvc.ReplaceById(s.Id, s)
}

func (svc *Service) Update() {
	for {
		if svc.stopped {
			return
		}

		svc.update()

		time.Sleep(svc.updateInterval)
	}
}

func (svc *Service) GetCron() (c *cron.Cron) {
	return svc.cron
}

func (svc *Service) update() {
	// fetch enabled schedules
	if err := svc.fetch(); err != nil {
		svc.Errorf("failed to fetch schedules: %v", err)
		return
	}

	// entry id map
	entryIdsMap := svc.getEntryIdsMap()

	// iterate enabled schedules
	for _, s := range svc.schedules {
		_, ok := entryIdsMap[s.EntryId]
		if ok {
			entryIdsMap[s.EntryId] = true
		} else {
			if !s.Enabled {
				err := svc.Enable(s, s.GetCreatedBy())
				if err != nil {
					svc.Errorf("failed to enable schedule: %v", err)
					continue
				}
			}
		}
	}

	// remove non-existent entries
	for id, ok := range entryIdsMap {
		if !ok {
			svc.cron.Remove(id)
		}
	}
}

func (svc *Service) getEntryIdsMap() (res map[cron.EntryID]bool) {
	res = map[cron.EntryID]bool{}
	for _, e := range svc.cron.Entries() {
		res[e.ID] = false
	}
	return res
}

func (svc *Service) fetch() (err error) {
	query := bson.M{
		"enabled": true,
	}
	svc.schedules, err = svc.modelSvc.GetMany(query, nil)
	if err != nil {
		return err
	}
	return nil
}

func (svc *Service) schedule(id primitive.ObjectID) (fn func()) {
	return func() {
		// schedule
		s, err := svc.modelSvc.GetById(id)
		if err != nil {
			svc.Errorf("failed to get schedule: %v", err)
			return
		}

		// spider
		spider, err := service.NewModelService[models.Spider]().GetById(s.SpiderId)
		if err != nil {
			svc.Errorf("failed to get spider: %v", err)
			return
		}

		// options
		opts := &interfaces.SpiderRunOptions{
			Mode:       s.Mode,
			NodeIds:    s.NodeIds,
			Cmd:        s.Cmd,
			Param:      s.Param,
			Priority:   s.Priority,
			ScheduleId: s.Id,
			UserId:     s.GetCreatedBy(),
		}

		// normalize options
		if opts.Mode == "" {
			opts.Mode = spider.Mode
		}
		if len(opts.NodeIds) == 0 {
			opts.NodeIds = spider.NodeIds
		}
		if opts.Cmd == "" {
			opts.Cmd = spider.Cmd
		}
		if opts.Param == "" {
			opts.Param = spider.Param
		}
		if opts.Priority == 0 {
			if spider.Priority > 0 {
				opts.Priority = spider.Priority
			} else {
				opts.Priority = 5
			}
		}

		// schedule or assign a task in the task queue
		if _, err := svc.adminSvc.Schedule(s.SpiderId, opts); err != nil {
			svc.Errorf("failed to schedule spider: %v", err)
			return
		}
	}
}

func newScheduleService() *Service {
	// service
	svc := &Service{
		loc: time.Local,
		// TODO: implement delay and skip
		delay:          false,
		skip:           false,
		updateInterval: 1 * time.Minute,
		adminSvc:       admin.GetSpiderAdminService(),
		modelSvc:       service.NewModelService[models.Schedule](),
		Logger:         utils.NewLogger("ScheduleService"),
	}

	// logger
	svc.logger = NewCronLogger()

	// cron
	svc.cron = cron.New(
		cron.WithLogger(svc.logger),
		cron.WithLocation(svc.loc),
		cron.WithChain(cron.Recover(svc.logger)),
	)

	// initialize
	if err := svc.Init(); err != nil {
		panic(err)
	}

	return svc
}

var _service *Service
var _serviceOnce sync.Once

func GetScheduleService() *Service {
	_serviceOnce.Do(func() {
		_service = newScheduleService()
	})
	return _service
}
