package farm

import (
	"sync"
	"time"
	"context"
)

import (
	"github.com/vouquet/shop"
)

type ShipRecorder struct {
	symbol string
	soil   shop.Shop

	p_idx  map[string]int64
	f_idx  map[string]int64

	log    logger

	ctx  context.Context
	cancel context.CancelFunc
	mtx    *sync.Mutex
}

func OpenShipRecorder(soil_name string, symbol string, c_path string, ctx context.Context, log logger) (*ShipRecorder, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c_ctx, cancel := context.WithCancel(ctx)

	c, err := loadConfig(c_path)
	if err != nil {
		return nil, err
	}
	s, err := openShop(soil_name, c, c_ctx)
	if err != nil {
		return nil, err
	}

	self := &ShipRecorder{
		symbol: symbol,
		soil: s,

		p_idx: make(map[string]int64),
		f_idx: make(map[string]int64),

		log: log,

		ctx: c_ctx,
		cancel: cancel,

		mtx: new(sync.Mutex),
	}

	_, err = self.updateShipRecord()
	if err != nil {
		return nil, err
	}
	return self, nil
}

func (self *ShipRecorder) StreamRead() (<- chan *ShipRecord, error) {
	sr_ch := make(chan *ShipRecord)

	u_t := time.NewTicker(time.Second * 30)
	c_t := time.NewTicker(time.Hour)
	mtx := new(sync.Mutex)
	go func() {
		defer close(sr_ch)

		for {
			select {
			case <- self.ctx.Done():
				return
			case <- c_t.C:
				go func() {
					self.gcIndex()
				}()
			case <- u_t.C:
				go func() {
					mtx.Lock()
					defer mtx.Unlock()

					srs, err := self.updateShipRecord()
					if err != nil {
						self.log.WriteErr("Failed update ship record.'%s'", err)
						return
					}
					for _, sr := range srs {
						select {
						case <- self.ctx.Done():
							return
						case sr_ch <- sr:
						}
					}
				}()
			}
		}
	}()

	return sr_ch, nil
}

func (self *ShipRecorder) gcIndex() {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	now_utime := time.Now().Unix()
	limit := now_utime - int64(60 * 60 * 24 * 7)

	for key, utime := range self.p_idx {
		if utime > limit {
			continue
		}
		delete(self.p_idx, key)
	}
	for key, utime := range self.f_idx {
		if utime > limit {
			continue
		}
		delete(self.f_idx, key)
	}
}

func (self *ShipRecorder) updateShipRecord() ([]*ShipRecord, error) {
	self.mtx.Lock()
	defer self.mtx.Unlock()

	srs := []*ShipRecord{}

	poss, err := self.soil.GetPositions(self.symbol)
	if err != nil {
		return nil, err
	}
	for _, pos := range poss {
		if _, ok := self.p_idx[pos.Id()]; ok {
			continue
		}

		sr := &ShipRecord{
			id: pos.Id(),
			o_type: pos.OrderType(),
			price: pos.Price(),
			yield: float64(0),
			isOpenOrder: true,
		}
		srs = append(srs, sr)
		self.p_idx[pos.Id()] = time.Now().Unix()
	}

	fs, err := self.soil.GetFixes(self.symbol)
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		if _, ok := self.f_idx[f.Id()]; ok {
			continue
		}

		date, err := f.Date()
		if err != nil {
			self.log.WriteErr("ShipRecorder: cannot convert date: '%s'", err)
			continue
		}
		yield, err := f.Yield()
		if err != nil {
			self.log.WriteErr("ShipRecorder: cannot convert yield: '%s'", err)
			continue
		}
		sr := &ShipRecord{
			id: f.Id(),
			o_type: f.OrderType(),
			price: f.Price(),
			yield: yield,
			isOpenOrder: false,
		}
		srs = append(srs, sr)
		self.f_idx[f.Id()] = date.Unix()
	}
	return srs, nil
}

type ShipRecord struct {
	id     string
	o_type string
	price  float64
	yield  float64

	isOpenOrder bool
}

func (self *ShipRecord) Id() string {
	return self.id
}

func (self *ShipRecord) OrderType() string {
	return self.o_type
}

func (self *ShipRecord) Price() float64 {
	return self.price
}

func (self *ShipRecord) Yield() float64 {
	return self.yield
}

func (self *ShipRecord) IsOpenOrder() bool {
	return self.isOpenOrder
}
