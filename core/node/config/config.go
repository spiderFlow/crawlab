package config

import (
	"github.com/crawlab-team/crawlab/core/entity"
	"github.com/crawlab-team/crawlab/core/utils"
)

func newConfig() (cfg *entity.NodeInfo) {
	return &entity.NodeInfo{
		Key:        utils.GetNodeKey(),
		Name:       utils.GetNodeName(),
		IsMaster:   utils.IsMaster(),
		MaxRunners: utils.GetNodeMaxRunners(),
	}
}
