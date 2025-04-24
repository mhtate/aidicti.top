package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aidicti.top/api/protogen_oxf"
	"aidicti.top/oxf/internal/controller"
	gateway "aidicti.top/oxf/internal/gateway/oxf"
	"aidicti.top/oxf/internal/handler"
	"aidicti.top/oxf/internal/service"
	"aidicti.top/pkg/discovery"
	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/gatewaydialog/processor/retry"
	"aidicti.top/pkg/initial"
	"aidicti.top/pkg/socket"
	"aidicti.top/pkg/utils"

	pkgmodel "aidicti.top/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := socket.NewRandomPort()
	initial.Init(pkgmodel.ServiceOXF, port, func(reg discovery.Registry, addr string) {

		ctx, cancel := context.WithCancel(context.Background())

		gtw := gateway.New()
		gtw.SetProcessor(
			func(p gatewaydialog.Processor[gateway.Req]) gatewaydialog.Processor[gateway.Req] {
				return retry.New(p, 1*time.Second, true, 3)
			})
		go gtw.Run(ctx)

		srvc := service.New(gtw)
		go srvc.Run(ctx)

		ctrl := controller.New(srvc)

		h := handler.New(ctrl)

		srv := grpc.NewServer()
		reflection.Register(srv)
		protogen_oxf.RegisterServiceOXFServer(srv, h)
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
