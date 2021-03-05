package soil

import (
	"fmt"
	"sync"
	"context"
)

import (
	"github.com/vouquet/shop"
	"github.com/vouquet/go-gmo-coin/gomocoin"
)

func NewThemograpy(name string, ctx context.Context) (*Themography, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c_ctx, cancel := context.WithCancel(ctx)

	var s shop.Shop
	switch name {
	case SOIL_GMO:
		g_shop, err := gomocoin.NewGoMOcoin("", "", c_ctx)
		if err != nil {
			return nil, err
		}
		s = g_shop

	default:
		return nil, fmt.Errorf("undefined type '%s'", name)
	}

	return &Themography{name:name, shop:s, ctx:c_ctx, cancel:cancel, mtx:new(sync.Mutex)}, nil
}

type Themography struct {
	name  string
	shop  shop.Shop

	ctx    context.Context
	cancel context.CancelFunc
	mtx   *sync.Mutex
}

func (self *Themography) State() (*State, error) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	rates, err := self.shop.GetRate()
	if err != nil {
		return nil, err
	}
	return &State{name:self.name, rates:rates}, nil
}

func (self *Themography) Release() error {
	self.cancel()
	return self.shop.Close()
}

type State struct {
	name  string
	rates map[string]shop.Rate
}
