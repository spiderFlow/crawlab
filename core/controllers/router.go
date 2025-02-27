package controllers

import (
	"net/http"

	"github.com/crawlab-team/crawlab/core/middlewares"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/gin-gonic/gin"
)

// RouterGroups defines the different authentication levels for API routes
type RouterGroups struct {
	AuthGroup      *gin.RouterGroup // Routes requiring full authentication
	SyncAuthGroup  *gin.RouterGroup // Routes for sync operations with special auth
	AnonymousGroup *gin.RouterGroup // Public routes that don't require auth
}

// NewRouterGroups initializes the router groups with their respective middleware
func NewRouterGroups(app *gin.Engine) (groups *RouterGroups) {
	return &RouterGroups{
		AuthGroup:      app.Group("/", middlewares.AuthorizationMiddleware()),
		SyncAuthGroup:  app.Group("/", middlewares.SyncAuthorizationMiddleware()),
		AnonymousGroup: app.Group("/"),
	}
}

// RegisterController registers a generic controller with standard CRUD endpoints
// and any additional custom actions
func RegisterController[T any](group *gin.RouterGroup, basePath string, ctr *BaseController[T]) {
	// Track registered paths to avoid duplicates
	actionPaths := make(map[string]bool)
	for _, action := range ctr.actions {
		group.Handle(action.Method, basePath+action.Path, action.HandlerFunc)
		path := basePath + action.Path
		key := action.Method + " - " + path
		actionPaths[key] = true
	}
	registerBuiltinHandler(group, http.MethodGet, basePath+"", ctr.GetList, actionPaths)
	registerBuiltinHandler(group, http.MethodGet, basePath+"/:id", ctr.GetById, actionPaths)
	registerBuiltinHandler(group, http.MethodPost, basePath+"", ctr.Post, actionPaths)
	registerBuiltinHandler(group, http.MethodPut, basePath+"/:id", ctr.PutById, actionPaths)
	registerBuiltinHandler(group, http.MethodPatch, basePath+"", ctr.PatchList, actionPaths)
	registerBuiltinHandler(group, http.MethodDelete, basePath+"/:id", ctr.DeleteById, actionPaths)
	registerBuiltinHandler(group, http.MethodDelete, basePath+"", ctr.DeleteList, actionPaths)
}

// RegisterActions registers a list of custom action handlers to a route group
func RegisterActions(group *gin.RouterGroup, basePath string, actions []Action) {
	for _, action := range actions {
		group.Handle(action.Method, basePath+action.Path, action.HandlerFunc)
	}
}

// registerBuiltinHandler registers a standard handler if it hasn't been overridden
// by a custom action
func registerBuiltinHandler(group *gin.RouterGroup, method, path string, handlerFunc gin.HandlerFunc, existingActionPaths map[string]bool) {
	key := method + " - " + path
	_, ok := existingActionPaths[key]
	if ok {
		return
	}
	group.Handle(method, path, handlerFunc)
}

// InitRoutes configures all API routes for the application
func InitRoutes(app *gin.Engine) (err error) {
	// Initialize route groups with different auth levels
	groups := NewRouterGroups(app)

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

	return nil
}
