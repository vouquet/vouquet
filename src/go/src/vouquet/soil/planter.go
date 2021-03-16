package soil

import (
	"fmt"
	"sync"
	"time"
	"context"
)

import (
	"github.com/vouquet/shop"
)

type Planter interface {
	Symbol() string
	SetSeed(string, float64, float64) error //TODO: not use rate, only stream order.
	ShowSproutList() ([]*Sprout, error)
	Harvest(*Sprout, float64) error //TODO: not use rate, only stream order.
	HarvestCnt() int64
	Yield() float64
}

type Flowerpot struct {
	symbol string

	log    logger

	ctx    context.Context
	cancel context.CancelFunc
	mtx    *sync.Mutex
}

func NewFlowerpot(name string, symbol string, c_path string, ctx context.Context, log logger) (*Flowerpot, error) {
	return nil, nil
}

func (self *Flowerpot) Symbol() string {
	return self.symbol
}

func (self *Flowerpot) SetSeed(o_type string, size float64, price float64) error {
	return nil
}

func (self *Flowerpot) ShowSproutList() ([]*Sprout, error) {
	return nil, nil
}

func (self *Flowerpot) Harvest(sp *Sprout, price float64) error {
	return nil
}

func (self *Flowerpot) HarvestCnt() int64 {
	return 0
}

func (self *Flowerpot) Yield() float64 {
	return float64(0)
}

type Sprout struct {
	date   time.Time

	price   float64
	size   float64
	o_type string

	pos  shop.Position
}

func (self *Sprout) CreateTime() time.Time {
	return self.date
}

func (self *Sprout) HasPosition() bool {
	return self.pos != nil
}

func (self *Sprout) Symbol() string {
	if self.pos == nil {
		return ""
	}
	return self.pos.Symbol()
}

func (self *Sprout) Size() float64 {
	if self.pos == nil {
		return float64(0)
	}
	return self.pos.Size()
}

func (self *Sprout) Price() float64 {
	if self.pos == nil {
		return float64(0)
	}
	return self.pos.Price()
}

func (self *Sprout) OrderType() string {
	if self.pos == nil {
		return ""
	}
	return self.pos.OrderType()
}

func (self *Sprout) equal(sp *Sprout) bool {
	if self.pos == nil {
		return false
	}
	if sp.pos == nil {
		return false
	}
	return self.pos.Id() == sp.pos.Id()
}

type testPosition struct {
	id     string
	symbol string
	size   float64
	price  float64
	o_type string
}

func newTestPosition(o_type string, symbol string, size float64, price float64) *testPosition {
	id := fmt.Sprintf("%v", time.Now().Unix())
	return &testPosition{id:id, symbol:symbol, size:size, price:price, o_type:o_type}
}

func (self *testPosition) Id() string {
	return self.id
}

func (self *testPosition) Symbol() string {
	return self.symbol
}

func (self *testPosition) Size() float64 {
	return self.size
}

func (self *testPosition) Price() float64 {
	return self.price
}

func (self *testPosition) OrderType() string {
	return self.o_type
}

type TestPlanter struct {
	name       string
	symbol     string

	sp_list    []*Sprout
	yield_cnt  float64
	hvst_cnt   int64

	log     logger
	mtx     *sync.Mutex
}

func NewTestPlanter(name string, symbol string, log logger) *TestPlanter {
	return &TestPlanter{
		name: name,
		symbol: symbol,
		sp_list: []*Sprout{},
		log: log,
		mtx: new(sync.Mutex),
	}
}

func (self *TestPlanter) Symbol() string {
	return self.symbol
}

func (self *TestPlanter) SetSeed(o_type string, size float64, price float64) error {
	self.lock()
	defer self.unlock()

	tpos := newTestPosition(o_type, self.symbol, size, price)
	sp := &Sprout{
		date: time.Now(),
		price: price,
		size: size,
		o_type: o_type,
		pos: tpos,
	}

	self.sp_list = append(self.sp_list, sp)
	return nil
}

func (self *TestPlanter) ShowSproutList() ([]*Sprout, error) {
	self.lock()
	defer self.unlock()

	return self.sp_list, nil
}

func (self *TestPlanter) Harvest(sp *Sprout, price float64) error {
	self.lock()
	defer self.unlock()

	in_val := sp.Price()
	out_val := price
	var yield float64
	switch sp.OrderType() {
	case shop.ORDER_TYPE_BUY:
		yield = (sp.Size() * out_val) - (sp.Size() * in_val)
	case shop.ORDER_TYPE_SELL:
		yield = (sp.Size() * in_val) - (sp.Size() * out_val)
	default:
		return fmt.Errorf("undefined type of order: '%s'", sp.OrderType())
	}

	self.yield_cnt += yield
	self.hvst_cnt ++

	sp_list := []*Sprout{}
	for i, bsp := range self.sp_list {
		if !bsp.equal(sp) {
			continue
		}

		sp_list = append(self.sp_list[:i], self.sp_list[i+1:]...)
		break
	}
	self.sp_list = sp_list

	self.log.WriteMsg("%s.Harvested(%s) orderIn: %f -> orderOut: %f, win: %f",
							self.name, sp.OrderType(), in_val, out_val, yield)
	return nil
}

func (self *TestPlanter) HarvestCnt() int64 {
	self.lock()
	defer self.unlock()

	return self.hvst_cnt
}

func (self *TestPlanter) Yield() float64 {
	self.lock()
	defer self.unlock()

	return self.yield_cnt
}

func (self *TestPlanter) lock() {
	self.mtx.Lock()
}

func (self *TestPlanter) unlock() {
	self.mtx.Unlock()
}
