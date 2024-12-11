package common

import (
	"encoding/json"
	"fmt"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

// getIndexKeyString converts index keys to a consistent string representation
func getIndexKeyString(key interface{}) string {
	switch v := key.(type) {
	case mongo2.IndexModel:
		return normalizeIndexKey(v.Keys)
	default:
		return normalizeIndexKey(key)
	}
}

// normalizeIndexKey converts an index key to a standardized string representation
func normalizeIndexKey(key interface{}) string {
	// Convert to bson.D to maintain order
	var doc bson.D

	switch v := key.(type) {
	case bson.D:
		doc = v
	default:
		// Convert to JSON bytes first
		jsonBytes, err := json.Marshal(key)
		if err != nil {
			return fmt.Sprintf("%v", key)
		}

		// Try to unmarshal as array of key-value pairs first
		var keyArray []struct {
			Key   string      `json:"Key"`
			Value interface{} `json:"Value"`
		}
		if err := json.Unmarshal(jsonBytes, &keyArray); err == nil {
			// Convert array format to bson.D
			doc = make(bson.D, len(keyArray))
			for i, kv := range keyArray {
				doc[i] = bson.E{Key: kv.Key, Value: kv.Value}
			}
		} else {
			// Try to unmarshal as regular map
			if err := bson.UnmarshalExtJSON(jsonBytes, true, &doc); err != nil {
				return string(jsonBytes)
			}
		}
	}

	// Convert bson.D to consistent string representation
	pairs := make([]string, 0, len(doc))
	for _, elem := range doc {
		pairs = append(pairs, fmt.Sprintf("%q:%v", elem.Key, elem.Value))
	}

	res, _ := bson.Marshal(pairs)
	return string(res)
}

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
				keyStr := getIndexKeyString(key)
				existingKeys[keyStr] = true
			}
		}
	}

	// Check if desired indexes exist
	for _, idx := range desiredIndexes {
		keyStr := getIndexKeyString(idx.Keys)
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
