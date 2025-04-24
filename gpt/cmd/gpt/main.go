package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aidicti.top/api/protogen_gpt"
	"aidicti.top/gpt/internal/controller"
	"aidicti.top/gpt/internal/gateway"
	"aidicti.top/gpt/internal/gateway/gpt"
	"aidicti.top/gpt/internal/handler"
	"aidicti.top/gpt/internal/repository/memory"
	"aidicti.top/gpt/internal/service"
	"aidicti.top/pkg/discovery"
	"aidicti.top/pkg/initial"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/socket"
	"aidicti.top/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := socket.NewRandomPort()
	initial.Init(pkgmodel.ServiceGPT, port, func(reg discovery.Registry, addr string) {

		repos := memory.NewTemporarilyDialogStorage(time.Hour)
		gtw := gateway.New()
		gtwGPT := gpt.New()

		srvc := service.NewService(gtw, gtwGPT)
		ctrl := controller.New(srvc, repos)

		h := handler.New(ctrl)

		srv := grpc.NewServer()
		reflection.Register(srv)
		protogen_gpt.RegisterAIProviderServiceServer(srv, h)

		lis := socket.NewListener(addr)
		go func() {
			err := srv.Serve(lis)
			utils.Assert(err == nil, "serve port failed")
		}()

		ctx := context.Background()

		go srvc.Run(ctx)
		go gtwGPT.Run(ctx)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s, shutting down...\n", sig)
	})
}
