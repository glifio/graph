package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/filecoin-project/go-address"
	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin"
	builtin1 "github.com/filecoin-project/specs-actors/actors/builtin"
	util "github.com/glifio/graph/internal/utils"
	"github.com/spf13/cobra"
)

func init() {
	lotusCmd.Flags().StringVarP(&flag_address, "address", "a", "", "The address of the txn")
	lotusCmd.Flags().Uint64VarP(&flag_method, "method", "m", 0, "The method used")
	rootCmd.AddCommand(lotusCmd)
}

var lotusCmd = &cobra.Command{
  Use:   "lotus",
  Short: "Test lotus",
  Long:  `Test lotus`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Glif Graph - Lotus")

	// handle any missing args
	switch {
	case flag_address == "":
		fmt.Fprintln(os.Stderr, "Missing address - please provide the address for the record you'd like to create")
		return
	case flag_method == 0:
		fmt.Fprintln(os.Stderr, "Missing method - please provide the method that you would like to create")
		return
	}

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	
	authToken := "<value found in ~/.lotus/token>"
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	
	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), config.LotusAddress+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		log.Fatalf("connecting with lotus failed: %s", err)
	}
	defer closer()

	addr, err := address.NewFromString(flag_address)
	fmt.Println("address: ", addr)
	   // Now you can call any API you're interested in.
	tipset, err := api.ChainHead(context.Background())
	actor, err := api.StateGetActor(context.Background(), addr, tipset.Key() )
	if err != nil {
		log.Fatalf("calling chain head: %s", err)
	}

	fmt.Println("Current actor cid: ", actor.Code)
	fmt.Println("Current actor bal: ", actor.Balance)
	// test, err := cid.Decode("bafkqadlgnfwc6nzpmfrwg33vnz2a")

	// bafkqadlgnfwc6nrpmfrwg33vnz2a
	//bafkqadlgnfwc6nzpmfrwg33vnz2a
	//bafkqadlgnfwc6nrpmfrwg33vnz2a

	m := make(map[string]interface{})
	n := make(map[abi.MethodNum]string)
	elem := reflect.ValueOf(&builtin1.MethodsMultisig).Elem()
  	relType := elem.Type()
  	for i := 0; i < relType.NumField(); i++ {
    	m[relType.Field(i).Name] = elem.Field(i).Interface().(abi.MethodNum)
		n[elem.Field(i).Interface().(abi.MethodNum)] = relType.Field(i).Name
  	}
//  	fmt.Println(m)
//  	fmt.Println(n)

//	test := structs.Map(builtin1.MethodsMultisig)
//	fmt.Println("constructor: ",test)
	switch(strings.Split(builtin.ActorNameByCode(actor.Code), "/")[2]){
	case "account":
		fmt.Println("Method: ", n[abi.MethodNum(flag_method)]);	
		break;
	}

	fmt.Println("Current actor name: ", strings.Split(builtin.ActorNameByCode(actor.Code), "/")[2])
  },
}

func findByValue(m map[string]string, value string) string {
	for key, val := range m {
		if val == value {
			return key
		}
	}
	return ""
}

// Address + Method => MethodName