package controllers

import (
	"time"

	"github.com/crawlab-team/crawlab/core/stats"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

var statsDefaultQuery = bson.M{
	"created_ts": bson.M{
		"$gte": time.Now().Add(-30 * 24 * time.Hour),
	},
}

type GetStatsOverviewParams struct {
	Query bson.M `json:"query"`
}

func GetStatsOverview(_ *gin.Context, params *GetStatsOverviewParams) (response *Response[bson.M], err error) {
	query := statsDefaultQuery
	if params.Query != nil {
		query = params.Query
	}

	data, err := stats.GetStatsService().GetOverviewStats(query)
	if err != nil {
		return GetErrorResponse[bson.M](err)
	}
	return GetDataResponse(data.(bson.M))
}

type GetStatsDailyParams struct {
	Query bson.M `json:"query"`
}

func GetStatsDaily(_ *gin.Context, params *GetStatsDailyParams) (response *Response[bson.M], err error) {
	query := statsDefaultQuery
	if params.Query != nil {
		query = params.Query
	}

	data, err := stats.GetStatsService().GetDailyStats(query)
	if err != nil {
		return GetErrorResponse[bson.M](err)
	}
	return GetDataResponse(data.(bson.M))
}

type GetStatsTasksParams struct {
	Query bson.M `json:"query"`
}

func GetStatsTasks(_ *gin.Context, params *GetStatsTasksParams) (response *Response[bson.M], err error) {
	query := statsDefaultQuery
	if params.Query != nil {
		query = params.Query
	}

	data, err := stats.GetStatsService().GetTaskStats(query)
	if err != nil {
		return GetErrorResponse[bson.M](err)
	}
	return GetDataResponse(data.(bson.M))
}
