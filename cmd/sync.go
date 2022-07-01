package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/glifio/graph/pkg/daemon"
	"github.com/glifio/graph/pkg/node"
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.AddCommand(syncLilyCmd)
	syncCmd.AddCommand(syncValidateCmd)
	syncCmd.PersistentFlags().Uint64("height", 0, "Tipset height to start sync desc from")
	syncCmd.PersistentFlags().Uint64("length", 0, "Number of Tipsets")
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tipsets from lotus node",
	Long:  `Sync tipset data`,
	Run: func(cmd *cobra.Command, args []string) {
		height, _ := cmd.Flags().GetUint64("height")
		length, _ := cmd.Flags().GetUint64("length")

		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", viper.GetViper().GetUint("rpc")), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewDaemonClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.Sync(ctx, &pb.SyncRequest{Height: height, Length: length})
		if err != nil {
			log.Fatalf("could not sync: %v", err)
		}
		log.Printf("Sync: %s", r.GetMessage())

	},
}

var syncLilyCmd = &cobra.Command{
	Use:   "lily",
	Short: "Sync tipsets from lily database",
	Long:  `Sync lily tipset data`,
	Run: func(cmd *cobra.Command, args []string) {
		height, _ := cmd.Flags().GetUint64("height")
		length, _ := cmd.Flags().GetUint64("length")

		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", viper.GetViper().GetUint("rpc")), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewDaemonClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.SyncLily(ctx, &pb.SyncRequest{Height: height, Length: length})
		if err != nil {
			log.Fatalf("could not sync: %v", err)
		}
		log.Printf("Sync Lily: %s", r.GetMessage())
	},
}

var syncValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate indexes",
	Long:  `Iterate through all messages and validate indexes`,
	Run: func(cmd *cobra.Command, args []string) {
		height, _ := cmd.Flags().GetUint64("height")

		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", viper.GetViper().GetUint("rpc")), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewDaemonClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.SyncValidate(ctx, &pb.SyncRequest{Height: height})
		if err != nil {
			log.Fatalf("could not validate: %v", err)
		}
		log.Printf("Validate: %s", r.GetMessage())
	},
}

// Sync implements daemon.Sync
func (s *server) Sync(ctx context.Context, in *pb.SyncRequest) (*pb.SyncReply, error) {
	height := in.GetHeight()
	log.Printf("grpc -> sync: %v", height)

	go node.Sync(context.Background(), 50, in.Height, in.Length)

	str := fmt.Sprintf("Sync started desc from: %d", height)
	return &pb.SyncReply{Message: str}, nil
}

// SyncLily implements daemon.SyncLily
func (s *server) SyncLily(ctx context.Context, in *pb.SyncRequest) (*pb.SyncReply, error) {
	log.Printf("grpc -> sync lily: %v %d", in.Height, in.Length)

	go node.SyncLily(context.Background(), in.Height, in.Length)

	str := fmt.Sprintf("Syncing desc from: %d", in.Height)
	return &pb.SyncReply{Message: str}, nil
}

func (s *server) SyncValidate(ctx context.Context, in *pb.SyncRequest) (*pb.SyncReply, error) {
	height := in.GetHeight()
	log.Printf("grpc -> sync validate: %v", height)

	go node.ValidateMessages(uint64(height))

	str := fmt.Sprintf("Validate: %d", height)
	return &pb.SyncReply{Message: str}, nil
}
