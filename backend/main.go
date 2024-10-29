package main

import (
	"github.com/crawlab-team/crawlab/core/cmd"
	"github.com/crawlab-team/crawlab/core/utils"
)

func main() {
	go func() {
		err := cmd.Execute()
		if err != nil {
			panic(err)
		}
	}()
	utils.DefaultWait()
}
