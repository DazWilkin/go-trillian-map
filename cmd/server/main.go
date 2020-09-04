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

	client := NewClient(c, *tMapID)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rev := int64(2)
	for k, v := range map[string]string{
		"Freddie":  "Border Collie",
		"Noisette": "Australian Shepherd",
		"Artie":    "Weimeraner",
		"Louie":    "Formosan Mountain Dog",
		"Luna":     "Yorkiepoo",
		"Bertie":   "Cockapoo",
		"Unnamed":  "Scottish Terrier",
	} {
		log.Printf("[main] %s", k)

		hasher := sha256.New()
		hasher.Write([]byte(k))
		index := hasher.Sum(nil)

		log.Printf("[main] %s: %x", k, index)

		leaf := &trillian.MapLeaf{
			Index:     index,
			LeafValue: []byte(v),
		}
		log.Printf("[main] Leaf:\n%+v", leaf)

		// Add
		{
			log.Printf("[main:Add] %s", k)
			err := client.Add(ctx, leaf, rev)
			if err != nil {
				log.Fatal(err)
			}
		}
		// Get
		{
			log.Printf("[main:Get] %s", k)
			err := client.Get(ctx, leaf, rev)
			if err != nil {
				log.Fatal(err)
			}
		}
		rev++
		log.Print("[main] Sleeping 5 seconds")
		time.Sleep(5 * time.Second)
	}
}
