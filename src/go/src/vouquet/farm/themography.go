package farm

import (
	"sync"
	"context"
)

import (
	"vouquet/shop"
)

func NewThemograpy(c_path string, soil_name string, ctx context.Context) (*Themography, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c_ctx, cancel := context.WithCancel(ctx)

	cfg, err := loadConfig(c_path)
	if err != nil {
		return nil, err
	}

	s, err := shop.New(soil_name, cfg, c_ctx)
	if err != nil {
		return nil, err
	}

	return &Themography{soil_name:soil_name, soil:s, ctx:c_ctx, cancel:cancel, mtx:new(sync.Mutex)}, nil
}

type Themography struct {
	soil_name  string
	soil       shop.Handler

	ctx    context.Context
	cancel context.CancelFunc
	mtx   *sync.Mutex
}

func (self *Themography) Status() (*Status, error) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	rates, err := self.soil.GetRate()
	if err != nil {
		return nil, err
	}
	return &Status{soil_name:self.soil_name, rates:rates}, nil
}

func (self *Themography) Release() error {
	self.cancel()
	return self.soil.Release()
}

type Status struct {
	soil_name  string
	rates map[string]shop.Rate
}
