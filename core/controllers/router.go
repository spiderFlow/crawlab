package controllers

import (
	"net/http"
	"strings"

	"github.com/crawlab-team/crawlab/core/middlewares"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/openapi"
	"github.com/gin-gonic/gin"
)

// RouterGroups defines the different authentication levels for API routes
type RouterGroups struct {
	AuthGroup      *gin.RouterGroup     // Routes requiring full authentication
	SyncAuthGroup  *gin.RouterGroup     // Routes for sync operations with special auth
	AnonymousGroup *gin.RouterGroup     // Public routes that don't require auth
	Wrapper        *openapi.FizzWrapper // OpenAPI wrapper for documentation
}

// Global variable to store the OpenAPI wrapper
// This is a workaround since we can't easily pass it through the Gin context
var globalWrapper *openapi.FizzWrapper

func GetGlobalFizzWrapper() *openapi.FizzWrapper {
	return globalWrapper
}

// NewRouterGroups initializes the router groups with their respective middleware
func NewRouterGroups(app *gin.Engine) (groups *RouterGroups) {
	// Create OpenAPI wrapper
	globalWrapper = openapi.NewFizzWrapper(app)

	return &RouterGroups{
		AuthGroup:      app.Group("/", middlewares.AuthorizationMiddleware()),
		SyncAuthGroup:  app.Group("/", middlewares.SyncAuthorizationMiddleware()),
		AnonymousGroup: app.Group("/"),
		Wrapper:        globalWrapper,
	}
}

// RegisterController registers a generic controller with standard CRUD endpoints
// and any additional custom actions
func RegisterController[T any](group *gin.RouterGroup, basePath string, ctr *BaseController[T]) {
	// Track registered paths to avoid duplicates
	actionPaths := make(map[string]bool)
	for _, action := range ctr.actions {
		path := basePath + action.Path
		key := action.Method + " - " + path
		actionPaths[key] = true

		// Create appropriate model response based on the action
		responses := globalWrapper.BuildModelResponse()

		id := getIDForAction(action.Method, path)
		summary := getSummaryForAction(action.Method, basePath, action.Path)
		description := getDescriptionForAction(action.Method, basePath, action.Path)

		globalWrapper.RegisterRoute(action.Method, path, action.HandlerFunc, id, summary, description, responses)
	}

	// Register built-in handlers if they haven't been overridden
	// Create a zero value of T to use as the model
	var model T
	registerBuiltinHandler(group, globalWrapper, basePath, http.MethodGet, "", ctr.GetList, actionPaths, "Get list", "Get a list of items", model)
	registerBuiltinHandler(group, globalWrapper, basePath, http.MethodGet, "/:id", ctr.GetById, actionPaths, "Get by ID", "Get a single item by ID", model)
	registerBuiltinHandler(group, globalWrapper, basePath, http.MethodPost, "", ctr.Post, actionPaths, "Create", "Create a new item", model)
	registerBuiltinHandler(group, globalWrapper, basePath, http.MethodPut, "/:id", ctr.PutById, actionPaths, "Update by ID", "Update an item by ID", model)
	registerBuiltinHandler(group, globalWrapper, basePath, http.MethodPatch, "", ctr.PatchList, actionPaths, "Patch list", "Patch multiple items", model)
	registerBuiltinHandler(group, globalWrapper, basePath, http.MethodDelete, "/:id", ctr.DeleteById, actionPaths, "Delete by ID", "Delete an item by ID", model)
	registerBuiltinHandler(group, globalWrapper, basePath, http.MethodDelete, "", ctr.DeleteList, actionPaths, "Delete list", "Delete multiple items", model)
}

// RegisterActions registers a list of custom action handlers to a route group
func RegisterActions(group *gin.RouterGroup, basePath string, actions []Action) {
	for _, action := range actions {
		path := basePath + action.Path

		// Create generic response
		responses := globalWrapper.BuildModelResponse()

		id := getIDForAction(action.Method, path)
		summary := getSummaryForAction(action.Method, basePath, action.Path)
		description := getDescriptionForAction(action.Method, basePath, action.Path)

		globalWrapper.RegisterRoute(action.Method, path, action.HandlerFunc, id, summary, description, responses)
	}
}

// registerBuiltinHandler registers a standard handler if it hasn't been overridden
// by a custom action
func registerBuiltinHandler[T any](group *gin.RouterGroup, wrapper *openapi.FizzWrapper, basePath, method, pathSuffix string, handlerFunc interface{}, existingActionPaths map[string]bool, summary, description string, model T) {
	path := basePath + pathSuffix
	key := method + " - " + path
	_, ok := existingActionPaths[key]
	if ok {
		return
	}

	id := getIDForAction(method, path)

	// Create appropriate response based on the method
	responses := wrapper.BuildModelResponse()

	wrapper.RegisterRoute(method, path, handlerFunc, id, summary, description, responses)
}

// Helper functions to generate OpenAPI documentation
func getIDForAction(method, path string) string {
	// Remove leading slash and convert remaining slashes to underscores
	cleanPath := strings.TrimPrefix(path, "/")
	cleanPath = strings.ReplaceAll(cleanPath, "/", "_")
	cleanPath = strings.ReplaceAll(cleanPath, ":", "_")
	cleanPath = strings.ReplaceAll(cleanPath, "__", "_")

	return method + "_" + cleanPath
}

func getSummaryForAction(method, basePath, path string) string {
	resource := getResourceName(basePath)

	switch method {
	case http.MethodGet:
		if path == "" {
			return "List " + resource
		} else if path == "/:id" {
			return "Get " + resource + " by ID"
		}
	case http.MethodPost:
		if path == "" {
			return "Create " + resource
		}
	case http.MethodPut:
		if path == "/:id" {
			return "Update " + resource + " by ID"
		}
	case http.MethodPatch:
		if path == "" {
			return "Patch " + resource + " list"
		}
	case http.MethodDelete:
		if path == "/:id" {
			return "Delete " + resource + " by ID"
		} else if path == "" {
			return "Delete " + resource + " list"
		}
	}

	// For custom actions, use a more descriptive summary
	if path != "" && path != "/:id" {
		return method + " " + resource + path
	}

	return method + " " + resource
}

func getDescriptionForAction(method, basePath, path string) string {
	resource := getResourceName(basePath)

	switch method {
	case http.MethodGet:
		if path == "" {
			return "Get a list of " + resource + " items"
		} else if path == "/:id" {
			return "Get a single " + resource + " by ID"
		}
	case http.MethodPost:
		if path == "" {
			return "Create a new " + resource
		}
	case http.MethodPut:
		if path == "/:id" {
			return "Update a " + resource + " by ID"
		}
	case http.MethodPatch:
		if path == "" {
			return "Patch multiple " + resource + " items"
		}
	case http.MethodDelete:
		if path == "/:id" {
			return "Delete a " + resource + " by ID"
		} else if path == "" {
			return "Delete multiple " + resource + " items"
		}
	}

	// For custom actions, use a more descriptive description
	if path != "" && path != "/:id" {
		return "Perform " + method + " operation on " + resource + " with path " + path
	}

	return "Perform " + method + " operation on " + resource
}

func getResourceName(basePath string) string {
	// Remove leading slash and get the last part of the path
	if len(basePath) > 0 && basePath[0] == '/' {
		basePath = basePath[1:]
	}

	// Remove trailing slash if present
	if len(basePath) > 0 && basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}

	// If path is empty, return "resource"
	if basePath == "" {
		return "resource"
	}

	return basePath
}

// InitRoutes configures all API routes for the application
func InitRoutes(app *gin.Engine) (err error) {
	// Initialize route groups with different auth levels
	groups := NewRouterGroups(app)

	// Store the wrapper in the global variable for later use
	globalWrapper = groups.Wrapper

	// Register resource controllers with their respective endpoints
	// Each RegisterController call sets up standard CRUD operations
	// Additional custom actions can be specified in the controller initialization
	RegisterController(groups.AuthGroup, "/data/collections", NewController[models.DataCollection]())
	RegisterController(groups.AuthGroup, "/environments", NewController[models.Environment]())
	RegisterController(groups.AuthGroup, "/nodes", NewController[models.Node]())
	RegisterController(groups.AuthGroup, "/projects", NewController[models.Project]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetProjectList,
		},
	}...))
	RegisterController(groups.AuthGroup, "/spiders", NewController[models.Spider]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id",
			HandlerFunc: GetSpiderById,
		},
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetSpiderList,
		},
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostSpider,
		},
		{
			Method:      http.MethodPut,
			Path:        "/:id",
			HandlerFunc: PutSpiderById,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/:id",
			HandlerFunc: DeleteSpiderById,
		},
		{
			Method:      http.MethodDelete,
			Path:        "",
			HandlerFunc: DeleteSpiderList,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/files/list",
			HandlerFunc: GetSpiderListDir,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/files/get",
			HandlerFunc: GetSpiderFile,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/files/info",
			HandlerFunc: GetSpiderFileInfo,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/save",
			HandlerFunc: PostSpiderSaveFile,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/save/batch",
			HandlerFunc: PostSpiderSaveFiles,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/save/dir",
			HandlerFunc: PostSpiderSaveDir,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/rename",
			HandlerFunc: PostSpiderRenameFile,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/:id/files",
			HandlerFunc: DeleteSpiderFile,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/copy",
			HandlerFunc: PostSpiderCopyFile,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/export",
			HandlerFunc: PostSpiderExport,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/run",
			HandlerFunc: PostSpiderRun,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/results",
			HandlerFunc: GetSpiderResults,
		},
	}...))
	groups.AnonymousGroup.GET("/openapi.json", GetOpenAPI)
	return
	RegisterController(groups.AuthGroup, "/schedules", NewController[models.Schedule]([]Action{
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostSchedule,
		},
		{
			Method:      http.MethodPut,
			Path:        "/:id",
			HandlerFunc: PutScheduleById,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/enable",
			HandlerFunc: PostScheduleEnable,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/disable",
			HandlerFunc: PostScheduleDisable,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/run",
			HandlerFunc: PostScheduleRun,
		},
	}...))
	RegisterController(groups.AuthGroup, "/tasks", NewController[models.Task]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id",
			HandlerFunc: GetTaskById,
		},
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetTaskList,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/:id",
			HandlerFunc: DeleteTaskById,
		},
		{
			Method:      http.MethodDelete,
			Path:        "",
			HandlerFunc: DeleteList,
		},
		{
			Method:      http.MethodPost,
			Path:        "/run",
			HandlerFunc: PostTaskRun,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/restart",
			HandlerFunc: PostTaskRestart,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/cancel",
			HandlerFunc: PostTaskCancel,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/logs",
			HandlerFunc: GetTaskLogs,
		},
	}...))
	RegisterController(groups.AuthGroup, "/tokens", NewController[models.Token]([]Action{
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostToken,
		},
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetTokenList,
		},
	}...))
	RegisterController(groups.AuthGroup, "/users", NewController[models.User]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id",
			HandlerFunc: GetUserById,
		},
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetUserList,
		},
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostUser,
		},
		{
			Method:      http.MethodPut,
			Path:        "/:id",
			HandlerFunc: PutUserById,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/change-password",
			HandlerFunc: PostUserChangePassword,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/:id",
			HandlerFunc: DeleteUserById,
		},
		{
			Method:      http.MethodDelete,
			Path:        "",
			HandlerFunc: DeleteUserList,
		},
		{
			Method:      http.MethodGet,
			Path:        "/me",
			HandlerFunc: GetUserMe,
		},
		{
			Method:      http.MethodPut,
			Path:        "/me",
			HandlerFunc: PutUserMe,
		},
		{
			Method:      http.MethodPost,
			Path:        "/me/change-password",
			HandlerFunc: PostUserMeChangePassword,
		},
	}...))

	// Register standalone action routes that don't fit the standard CRUD pattern
	RegisterActions(groups.AuthGroup, "/export", []Action{
		{
			Method:      http.MethodPost,
			Path:        "/:type",
			HandlerFunc: PostExport,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:type/:id",
			HandlerFunc: GetExport,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:type/:id/download",
			HandlerFunc: GetExportDownload,
		},
	})
	RegisterActions(groups.AuthGroup, "/filters", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/:col",
			HandlerFunc: GetFilterColFieldOptions,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:col/:value",
			HandlerFunc: GetFilterColFieldOptions,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:col/:value/:label",
			HandlerFunc: GetFilterColFieldOptions,
		},
	})
	RegisterActions(groups.AuthGroup, "/settings", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/:key",
			HandlerFunc: GetSetting,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:key",
			HandlerFunc: PostSetting,
		},
		{
			Method:      http.MethodPut,
			Path:        "/:key",
			HandlerFunc: PutSetting,
		},
	})
	RegisterActions(groups.AuthGroup, "/stats", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/overview",
			HandlerFunc: GetStatsOverview,
		},
		{
			Method:      http.MethodGet,
			Path:        "/daily",
			HandlerFunc: GetStatsDaily,
		},
		{
			Method:      http.MethodGet,
			Path:        "/tasks",
			HandlerFunc: GetStatsTasks,
		},
	})

	// Register sync routes that require special authentication
	RegisterActions(groups.SyncAuthGroup, "/sync", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id/scan",
			HandlerFunc: GetSyncScan,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/download",
			HandlerFunc: GetSyncDownload,
		},
	})

	// Register public routes that don't require authentication
	RegisterActions(groups.AnonymousGroup, "/health", []Action{
		{
			Path:        "",
			Method:      http.MethodGet,
			HandlerFunc: GetHealthFn(func() bool { return true }),
		},
	})
	RegisterActions(groups.AnonymousGroup, "/system-info", []Action{
		{
			Path:        "",
			Method:      http.MethodGet,
			HandlerFunc: GetSystemInfo,
		},
	})
	RegisterActions(groups.AnonymousGroup, "/", []Action{
		{
			Method:      http.MethodPost,
			Path:        "/login",
			HandlerFunc: PostLogin,
		},
		{
			Method:      http.MethodPost,
			Path:        "/logout",
			HandlerFunc: PostLogout,
		},
	})

	// Register OpenAPI documentation route
	groups.AnonymousGroup.GET("/openapi.json", GetOpenAPI)

	return nil
}
