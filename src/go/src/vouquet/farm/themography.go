package farm

import (
	"sync"
	"context"
)

import (
	"github.com/vouquet/shop"
)

func NewThemograpy(soil_name string, ctx context.Context) (*Themography, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c_ctx, cancel := context.WithCancel(ctx)

	s, err := openShop(soil_name, nil, c_ctx)
	if err != nil {
		return nil, err
	}

	return &Themography{soil_name:soil_name, shop:s, ctx:c_ctx, cancel:cancel, mtx:new(sync.Mutex)}, nil
}

type Themography struct {
	soil_name  string
	shop  shop.Shop

	ctx    context.Context
	cancel context.CancelFunc
	mtx   *sync.Mutex
}

func (self *Themography) Status() (*Status, error) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	rates, err := self.shop.GetRate()
	if err != nil {
		return nil, err
	}
	return &Status{soil_name:self.soil_name, rates:rates}, nil
}

func (self *Themography) Release() error {
	self.cancel()
	return self.shop.Close()
}

type Status struct {
	soil_name  string
	rates map[string]shop.Rate
}
