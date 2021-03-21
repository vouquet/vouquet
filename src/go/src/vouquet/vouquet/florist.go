package vouquet

import (
	"context"
)

import (
	"vouquet/farm"
)

type Florist interface {
	Run(context.Context, float64, <- chan *farm.State) error
}
