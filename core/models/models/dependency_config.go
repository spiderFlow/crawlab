package models

type DependencyConfig struct {
	any                         `collection:"dependency_configs"`
	BaseModel[DependencyConfig] `bson:",inline"`
	Key                         string `json:"key" bson:"key"`
	Name                        string `json:"name" bson:"name"`
	ExecCmd                     string `json:"exec_cmd" bson:"exec_cmd"`
	PkgCmd                      string `json:"pkg_cmd" bson:"pkg_cmd"`
	Proxy                       string `json:"proxy" bson:"proxy"`
	Setup                       bool   `json:"setup" bson:"-"`
}
