package common

import (
	"fmt"
	"testing"

	"github.com/crawlab-team/crawlab/db/mongo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

func TestRecreateIndexes(t *testing.T) {
	// Setup test collection
	testCol := mongo.GetMongoCol("test_collection")
	defer func() {
		_ = testCol.GetCollection().Drop(testCol.GetContext())
	}()

	// Test cases
	tests := []struct {
		name           string
		desiredIndexes []mongo2.IndexModel
		expectedCount  int64
	}{
		{
			name: "Create new indexes",
			desiredIndexes: []mongo2.IndexModel{
				{Keys: bson.M{"field1": 1}},
				{Keys: bson.M{"field2": -1}},
			},
			expectedCount: 3, // Including _id index
		},
		{
			name: "Update existing indexes",
			desiredIndexes: []mongo2.IndexModel{
				{Keys: bson.M{"field1": 1}},
				{Keys: bson.M{"field3": 1}},
			},
			expectedCount: 3, // Including _id index
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute RecreateIndexes
			RecreateIndexes(testCol, tt.desiredIndexes)

			// Verify indexes
			cur, err := testCol.GetCollection().Indexes().List(testCol.GetContext())
			assert.NoError(t, err)

			var indexes []bson.M
			err = cur.All(testCol.GetContext(), &indexes)
			assert.NoError(t, err)

			// Check total number of indexes (including _id)
			assert.Equal(t, tt.expectedCount, int64(len(indexes)))

			// Verify each desired index exists
			for _, desiredIdx := range tt.desiredIndexes {
				found := false
				desiredKeyStr := fmt.Sprintf("%v", desiredIdx.Keys)
				for _, existingIdx := range indexes {
					if existingIdx["name"].(string) != "_id_" {
						key := existingIdx["key"].(bson.M)
						if fmt.Sprintf("%v", key) == desiredKeyStr {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Index not found: %v", desiredIdx.Keys)
			}
		})
	}
}
