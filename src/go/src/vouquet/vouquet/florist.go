package vouquet

import (
	"context"
)

import (
	"vouquet/farm"
)

type Florist interface {
	Run(context.Context, <- chan *farm.State) error
	SetSize(float64)
	Action(*farm.State) error
}
