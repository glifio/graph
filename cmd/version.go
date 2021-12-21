package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number of the Graph",
  Long:  `Print the version number of the Graph`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Glif Graph v0.0.1")
  },
}