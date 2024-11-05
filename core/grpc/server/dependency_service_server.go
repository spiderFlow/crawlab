package server

import (
	"context"
	"errors"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	mongo2 "github.com/crawlab-team/crawlab/db/mongo"
	"github.com/crawlab-team/crawlab/grpc"
	"github.com/crawlab-team/crawlab/trace"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"sync"
	"time"
)

type DependencyServiceServer struct {
	grpc.UnimplementedDependencyServiceServer
	mu      *sync.Mutex
	streams map[string]*grpc.DependencyService_ConnectServer
}

func (svr DependencyServiceServer) Connect(req *grpc.DependencyServiceConnectRequest, stream grpc.DependencyService_ConnectServer) (err error) {
	svr.mu.Lock()
	svr.streams[req.NodeKey] = &stream
	svr.mu.Unlock()
	log.Info("[DependencyServiceServer] connected: " + req.NodeKey)

	// Keep this scope alive because once this scope exits - the stream is closed
	for {
		select {
		case <-stream.Context().Done():
			log.Info("[DependencyServiceServer] disconnected: " + req.NodeKey)
			return nil
		}
	}
}

func (svr DependencyServiceServer) Sync(ctx context.Context, request *grpc.DependencyServiceSyncRequest) (response *grpc.Response, err error) {
	n, err := service.NewModelService[models.Node]().GetOne(bson.M{"key": request.NodeKey}, nil)
	if err != nil {
		return nil, err
	}

	depsDb, err := service.NewModelService[models.Dependency]().GetMany(bson.M{
		"node_id": n.Id,
		"type":    request.Lang,
	}, nil)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			log.Errorf("[DependencyService] get dependencies from db error: %v", err)
			return nil, err
		}
	}

	depsDbMap := make(map[string]*models.Dependency)
	for _, d := range depsDb {
		depsDbMap[d.Name] = &d
	}

	var depsToInsert []models.Dependency
	depsMap := make(map[string]*models.Dependency)
	for _, dep := range request.Dependencies {
		d := models.Dependency{
			Name:    dep.Name,
			NodeId:  n.Id,
			Type:    request.Lang,
			Version: dep.Version,
		}
		d.SetCreatedAt(time.Now())

		depsMap[d.Name] = &d

		_, ok := depsDbMap[d.Name]
		if !ok {
			depsToInsert = append(depsToInsert, d)
		}
	}

	var depIdsToDelete []primitive.ObjectID
	for _, d := range depsDb {
		_, ok := depsMap[d.Name]
		if !ok {
			depIdsToDelete = append(depIdsToDelete, d.Id)
		}
	}

	err = mongo2.RunTransaction(func(ctx mongo.SessionContext) (err error) {
		if len(depIdsToDelete) > 0 {
			err = service.NewModelService[models.Dependency]().DeleteMany(bson.M{
				"_id": bson.M{"$in": depIdsToDelete},
			})
			if err != nil {
				log.Errorf("[DependencyServiceServer] delete dependencies in db error: %v", err)
				trace.PrintError(err)
				return err
			}
		}

		if len(depsToInsert) > 0 {
			_, err = service.NewModelService[models.Dependency]().InsertMany(depsToInsert)
			if err != nil {
				log.Errorf("[DependencyServiceServer] insert dependencies in db error: %v", err)
				trace.PrintError(err)
				return err
			}
		}

		return nil
	})

	return nil, err
}

func (svr DependencyServiceServer) UpdateLogs(stream grpc.DependencyService_UpdateLogsServer) (err error) {
	var n *models.Node
	var dep *models.Dependency
	for {
		// receive message
		req, err := stream.Recv()
		if err == io.EOF {
			// all messages have been received
			return stream.SendAndClose(&grpc.Response{Message: "update task log finished"})
		}
		if err != nil {
			return err
		}

		// if node is nil, get node
		if n == nil {
			n, err = service.NewModelService[models.Node]().GetOne(bson.M{"key": req.NodeKey}, nil)
			if err != nil {
				log.Errorf("[DependencyServiceServer] get node error: %v", err)
				return err
			}
		}

		// if dependency is nil, get dependency
		if dep == nil {
			dep, err = service.NewModelService[models.Dependency]().GetOne(bson.M{
				"node_id": n.Id,
				"name":    req.Name,
				"type":    req.Lang,
			}, nil)
			if err != nil {
				if !errors.Is(err, mongo.ErrNoDocuments) {
					log.Errorf("[DependencyServiceServer] get dependency error: %v", err)
					return err
				}
				// create dependency if not found
				dep = &models.Dependency{
					NodeId: n.Id,
					Name:   req.Name,
					Type:   req.Lang,
				}
				dep.SetCreatedAt(time.Now())
				dep.SetUpdatedAt(time.Now())
				dep.Id, err = service.NewModelService[models.Dependency]().InsertOne(*dep)
				if err != nil {
					log.Errorf("[DependencyServiceServer] insert dependency error: %v", err)
					return err
				}
			}
		}

	}
}

func (svr DependencyServiceServer) GetStream(key string) (stream *grpc.DependencyService_ConnectServer, err error) {
	svr.mu.Lock()
	defer svr.mu.Unlock()
	stream, ok := svr.streams[key]
	if !ok {
		return nil, errors.New("stream not found")
	}
	return stream, nil
}

func newDependencyServer() *DependencyServiceServer {
	return &DependencyServiceServer{
		mu:      new(sync.Mutex),
		streams: make(map[string]*grpc.DependencyService_ConnectServer),
	}
}

var depSvc *DependencyServiceServer
var depSvcOnce sync.Once

func GetDependencyServer() *DependencyServiceServer {
	depSvcOnce.Do(func() {
		depSvc = newDependencyServer()
	})
	return depSvc
}
