package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aidicti.top/api/protogen_tlg"
	"aidicti.top/api/protogen_uis"
	"aidicti.top/pkg/discovery"
	"aidicti.top/pkg/discoveryclient"
	discoverygrpc "aidicti.top/pkg/discoveryclient/grpc/client"
	"aidicti.top/pkg/discoveryclient/grpc/reconnectclient"
	"aidicti.top/pkg/initial"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/socket"
	"aidicti.top/pkg/utils"
	"aidicti.top/uis/internal/controller"
	"aidicti.top/uis/internal/gateway/dbs"
	"aidicti.top/uis/internal/gateway/scn"
	"aidicti.top/uis/internal/handler"
	"aidicti.top/uis/internal/repo/memory"
	"aidicti.top/uis/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := socket.NewRandomPort()
	initial.Init(pkgmodel.ServiceUIS, port, func(reg discovery.Registry, addr string) {

		ctx, cancel := context.WithCancel(context.Background())

		// uis
		gatewayDBS := dbs.New("aidicti_postgres", "dKhfb4Uhfy4", "fbS5fb44UhC")
		go gatewayDBS.Run(ctx)

		connSCN, err := reconnectclient.New(
			func(ctx context.Context) (discoveryclient.Client, error) {
				return discoverygrpc.New(reg, pkgmodel.ServiceSCN)
			})

		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer connSCN.Close()

		client := protogen_tlg.NewTelegramServiceClient(connSCN)

		gatewaySCN := scn.New(client)
		go gatewaySCN.Run(ctx)

		repo := memory.New()

		srvc := service.New(repo, gatewayDBS, gatewaySCN)
		go srvc.Run(ctx)

		ctrl := controller.New(srvc)

		h := handler.New(ctrl)

		srv := grpc.NewServer()
		reflection.Register(srv)
		protogen_uis.RegisterServiceUISServer(srv, h)
		lis := socket.NewListener(addr)

		go func() {
			err := srv.Serve(lis)
			utils.Assert(err == nil, "serve port failed")
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		cancel()
		time.Sleep(1 * time.Second)

		fmt.Printf("\nReceived signal: %s, shutting down...\n", sig)
	})
}
