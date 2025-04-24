package button

import "aidicti.top/uis/internal/model"

type Button interface {
	Info() model.ButtonInfo
	Click()
}
