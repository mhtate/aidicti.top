package repo

import (
	"aidicti.top/uis/internal/model"
)

type Repository interface {
	CreateInfo(model.Button) (model.ButtonId, error)
	Info(model.ButtonId) (model.Button, error)
	UpdateInfo(model.ButtonId, model.Button) error
	//TODO idiomatic go way not to use Get_methodname
	DeleteInfo(model.ButtonId) error
}
