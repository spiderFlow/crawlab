package controllers

import (
	"net/http"
	"strings"

	"github.com/crawlab-team/fizz"

	"github.com/crawlab-team/crawlab/core/middlewares"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/openapi"
	"github.com/gin-gonic/gin"
)

// RouterGroups defines the different authentication levels for API routes
type RouterGroups struct {
	AuthGroup      *fizz.RouterGroup    // Routes requiring full authentication
	AnonymousGroup *fizz.RouterGroup    // Public routes that don't require auth
	SyncAuthGroup  *gin.RouterGroup     // Routes for sync operations with special auth
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
	globalWrapper = openapi.GetFizzWrapper(app)

	f := globalWrapper.GetFizz()

	return &RouterGroups{
		AuthGroup:      f.Group("/", "AuthGroup", "Router group that requires authentication", middlewares.AuthorizationMiddleware()),
		AnonymousGroup: f.Group("/", "AnonymousGroup", "Router group that doesn't require authentication"),
		SyncAuthGroup:  app.Group("/", middlewares.SyncAuthorizationMiddleware()),
		Wrapper:        globalWrapper,
	}
}

// RegisterController registers a generic controller with standard CRUD endpoints
// and any additional custom actions
func RegisterController[T any](group *fizz.RouterGroup, basePath string, ctr *BaseController[T]) {
	// Track registered paths to avoid duplicates
	actionPaths := make(map[string]bool)
	for _, action := range ctr.actions {
		fullPath := basePath + action.Path
		key := action.Method + " - " + fullPath
		actionPaths[key] = true

		// Create appropriate model response based on the action
		responses := globalWrapper.BuildModelResponse()

		id := getIDForAction(action.Method, fullPath)
		summary := getSummaryForAction(action.Method, basePath, action.Path)
		description := getDescriptionForAction(action.Method, basePath, action.Path)

		globalWrapper.RegisterRoute(action.Method, fullPath, group, action.HandlerFunc, id, summary, description, responses)
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
func RegisterActions(group *fizz.RouterGroup, basePath string, actions []Action) {
	for _, action := range actions {
		fullPath := basePath + action.Path

		// Create generic response
		responses := globalWrapper.BuildModelResponse()

		id := getIDForAction(action.Method, fullPath)
		summary := getSummaryForAction(action.Method, basePath, action.Path)
		description := getDescriptionForAction(action.Method, basePath, action.Path)

		globalWrapper.RegisterRoute(action.Method, fullPath, group, action.HandlerFunc, id, summary, description, responses)
	}
}

// registerBuiltinHandler registers a standard handler if it hasn't been overridden
// by a custom action
func registerBuiltinHandler[T any](group *fizz.RouterGroup, wrapper *openapi.FizzWrapper, basePath, method, pathSuffix string, handlerFunc interface{}, existingActionPaths map[string]bool, summary, description string, model T) {
	path := basePath + pathSuffix
	key := method + " - " + path
	_, ok := existingActionPaths[key]
	if ok {
		return
	}

	id := getIDForAction(method, path)

	// Create appropriate response based on the method
	responses := wrapper.BuildModelResponse()

	wrapper.RegisterRoute(method, path, group, handlerFunc, id, summary, description, responses)
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
	RegisterController(groups.AuthGroup.Group("", "Data Collections", "APIs for data collections management"), "/data/collections", NewController[models.DataCollection]())
	RegisterController(groups.AuthGroup.Group("", "Environments", "APIs for environment variables management"), "/environments", NewController[models.Environment]())
	RegisterController(groups.AuthGroup.Group("", "Nodes", "APIs for nodes management"), "/nodes", NewController[models.Node]())
	RegisterController(groups.AuthGroup.Group("", "Projects", "APIs for projects management"), "/projects", NewController[models.Project]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetProjectList,
		},
	}...))
	RegisterController(groups.AuthGroup.Group("", "Spiders", "APIs for spiders management"), "/spiders", NewController[models.Spider]([]Action{
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
			HandlerFunc: GetSpiderFileContent,
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
	RegisterController(groups.AuthGroup.Group("", "Schedules", "APIs for schedules management"), "/schedules", NewController[models.Schedule]([]Action{
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
	RegisterController(groups.AuthGroup.Group("", "Tasks", "APIs for tasks management"), "/tasks", NewController[models.Task]([]Action{
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
	RegisterController(groups.AuthGroup.Group("", "Tokens", "APIs for PAT management"), "/tokens", NewController[models.Token]([]Action{
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
	RegisterController(groups.AuthGroup.Group("", "Users", "APIs for users management"), "/users", NewController[models.User]([]Action{
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
	RegisterActions(groups.AuthGroup.Group("", "Export", "APIs for exporting data"), "/export", []Action{
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
	RegisterActions(groups.AuthGroup.Group("", "Filters", "APIs for data collections filters management"), "/filters", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/:col",
			HandlerFunc: GetFilterColFieldOptions,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:col/:value",
			HandlerFunc: GetFilterColFieldOptionsWithValue,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:col/:value/:label",
			HandlerFunc: GetFilterColFieldOptionsWithValueLabel,
		},
	})
	RegisterActions(groups.AuthGroup.Group("", "Settings", "APIs for settings management"), "/settings", []Action{
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
	RegisterActions(groups.AuthGroup.Group("", "Stats", "APIs for data stats"), "/stats", []Action{
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

	// Register public routes that don't require authentication
	RegisterActions(groups.AnonymousGroup.Group("", "System", "APIs for system info"), "/system-info", []Action{
		{
			Path:        "",
			Method:      http.MethodGet,
			HandlerFunc: GetSystemInfo,
		},
	})
	RegisterActions(groups.AnonymousGroup.Group("", "Auth", "APIs for authentication"), "/", []Action{
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

	// Register sync routes that require special authentication
	groups.SyncAuthGroup.GET("/sync/:id/scan", GetSyncScan)
	groups.SyncAuthGroup.GET("/sync/:id/download", GetSyncDownload)

	// Register health check route
	groups.AnonymousGroup.GinRouterGroup().GET("/health", GetHealthFn(func() bool { return true }))

	// Register OpenAPI documentation route
	groups.AnonymousGroup.GinRouterGroup().GET("/openapi.json", GetOpenAPI)

	return nil
}
