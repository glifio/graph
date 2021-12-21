package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	graph "github.com/glifio/graph/gql"
	"github.com/glifio/graph/gql/generated"
	util "github.com/glifio/graph/internal/utils"
	"github.com/glifio/graph/pkg/node"
	"github.com/glifio/graph/pkg/postgres"
	"github.com/spf13/cobra"
)

func init() {
  rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
  Use:   "daemon",
  Short: "Start the graphQL server",
  Long:  `Start the graphQL server`,
  Run: func(cmd *cobra.Command, args []string) {
	config, err := util.LoadConfig(".")

    fmt.Println("start the graphQL server")
	// Create a new connection to our pg database
	var db postgres.Database
	err = db.New(
		db.ConnString(config.DbHost, config.DbPort, config.DbUser, config.DbPassword, config.DbDatabase),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	nodeService := &node.Node{}
	nodeService.Connect(config.LotusAddress)
	defer nodeService.Close()

	// actorService := &postgres.Actor{}
	// actorService.Init(db)
	messageService := &postgres.Message{}
	messageService.Init(db)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{NodeService: nodeService, MessageService: messageService}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", config.Port)
	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(config.Port), nil))
  },
}