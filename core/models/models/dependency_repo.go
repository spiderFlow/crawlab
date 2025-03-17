package models

type DependencyRepo struct {
	any           `collection:"dependency_repos"`
	BaseModel     `bson:",inline"`
	Name          string   `json:"name" bson:"name"`
	Type          string   `json:"type" bson:"type"`
	LatestVersion string   `json:"latest_version" bson:"latest_version"`
	AllVersions   []string `json:"all_versions" bson:"all_versions"`
}
