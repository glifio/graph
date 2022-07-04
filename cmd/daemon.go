package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/getsentry/sentry-go"
	graphql "github.com/glifio/graph/gql"
	"github.com/glifio/graph/gql/generated"
	pb "github.com/glifio/graph/pkg/daemon"
	"github.com/glifio/graph/pkg/kvdb"
	"github.com/glifio/graph/pkg/node"
	"github.com/glifio/graph/pkg/postgres"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var nodeService *node.Node

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.PersistentFlags().Bool("no-cache", false, "Don't start the cache")
	daemonCmd.PersistentFlags().Bool("sync", false, "Start sync")
	daemonCmd.PersistentFlags().Bool("no-timer", false, "Don't start sync timer")
}

func SetupCloseHandler(node *node.Node) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		node.Close()
		os.Exit(0)
	}()
}

// server is used to implement daemon.server
type server struct {
	pb.UnimplementedDaemonServer
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the graphQL server",
	Long:  `Start the graphQL server`,
	Run: func(cmd *cobra.Command, args []string) {
		//config, _ := util.LoadConfig(".")
		cacheDisabled, _ := cmd.Flags().GetBool("no-cache")
		syncEnabled, _ := cmd.Flags().GetBool("sync")
		timerDisabled, _ := cmd.Flags().GetBool("no-timer")

		// log.Println("cache -> init")
		// log.Println("lotus -> ", viper.GetString("lotus"))
		// log.Println("confidence -> ", viper.GetString("confidence"))
		// log.Println("sentry -> ", viper.GetString("sentry"))

		cache := node.GetCacheInstance().Cache()

		kvdb.Open()
		postgres.GetInstanceDB()
		node.GetCacheInstance()

		nodeService = &node.Node{}
		network, _ := nodeService.Connect(viper.GetString("lotus"), viper.GetString("lotus_token"))
		defer nodeService.Close()

		SetupCloseHandler(nodeService)

		messageService := &postgres.Message{}
		messageConfirmedService := &postgres.MessageConfirmed{}
		messageConfirmedService.Init(cache)
		blockService := &postgres.BlockHeader{}

		if !cacheDisabled {
			nodeService.StartCache()
		}

		if syncEnabled {
			go node.Sync(context.Background(), uint64(viper.GetInt("confidence")), 0, 0)
		}

		if !timerDisabled {
			nodeService.SyncTimerStart(uint32(viper.GetInt("confidence")))
		}

		err := sentry.Init(sentry.ClientOptions{
			Dsn:         viper.GetString("sentry"),
			Environment: string(network),
			Debug:       true,
			Release:     "current",
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}

		// Flush buffered events before the program terminates.
		defer sentry.Flush(2 * time.Second)

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetUint("rpc")))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterDaemonServer(s, &server{})
		log.Printf("rpc server listening at %v", lis.Addr())
		go func() {
			if err := s.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %v", err)
			}
		}()

		router := chi.NewRouter()

		// Add CORS middleware around every request
		// See https://github.com/rs/cors for full option listing
		router.Use(cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
			Debug:            false,
		}).Handler)

		srv := handler.New(generated.NewExecutableSchema(generated.Config{
			Resolvers: &graphql.Resolver{
				NodeService:             nodeService,
				MessageService:          messageService,
				MessageConfirmedService: messageConfirmedService,
				BlockService:            blockService,
			},
		}))
		srv.AddTransport(transport.Websocket{
			KeepAlivePingInterval: 10 * time.Second,
			Upgrader: websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					fmt.Printf("wss origin: %s\n", r.Host)
					return true
				},
			},
		})
		srv.AddTransport(transport.Options{})
		srv.AddTransport(transport.GET{})
		srv.AddTransport(transport.POST{})
		srv.AddTransport(transport.MultipartForm{})

		srv.SetQueryCache(lru.New(1000))

		srv.Use(extension.Introspection{})
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New(100),
		})

		log.Println("start -> graphQL server")

		router.Handle("/", playground.Handler("GraphQL playground", "/query"))
		router.Handle("/query", srv)

		log.Printf("connect to http://localhost:%s/ for GraphQL playground", viper.GetString("port"))
		log.Fatal(http.ListenAndServe(":"+viper.GetString("port"), router))
	},
}
