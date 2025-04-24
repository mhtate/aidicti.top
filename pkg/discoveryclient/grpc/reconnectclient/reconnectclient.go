package reconnectclient

import (
	"context"
	"fmt"

	"aidicti.top/pkg/discoveryclient"
	"aidicti.top/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type reconnectClient struct {
	create func(context.Context) (discoveryclient.Client, error)
	client discoveryclient.Client
}

func New(create func(context.Context) (discoveryclient.Client, error)) (*reconnectClient, error) {
	return &reconnectClient{
		create: create,
	}, nil
}

func (c *reconnectClient) Invoke(
	ctx context.Context,
	method string,
	args any,
	reply any,
	opts ...grpc.CallOption) error {

	if c.client == nil {
		c.reconnect(ctx)
		if c.client == nil {
			logging.Warn("create client fail")
			return fmt.Errorf("create client fail")
		}
	}

	err := c.client.Invoke(ctx, method, args, reply, opts...)

	if err == nil {
		return err
	}

	if !c.isReconnectCase(ctx, err) {
		logging.Debug("invoke fail [no reconnect]", "err", err)
		return err
	}

	logging.Debug("invoke fail [reconnect]", "err", err)

	c.reconnect(ctx)

	return err
}

func (c *reconnectClient) NewStream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {

	if c.client == nil {
		c.reconnect(ctx)
		if c.client == nil {
			logging.Warn("create client fail")
			return nil, fmt.Errorf("create client fail")
		}
	}

	stream, err := c.client.NewStream(ctx, desc, method, opts...)

	if err == nil {
		return stream, err
	}

	if !c.isReconnectCase(ctx, err) {
		logging.Debug("new stream fail [no reconnect]", "err", err)
		return stream, err
	}

	logging.Debug("new stream fail [reconnect]", "err", err)

	c.reconnect(ctx)

	return stream, err
}

func (c *reconnectClient) Close() error {
	if c.client != nil {
		logging.Debug("close client")

		client := c.client
		c.client = nil

		return client.Close()
	}

	logging.Debug("close client ok [no client]")
	return nil
}

func (c *reconnectClient) isReconnectCase(_ context.Context, err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		logging.Debug("check reconnect case fail [no gRPC error]", "err", err)
		return false
	}

	if st.Code() == codes.Unavailable {
		logging.Info("check reconnect case ok")
		return true
	}

	logging.Debug("check reconnect case fail [other gRPC error]", "err", err)
	return false
}

func (c *reconnectClient) reconnect(ctx context.Context) {
	if c.client != nil {
		err := c.Close()
		if err != nil {
			logging.Debug("close client fail", "err", err)
		} else {
			logging.Debug("close client ok")
		}
	}

	client, err := c.create(ctx)
	if err != nil {
		logging.Warn("reconnect client fail", "err", err)
	}

	c.client = client
	logging.Info("reconnect client ok")
}
