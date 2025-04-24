package proc

import (
	"context"
	"reflect"

	"aidicti.top/api/protogen_cmn"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetterId interface {
	GetId() *protogen_cmn.ReqData
}

type handlerProc[Req GetterId, Resp GetterId, ReqModel any, RespModel any] struct {
	toModel   func(Req) (ReqModel, error)
	fromModel func(RespModel) (Resp, error)
	call      func(context.Context, model.ReqData, ReqModel) (RespModel, error)
}

func New[Req GetterId, Resp GetterId, ReqModel any, RespModel any](
	toModel func(Req) (ReqModel, error),
	fromModel func(RespModel) (Resp, error),
	call func(context.Context, model.ReqData, ReqModel) (RespModel, error)) *handlerProc[Req, Resp, ReqModel, RespModel] {

	return &handlerProc[Req, Resp, ReqModel, RespModel]{
		toModel:   toModel,
		fromModel: fromModel,
		call:      call,
	}
}

func (p handlerProc[Req, Resp, ReqModel, RespModel]) Exec(ctx context.Context, req Req) (Resp, error) {
	var nullResp Resp

	logging.Info("proc req", "st", "start")

	if req.GetId().ReqId == 0 {
		err := status.Error(codes.InvalidArgument, "MUST: req.Id.ReqId != 0")

		logging.Warn("proc req", "st", "fail", "err", err)

		return nullResp, err
	}

	if req.GetId().DlgId == 0 {
		err := status.Error(codes.InvalidArgument, "MUST: req.Id.DlgId != 0")

		logging.Warn("proc req", "st", "fail", "err", err)

		return nullResp, err
	}

	in, err := p.toModel(req)
	if err != nil {
		err := status.Errorf(codes.InvalidArgument, "MUST: %v", err)

		logging.Warn("proc req", "st", "fail", "err", err)

		return nullResp, err
	}

	utils.Contract(!reflect.ValueOf(in).IsNil())

	reqData := model.ReqData{
		ReqId: model.ReqID(req.GetId().ReqId),
		DlgId: model.DlgID(req.GetId().DlgId),
	}

	logging.Debug("proc req", "req data", reqData)

	resp, err := p.call(ctx, reqData, in)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			logging.Warn("proc req", "st", "fail", "req data", reqData, "code", st.Code(), "msg", st.Message())
		} else {
			logging.Warn("proc req", "st", "fail", "req data", reqData, "err", err)
		}

		return nullResp, err
	}

	out, err := p.fromModel(resp)
	if err != nil {
		err := status.Errorf(codes.Internal, "FAIL: %v", err)

		logging.Warn("proc req", "st", "fail", "req data", reqData, "err", err)

		return nullResp, err
	}

	utils.Contract(!reflect.ValueOf(out).IsNil())

	utils.Contract(out.GetId() != nil)

	*out.GetId() = protogen_cmn.ReqData{
		ReqId: uint64(reqData.ReqId),
		DlgId: uint64(reqData.DlgId),
	}

	logging.Info("proc req", "st", "ok", "req data", reqData)
	logging.Debug("proc req", "st", "ok", "resp", out)

	return out, nil
}
