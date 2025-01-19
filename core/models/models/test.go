package models

type TestModel struct {
	any                  `collection:"testmodels"`
	BaseModel[TestModel] `bson:",inline"`
	Name                 string `json:"name" bson:"name"`
}
