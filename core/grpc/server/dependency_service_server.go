package server

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	mongo2 "github.com/crawlab-team/crawlab/db/mongo"
	"github.com/crawlab-team/crawlab/grpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	<-stream.Context().Done()
	log.Info("[DependencyServiceServer] disconnected: " + req.NodeKey)

	return nil
}

// Sync handles synchronization of dependencies between nodes and the database
func (svr DependencyServiceServer) Sync(_ context.Context, request *grpc.DependencyServiceSyncRequest) (response *grpc.Response, err error) {
	// Get node by node key
	n, err := service.NewModelService[models.Node]().GetOne(bson.M{"key": request.NodeKey}, nil)
	if err != nil {
		return nil, err
	}

	// Get existing dependencies from database for this node and language
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

	// Create map of existing dependencies for faster lookup
	depsDbMap := make(map[string]*models.Dependency)
	for _, d := range depsDb {
		depsDbMap[d.Name] = &d
	}

	// Process new dependencies from request
	var depsToInsert []models.Dependency
	var depsToUpdate []models.Dependency
	depsMap := make(map[string]*models.Dependency)
	for _, dep := range request.Dependencies {
		// Create dependency model
		d := models.Dependency{
			Name:    dep.Name,
			NodeId:  n.Id,
			Type:    request.Lang,
			Version: dep.Version,
			Status:  constants.DependencyStatusInstalled,
		}
		d.SetCreatedAt(time.Now())
		d.SetUpdatedAt(time.Now())

		// Add to map
		depsMap[d.Name] = &d

		// Check if dependency exists in DB
		existingDep, ok := depsDbMap[d.Name]
		if !ok {
			// If dependency doesn't exist, add to insert list
			depsToInsert = append(depsToInsert, d)
		} else if existingDep.Version != d.Version || existingDep.Status != constants.DependencyStatusInstalled {
			// If dependency exists but version is different or status is not installed, add to update list
			d.Id = existingDep.Id
			d.SetUpdatedAt(time.Now())
			depsToUpdate = append(depsToUpdate, d)
		}
	}

	// Find dependencies to delete (exist in DB but not in request)
	var depIdsToDelete []primitive.ObjectID
	for _, d := range depsDb {
		_, ok := depsMap[d.Name]
		if !ok {
			depIdsToDelete = append(depIdsToDelete, d.Id)
		}
	}

	// Run database operations in a transaction
	err = mongo2.RunTransaction(func(ctx mongo.SessionContext) (err error) {
		// Delete old dependencies if any
		if len(depIdsToDelete) > 0 {
			err = service.NewModelService[models.Dependency]().DeleteMany(bson.M{
				"_id": bson.M{"$in": depIdsToDelete},
			})
			if err != nil {
				log.Errorf("[DependencyServiceServer] delete dependencies in db error: %v", err)
				return err
			}
		}

		// Insert new dependencies if any
		if len(depsToInsert) > 0 {
			_, err = service.NewModelService[models.Dependency]().InsertMany(depsToInsert)
			if err != nil {
				log.Errorf("[DependencyServiceServer] insert dependencies in db error: %v", err)
				return err
			}
		}

		// Update dependencies with different versions
		for _, d := range depsToUpdate {
			err = service.NewModelService[models.Dependency]().ReplaceById(d.Id, d)
			if err != nil {
				log.Errorf("[DependencyServiceServer] update dependency in db error: %v", err)
				return err
			}
		}

		return nil
	})

	return nil, err
}

func (svr DependencyServiceServer) UpdateLogs(stream grpc.DependencyService_UpdateLogsServer) (err error) {
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

		// if dependency is nil, get dependency
		if dep == nil {
			id, err := primitive.ObjectIDFromHex(req.DependencyId)
			if err != nil {
				log.Errorf("[DependencyServiceServer] convert dependency id error: %v", err)
				return err
			}
			dep, err = service.NewModelService[models.Dependency]().GetById(id)
			if err != nil {
				log.Errorf("[DependencyServiceServer] get dependency error: %v", err)
				return err
			}
		}

		// insert dependency logs
		var depLogs []models.DependencyLog
		for _, line := range req.Logs {
			depLog := models.DependencyLog{
				DependencyId: dep.Id,
				Content:      line,
			}
			depLogs = append(depLogs, depLog)
		}
		_, err = service.NewModelService[models.DependencyLog]().InsertMany(depLogs)
		if err != nil {
			log.Errorf("[DependencyServiceServer] insert dependency logs error: %v", err)
			return err
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
