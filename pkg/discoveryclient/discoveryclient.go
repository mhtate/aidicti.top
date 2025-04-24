package discoveryclient

import (
	"context"

	"google.golang.org/grpc"
)

type Client interface {
	Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
	Close() error
}
