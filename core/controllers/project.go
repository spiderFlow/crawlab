package controllers

import (
	"errors"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

func GetProjectList(c *gin.Context) {
	// get all list
	all := MustGetFilterAll(c)
	if all {
		NewController[models.Project]().getAll(c)
		return
	}

	// params
	pagination := MustGetPagination(c)
	query := MustGetFilterQuery(c)
	sort := MustGetSortOption(c)

	// get list
	projects, err := service.NewModelService[models.Project]().GetMany(query, &mongo.FindOptions{
		Sort:  sort,
		Skip:  pagination.Size * (pagination.Page - 1),
		Limit: pagination.Size,
	})
	if err != nil {
		if err.Error() != mongo2.ErrNoDocuments.Error() {
			HandleErrorInternalServerError(c, err)
		}
		return
	}
	if len(projects) == 0 {
		HandleSuccessWithListData(c, []models.Project{}, 0)
		return
	}

	// total count
	total, err := service.NewModelService[models.Project]().Count(query)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}

	// project ids
	var ids []primitive.ObjectID

	// count cache
	cache := map[primitive.ObjectID]int{}
	for _, p := range projects {
		ids = append(ids, p.Id)
		cache[p.Id] = 0
	}

	// spiders
	spiders, err := service.NewModelService[models.Spider]().GetMany(bson.M{
		"project_id": bson.M{
			"$in": ids,
		},
	}, nil)
	if err != nil {
		HandleErrorInternalServerError(c, err)
		return
	}
	for _, s := range spiders {
		_, ok := cache[s.ProjectId]
		if !ok {
			HandleErrorInternalServerError(c, errors.New("project id not found"))
			return
		}
		cache[s.ProjectId]++
	}

	// assign
	for i, p := range projects {
		projects[i].Spiders = cache[p.Id]
	}

	HandleSuccessWithListData(c, projects, total)
}
