package memory

import (
	"fmt"

	"aidicti.top/pkg/utils"
	"aidicti.top/uis/internal/model"
	"github.com/google/uuid"
)

type repo struct {
	mp map[model.ButtonId]model.Button
}

func New() *repo {
	return &repo{mp: map[model.ButtonId]model.Button{}}
}

func (r repo) CreateInfo(btn model.Button) (model.ButtonId, error) {
	//TODO use must pattern
	id, err := uuid.NewUUID()
	utils.Assert(err == nil, "UUID creating problem")

	buttonId := model.ButtonId(id.ID())

	r.mp[buttonId] = btn

	return buttonId, nil
}

func (r repo) Info(id model.ButtonId) (model.Button, error) {
	btn, ok := r.mp[id]
	if !ok {
		return model.Button{}, fmt.Errorf("id not found")
	}

	return btn, nil
}

func (r repo) UpdateInfo(id model.ButtonId, btn model.Button) error {
	_, ok := r.mp[id]
	if !ok {
		return fmt.Errorf("id not found")
	}

	r.mp[id] = btn

	return nil
}

func (r repo) DeleteInfo(id model.ButtonId) error {
	_, ok := r.mp[id]
	if !ok {
		return fmt.Errorf("id not found")
	}

	delete(r.mp, id)

	return nil
}
