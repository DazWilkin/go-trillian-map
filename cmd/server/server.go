package main

import (
	"context"
	"log"

	"github.com/google/trillian"
)

type Client struct {
	client trillian.TrillianMapWriteClient
	mapID  int64
}

func NewClient(client trillian.TrillianMapWriteClient, mapID int64) *Client {
	return &Client{
		client: client,
		mapID:  mapID,
	}
}

func (c *Client) Add(ctx context.Context, leaf *trillian.MapLeaf, revision int64) error {
	rqst := &trillian.WriteMapLeavesRequest{
		MapId:          c.mapID,
		Leaves:         []*trillian.MapLeaf{leaf},
		ExpectRevision: revision,
	}
	resp, err := c.client.WriteLeaves(ctx, rqst)
	if err != nil {
		return err
	}

	log.Printf("[Client:Add] %+v", resp)
	return nil
}
func (c *Client) Get(ctx context.Context, leaf *trillian.MapLeaf, revision int64) error {
	rqst := &trillian.GetMapLeavesByRevisionRequest{
		MapId: c.mapID,
		Index: [][]byte{
			leaf.Index,
		},
		Revision: revision,
	}
	resp, err := c.client.GetLeavesByRevision(ctx, rqst)
	if err != nil {
		return err
	}

	log.Printf("[Client:Get] %+v", resp)
	return nil
}
