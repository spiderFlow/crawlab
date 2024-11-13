package controllers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crawlab-team/crawlab/core/controllers"
	"github.com/crawlab-team/crawlab/core/middlewares"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/crawlab-team/crawlab/core/utils"
)

func TestGetUserById_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	// Create test user with required fields
	modelSvc := service.NewModelService[models.User]()
	u := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: utils.EncryptMd5("testpassword"), // Add password
	}
	id, err := modelSvc.InsertOne(u)
	require.Nil(t, err)
	u.SetId(id)

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.GET("/users/:id", controllers.GetUserById)

	// Test valid ID
	req, err := http.NewRequest(http.MethodGet, "/users/"+id.Hex(), nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test invalid ID format
	req, err = http.NewRequest(http.MethodGet, "/users/invalid-id", nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserList_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	modelSvc := service.NewModelService[models.User]()

	// Create test users with required fields
	users := []models.User{
		{Username: "user1", Email: "user1@example.com", Password: utils.EncryptMd5("password1")},
		{Username: "user2", Email: "user2@example.com", Password: utils.EncryptMd5("password2")},
		{Username: "user3", Email: "user3@example.com", Password: utils.EncryptMd5("password3")},
	}

	for _, u := range users {
		_, err := modelSvc.InsertOne(u)
		assert.Nil(t, err)
	}

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.GET("/users", controllers.GetUserList)

	// Test default pagination
	req, err := http.NewRequest(http.MethodGet, "/users", nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with pagination parameters
	req, err = http.NewRequest(http.MethodGet, "/users?page=1&size=2", nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostUser_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.POST("/users", controllers.PostUser)

	// Test creating a new user with valid data
	reqBody := strings.NewReader(`{
		"username": "newuser",
		"password": "password123",
		"email": "newuser@example.com"
	}`)
	req, err := http.NewRequest(http.MethodPost, "/users", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user was created
	modelSvc := service.NewModelService[models.User]()
	u, err := modelSvc.GetOne(bson.M{"username": "newuser"}, nil)
	assert.Nil(t, err)
	assert.Equal(t, "newuser", u.Username)
	assert.Equal(t, "newuser@example.com", u.Email)

	// Test creating a user with invalid data
	reqBody = strings.NewReader(`{
		"username": "",
		"password": "",
		"email": "invalid-email"
	}`)
	req, err = http.NewRequest(http.MethodPost, "/users", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPutUserById_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	modelSvc := service.NewModelService[models.User]()
	u := models.User{}
	id, err := modelSvc.InsertOne(u)
	require.Nil(t, err)
	u.SetId(id)

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.PUT("/users/:id", controllers.PutUserById)

	// Test case 1: Regular user update
	reqBody := strings.NewReader(`{
		"id":"` + id.Hex() + `",
		"username":"newUsername",
		"email":"newEmail@test.com"
	}`)
	req, _ := http.NewRequest(http.MethodPut, "/users/"+id.Hex(), reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	// Make request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test case 2: Root admin user update (should not change username)
	u.RootAdmin = true
	err = modelSvc.ReplaceById(id, u)
	assert.Nil(t, err)

	reqBody = strings.NewReader(`{
		"id":"` + id.Hex() + `",
		"username":"attemptedNewUsername",
		"email":"newEmail@test.com"
	}`)
	req, _ = http.NewRequest(http.MethodPut, "/users/"+id.Hex(), reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify username wasn't changed for root admin
	updatedUser, err := modelSvc.GetById(id)
	assert.Nil(t, err)
	assert.NotEqual(t, "attemptedNewUsername", updatedUser.Username)
}

func TestPostUserChangePassword_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	modelSvc := service.NewModelService[models.User]()
	u := models.User{}
	id, err := modelSvc.InsertOne(u)
	require.Nil(t, err)
	u.SetId(id)

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.POST("/users/:id/change-password", controllers.PostUserChangePassword)

	// Add validation for minimum password length
	// Test case 1: Valid password
	password := "validPassword123"
	reqBody := strings.NewReader(`{"password":"` + password + `"}`)
	req, _ := http.NewRequest(http.MethodPost, "/users/"+id.Hex()+"/change-password", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test case 2: Password too short
	shortPassword := "1234"
	reqBody = strings.NewReader(`{"password":"` + shortPassword + `"}`)
	req, _ = http.NewRequest(http.MethodPost, "/users/"+id.Hex()+"/change-password", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserMe_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	modelSvc := service.NewModelService[models.User]()
	u := models.User{}
	id, err := modelSvc.InsertOne(u)
	require.Nil(t, err)
	u.SetId(id)

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.GET("/users/me", controllers.GetUserMe)

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPutUserMe_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	// Create test user with required fields
	modelSvc := service.NewModelService[models.User]()
	u := models.User{
		Username: "originaluser",
		Email:    "original@example.com",
		Password: utils.EncryptMd5("testpassword"),
	}
	id, err := modelSvc.InsertOne(u)
	require.Nil(t, err)
	u.SetId(id)

	// Create token for user
	userSvc, err := user.GetUserService()
	require.Nil(t, err)
	token, err := userSvc.MakeToken(&u)
	require.Nil(t, err)

	// Create router
	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.PUT("/users/me", controllers.PutUserMe)

	// Test valid update
	reqBody := strings.NewReader(`{
		"username": "updateduser",
		"email": "updated@example.com"
	}`)
	req, err := http.NewRequest(http.MethodPut, "/users/me", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the update
	updatedUser, err := modelSvc.GetById(id)
	assert.Nil(t, err)
	assert.Equal(t, "updateduser", updatedUser.Username)
	assert.Equal(t, "updated@example.com", updatedUser.Email)

	// Verify password wasn't changed
	assert.Equal(t, utils.EncryptMd5("testpassword"), updatedUser.Password)
}

func TestPostUserMeChangePassword_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	// Create test user with initial password
	modelSvc := service.NewModelService[models.User]()
	u := models.User{
		Username: "testuser",
		Password: utils.EncryptMd5("initialpassword"),
		Email:    "test@example.com",
	}
	id, err := modelSvc.InsertOne(u)
	require.Nil(t, err)
	u.SetId(id)

	// Create token for user
	userSvc, err := user.GetUserService()
	require.Nil(t, err)
	token, err := userSvc.MakeToken(&u)
	require.Nil(t, err)

	// Create router
	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.POST("/users/me/change-password", controllers.PostUserMeChangePassword)

	// Test valid password change
	password := "newValidPassword123"
	reqBody := strings.NewReader(`{"password":"` + password + `"}`)
	req, err := http.NewRequest(http.MethodPost, "/users/me/change-password", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	// Make request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify password was changed
	updatedUser, err := modelSvc.GetById(id)
	assert.Nil(t, err)
	assert.Equal(t, utils.EncryptMd5(password), updatedUser.Password)

	// Test invalid password (too short)
	shortPassword := "123"
	reqBody = strings.NewReader(`{"password":"` + shortPassword + `"}`)
	req, err = http.NewRequest(http.MethodPost, "/users/me/change-password", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserById_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	// Create test user
	modelSvc := service.NewModelService[models.User]()
	u := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: utils.EncryptMd5("testpassword"),
	}
	id, err := modelSvc.InsertOne(u)
	require.Nil(t, err)
	u.SetId(id)

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.DELETE("/users/:id", controllers.DeleteUserById)

	// Test deleting normal user
	req, err := http.NewRequest(http.MethodDelete, "/users/"+id.Hex(), nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user was deleted
	_, err = modelSvc.GetById(id)
	assert.NotNil(t, err)

	// Test deleting root admin user
	rootAdmin := models.User{
		Username:  "rootadmin",
		Email:     "root@example.com",
		Password:  utils.EncryptMd5("rootpass"),
		RootAdmin: true,
	}
	rootId, err := modelSvc.InsertOne(rootAdmin)
	require.Nil(t, err)

	req, err = http.NewRequest(http.MethodDelete, "/users/"+rootId.Hex(), nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test deleting with invalid ID
	req, err = http.NewRequest(http.MethodDelete, "/users/invalid-id", nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteUserList_Success(t *testing.T) {
	SetupTestDB()
	defer CleanupTestDB()

	modelSvc := service.NewModelService[models.User]()

	// Create test users
	users := []models.User{
		{Username: "user1", Email: "user1@example.com", Password: utils.EncryptMd5("pass1")},
		{Username: "user2", Email: "user2@example.com", Password: utils.EncryptMd5("pass2")},
		{Username: "rootadmin", Email: "root@example.com", Password: utils.EncryptMd5("rootpass"), RootAdmin: true},
	}

	var userIds []primitive.ObjectID
	var normalUserIds []primitive.ObjectID
	for _, user := range users {
		id, err := modelSvc.InsertOne(user)
		require.Nil(t, err)
		userIds = append(userIds, id)
		if !user.RootAdmin {
			normalUserIds = append(normalUserIds, id)
		}
	}

	router := gin.Default()
	router.Use(middlewares.AuthorizationMiddleware())
	router.DELETE("/users", controllers.DeleteUserList)

	// Test deleting normal users
	reqBody := strings.NewReader(fmt.Sprintf(`{"ids":["%s","%s"]}`,
		normalUserIds[0].Hex(),
		normalUserIds[1].Hex()))
	req, err := http.NewRequest(http.MethodDelete, "/users", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify users were deleted
	for _, id := range normalUserIds {
		_, err = modelSvc.GetById(id)
		assert.NotNil(t, err)
	}

	// Test attempting to delete list including root admin
	reqBody = strings.NewReader(fmt.Sprintf(`{"ids":["%s"]}`, userIds[2].Hex()))
	req, err = http.NewRequest(http.MethodDelete, "/users", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test with mix of valid and invalid ids
	reqBody = strings.NewReader(fmt.Sprintf(`{"ids":["%s","invalid-id"]}`, normalUserIds[0].Hex()))
	req, err = http.NewRequest(http.MethodDelete, "/users", reqBody)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", TestToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
