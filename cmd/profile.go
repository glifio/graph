package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/glifio/graph/pkg/daemon"
)

func init() {
	rootCmd.AddCommand(profileMemCmd)
}

var profileMemCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile memory",
	Run: func(cmd *cobra.Command, args []string) {
		client, conn := RpcDial()
		defer conn.Close()

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r, err := client.ProfileMemory(ctx, &pb.ProfileRequest{Action: pb.ProfileRequest_START})
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(r.GetMessage())
	},
}

func (s *server) ProfileMemory(ctx context.Context, in *pb.ProfileRequest) (*pb.Reply, error) {
	str := ""
	switch in.Action {
	case pb.ProfileRequest_START:
		f, err := os.Create("memprofile")
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()

		str = fmt.Sprintf("profile done")
		return &pb.Reply{Message: str}, nil
	}
	return nil, nil
}
