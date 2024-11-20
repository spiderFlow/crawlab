package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "crawlab",
	Short: "Crawlab CLI tool",
	Long:  `Command line interface for Crawlab, the distributed web crawler admin platform`,
}

func init() {
	RootCmd.AddCommand(protoCmd)
}
