package models

import (
	"time"
)

type TaskStat struct {
	any             `collection:"task_stats"`
	BaseModel       `bson:",inline"`
	StartTs         time.Time `json:"start_ts" bson:"start_ts,omitempty" description:"Start time"`
	EndTs           time.Time `json:"end_ts" bson:"end_ts,omitempty" description:"End time"`
	WaitDuration    int64     `json:"wait_duration" bson:"wait_duration,omitempty" description:"Wait duration (in millisecond)"`
	RuntimeDuration int64     `json:"runtime_duration" bson:"runtime_duration,omitempty" description:"Runtime duration (in millisecond)"`
	TotalDuration   int64     `json:"total_duration" bson:"total_duration,omitempty" description:"Total duration (in millisecond)"`
	ResultCount     int64     `json:"result_count" bson:"result_count" description:"Result count"`
}
