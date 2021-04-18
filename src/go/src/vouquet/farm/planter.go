package farm

import (
	"fmt"
	"sync"
	"time"
	"context"
)

import (
	"vouquet/shop"
)

type OpeOption struct {
	Price  float64
	Stream bool //TODO: not use false ("not stream"), only stream order.
}

type Planter interface {
	Seed() string
	SetSeed(string, float64, *OpeOption) error
	ShowSproutList() ([]*Sprout, error)
	Harvest(*Sprout, *OpeOption) error
	Win()     float64
	WinCnt()  int64
	Lose()    float64
	LoseCnt() int64

	Release() error
}

type Flowerpot struct {
	seed    string
	soil    shop.Handler
	sp_list []*Sprout

	win      float64
	win_cnt  int64
	lose     float64
	lose_cnt int64

	log     logger

	ctx     context.Context
	cancel  context.CancelFunc
	mtx     *sync.Mutex
}

func NewFlowerpot(soil_name string, seed string, c *Config, ctx context.Context, log logger) (*Flowerpot, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	c_ctx, cancel := context.WithCancel(ctx)

	s, err := shop.New(soil_name, c, c_ctx)
	if err != nil {
		return nil, err
	}

	self := &Flowerpot{
		seed:seed,
		soil:s,
		sp_list: []*Sprout{},

		log: log,

		ctx: ctx,
		cancel: cancel,
		mtx: new(sync.Mutex),
	}

	if err := self.updateSproutList(true); err != nil {
		return nil, err
	}
	return self, nil
}

func (self *Flowerpot) Release() error {
	self.cancel()
	return self.soil.Release()
}

func (self *Flowerpot) Seed() string {
	return self.seed
}

func (self *Flowerpot) SetSeed(o_type string, size float64, opt *OpeOption) error {
	self.lock()
	defer self.unlock()

	if opt == nil {
		opt = DEFAULT_OpeOption
	}

	if err := self.soil.Order(o_type, self.seed, size, opt.Stream, opt.Price); err != nil {
		return err
	}
	sp := &Sprout{
		date: time.Now(),
		price: opt.Price,
		size: size,
		o_type: o_type,
	}

	self.log.WriteMsg("[SetSeed] %s, size: %.3f, price: %.3f", o_type, size, opt.Price)

	self.sp_list = append(self.sp_list, sp)
	return nil
}

func (self *Flowerpot) ShowSproutList() ([]*Sprout, error) {
	self.lock()
	defer self.unlock()

	if err := self.updateSproutList(false); err != nil {
		return nil, err
	}
	return self.getSproutList()
}

func (self *Flowerpot) getSproutList() ([]*Sprout, error) {
	if self.sp_list == nil {
		return nil, fmt.Errorf("sprout list is nil.")
	}

	ret_spl := make([]*Sprout, len(self.sp_list))
	copy(ret_spl, self.sp_list)

	return ret_spl, nil
}

func (self *Flowerpot) updateSproutList(always_update bool) error {
	has_pos_idx := make(map[string]interface{})
	no_pos := []*Sprout{}
	for _, sp := range self.sp_list {
		if sp.posId() == "" {
			no_pos = append(no_pos, sp)
			continue
		}
		has_pos_idx[sp.posId()] = nil
	}

	if !always_update {
		if len(has_pos_idx) == len(self.sp_list) {
			return nil
		}
	}

	poss, err := self.soil.GetPositions(self.seed)
	if err != nil {
		return err
	}
	for _, pos := range poss {
		if _, ok := has_pos_idx[pos.Id()]; ok {
			continue
		}

		mapped := false
		for _, sp := range no_pos {
			if sp.pos != nil {
				continue
			}

			if sp.o_type != pos.OrderType() {
				continue
			}

			upper := sp.price * 1.02
			lower := sp.price * 0.98
			if pos.Price() > upper || lower > pos.Price() {
				continue
			}

			sp.pos = pos
			mapped = true
		}

		if mapped {
			continue
		}
		sp := &Sprout{
			date: time.Now(),//TODO: want to set datetime where shop.position.
			price: pos.Price(),
			size: pos.Size(),
			o_type: pos.OrderType(),
			pos: pos,
		}
		self.sp_list = append(self.sp_list, sp)
	}

	return nil
}

func (self *Flowerpot) Harvest(h_sp *Sprout, opt *OpeOption) error {
	self.lock()
	defer self.unlock()

	if opt == nil {
		opt = DEFAULT_OpeOption
	}

	if h_sp.pos == nil {
		for i, sp := range self.sp_list {
			if !sp.equal(h_sp) {
				continue
			}

			self.sp_list = append(self.sp_list[:i], (self.sp_list)[i+1:]...)
			break
		}
		return fmt.Errorf("nil pointer error, doesn't get a position pointer.")
	}
	if err := self.soil.OrderFix(h_sp.pos, opt.Stream, opt.Price); err != nil {
		return err
	}

	var yield float64
	switch h_sp.OrderType() {
	case TYPE_SELL:
		yield = (h_sp.Price() * h_sp.Size()) - (opt.Price * h_sp.Size())
	case TYPE_BUY:
		yield = (opt.Price * h_sp.Size()) - (h_sp.Price() * h_sp.Size())
	default:
		return fmt.Errorf("unkown operation, '%s'", h_sp.OrderType())
	}

	if yield < 0 {
		self.lose += yield
		self.lose_cnt++
	} else {
		self.win += yield
		self.win_cnt++
	}

	for i, sp := range self.sp_list {
		if !sp.equal(h_sp) {
			continue
		}

		self.sp_list = append(self.sp_list[:i], (self.sp_list)[i+1:]...)
		break
	}

	self.log.WriteMsg("[Harvest] %s, size: %.3f, price: %.3f -> %.3f, win: %.3f(/%.3f)",
							h_sp.OrderType(), h_sp.Size(), h_sp.Price(), opt.Price,
							yield, self.win + self.lose)
	return nil
}

func (self *Flowerpot) Win() float64 {
	self.lock()
	defer self.unlock()

	return self.win
}

func (self *Flowerpot) WinCnt() int64 {
	self.lock()
	defer self.unlock()

	return self.win_cnt
}

func (self *Flowerpot) Lose() float64 {
	self.lock()
	defer self.unlock()

	return self.lose
}

func (self *Flowerpot) LoseCnt() int64 {
	self.lock()
	defer self.unlock()

	return self.lose_cnt
}

func (self *Flowerpot) lock() {
	self.mtx.Lock()
}

func (self *Flowerpot) unlock() {
	self.mtx.Unlock()
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

func (self *Sprout) Seed() string {
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

func (self *Sprout) posId() string {
	if self.pos == nil {
		return ""
	}
	return self.pos.Id()
}

type testPosition struct {
	id     string
	seed   string
	size   float64
	price  float64
	o_type string
}

func newTestPosition(o_type string, seed string, size float64, price float64) *testPosition {
	id := fmt.Sprintf("%v", time.Now().Unix())
	return &testPosition{id:id, seed:seed, size:size, price:price, o_type:o_type}
}

func (self *testPosition) Id() string {
	return self.id
}

func (self *testPosition) Symbol() string {
	return self.seed
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
	seed      string

	now_state *State

	sp_list   []*Sprout

	win       float64
	win_cnt   int64
	lose      float64
	lose_cnt  int64

	log     logger
	mtx     *sync.Mutex
}

func NewTestPlanter(seed string, log logger) *TestPlanter {
	return &TestPlanter{
		seed: seed,
		sp_list: []*Sprout{},
		log: log,
		mtx: new(sync.Mutex),
	}
}

func (self *TestPlanter) Release() error {
	return nil
}

func (self *TestPlanter) Seed() string {
	return self.seed
}

func (self *TestPlanter) SetState(state *State) {
	self.lock()
	defer self.unlock()

	self.now_state = state
}

func (self *TestPlanter) SetSeed(o_type string, size float64, opt *OpeOption) error {
	self.lock()
	defer self.unlock()

	if opt == nil {
		opt = DEFAULT_OpeOption
	}

	price := (self.now_state.Ask() + self.now_state.Bid()) / float64(2)
	tpos := newTestPosition(o_type, self.seed, size, price)
	sp := &Sprout{
		date: self.now_state.Date(),
		price: opt.Price,
		size: size,
		o_type: o_type,
		pos: tpos,
	}

	self.log.WriteDebug("[SetSeed] %s, size: %.3f, price: %.3f", o_type, size, price)

	self.sp_list = append(self.sp_list, sp)
	return nil
}

func (self *TestPlanter) ShowSproutList() ([]*Sprout, error) {
	self.lock()
	defer self.unlock()

	c_sp_list := make([]*Sprout, len(self.sp_list))
	copy(c_sp_list, self.sp_list)

	return c_sp_list, nil
}

func (self *TestPlanter) Harvest(sp *Sprout, opt *OpeOption) error {
	self.lock()
	defer self.unlock()

	if opt == nil {
		opt = DEFAULT_OpeOption
	}

	in_val := sp.Price()
	out_val := (self.now_state.Ask() + self.now_state.Bid()) / float64(2)
	var yield float64
	switch sp.OrderType() {
	case TYPE_BUY:
		yield = (sp.Size() * out_val) - (sp.Size() * in_val)
	case TYPE_SELL:
		yield = (sp.Size() * in_val) - (sp.Size() * out_val)
	default:
		return fmt.Errorf("undefined type of order: '%s'", sp.OrderType())
	}

	if yield < 0 {
		self.lose += yield
		self.lose_cnt++
	} else {
		self.win += yield
		self.win_cnt++
	}

	sp_list := []*Sprout{}
	for i, bsp := range self.sp_list {
		if !bsp.equal(sp) {
			continue
		}

		sp_list = append(self.sp_list[:i], self.sp_list[i+1:]...)
		break
	}
	self.sp_list = sp_list

	t_fmt := "2006/01/02 15:04"
	self.log.WriteDebug("[Harvested(%s)] orderIn: %f(%s) -> orderOut: %f(%s), win: %f",
							sp.OrderType(), in_val, sp.CreateTime().Format(t_fmt),
							out_val, self.now_state.Date().Format(t_fmt),  yield)
	return nil
}

func (self *TestPlanter) Win() float64 {
	self.lock()
	defer self.unlock()

	return self.win
}

func (self *TestPlanter) WinCnt() int64 {
	self.lock()
	defer self.unlock()

	return self.win_cnt
}

func (self *TestPlanter) Lose() float64 {
	self.lock()
	defer self.unlock()

	return self.lose
}

func (self *TestPlanter) LoseCnt() int64 {
	self.lock()
	defer self.unlock()

	return self.lose_cnt
}

func (self *TestPlanter) lock() {
	self.mtx.Lock()
}

func (self *TestPlanter) unlock() {
	self.mtx.Unlock()
}
