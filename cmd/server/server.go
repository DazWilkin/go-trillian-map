package main

import (
	"context"
	"log"

	"github.com/google/trillian"
)

// Client is a type that represents a Trillian Map Client
type Client struct {
	client trillian.TrillianMapWriteClient
	mapID  int64
}

// NewClient is a function that creates a new Client
func NewClient(client trillian.TrillianMapWriteClient, mapID int64) *Client {
	return &Client{
		client: client,
		mapID:  mapID,
	}
}

// Add is a function that adds leaves to a Map
func (c *Client) Add(ctx context.Context, leaves []*trillian.MapLeaf, revision int64) error {
	rqst := &trillian.WriteMapLeavesRequest{
		MapId:          c.mapID,
		Leaves:         leaves,
		ExpectRevision: revision,
	}
	resp, err := c.client.WriteLeaves(ctx, rqst)
	if err != nil {
		return err
	}

	log.Printf("[Client:Add] %+v", resp)
	return nil
}

// Get is a function that gets specific leaf-revisions from a Map
func (c *Client) Get(ctx context.Context, leaf *trillian.MapLeaf, revision int64) ([]*trillian.MapLeaf, error) {
	rqst := &trillian.GetMapLeavesByRevisionRequest{
		MapId: c.mapID,
		Index: [][]byte{
			leaf.Index,
		},
		Revision: revision,
	}
	resp, err := c.client.GetLeavesByRevision(ctx, rqst)
	if err != nil {
		return nil, err
	}

	log.Printf("[Client:Get] %+v", resp)
	return resp.GetLeaves(), nil
}
