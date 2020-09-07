package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"log"
	"time"

	"github.com/google/trillian"
	"google.golang.org/grpc"
)

// Build info
var (
	buildTime string
	gitCommit string
)

var (
	tMapEndpoint = flag.String("tmap_endpoint", "", "The gRPC endpoint of the Trillian Map Server")
	tMapID       = flag.Int64("tmap_id", 0, "Trillian Map ID")
	tMapRevision = flag.Int64("tmap_rev", 0, "Trillian Map ID current revision")
)

func main() {
	// Print build info
	log.Printf("[main] BuildTime: %s", buildTime)
	log.Printf("[main] GitCommit: %s", gitCommit)

	flag.Parse()
	if *tMapID == 0 {
		log.Fatal("--tmap_id must be non-zero")
	}
	if *tMapRevision == 0 {
		log.Fatal("--tmap_rev must be non-zero")
	}

	// Create Trillian Map (gRPC) Client
	conn, err := grpc.Dial(*tMapEndpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := trillian.NewTrillianMapWriteClient(conn)

	client := NewClient(c, *tMapID)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// `rev` must start at +1 previous `rev` values for a given map
	// There are 7 examples, but these are added as a single batch
	// So, each iteration should be `rev`+1 (one)
	rev := int64(*tMapRevision)

	// Add many
	{
		leaves := make([]*trillian.MapLeaf, len(Examples))
		i := 0
		for k, v := range Examples {
			log.Printf("[main:add] %s", k)

			hasher := sha256.New()
			hasher.Write([]byte(k))
			index := hasher.Sum(nil)

			log.Printf("[main:add] %s: %x", k, index)

			leaf := &trillian.MapLeaf{
				Index:     index,
				LeafValue: []byte(v),
			}
			log.Printf("[main:add] Leaf #%02d:\n%+v", i, leaf)
			leaves[i] = leaf
			i = i + 1
		}
		{
			log.Print("[main:add] Add'ing")
			err := client.Add(ctx, leaves, rev)
			if err != nil {
				log.Fatal(err)
			}
		}

	}
	// Get each
	{
		for k, v := range Examples {
			log.Printf("[main:get] %s", k)

			hasher := sha256.New()
			hasher.Write([]byte(k))
			index := hasher.Sum(nil)

			log.Printf("[main:get] %s: %x", k, index)

			leaf := &trillian.MapLeaf{
				Index:     index,
				LeafValue: []byte(v),
			}
			log.Printf("[main:get] Leaf:\n%+v", leaf)

			{
				log.Printf("[main:get] Get'ing %s", k)
				leaves, err := client.Get(ctx, leaf, rev)
				if err != nil {
					log.Fatal(err)
				}
				for i, leaf := range leaves {
					log.Printf("[main:get] Leaf #%02d:\n%+v\n", i, leaf)
				}
			}
			log.Print("[main:get] Sleeping 1 second")
			time.Sleep(1 * time.Second)
		}
	}
}
