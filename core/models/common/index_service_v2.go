package common

import (
	"fmt"

	"github.com/apex/log"
	models2 "github.com/crawlab-team/crawlab/core/models/models/v2"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateIndexesV2() {
	// Helper function to recreate indexes only if different
	recreateIndexes := func(col *mongo.Col, desiredIndexes []mongo2.IndexModel) {
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

	// nodes
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.NodeV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"key": 1}},       // key
		{Keys: bson.M{"name": 1}},      // name
		{Keys: bson.M{"is_master": 1}}, // is_master
		{Keys: bson.M{"status": 1}},    // status
		{Keys: bson.M{"enabled": 1}},   // enabled
		{Keys: bson.M{"active": 1}},    // active
	})

	// projects
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.ProjectV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"name": 1}},
	})

	// spiders
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.SpiderV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"name": 1}},
		{Keys: bson.M{"type": 1}},
		{Keys: bson.M{"col_id": 1}},
		{Keys: bson.M{"project_id": 1}},
	})

	// tasks
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.TaskV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"spider_id": 1}},
		{Keys: bson.M{"status": 1}},
		{Keys: bson.M{"node_id": 1}},
		{Keys: bson.M{"schedule_id": 1}},
		{Keys: bson.M{"type": 1}},
		{Keys: bson.M{"mode": 1}},
		{Keys: bson.M{"priority": 1}},
		{Keys: bson.M{"parent_id": 1}},
		{Keys: bson.M{"has_sub": 1}},
		{Keys: bson.M{"created_ts": -1}, Options: (&options.IndexOptions{}).SetExpireAfterSeconds(60 * 60 * 24 * 30)},
	})

	// task stats
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.TaskStatV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"created_ts": -1}, Options: (&options.IndexOptions{}).SetExpireAfterSeconds(60 * 60 * 24 * 30)},
	})

	// schedules
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.ScheduleV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"name": 1}},
		{Keys: bson.M{"spider_id": 1}},
		{Keys: bson.M{"enabled": 1}},
	})

	// users
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.UserV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"username": 1}},
		{Keys: bson.M{"role": 1}},
		{Keys: bson.M{"email": 1}},
	})

	// settings
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.SettingV2{})), []mongo2.IndexModel{
		{Keys: bson.D{{"key", 1}}, Options: options.Index().SetUnique(true)},
	})

	// tokens
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.TokenV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"name": 1}},
	})

	// data sources
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DatabaseV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"name": 1}},
	})

	// data collections
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DataCollectionV2{})), []mongo2.IndexModel{
		{Keys: bson.M{"name": 1}},
	})

	// roles
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.RoleV2{})), []mongo2.IndexModel{
		{Keys: bson.D{{"key", 1}}, Options: options.Index().SetUnique(true)},
	})

	// user role relations
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.UserRoleV2{})), []mongo2.IndexModel{
		{Keys: bson.D{{"user_id", 1}, {"role_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"role_id", 1}, {"user_id", 1}}, Options: options.Index().SetUnique(true)},
	})

	// permissions
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.PermissionV2{})), []mongo2.IndexModel{
		{Keys: bson.D{{"key", 1}}, Options: options.Index().SetUnique(true)},
	})

	// role permission relations
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.RolePermissionV2{})), []mongo2.IndexModel{
		{Keys: bson.D{{"role_id", 1}, {"permission_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"permission_id", 1}, {"role_id", 1}}, Options: options.Index().SetUnique(true)},
	})

	// dependencies
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DependencyV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{
				{"type", 1},
				{"node_id", 1},
				{"name", 1},
			},
			Options: (&options.IndexOptions{}).SetUnique(true),
		},
	})

	// dependency settings
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DependencySettingV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{
				{"type", 1},
				{"node_id", 1},
				{"name", 1},
			},
			Options: (&options.IndexOptions{}).SetUnique(true),
		},
	})

	// dependency logs
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DependencyLogV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{{"task_id", 1}},
		},
		{
			Keys:    bson.D{{"update_ts", 1}},
			Options: (&options.IndexOptions{}).SetExpireAfterSeconds(60 * 60 * 24),
		},
	})

	// dependency tasks
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DependencyTaskV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{
				{"update_ts", 1},
			},
			Options: (&options.IndexOptions{}).SetExpireAfterSeconds(60 * 60 * 24),
		},
	})

	// metrics
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.MetricV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{
				{"created_ts", -1},
			},
			Options: (&options.IndexOptions{}).SetExpireAfterSeconds(60 * 60 * 24 * 30),
		},
		{
			Keys: bson.D{
				{"node_id", 1},
			},
		},
		{
			Keys: bson.D{
				{"type", 1},
			},
		},
	})

	// notification requests
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.NotificationRequestV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{
				{"created_ts", -1},
			},
			Options: (&options.IndexOptions{}).SetExpireAfterSeconds(60 * 60 * 24 * 7),
		},
		{
			Keys: bson.D{
				{"channel_id", 1},
			},
		},
		{
			Keys: bson.D{
				{"setting_id", 1},
			},
		},
		{
			Keys: bson.D{
				{"status", 1},
			},
		},
	})

	// databases
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DatabaseV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{
				{"data_source_id", 1},
			},
		},
	})

	// database metrics
	recreateIndexes(mongo.GetMongoCol(service.GetCollectionNameByInstance(models2.DatabaseMetricV2{})), []mongo2.IndexModel{
		{
			Keys: bson.D{
				{"created_ts", -1},
			},
			Options: (&options.IndexOptions{}).SetExpireAfterSeconds(60 * 60 * 24 * 30),
		},
		{
			Keys: bson.D{
				{"database_id", 1},
			},
		},
		{
			Keys: bson.D{
				{"type", 1},
			},
		},
	})
}
