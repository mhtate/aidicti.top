package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"aidicti.top/api/protogen_stt"
	"aidicti.top/pkg/discovery"
	"aidicti.top/pkg/initial"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/socket"
	"aidicti.top/pkg/utils"
	"aidicti.top/stt/internal/controller"
	"aidicti.top/stt/internal/gateway"
	"aidicti.top/stt/internal/handler"
	"aidicti.top/stt/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := socket.NewRandomPort()
	initial.Init(pkgmodel.ServiceSTT, port, func(reg discovery.Registry, addr string) {
		gtw := gateway.New()
		defer gtw.Close()

		srvc := service.NewService(gtw)
		ctrl := controller.New(srvc)

		h := handler.New(ctrl)

		srv := grpc.NewServer()
		reflection.Register(srv)
		protogen_stt.RegisterSTTServiceServer(srv, h)
		fmt.Println("Listenning to a port")

		lis := socket.NewListener(addr)
		go func() {
			err := srv.Serve(lis)
			utils.Assert(err == nil, "serve port failed")
		}()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s, shutting down...\n", sig)
	})
}
