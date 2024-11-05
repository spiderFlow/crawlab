package models

type Permission struct {
	any                   `collection:"permissions"`
	BaseModel[Permission] `bson:",inline"`
	Key                   string   `json:"key" bson:"key"`
	Name                  string   `json:"name" bson:"name"`
	Description           string   `json:"description" bson:"description"`
	Type                  string   `json:"type" bson:"type"`
	Target                []string `json:"target" bson:"target"`
	Allow                 []string `json:"allow" bson:"allow"`
	Deny                  []string `json:"deny" bson:"deny"`
}
