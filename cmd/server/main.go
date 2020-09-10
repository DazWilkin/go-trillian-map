package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
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
			log.Printf("[main:add] Leaf #%02d: %s", i, toString(leaf))
			leaves[i] = leaf
			i++
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
		for k := range Examples {
			log.Printf("[main:get] %s", k)

			hasher := sha256.New()
			hasher.Write([]byte(k))
			index := hasher.Sum(nil)

			log.Printf("[main:get] %s: %x", k, index)

			{
				log.Printf("[main:get] Get'ing %s", k)
				indexes := [][]byte{
					index,
				}
				leaves, err := client.Get(ctx, indexes, rev)
				if err != nil {
					log.Fatal(err)
				}
				for i, leaf := range leaves {
					log.Printf("[main:get] Leaf #%02d: %s", i, toString(leaf))
				}
			}
			log.Print("[main:get] Sleeping 1 second")
			time.Sleep(1 * time.Second)
		}
	}

	// Get prior revisions
	{
		log.Print("[main:revisions] Build indexes")
		indexes := make([][]byte, len(Examples))
		i := 0
		for k := range Examples {
			log.Printf("[main:revisions] %s", k)

			hasher := sha256.New()
			hasher.Write([]byte(k))
			index := hasher.Sum(nil)
			indexes[i] = index
			i++
		}

		log.Print("[main:revisions] Iterate")
		for j := rev; j >= 1; j-- {
			log.Printf("[main:revisions] Rev: %02d", j)
			leaves, err := client.Get(ctx, indexes, j)
			if err != nil {
				log.Println(err)
			}
			for k, leaf := range leaves {
				log.Printf("[main:revisions] Rev %02d Leaf #%02d: %s", j, k, toString(leaf))
			}
		}
	}
}
func toString(l *trillian.MapLeaf) string {
	return fmt.Sprintf("Index: %x, Value: %s", l.GetIndex(), l.GetLeafValue())
}
