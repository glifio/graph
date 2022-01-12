package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	graph "github.com/glifio/graph/gql"
	"github.com/glifio/graph/gql/generated"
	util "github.com/glifio/graph/internal/utils"
	"github.com/glifio/graph/pkg/node"
	"github.com/glifio/graph/pkg/postgres"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
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
	err = db.New(config.DbHost, config.DbPort, config.DbUser, config.DbPassword, config.DbDatabase, config.DbSchema)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	nodeService := &node.Node{}
	nodeService.Connect(config.LotusAddress, config.LotusToken)
	defer nodeService.Close()

	// actorService := &postgres.Actor{}
	// actorService.Init(db)
	messageService := &postgres.Message{}
	messageService.Init(db)
	messageConfirmedService := &postgres.MessageConfirmed{}
	messageConfirmedService.Init(db)
	blockService := &postgres.BlockHeader{}
	blockService.Init(db)

	router := chi.NewRouter()

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			NodeService: nodeService, 
			MessageService: messageService, 
			MessageConfirmedService: messageConfirmedService,
			BlockService: blockService, 
		},
	}))
	srv.AddTransport(&transport.Websocket{
        Upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                // Check against your desired domains here
                // return r.Host == "example.org"
				 return r.Host == "*"
            },
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
        },
    })

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%d/ for GraphQL playground", config.Port)
	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(config.Port), router))
  },
}