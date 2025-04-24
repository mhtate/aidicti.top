package rds

import (
	"context"
	"fmt"
	"strconv"

	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"github.com/go-redis/redis/v7"
)

const (
	Get = iota
	Set
)

type Method int

type Req struct {
	pkgmodel.ReqData
	Method Method
	Mp     map[string]string
}

type Resp struct {
	pkgmodel.ReqData
	Mp map[string]string
}

type gateway struct {
	clnt  *redis.Client
	reqs  chan Req
	resps chan Resp
	// initPrF func(gatewaydialog.Processor[Req]) gatewaydialog.Processor[Req]
}

func (g gateway) Reqs() chan<- Req {
	return g.reqs
}

func (g gateway) Resps() <-chan Resp {
	return g.resps
}

func New(client *redis.Client) *gateway {
	return &gateway{
		clnt:  client,
		reqs:  make(chan Req, 64),
		resps: make(chan Resp, 64),
	}
}

type reqProcessor struct {
	resps chan<- Resp
	clnt  *redis.Client
}

func NewProcessor() *reqProcessor {
	return &reqProcessor{}
}

func (p reqProcessor) Process(ctx context.Context, req Req) error {
	utils.Assert(p.clnt != nil, "")

	resp := Resp{ReqData: req.ReqData}

	if req.Method == Get {
		mp, err := p.clnt.HGetAll(strconv.FormatUint(uint64(req.ReqId), 10)).Result()
		if err != nil {
			return fmt.Errorf("errror")
		}

		resp.Mp = mp

		select {
		case p.resps <- resp:
			logging.Debug("redis req ok")

		default:
			logging.Debug("redis req ok dont listen")
			return fmt.Errorf("errror")
		}
	}

	if req.Method == Set {
		utils.Assert(req.Mp != nil, "")
		v := strconv.FormatUint(uint64(req.ReqId), 10)

		g := make([]string, 0)
		for i, value := range req.Mp {
			g = append(g, i)
			g = append(g, value)
		}

		_, err := p.clnt.HSet(v, g).Result()
		if err != nil {
			return err
		}
	}

	return nil
}

func (g gateway) Run(ctx context.Context) {
	utils.Assert(g.clnt != nil, "req word size == 0")

	reqPr := reqProcessor{g.resps, g.clnt}

	reqLstn := listener.New(g.reqs, reqPr)

	reqLstn.Run(ctx)
}
