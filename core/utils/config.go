package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	DefaultWorkspace           = "crawlab_workspace"
	DefaultTaskLogPath         = "/var/log/crawlab/tasks"
	DefaultServerHost          = "0.0.0.0"
	DefaultServerPort          = 8000
	DefaultGrpcHost            = "localhost"
	DefaultGrpcPort            = 9666
	DefaultGrpcServerHost      = "127.0.0.1"
	DefaultGrpcServerPort      = 9666
	DefaultAuthKey             = "Crawlab2024!"
	DefaultApiEndpoint         = "http://localhost:8000"
	DefaultApiAllowOrigin      = "*"
	DefaultApiAllowCredentials = "true"
	DefaultApiAllowMethods     = "DELETE, POST, OPTIONS, GET, PUT"
	DefaultApiAllowHeaders     = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"
)

func IsDev() bool {
	return viper.GetBool("dev")
}

func IsAuthDisabled() bool {
	return viper.GetBool("auth.disabled")
}

func GetAllowOrigin() string {
	if res := viper.GetString("api.allow.origin"); res != "" {
		return res
	}
	return DefaultApiAllowOrigin
}

func GetAllowCredentials() string {
	if res := viper.GetString("api.allow.credentials"); res != "" {
		return res
	}
	return DefaultApiAllowCredentials
}

func GetAllowMethods() string {
	if res := viper.GetString("api.allow.methods"); res != "" {
		return res
	}
	return DefaultApiAllowMethods
}

func GetAllowHeaders() string {
	if res := viper.GetString("api.allow.headers"); res != "" {
		return res
	}
	return DefaultApiAllowHeaders
}

func GetGinMode() string {
	if res := viper.GetString("gin.mode"); res != "" {
		return res
	}
	if IsDev() {
		return gin.DebugMode
	} else {
		return gin.ReleaseMode
	}
}

func IsPro() bool {
	return viper.GetString("edition") == "global.edition.pro"
}

func GetWorkspace() string {
	if res := viper.GetString("workspace"); res != "" {
		return res
	}
	return DefaultWorkspace
}

func GetTaskLogPath() string {
	if res := viper.GetString("log.path"); res != "" {
		return res
	}
	return DefaultTaskLogPath
}

func GetServerAddress() string {
	host := viper.GetString("server.host")
	if host == "" {
		host = DefaultServerHost
	}

	port := viper.GetInt("server.port")
	if port == 0 {
		port = DefaultServerPort
	}

	return fmt.Sprintf("%s:%d", host, port)
}

func GetMasterHost() string {
	return viper.GetString("master.host")
}

func GetGrpcAddress() string {
	host := viper.GetString("grpc.host")
	if host == "" {
		masterHost := GetMasterHost()
		if masterHost != "" {
			host = masterHost
		} else {
			host = DefaultGrpcHost
		}
	}

	port := viper.GetInt("grpc.port")
	if port == 0 {
		port = DefaultGrpcPort
	}

	return fmt.Sprintf("%s:%d", host, port)
}

func GetGrpcServerAddress() string {
	host := viper.GetString("grpc.server.host")
	if host == "" {
		host = DefaultGrpcServerHost
	}

	port := viper.GetInt("grpc.server.port")
	if port == 0 {
		port = DefaultGrpcServerPort
	}

	return fmt.Sprintf("%s:%d", host, port)
}

func GetAuthKey() string {
	if res := viper.GetString("auth.key"); res != "" {
		return res
	}
	return DefaultAuthKey
}

func GetApiEndpoint() string {
	if res := viper.GetString("api.endpoint"); res != "" {
		return res
	}
	masterHost := GetMasterHost()
	if masterHost != "" {
		apiHttps := viper.GetBool("api.https")
		if apiHttps {
			return "https://" + masterHost + "/api"
		}
		return "http://" + masterHost + "/api"
	}
	return DefaultApiEndpoint
}
