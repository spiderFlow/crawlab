package common

import (
	"fmt"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/db/mongo"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

func RecreateIndexes(col *mongo.Col, desiredIndexes []mongo2.IndexModel) {
	existingIndexes, err := col.ListIndexes()
	if err != nil {
		log.Errorf("error listing indexes: %v", err)
		return
	}

	// Compare and recreate only if different
	needsUpdate := false
	existingKeys := make(map[string]bool)

	// Skip _id index when comparing
	for _, idx := range existingIndexes {
		name, ok := idx["name"].(string)
		if ok && name != "_id_" {
			if key, ok := idx["key"]; ok {
				keyStr := fmt.Sprintf("%v", key)
				existingKeys[keyStr] = true
			}
		}
	}

	// Check if desired indexes exist
	for _, idx := range desiredIndexes {
		keyStr := fmt.Sprintf("%v", idx.Keys)
		if !existingKeys[keyStr] {
			needsUpdate = true
			break
		}
	}

	if needsUpdate {
		// Drop all existing indexes (except _id)
		err := col.DeleteAllIndexes()
		if err != nil {
			log.Errorf("error dropping indexes: %v", err)
		}

		// Create new indexes
		err = col.CreateIndexes(desiredIndexes)
		if err != nil {
			log.Errorf("error creating indexes: %v", err)
			return
		}
		log.Infof("recreated indexes for collection: %s", col.GetCollection().Name())
	}
}
