package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-rich-sdk-go/rpc"

	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/services"
)

func NewBlockchainRouter(
	network *types.NetworkIdentifier,
	asserter *asserter.Asserter,
	client rpc.Client,
) http.Handler {
	networkAPIService := services.NewNetworkAPIService(network, client)
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	blockAPIService := services.NewBlockAPIService(network, client)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	accountAPIService := services.NewAccountAPIService(network, client)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	constructionAPIService := services.NewConstructionAPIService(network, client)
	constructionAPIController := server.NewConstructionAPIController(
		constructionAPIService,
		asserter,
	)

	return server.NewRouter(networkAPIController, blockAPIController, accountAPIController, constructionAPIController)
}

func main() {
	c, err := config.Init("config.yaml")
	if err != nil {
		log.Fatalf("initial config error: %v", err)
	}

	client, err := rpc.Dial(c.RichNodeRpc+"/rpc", c.RichNodeRpc+"/indexer")
	if err != nil {
		log.Fatalf("dial rich node rpc error: %v", err)
	}

	network := &types.NetworkIdentifier{
		Blockchain: "CKB",
		Network:    c.Network,
	}

	serverAsserter, err := asserter.NewServer(services.SupportedOperationTypes, false, []*types.NetworkIdentifier{network})
	if err != nil {
		log.Fatalf("initial server error: %v", err)
	}

	router := NewBlockchainRouter(network, serverAsserter, client)
	log.Printf("Listening on port %d\n", c.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.Port), router))
}
