package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	pb "github.com/glifio/graph/pkg/daemon"
	"github.com/glifio/graph/pkg/graph"
	"github.com/glifio/graph/pkg/kvdb"
	"github.com/glifio/graph/pkg/node"
)

func init() {
	rootCmd.AddCommand(kvCmd)
	kvCmd.AddCommand(kvGetCmd)
	kvCmd.AddCommand(kvDelCmd)
	kvCmd.AddCommand(kvMatchCmd)
}

var kvCmd = &cobra.Command{
	Use:   "kv",
	Short: "",
	Long:  ``,
}

var kvDelCmd = &cobra.Command{
	Use:   "del",
	Short: "del key/value",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", viper.GetViper().GetUint("rpc")), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewDaemonClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.KvDel(ctx, &pb.KvRequest{Key: args[0]})
		if err != nil {
			log.Fatalf("could not del: %v", err)
		}
		log.Printf("%s", r.GetMessage())
	},
}

var kvGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", viper.GetViper().GetUint("rpc")), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewDaemonClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := c.KvGet(ctx, &pb.KvRequest{Key: args[0]})
		if err != nil {
			log.Fatalf("could not get: %v", err)
		}
		log.Printf("Get: %s", r.GetMessage())
	},
}

var kvMatchCmd = &cobra.Command{
	Use:   "match [key prefix] [length] [offset]",
	Short: "",
	Long:  ``,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", viper.GetViper().GetUint("rpc")), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewDaemonClient(conn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		length, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			log.Fatalf("invalid length: %v", err)
		}
		offset, err := strconv.ParseUint(args[2], 10, 32)
		if err != nil {
			log.Fatalf("invalid offset: %v", err)
		}

		r, err := c.KvMatch(ctx, &pb.KvRequest{Key: args[0], Lenght: length, Offset: offset})
		if err != nil {
			log.Fatalf("could not validate: %v", err)
		}
		log.Printf("Validate: %s", r.GetMessage())
	},
}

func (s *server) KvDel(ctx context.Context, in *pb.KvRequest) (*pb.KvReply, error) {
	log.Printf("grpc -> kv del: %v", in.Key)

	err := kvdb.Open().Del([]byte(in.Key))
	str := fmt.Sprintf("del: %s %s", in.Key, err)
	return &pb.KvReply{Message: str}, nil
}

func (s *server) KvGet(ctx context.Context, in *pb.KvRequest) (*pb.KvReply, error) {
	log.Printf("grpc -> kv get: %v", in.Key)
	var str string

	val, err := kvdb.Open().Get([]byte(in.Key))

	if err != nil {
		return &pb.KvReply{Message: err.Error()}, err
	}

	switch {
	case strings.HasPrefix(in.Key, "cid:"):
		str = fmt.Sprintln("Message:")
		gmsg := &graph.Message{}
		if err := proto.Unmarshal(val, gmsg); err != nil {
			log.Fatalln("Failed to parse message:", err)
			return nil, err
		}
		var msg types.Message
		if err := msg.UnmarshalCBOR(bytes.NewReader(gmsg.MessageCbor)); err != nil {
			return nil, err
		}
		cid, _ := cid.Decode(in.Key[4:])
		res := &node.SearchStateStruct{Message: api.InvocResult{MsgCid: cid, Msg: &msg}}
		item := res.CreateMessage()
		item.Height = gmsg.Height
		str = str + fmt.Sprintln(item)
	case strings.HasPrefix(in.Key, "tm:"):
		str = fmt.Sprintln("Tipset messages")
	default:
		str = fmt.Sprintln("Not sure")
	}

	return &pb.KvReply{Message: str}, nil
}

func (s *server) KvMatch(ctx context.Context, in *pb.KvRequest) (*pb.KvReply, error) {
	log.Printf("grpc -> kv match: %v", in.Key)
	var str string

	res, _ := kvdb.Open().Search([]byte(in.Key), uint(in.Lenght), uint(in.Offset))
	str = "result:\n"
	for _, item := range res {
		msg := string(item)
		str = str + msg + "\n"
	}

	return &pb.KvReply{Message: str}, nil
}
