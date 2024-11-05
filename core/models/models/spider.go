package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Spider struct {
	any               `collection:"spiders"`
	BaseModel[Spider] `bson:",inline"`
	Name              string               `json:"name" bson:"name"`                     // spider name
	ColId             primitive.ObjectID   `json:"col_id" bson:"col_id"`                 // data collection id (deprecated) # TODO: remove this field in the future
	ColName           string               `json:"col_name,omitempty" bson:"col_name"`   // data collection name
	DbName            string               `json:"db_name,omitempty" bson:"db_name"`     // database name
	DataSourceId      primitive.ObjectID   `json:"data_source_id" bson:"data_source_id"` // data source id
	DataSource        *Database            `json:"data_source,omitempty" bson:"-"`       // data source
	Description       string               `json:"description" bson:"description"`       // description
	ProjectId         primitive.ObjectID   `json:"project_id" bson:"project_id"`         // Project.Id
	Mode              string               `json:"mode" bson:"mode"`                     // default Task.Mode
	NodeIds           []primitive.ObjectID `json:"node_ids" bson:"node_ids"`             // default Task.NodeIds
	GitId             primitive.ObjectID   `json:"git_id" bson:"git_id"`                 // related Git.Id
	GitRootPath       string               `json:"git_root_path" bson:"git_root_path"`
	Git               *Git                 `json:"git,omitempty" bson:"-"`

	// stats
	Stat *SpiderStat `json:"stat,omitempty" bson:"-"`

	// execution
	Cmd         string `json:"cmd" bson:"cmd"`     // execute command
	Param       string `json:"param" bson:"param"` // default task param
	Priority    int    `json:"priority" bson:"priority"`
	AutoInstall bool   `json:"auto_install" bson:"auto_install"`
}
