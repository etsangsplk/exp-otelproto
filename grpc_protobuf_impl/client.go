package grpc_protobuf_impl

import (
	"context"
	"log"

	"github.com/tigrannajaryan/exp-otelproto/traceprotobuf"

	"google.golang.org/grpc"

	"github.com/tigrannajaryan/exp-otelproto/core"
)

// Client can connect to a server and send a batch of spans.
type Client struct {
	client traceprotobuf.TracerClient
}

func (c *Client) Connect(server string) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c.client = traceprotobuf.NewTracerClient(conn)
	return nil
}

func (c *Client) Export(batch core.SpanBatch) {
	c.client.SendBatch(context.Background(), batch.(*traceprotobuf.SpanBatch))
}