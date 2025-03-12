package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	any        `collection:"tasks"`
	BaseModel  `bson:",inline"`
	SpiderId   primitive.ObjectID   `json:"spider_id" bson:"spider_id"`
	Status     string               `json:"status" bson:"status"`
	NodeId     primitive.ObjectID   `json:"node_id" bson:"node_id"`
	Cmd        string               `json:"cmd" bson:"cmd"`
	Param      string               `json:"param" bson:"param"`
	Error      string               `json:"error" bson:"error"`
	Pid        int                  `json:"pid" bson:"pid"`
	ScheduleId primitive.ObjectID   `json:"schedule_id" bson:"schedule_id"`
	Type       string               `json:"type" bson:"type"`
	Mode       string               `json:"mode" bson:"mode"`
	Priority   int                  `json:"priority" bson:"priority"`
	NodeIds    []primitive.ObjectID `json:"node_ids,omitempty" bson:"-"`
	Stat       *TaskStat            `json:"stat,omitempty" bson:"-"`
	Spider     *Spider              `json:"spider,omitempty" bson:"-"`
	Schedule   *Schedule            `json:"schedule,omitempty" bson:"-"`
	Node       *Node                `json:"node,omitempty" bson:"-"`
	UserId     primitive.ObjectID   `json:"-" bson:"-"`
}
