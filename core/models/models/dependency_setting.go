package models

type DependencySetting struct {
	any                          `collection:"dependency_settings"`
	BaseModel[DependencySetting] `bson:",inline"`
	Key                          string `json:"key" bson:"key"`
	Name                         string `json:"name" bson:"name"`
	Description                  string `json:"description" bson:"description"`
	Enabled                      bool   `json:"enabled" bson:"enabled"`
	Cmd                          string `json:"cmd" bson:"cmd"`
	Proxy                        string `json:"proxy" bson:"proxy"`
}
