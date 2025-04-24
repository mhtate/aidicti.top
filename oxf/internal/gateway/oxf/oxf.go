package oxf

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"aidicti.top/oxf/internal/gateway/oxf/parse"
	"aidicti.top/oxf/internal/model"
	"aidicti.top/pkg/gatewaydialog"
	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/logging"
	pkgmodel "aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
)

type Req struct {
	pkgmodel.ReqData
	model.Word
}

type Resp struct {
	pkgmodel.ReqData
	model.DictionaryEntry
}

type gateway struct {
	reqs    chan Req
	resps   chan Resp
	initPrF func(gatewaydialog.Processor[Req]) gatewaydialog.Processor[Req]
}

func (g gateway) Reqs() chan<- Req {
	return g.reqs
}

func (g gateway) Resps() <-chan Resp {
	return g.resps
}

func (g *gateway) SetProcessor(
	pr func(gatewaydialog.Processor[Req]) gatewaydialog.Processor[Req]) {
	g.initPrF = pr
}

func New() *gateway {
	return &gateway{
		reqs:  make(chan Req, 64),
		resps: make(chan Resp, 64),
	}
}

type reqProcessor struct {
	resps chan<- Resp
}

func NewProcessor() *reqProcessor {
	return &reqProcessor{}
}

func (p reqProcessor) Process(ctx context.Context, req Req) error {
	utils.Assert(len(req.Text) != 0, "req word size == 0")

	const urlTmpl = "https://www.oxfordlearnersdictionaries.com/definition/english/%s"

	reqWord := strings.ReplaceAll(utils.Sanitize(req.Text), " ", "-")

	url := fmt.Sprintf(urlTmpl, reqWord)

	realURL, html, err := fetchHTML(url)
	if err != nil {
		logging.Warn("get html fail", "err", err, "url", url)
		return err
	}

	entry, err := parse.GetEntry(html)
	if err != nil {
		logging.Warn("parse html fail", "err", err, "url", url)
		return err
	}

	entry.Link = string(realURL)

	select {
	case p.resps <- Resp{ReqData: req.ReqData, DictionaryEntry: *entry}:
		// logging.Debug("parse html fail", "err", err, "url", url)

	default:
		logging.Debug("parse html fail", "err", err, "url", url)
	}

	return nil
}

func (g gateway) Run(ctx context.Context) {
	utils.Assert(g.initPrF != nil, "req word size == 0")

	reqLstn := listener.New(g.reqs, g.initPrF(reqProcessor{g.resps}))

	reqLstn.Run(ctx)
}

type url_ string

func fetchHTML(url string) (url_, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set(
		"User-Agent",
		"Safari/537.36")
	req.Header.Set(
		"Accept-Language",
		"en-US,en;q=0.9")
	req.Header.Set("Accept",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	return url_(resp.Request.URL.Host + resp.Request.URL.Path), string(body), nil
}
