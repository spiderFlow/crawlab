package main

import (
	"github.com/crawlab-team/crawlab/core/cmd"
	"github.com/crawlab-team/crawlab/core/config"
	"github.com/crawlab-team/crawlab/core/utils"
)

func init() {
	config.InitConfig()
}

func main() {
	go func() {
		err := cmd.Execute()
		if err != nil {
			panic(err)
		}
	}()
	utils.DefaultWait()
}
