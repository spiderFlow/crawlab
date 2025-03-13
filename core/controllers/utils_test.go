package controllers_test

import (
	"testing"

	"github.com/crawlab-team/crawlab/core/controllers"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetFilterFromConditionString(t *testing.T) {
	// Simple condition with string value
	condStr := `[{"key":"name","op":"eq","value":"test"}]`
	filter, err := controllers.GetFilterFromConditionString(condStr)
	require.NoError(t, err)
	require.NotNil(t, filter)
	require.Len(t, filter.Conditions, 1)
	assert.Equal(t, "name", filter.Conditions[0].Key)
	assert.Equal(t, "eq", filter.Conditions[0].Op)
	assert.Equal(t, "test", filter.Conditions[0].Value)

	// Multiple conditions with different types
	condStr = `[{"key":"name","op":"eq","value":"test"},{"key":"priority","op":"gt","value":5}]`
	filter, err = controllers.GetFilterFromConditionString(condStr)
	require.NoError(t, err)
	require.NotNil(t, filter)
	require.Len(t, filter.Conditions, 2)
	assert.Equal(t, "name", filter.Conditions[0].Key)
	assert.Equal(t, "eq", filter.Conditions[0].Op)
	assert.Equal(t, "test", filter.Conditions[0].Value)
	assert.Equal(t, "priority", filter.Conditions[1].Key)
	assert.Equal(t, "gt", filter.Conditions[1].Op)
	assert.Equal(t, float64(5), filter.Conditions[1].Value) // JSON parses numbers as float64

	// Invalid JSON should return error
	condStr = `[{"key":"name","op":"eq","value":"test"`
	_, err = controllers.GetFilterFromConditionString(condStr)
	assert.Error(t, err)
}

func TestGetFilterQueryFromConditionString(t *testing.T) {
	// Simple equality condition
	condStr := `[{"key":"name","op":"eq","value":"test"}]`
	query, err := controllers.GetFilterQueryFromConditionString(condStr)
	require.NoError(t, err)
	require.NotNil(t, query)
	expected := bson.M{"name": "test"}
	assert.Equal(t, expected, query)

	// Greater than condition
	condStr = `[{"key":"priority","op":"gt","value":5}]`
	query, err = controllers.GetFilterQueryFromConditionString(condStr)
	require.NoError(t, err)
	require.NotNil(t, query)
	expected = bson.M{"priority": bson.M{"$gt": float64(5)}}
	assert.Equal(t, expected, query)

	// Multiple conditions
	condStr = `[{"key":"name","op":"eq","value":"test"},{"key":"priority","op":"gt","value":5}]`
	query, err = controllers.GetFilterQueryFromConditionString(condStr)
	require.NoError(t, err)
	require.NotNil(t, query)
	expected = bson.M{"name": "test", "priority": bson.M{"$gt": float64(5)}}
	assert.Equal(t, expected, query)

	// Contains operator
	condStr = `[{"key":"name","op":"contains","value":"test"}]`
	query, err = controllers.GetFilterQueryFromConditionString(condStr)
	require.NoError(t, err)
	require.NotNil(t, query)
	expectedRegex := bson.M{"name": bson.M{"$regex": "test", "$options": "i"}}
	assert.Equal(t, expectedRegex, query)

	// Invalid condition should return error
	condStr = `[{"key":"name","op":"invalid_op","value":"test"}]`
	_, err = controllers.GetFilterQueryFromConditionString(condStr)
	assert.Error(t, err)
}

func TestGetFilterQueryFromListParams(t *testing.T) {
	// No conditions
	params := &controllers.GetListParams{}
	query, err := controllers.GetFilterQueryFromListParams(params)
	require.NoError(t, err)
	assert.Nil(t, query)

	// With conditions
	params.Conditions = `[{"key":"name","op":"eq","value":"test"}]`
	query, err = controllers.GetFilterQueryFromListParams(params)
	require.NoError(t, err)
	require.NotNil(t, query)
	expected := bson.M{"name": "test"}
	assert.Equal(t, expected, query)
}

func TestGetUserFromContext(t *testing.T) {
	// Empty context should return nil
	c := &gin.Context{}
	user := controllers.GetUserFromContext(c)
	assert.Nil(t, user)

	// Context with non-user value should return nil
	c = &gin.Context{}
	c.Set("user", "not a user object")
	user = controllers.GetUserFromContext(c)
	assert.Nil(t, user)

	// Context with user should return the user
	c = &gin.Context{}
	expectedUser := &models.User{Username: "test_user"}
	expectedUser.Id = primitive.NewObjectID()
	c.Set("user", expectedUser)
	user = controllers.GetUserFromContext(c)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Id, user.Id)
	assert.Equal(t, expectedUser.Username, user.Username)
}

func TestGetErrorResponse(t *testing.T) {
	// Error response test
	err := assert.AnError
	resp, _ := controllers.GetErrorResponse[models.Task](err)
	assert.Equal(t, err.Error(), resp.Error)
	assert.Equal(t, models.Task{}, resp.Data)
}

func TestGetDataResponse(t *testing.T) {
	// Data response test
	task := models.Task{
		Status: "running",
		Cmd:    "python main.py",
		Param:  "test param",
	}
	task.Id = primitive.NewObjectID()

	resp, err := controllers.GetDataResponse(task)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, task, resp.Data)
	assert.Empty(t, resp.Error)
}

func TestGetListResponse(t *testing.T) {
	// List response test
	tasks := []models.Task{
		{
			Status: "running",
			Cmd:    "python main.py",
		},
		{
			Status: "pending",
			Cmd:    "python main.py",
		},
	}
	tasks[0].Id = primitive.NewObjectID()
	tasks[1].Id = primitive.NewObjectID()

	total := 2
	resp, err := controllers.GetListResponse(tasks, total)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, tasks, resp.Data)
	assert.Equal(t, total, resp.Total)
	assert.Empty(t, resp.Error)
}

func TestGetErrorListResponse(t *testing.T) {
	// Error list response test
	err := assert.AnError
	resp, _ := controllers.GetErrorListResponse[models.Task](err)
	assert.Equal(t, err.Error(), resp.Error)
	assert.Nil(t, resp.Data)
	assert.Equal(t, 0, resp.Total)
}
