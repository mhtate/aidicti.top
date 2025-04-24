package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"aidicti.top/api/protogen_gpt"
	"aidicti.top/api/protogen_oxf"
	"aidicti.top/api/protogen_stt"
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
	"aidicti.top/scn/internal/controller"
	"aidicti.top/scn/internal/gateway/ai"
	"aidicti.top/scn/internal/gateway/dbs"
	"aidicti.top/scn/internal/gateway/oxf"
	"aidicti.top/scn/internal/gateway/sttg"
	"aidicti.top/scn/internal/gateway/uis"
	"aidicti.top/scn/internal/handler"
	"aidicti.top/scn/internal/service"
	"google.golang.org/grpc"
)

func main() {
	port := socket.NewRandomPort()
	initial.Init(pkgmodel.ServiceSCN, port, func(reg discovery.Registry, addr string) {
		s := grpc.NewServer()

		ctrl := controller.New()

		conn, err := reconnectclient.New(
			func(ctx context.Context) (discoveryclient.Client, error) {
				return discoverygrpc.New(reg, pkgmodel.ServiceSTT)
			})

		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := protogen_stt.NewSTTServiceClient(conn)
		sttGtw := sttg.New(client)

		connAI, err := reconnectclient.New(
			func(ctx context.Context) (discoveryclient.Client, error) {
				return discoverygrpc.New(reg, pkgmodel.ServiceGPT)
			})

		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer connAI.Close()

		clientAI := protogen_gpt.NewAIProviderServiceClient(connAI)
		aiGtw := ai.New(clientAI)

		// oxf
		connOxf, err := reconnectclient.New(
			func(ctx context.Context) (discoveryclient.Client, error) {
				return discoverygrpc.New(reg, pkgmodel.ServiceOXF)
			})

		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer connOxf.Close()

		clientOxf := protogen_oxf.NewServiceOXFClient(connOxf)
		oxfGtw := oxf.New(clientOxf)

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

		gatewayDBS := dbs.New("aidicti_postgres", "dKhfb4Uhfy4", "fbS5fb44UhC")
		// gatewayDBS := dbs.New("172.19.0.4", "dKhfb4Uhfy4", "fbS5fb44UhC")

		protogen_tlg.RegisterTelegramServiceServer(s, handler.New(ctrl))
		srvc := service.New(sttGtw, aiGtw, ctrl, oxfGtw, gatewayUIS, gatewayDBS)

		go gatewayDBS.Run(context.Background())
		go srvc.Run(context.Background())
		go sttGtw.Run(context.Background())
		go aiGtw.Run(context.Background())
		go oxfGtw.Run(context.Background())
		go gatewayUIS.Run(context.Background())

		lis := socket.NewListener(addr)
		go func() {
			err := s.Serve(lis)
			utils.Assert(err == nil, "serve port failed")
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s, shutting down...\n", sig)
	})
}
