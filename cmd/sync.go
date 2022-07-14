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
	syncCmd.AddCommand(syncStopCmd)
	syncCmd.AddCommand(syncValidateCmd)
	syncCmd.PersistentFlags().Uint64("height", 0, "Tipset height to start sync desc from")
	syncCmd.PersistentFlags().Uint64("length", 0, "Number of Tipsets")
}

var syncCmdxx = &cobra.Command{
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

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tipset data",
	Long:  `Fetch all tipsets from lotus node top down`,
	Run: func(cmd *cobra.Command, args []string) {
		height, _ := cmd.Flags().GetUint64("height")
		length, _ := cmd.Flags().GetUint64("length")

		client, conn := RpcDial()
		defer conn.Close()

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r, err := client.SyncTipset(ctx, &pb.SyncRequest{Action: pb.SyncRequest_START, Height: height, Length: length})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r.GetMessage())
	},
}

func RpcDial() (pb.DaemonClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", viper.GetViper().GetUint("rpc")), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return pb.NewDaemonClient(conn), conn
}

var syncStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Sync stop",
	Run: func(cmd *cobra.Command, args []string) {
		client, conn := RpcDial()
		defer conn.Close()

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r, err := client.SyncTipset(ctx, &pb.SyncRequest{Action: pb.SyncRequest_STOP})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r.GetMessage())
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

var ctxTipset context.Context = nil
var ctxCancel context.CancelFunc = nil

func (s *server) SyncTipset(ctx context.Context, in *pb.SyncRequest) (*pb.SyncReply, error) {
	str := ""
	height := in.GetHeight()
	switch in.Action {
	case pb.SyncRequest_START:
		if ctxCancel == nil {
			ctxTipset, ctxCancel = context.WithCancel(context.Background())
			go node.SyncTipsetStart(ctxTipset, 50, in.Height, in.Length)
			str = fmt.Sprintf("Sync started desc from: %d", height)
		} else {
			str = "Sync already started"
		}
	case pb.SyncRequest_STOP:
		if ctxCancel != nil {
			ctxCancel()
			str = fmt.Sprintf("Sync tipset stopped: %d", height)
			ctxCancel = nil
		} else {
			str = "Sync not started"
		}

	}
	return &pb.SyncReply{Message: str}, nil
}

func (s *server) SyncMessages(ctx context.Context, in *pb.SyncRequest) (*pb.SyncReply, error) {
	height := in.GetHeight()
	log.Printf("grpc -> sync message: %v", height)

	go node.SyncMessages(context.Background(), 50, in.Height, in.Length)

	str := fmt.Sprintf("Sync started desc from: %d", height)
	return &pb.SyncReply{Message: str}, nil
}

// SyncLily implements daemon.SyncLily
func (s *server) SyncLily(ctx context.Context, in *pb.SyncRequest) (*pb.SyncReply, error) {
	log.Printf("grpc -> sync lily: %v %d", in.Height, in.Length)

	// go node.SyncLily(context.Background(), in.Height, in.Length)

	str := fmt.Sprintf("Syncing desc from: %d", in.Height)
	return &pb.SyncReply{Message: str}, nil
}

func (s *server) SyncValidate(ctx context.Context, in *pb.SyncRequest) (*pb.SyncReply, error) {
	height := in.GetHeight()
	log.Printf("grpc -> sync validate")

	go node.ValidateMessages(uint64(height))

	str := fmt.Sprintln("Validate: started")
	return &pb.SyncReply{Message: str}, nil
}
