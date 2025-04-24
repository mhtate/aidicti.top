package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"aidicti.top/tlg/internal/gateway"
	"aidicti.top/tlg/internal/gateway/rds"
	"aidicti.top/tlg/internal/gateway/uis"
	"aidicti.top/tlg/internal/model"
	"github.com/google/uuid"
)

type TGGateway interface {
	Requests() chan<- gateway.Req
	Responses() <-chan gateway.Resp
}

type gatewayRDS interface {
	Reqs() chan<- rds.Req
	Resps() <-chan rds.Resp
}

type gatewayUIS interface {
	Reqs() chan<- uis.Req
	Resps() <-chan uis.Resp
}

type service struct {
	g   TGGateway
	uis gatewayUIS
	rds gatewayRDS
	f   func(context.Context, model.Message) error
}

func processMDdata(s *service, resp *gateway.Resp, md *gateway.ModifyData) (map[string]string, error) {
	id, err := strconv.Atoi(md.Data[7:])
	if err != nil {
		return nil, err
		// panic(err)
	}

	req := rds.Req{
		ReqData: pkgmodel.ReqData{ReqId: pkgmodel.ReqID(id), DlgId: 0},
		Method:  rds.Get,
		Mp:      make(map[string]string),
	}

	s.rds.Reqs() <- req

	select {
	case resp := <-s.rds.Resps():
		logging.Debug("get reso from rds", "map", resp.Mp, "ids", resp.ReqData)

		return resp.Mp, nil

	case <-time.After(3 * time.Second):
		logging.Debug("get reso from rds", "st", "finish", "rsn", "timeout")
	}

	return nil, fmt.Errorf("asdqwe")

}

func New(g TGGateway, ruis gatewayUIS, rdsg gatewayRDS, f func(context.Context, model.Message) error) *service {
	s := &service{g, ruis, rdsg, f}

	go func() {
		//TODO a way to go out
		for {
			select {
			case resp := <-s.g.Responses():

				// utils.Contract(resp.ReqId != 0)
				// utils.Contract(resp.DlgId != 0)

				newreqData := pkgmodel.ReqData{DlgId: pkgmodel.DlgID(resp.Message.Id)}
				if resp.Message.Id == 0 {
					newreqData.DlgId = resp.DlgId
				}

				uuidId, err := uuid.NewRandom()
				utils.Assert(err == nil, "cant create uuid")

				newreqData.ReqId = pkgmodel.ReqID(uuidId.ID())

				logging.Info("Responce got")

				m := model.Message{Id: resp.Id, Message: resp.Message.Message}

				if resp.UICall != nil {
					select {
					case s.uis.Reqs() <- uis.Req{
						ReqData:  newreqData,
						ButtonId: model.ButtonId(resp.UICall.Id),
						Clicked:  true,
					}:
						logging.Info("send click to uis")
					}

					select {
					case uis_resp := <-s.uis.Resps():
						s.g.Requests() <- gateway.Req{
							ReqData: newreqData,
							Message: model.Message{},
							UIData: []gateway.UIObject{
								gateway.UIObject{
									Id:        uint64(uis_resp.ButtonInfo.Id),
									Text:      uis_resp.ButtonInfo.Text,
									IsModied:  true,
									MessageId: resp.UICall.MessageId},
							},
						}
					}

					break
				}

				//TODO it's blocking case
				if resp.MData != nil {

					mp, err := processMDdata(s, &resp, resp.MData)
					if err != nil {
						logging.Debug("data got with err", "st", "fail")
					}

					s.g.Requests() <- gateway.Req{
						ReqData: newreqData,
						Message: model.Message{
							Action: &model.Action{Message: mp["0"]},
						},
						MData: resp.MData,
					}

					break
				}

				if resp.Audio != nil {
					m.Audio = &model.AudioData{Data: resp.Audio.Data}
				}

				if s.SendMessage(context.Background(), m) != nil {
					s.g.Requests() <- gateway.Req{
						ReqData: newreqData,
						Message: model.Message{
							Id:      resp.Id,
							Message: "Oof, looks like our main server took a graceful dive. Please hang tight for a few minutes while we work our magic!",
						},
					}
				}
			}
		}
	}()

	return s
}

func (s service) GetMessage(ctx context.Context, reqData pkgmodel.ReqData, message model.Message) {
	//TODO put context further

	uuidId, err := uuid.NewRandom()
	utils.Assert(err == nil, "cant create uuid")

	reqData.ReqId = pkgmodel.ReqID(uuidId.ID())

	req := gateway.Req{reqData, message, nil, make([]gateway.UIObject, 0)}

	for _, id := range message.Actions {
		select {
		case s.uis.Reqs() <- uis.Req{
			ReqData:  reqData,
			ButtonId: model.ButtonId(id),
			Clicked:  false,
		}:
			logging.Info("Request")

		default:
			logging.Info("Defautl occured")
		}

		select {
		case resp := <-s.uis.Resps():
			logging.Info("<-time.After(2 * time.Second):")

			req.UIData = append(req.UIData,
				gateway.UIObject{Id: uint64(resp.ButtonInfo.Id), Text: resp.ButtonInfo.Text, IsModied: false})

		case <-time.After(800 * time.Millisecond):
			logging.Info("<-time.After(2 * time.Second):")
		}
	}

	if message.Action != nil {
		req := rds.Req{
			ReqData: req.ReqData,
			Method:  rds.Set,
			Mp:      make(map[string]string),
		}

		for i, val := range message.Action.Values {
			req.Mp[strconv.Itoa(i)] = val
		}

		s.rds.Reqs() <- req
	}

	select {
	case s.g.Requests() <- req:
		logging.Info("Request")

	default:
		logging.Info("Error occured")
	}
}

func (s service) SendMessage(ctx context.Context, message model.Message) error {
	logging.Info("Sent")
	return s.f(ctx, message)
}
