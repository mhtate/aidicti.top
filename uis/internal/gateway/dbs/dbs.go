package dbs

import (
	"context"
	"fmt"
	"time"

	"aidicti.top/pkg/gatewaydialog/listener"
	"aidicti.top/pkg/gatewaydialog/processor"
	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/model"
	"aidicti.top/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

type Req struct {
	model.ReqData
	DictWord
	Method
}

type Method uint

const (
	Get Method = 1
	Set Method = 2
)

type DictWord struct {
	ID     uint `gorm:"primarykey"`
	Word   string
	Def    string
	Usage  string
	Link   string
	UserId uint
	Pos    int
}

type Resp struct {
	model.ReqData
	Words []DictWord
}

type gateway struct {
	db    *gorm.DB
	reqs  chan Req
	resps chan Resp
}

func (g gateway) Reqs() chan<- Req {
	return g.reqs
}

func (g gateway) Resps() <-chan Resp {
	return g.resps
}

type dbsLogger struct{}

func (l dbsLogger) LogMode(gorm_logger.LogLevel) gorm_logger.Interface {
	return l
}

func (l dbsLogger) Info(_ context.Context, s string, i ...interface{}) {
	logging.Info(s, i...)
}

func (l dbsLogger) Warn(_ context.Context, s string, i ...interface{}) {
	logging.Warn(s, i...)
}

func (l dbsLogger) Error(_ context.Context, s string, i ...interface{}) {
	logging.Warn(s, i...)
}

func (l dbsLogger) Trace(context.Context, time.Time, func() (string, int64), error) {
}

func New(host, user, password string) *gateway {
	//TODO move to main
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=aidicti_db port=5432 sslmode=disable",
		host,
		user,
		password)

	db := utils.Must(gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbsLogger{},
	}))

	return &gateway{
		db:    db,
		reqs:  make(chan Req, 32),
		resps: make(chan Resp, 32),
	}
}

type reqProcessor struct {
	resps chan<- Resp
	db    *gorm.DB
}

func (p reqProcessor) Process(ctx context.Context, req Req) error {

	if req.Method == Set {
		req.DictWord.UserId = uint(req.DlgId)

		tx := p.db.Create(&req.DictWord)
		if tx.Error != nil {
			logging.Warn("set to db", "st", "fail", "err", tx.Error)
			return tx.Error
		}

		return nil
	}

	words := make([]DictWord, 0)
	if req.Method == Get {
		if req.Link == "" {
			return fmt.Errorf("invalid request req.Link not set")
		}

		if req.Pos == -1 {
			return fmt.Errorf("invalid request req.Pos not set")
		}

		// where := fmt.Sprintf("link = ? AND pos = ? AND userid = ?", req.Link, req.Pos, req.DlgId)
		// logging.Info("use where", "where", where)

		tx := p.db.Where("link = ? AND pos = ? AND user_id = ?", req.Link, req.Pos, req.DlgId).Find(&words)
		if tx.Error != nil {
			logging.Warn("get from db", "st", "fail", "err", tx.Error)
			return tx.Error
		}

		select {
		case p.resps <- Resp{ReqData: req.ReqData, Words: words}:
			logging.Info("get data from db ok")

		default:
			logging.Debug("get data from db fail")
		}
	}

	return fmt.Errorf("unknown method used")
}

func (g gateway) Run(ctx context.Context) {
	reqLstn := listener.New(
		g.reqs,
		processor.New(reqProcessor{resps: g.resps, db: g.db}, 16))

	g.db.AutoMigrate(&DictWord{})

	reqLstn.Run(ctx)
}
