package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
  Use:   "status",
  Short: "Check graph status",
  Long:  `Check graph status`,
  Run: func(cmd *cobra.Command, args []string) {
	// config, _ := util.LoadConfig(".")

	// // Create a new connection to our pg database
	// var db postgres.Database
	// err = db.New(config.DbHost, config.DbPort, config.DbUser, config.DbPassword, config.DbDatabase, config.DbSchema)

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// // nodeService.StartCache(maxheight)

	// // Flush buffered events before the program terminates.
	// defer sentry.Flush(2 * time.Second)
  },
}
