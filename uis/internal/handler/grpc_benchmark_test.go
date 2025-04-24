package handler_test

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"aidicti.top/api/oxf_gen_proto"
	"aidicti.top/oxf/internal/controller"
	gateway "aidicti.top/oxf/internal/gateway/oxf"
	"aidicti.top/oxf/internal/handler"
	"aidicti.top/oxf/internal/model"
	"aidicti.top/oxf/internal/service"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/processor"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/socket"
	"aidicti.top/pkg/utils"
	"github.com/bojand/ghz/printer"
	"github.com/bojand/ghz/runner"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type mockProcessor struct {
	Resps_ chan gateway.Resp
}

func (p mockProcessor) Process(ctx context.Context, req gateway.Req) error {
	minDelay := 30
	maxDelay := 600
	processDelay := rand.Intn(maxDelay-minDelay+1) + minDelay

	successRatePercentage := 80
	isSuccess := rand.Intn(100) < successRatePercentage

	time.Sleep(time.Millisecond * time.Duration(processDelay))

	if isSuccess {
		p.Resps_ <- gateway.Resp{
			ReqData:         req.ReqData,
			DictionaryEntry: model.DictionaryEntry{Word: req.Text},
		}

		return nil
	}

	return fmt.Errorf("")
}

type mockGateway struct {
	Reqs_  chan gateway.Req
	Resps_ chan gateway.Resp
}

func (g mockGateway) Reqs() chan<- gateway.Req {
	return g.Reqs_
}

func (g mockGateway) Resps() <-chan gateway.Resp {
	return g.Resps_
}

func Benchmark_Handler(b *testing.B) {
	logging.Set(slog.New(slog.NewJSONHandler(logging.NullWriter{}, nil)))

	ctx, cancel := context.WithCancel(context.Background())

	gtw := mockGateway{
		Reqs_:  make(chan gateway.Req, 64),
		Resps_: make(chan gateway.Resp, 64),
	}

	reqLstn := listener.New(gtw.Reqs_, mockProcessor{Resps_: gtw.Resps_})
	srvc := service.New(gtw)
	ctrl := controller.New(srvc)
	h := handler.New(ctrl)

	go reqLstn.Run(ctx)
	go srvc.Run(ctx)

	time.Sleep(10 * time.Millisecond)

	usersCount := 32
	batchesCount := 2

	sem := make(chan struct{}, batchesCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			w := oxf_gen_proto.Word{
				Id:     rand.Uint64(),
				UserId: rand.Uint64() % uint64(usersCount),
				Text:   strconv.FormatUint(rand.Uint64(), 10)}

			entry, err := h.GetDictEntry(ctx, &w)
			if (err != nil) && (entry.Word != w.Text) {
				panic("")
			}
		}()
	}

	_ = cancel
}

func Benchmark_Service(b *testing.B) {
	// logging.Set(slog.New(slog.NewJSONHandler(logging.NullWriter{}, nil)))
	// logging.Set(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	file, err := os.Create("output")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	logging.Set(slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := context.WithCancel(context.Background())

	gtw := mockGateway{
		Reqs_:  make(chan gateway.Req, 256),
		Resps_: make(chan gateway.Resp, 256),
	}

	reqLstn := listener.New(gtw.Reqs_,
		processor.New(mockProcessor{Resps_: gtw.Resps_}, processor.BatchCount(16)))
	srvc := service.New(gtw)
	ctrl := controller.New(srvc)
	h := handler.New(ctrl)

	addr := fmt.Sprintf("localhost:%d", socket.NewRandomPort())

	srv := grpc.NewServer()
	reflection.Register(srv)
	oxf_gen_proto.RegisterServiceOXFServer(srv, h)
	lis := socket.NewListener(addr)

	go func() {
		err := srv.Serve(lis)
		utils.Assert(err == nil, "serve port failed")
	}()
	go reqLstn.Run(ctx)
	go srvc.Run(ctx)

	time.Sleep(10 * time.Millisecond)

	// usersCount := 32
	// batchesCount := 2
	RPC := uint(8)

	report, err := runner.Run(
		"ServiceOXF.GetDictEntry",
		addr,
		runner.WithProtoFile("oxf.proto", []string{"./api", "../api", "../../api", "../../../api"}),
		runner.WithRunDuration(time.Duration(30*time.Second)),
		runner.WithInsecure(true),
		runner.WithRPS(RPC),
		runner.WithDataProvider(
			func(cd *runner.CallData) ([]*dynamic.Message, error) {
				const UsersCount = 128

				w := oxf_gen_proto.Word{
					Id:     rand.Uint64(),
					UserId: rand.Uint64() % uint64(UsersCount),
					Text:   strconv.FormatUint(rand.Uint64(), 10)}

				entry, err := h.GetDictEntry(ctx, &w)
				dynamicMsg, err := dynamic.AsDynamicMessage(entry)
				if err != nil {
					return nil, err
				}
				return []*dynamic.Message{dynamicMsg}, nil
			}),
		runner.WithAsync(true),
	)

	if err != nil {
		panic("")
	}

	printer := printer.ReportPrinter{
		Out:    os.Stdout,
		Report: report,
	}

	printer.Print("pretty")

	_ = cancel
}
