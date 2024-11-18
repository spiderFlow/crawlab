package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	DefaultServerHost          = "0.0.0.0"
	DefaultServerPort          = "8000"
	DefaultTaskLogPath         = "/var/log/crawlab/tasks"
	DefaultGrpcHost            = "localhost"
	DefaultGrpcPort            = "9666"
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
	return viper.GetString("workspace")
}

func GetTaskLogPath() string {
	if res := viper.GetString("log.path"); res != "" {
		return res
	}
	return DefaultTaskLogPath
}

func GetServerHost() string {
	if res := viper.GetString("server.host"); res != "" {
		return res
	}
	return DefaultServerHost
}

func GetServerPort() string {
	if res := viper.GetString("server.port"); res != "" {
		return res
	}
	return DefaultServerPort
}

func GetMasterHost() string {
	return viper.GetString("master.host")
}

func GetGrpcAddress() string {
	host := viper.GetString("grpc.host")
	port := viper.GetString("grpc.port")
	if host == "" {
		masterHost := GetMasterHost()
		if masterHost != "" {
			host = masterHost
		} else {
			host = DefaultGrpcHost
		}
	}
	if port == "" {
		port = DefaultGrpcPort
	}
	return host + ":" + port
}

func GetGrpcServerAddress() string {
	return viper.GetString("grpc.server.address")
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
