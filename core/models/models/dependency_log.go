package models

type DependencyLog struct {
	any                      `collection:"dependency_logs"`
	BaseModel[DependencyLog] `bson:",inline"`
	Content                  string `json:"content" bson:"content"`
}
