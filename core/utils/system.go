package utils

import "github.com/spf13/viper"

func GetVersion() string {
	return viper.GetString("version")

}

func GetEdition() string {
	return viper.GetString("edition")
}

func GetEditionLabel() string {
	if IsPro() {
		return "Crawlab Pro"
	} else {
		return "Crawlab Community"
	}
}

func GetNodeTypeLabel() string {
	if IsMaster() {
		return "Master"
	} else {
		return "Worker"
	}
}
