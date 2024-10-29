package common

import (
	"fmt"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

func RecreateIndexes(col *mongo.Col, desiredIndexes []mongo2.IndexModel) {
	cur, err := col.GetCollection().Indexes().List(col.GetContext())
	if err != nil {
		log.Errorf("error listing indexes: %v", err)
		return
	}

	var existingIndexes []bson.M
	err = cur.All(col.GetContext(), &existingIndexes)
	if err != nil {
		log.Errorf("error listing indexes: %v", err)
		return
	}

	// Compare and recreate only if different
	needsUpdate := false
	existingKeys := make(map[string]bool)

	// Skip _id index when comparing
	for _, idx := range existingIndexes {
		if name, ok := idx["name"].(string); ok && name != "_id_" {
			key := idx["key"].(bson.M)
			keyStr := fmt.Sprintf("%v", key)
			existingKeys[keyStr] = true
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
		col.MustCreateIndexes(desiredIndexes)
		log.Infof("recreated indexes for collection: %s", col.GetCollection().Name())
	}
}
