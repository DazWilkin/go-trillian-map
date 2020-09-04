package main

import (
	"context"
	"encoding/hex"
	"flag"
	"log"

	"github.com/google/trillian"
	"google.golang.org/grpc"
)

var (
	tMapEndpoint = flag.String("tmap_endpoint", "", "The gRPC endpoint of the Trillian Map Server")
	tMapID       = flag.Int64("tmap_id", 0, "Trillian Map ID")
)

func main() {
	flag.Parse()

	// Create Trillian Map (gRPC) Client
	conn, err := grpc.Dial(*tMapEndpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := trillian.NewTrillianMapWriteClient(conn)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := NewClient(c, *tMapID)

	index, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		log.Fatal(err)
	}

	leaf := &trillian.MapLeaf{
		Index:     index,
		LeafValue: []byte("Freddie"),
	}

	// Add
	{
		err := client.Add(ctx, leaf, 1)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Get
	{
		err := client.Get(ctx, leaf, 1)
		if err != nil {
			log.Fatal(err)
		}
	}
}
