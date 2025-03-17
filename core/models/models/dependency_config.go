package models

type DependencyConfig struct {
	any               `collection:"dependency_configs"`
	BaseModel         `bson:",inline"`
	Key               string `json:"key" bson:"key"`
	Name              string `json:"name" bson:"name"`
	ExecCmd           string `json:"exec_cmd" bson:"exec_cmd"`
	PkgCmd            string `json:"pkg_cmd" bson:"pkg_cmd"`
	PkgSrcURL         string `json:"pkg_src_url" bson:"pkg_src_url"`
	Setup             bool   `json:"setup" bson:"-"`
	TotalDependencies int    `json:"total_dependencies" bson:"-"`
	SearchReady       bool   `json:"search_ready" bson:"-"`
}
