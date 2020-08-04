package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func main() {
	ctx := context.Background()

	// Step 1: Create a client
	clientCfg := client.NewConfiguration(
		"http://localhost:8080",
		"ckb-rosetta-sdk-go",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)

	client := client.NewAPIClient(clientCfg)

	networkList, rosettaErr, err := client.NetworkAPI.NetworkList(
		ctx,
		&types.MetadataRequest{},
	)
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}

	if len(networkList.NetworkIdentifiers) == 0 {
		log.Fatal("no available networks")
	}

	primaryNetwork := networkList.NetworkIdentifiers[0]

	// Construction derive
	pub, _ := hex.DecodeString("020ea44dd70b0116ab44ade483609973adf5ce900d7365d988bc5f352b68abe50b")
	addr, rosettaErr, err := client.ConstructionAPI.ConstructionDerive(ctx, &types.ConstructionDeriveRequest{
		NetworkIdentifier: primaryNetwork,
		PublicKey: &types.PublicKey{
			Bytes:     pub,
			CurveType: types.Secp256k1,
		},
	})
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Construction derive: %s\n", addr.Address)

	// Construction preprocess
	preprocessRes, rosettaErr, err := client.ConstructionAPI.ConstructionPreprocess(ctx, &types.ConstructionPreprocessRequest{
		NetworkIdentifier: primaryNetwork,
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 0,
				},
				Type: "Transfer",
				Account: &types.AccountIdentifier{
					Address: "ckb1qyqt705jmfy3r7jlvg88k87j0sksmhgduazqrr2qt2",
				},
				Amount: &types.Amount{
					Value: "6100000000",
					Currency: &types.Currency{
						Symbol:   "CKB",
						Decimals: 8,
					},
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 1,
				},
				Type: "Transfer",
				Account: &types.AccountIdentifier{
					Address: "ckb1qyqwmndf2yl6qvxwgvyw9yj95gkqytgygwasshh9m8",
				},
				Amount: &types.Amount{
					Value: "-6200000000",
					Currency: &types.Currency{
						Symbol:   "CKB",
						Decimals: 8,
					},
				},
			},
		},
	})
	if rosettaErr != nil {
		log.Printf("Rosetta Error: %+v\n", rosettaErr)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Construction preprocess res: %v\n", preprocessRes)
}
