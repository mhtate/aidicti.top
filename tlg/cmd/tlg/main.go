package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	//TODO change to top

	"aidicti.top/api/protogen_tlg"
	"aidicti.top/api/protogen_uis"
	"aidicti.top/pkg/discovery"
	"aidicti.top/pkg/discoveryclient"
	discoverygrpc "aidicti.top/pkg/discoveryclient/grpc/client"
	"aidicti.top/pkg/discoveryclient/grpc/reconnectclient"
	"aidicti.top/pkg/initial"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/socket"
	"aidicti.top/pkg/utils"
	"aidicti.top/tlg/internal/controller"
	"aidicti.top/tlg/internal/gateway"
	"aidicti.top/tlg/internal/gateway/rds"
	"aidicti.top/tlg/internal/gateway/uis"
	"aidicti.top/tlg/internal/handler"
	"aidicti.top/tlg/internal/model"
	"aidicti.top/tlg/internal/service"
)

type MessageSender interface {
	SendMessage(message model.Message) error
}

type handlerDecorator struct {
	H MessageSender
}

func (h handlerDecorator) SendMessage(message model.Message) error {
	utils.Assert(h.H != nil, "asd")

	return h.H.SendMessage(message)
}

func main() {
	port := socket.NewRandomPort()
	initial.Init(pkgmodel.ServiceTLG, port, func(reg discovery.Registry, addr string) {

		gtw := gateway.New()

		ctx, cancel := context.WithCancel(context.Background())

		gtw.Run(ctx)

		defer func() {
			cancel()
		}()

		var d func(ctx context.Context, message model.Message) error = nil
		pd := &d

		//TODO what is it?
		f := func(ctx context.Context, message model.Message) error {
			return (*pd)(ctx, message)
		}

		// RDSclient := redis.NewClient(&redis.Options{
		// 	Addr:     "172.19.0.5:6379", // Redis server address
		// 	Password: "",                // No password by default
		// 	DB:       0,                 // Default DB
		// })

		// err := RDSclient.Ping().Err()
		// utils.Assert(err == nil, "")

		// rdsGtw := rds.New(RDSclient)
		rdsGtw := rds.New(nil)
		// go rdsGtw.Run(ctx)

		// uis
		connUIS, err := reconnectclient.New(
			func(ctx context.Context) (discoveryclient.Client, error) {
				return discoverygrpc.New(reg, pkgmodel.ServiceUIS)
			})
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer connUIS.Close()

		clientUIS := protogen_uis.NewServiceUISClient(connUIS)
		gatewayUIS := uis.New(clientUIS)

		go gatewayUIS.Run(context.Background())

		srvc := service.New(gtw, gatewayUIS, rdsGtw, f)

		hd := handlerDecorator{}

		ctrl := controller.New(*srvc, hd)

		d = func(ctx context.Context, message model.Message) error {
			logging.Info("d = func(ctx context.Context, message model.Message) {")
			return ctrl.SendMessage(message)
		}

		createNewClient := func(ctx context.Context) (discoveryclient.Client, error) {
			return discoverygrpc.New(reg, pkgmodel.ServiceSCN)
		}

		conn, err := reconnectclient.New(createNewClient)
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := protogen_tlg.NewTelegramServiceClient(conn)

		h := handler.New(context.Background(), ctrl, client)

		ctrl.H = h

		// hd.H = h

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s, shutting down...\n", sig)
	})
}
