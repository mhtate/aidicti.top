package idle

import (
	"context"
)

type scnr struct{}

func New() *scnr {

	return &scnr{}
}

func (s scnr) Run(context.Context) {}

func (s scnr) Name() string {
	return "idle"
}
