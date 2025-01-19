package models

type DependencyPypiProject struct {
	any                              `collection:"dependency_pypi_projects"`
	BaseModel[DependencyPypiProject] `bson:",inline"`
	Name                             string `json:"name" bson:"name"`
	Version                          string `json:"version" bson:"version"`
	LastSerial                       int    `json:"_last-serial" bson:"last_serial"`
}
