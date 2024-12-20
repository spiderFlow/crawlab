package server

import (
	"context"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/grpc"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
	"time"
)

type MetricServiceServer struct {
	grpc.UnimplementedMetricServiceServer
}

func (svr MetricServiceServer) Send(_ context.Context, req *grpc.MetricServiceSendRequest) (res *grpc.Response, err error) {
	log.Debugf("[MetricServiceServer] received metric from node: " + req.NodeKey)
	n, err := service.NewModelService[models.Node]().GetOne(bson.M{"key": req.NodeKey}, nil)
	if err != nil {
		log.Errorf("[MetricServiceServer] error getting node: %v", err)
		return HandleError(err)
	}
	metric := models.Metric{
		Type:                 req.Type,
		NodeId:               n.Id,
		CpuUsagePercent:      req.CpuUsagePercent,
		TotalMemory:          req.TotalMemory,
		AvailableMemory:      req.AvailableMemory,
		UsedMemory:           req.UsedMemory,
		UsedMemoryPercent:    req.UsedMemoryPercent,
		TotalDisk:            req.TotalDisk,
		AvailableDisk:        req.AvailableDisk,
		UsedDisk:             req.UsedDisk,
		UsedDiskPercent:      req.UsedDiskPercent,
		DiskReadBytesRate:    req.DiskReadBytesRate,
		DiskWriteBytesRate:   req.DiskWriteBytesRate,
		NetworkBytesSentRate: req.NetworkBytesSentRate,
		NetworkBytesRecvRate: req.NetworkBytesRecvRate,
	}
	metric.CreatedAt = time.Unix(req.Timestamp, 0)
	_, err = service.NewModelService[models.Metric]().InsertOne(metric)
	if err != nil {
		log.Errorf("[MetricServiceServer] error inserting metric: %v", err)
		return HandleError(err)
	}
	return HandleSuccess()
}

func newMetricsServer() *MetricServiceServer {
	return &MetricServiceServer{}
}

var metricsServer *MetricServiceServer
var metricsServerOnce = &sync.Once{}

func GetMetricsServer() *MetricServiceServer {
	metricsServerOnce.Do(func() {
		metricsServer = newMetricsServer()
	})
	return metricsServer
}
