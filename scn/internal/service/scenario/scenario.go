package scenario

import (
	"context"

	"aidicti.top/pkg/model/oxf"
	"aidicti.top/scn/internal/gateway/ai"
	"aidicti.top/scn/internal/gateway/dbs"
	"aidicti.top/scn/internal/model"
)

type DialogSTT interface {
	Reqs() chan model.AudioData
	Resps() chan string
}

type DialogUSR interface {
	Reqs() chan model.Message
	Resps() chan model.Message
}

type DialogGPT interface {
	Reqs() chan ai.ReqAI
	Resps() chan ai.RespAI
}

type DialogOXF interface {
	Reqs() chan oxf.Word
	Resps() chan oxf.DictionaryEntry
}

type DialogUIS interface {
	Reqs() chan model.Button
	Resps() chan model.ButtonId
}

type DialogDBS interface {
	Reqs() chan dbs.Req
	Resps() chan dbs.Resp
}

type Scenario interface {
	Run(context.Context)
	Name() string
}
