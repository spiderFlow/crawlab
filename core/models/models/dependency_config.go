package models

type DependencyConfig struct {
	any                         `collection:"dependency_configs"`
	BaseModel[DependencyConfig] `bson:",inline"`
	Key                         string   `json:"key" bson:"key"`
	Name                        string   `json:"name" bson:"name"`
	Cmd                         string   `json:"cmd" bson:"cmd"`
	Proxy                       string   `json:"proxy" bson:"proxy"`
	SetupNodeIds                []string `json:"setup_node_ids" bson:"setup_node_ids"`
	SetupScriptPath             string   `json:"setup_script_path" bson:"setup_script_path"`
}
