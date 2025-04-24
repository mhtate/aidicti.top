package client

import (
	"context"
	"fmt"
	"math/rand"

	"aidicti.top/pkg/discovery"
	"aidicti.top/pkg/discoveryclient"
	"aidicti.top/pkg/logging"
	"google.golang.org/grpc"
)

type grpcClient struct {
	reg         discovery.Registry
	serviceName string
	// client      grpc.ClientConnInterface
	client discoveryclient.Client
}

func New(reg discovery.Registry, serviceName string) (*grpcClient, error) {
	return &grpcClient{
		reg:         reg,
		serviceName: serviceName,
	}, nil
}

func (c *grpcClient) Invoke(
	ctx context.Context,
	method string,
	args any,
	reply any,
	opts ...grpc.CallOption) error {

	client, err := c.getClient(ctx)
	if err != nil {
		return err
	}

	return client.Invoke(ctx, method, args, reply, opts...)
}

func (c *grpcClient) NewStream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	opts ...grpc.CallOption) (grpc.ClientStream, error) {

	client, err := c.getClient(ctx)
	if err != nil {
		return nil, err
	}

	return client.NewStream(ctx, desc, method, opts...)
}

func (c *grpcClient) Close() error {
	if c.client != nil {
		logging.Debug("close client")
		return c.closeClient()
	}

	logging.Debug("close client ok [no client]")
	return nil
}

func (c *grpcClient) createClient(ctx context.Context) error {
	addrs, err := c.reg.ServiceAddresses(ctx, c.serviceName)
	if err != nil {
		logging.Warn("req registry failed", "srvc", c.serviceName, "err", err)
		return err
	}

	shuffle := func(arr []string) []string {
		rand.Shuffle(
			len(arr),
			func(i, j int) {
				arr[i], arr[j] = arr[j], arr[i]
			})

		return arr
	}

	for _, addr := range shuffle(addrs) {
		logging.Debug("try connect started", "srvc", c.serviceName, "addr", addr)

		//TODO we should fix unsecure behaviour
		conn, err := grpc.NewClient(addr, grpc.WithInsecure())
		if err == nil {
			c.client = conn

			logging.Info("try connect ok", "srvc", c.serviceName, "addr", addr)
			return nil
		}

		logging.Debug("try connect failed", "srvc", c.serviceName, "addr", addr, "err", err)
	}

	logging.Warn("try connect all addrs failed", "srvc", c.serviceName)

	return fmt.Errorf("create client failed")
}

func (c *grpcClient) closeClient() error {
	return c.client.Close()
}

// NOTE In case we have to change interface without using Cancel func
//      we may call it from runtime package
//
// func (c *grpcClient) closeClient() error {
// 	val := reflect.ValueOf(c.client)

// 	method := val.MethodByName("Close")
// 	if !method.IsValid() {
// 		logging.Warn("find method 'Close' failed")
// 		return fmt.Errorf("close client failed")
// 	}

// 	out := method.Call([]reflect.Value{})
// 	utils.Assert(len(out) == 1, "expect out param only err failed")

// 	iface := out[0].Interface()

// 	if iface == nil {
// 		return nil
// 	}

// 	err, ok := iface.(error)
// 	utils.Assert(ok, "expect out param cast err failed")

// 	return err
// }

func (c *grpcClient) getClient(ctx context.Context) (grpc.ClientConnInterface, error) {
	if c.client != nil {
		return c.client, nil
	}

	err := c.createClient(ctx)

	return c.client, err
}
